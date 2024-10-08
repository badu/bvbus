package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/golang/freetype"
	"github.com/golang/geo/s2"
	"github.com/mattn/go-sqlite3"
	"github.com/qedus/osmpbf"
)

func TestLatestPBF(t *testing.T) {
	file, err := os.Open("./../../data/romania.osm.pbf")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := osmpbf.NewDecoder(file)
	err = decoder.Start(runtime.GOMAXPROCS(0) - 1)
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	repo, err := NewRepository(logger, "./../../data/brasov_busses.sqlite")
	if err != nil {
		t.Fatalf("error creating repository:%#v", err)
	}

	existingBussesMap := make(map[int64]Busline)
	existingBusses, err := repo.GetBusses(false)
	if err != nil {
		panic(err)
	}

	for _, bus := range existingBusses {
		existingBussesMap[bus.OSMID] = bus
	}

	tx, err := repo.DB.Begin()
	if err != nil {
		t.Fatalf("error beginning transaction:%#v", err)
	}

	bussesStmt, err := tx.Prepare("INSERT INTO busses (id, dir, name, from_station, to_station, no, color, website) VALUES (?, ?, ?, ?, ?, ?, ?, ?);")
	if err != nil {
		t.Fatalf("error preparing busses statement:%#v", err)
	}
	defer bussesStmt.Close()

	stopsStmt, err := tx.Prepare("INSERT INTO bus_stops (bus_id, station_id, station_index) VALUES (?, ?, ?);")
	if err != nil {
		t.Fatalf("error preparing bus stops statement:%#v", err)
	}
	defer stopsStmt.Close()

	stationsStmt, err := tx.Prepare("INSERT INTO stations (id, name, lat, lng, outside, board) VALUES (?, ?, ?, ?, ?, ?);")
	if err != nil {
		t.Fatalf("error preparing stations statement:%#v", err)
	}
	defer stationsStmt.Close()

	includedBusIDs := make(map[int64]struct{})
	for _, busID := range goodBusses {
		includedBusIDs[busID] = struct{}{}
	}

	metropolitans := make([]int64, 0)
	urbans := make([]int64, 0)
	busses := make(map[int64]*Busline)
	stations := make(map[int64]*Station)
	uniqueWays := make(map[int64]Node)
	uniqueNodes := make(map[int64]Node)
	relationWays := make(map[int64][]Member)
	relationStops := make(map[int64][]Member)
	subRels := make(map[int64][]int64)
	for {
		if v, err := decoder.Decode(); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("error decoding PBF:%#v", err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
				node := Node{ID: v.ID, Lat: v.Lat, Lon: v.Lon}
				uniqueNodes[v.ID] = node
				if v.Tags["network"] != "RAT Brașov" {
					continue
				}

				station := Station{
					OSMID: v.ID,
					Lat:   v.Lat,
					Lon:   v.Lon,
				}
				for k, tv := range v.Tags {
					switch k {
					case "departures_board":
						if tv == "realtime" {
							station.HasBoard = true
						}
					case "fare_zone":
						if len(tv) > 0 {
							station.IsOutsideCity = true
						}
					case "name":
						station.Name = tv
					}
				}

				station.Name = replaceDiacritics(station.Name)
				_, err = stationsStmt.Exec(station.OSMID, station.Name, station.Lat, station.Lon, station.IsOutsideCity, station.HasBoard)
				if err != nil {
					var sqliteErr sqlite3.Error
					if errors.As(err, &sqliteErr) {
						if !errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
							if err := tx.Rollback(); err != nil {
								t.Fatalf("error rolling back transaction on insert station :%#v", err)
							}
							t.Fatalf("error inserting station :%#v", err)
						}
					}
				}

				stations[station.OSMID] = &station
				// logger.Info("station created", "name", station.Name, "metropolitan", station.IsOutsideCity)
			case *osmpbf.Way:
				uniqueWays[v.ID] = Node{Nodes: v.NodeIDs, ID: v.ID, Tags: v.Tags}
			case *osmpbf.Relation:
				// not something we are interested in ?
				if v.Tags["network"] != "RAT Brașov" {
					continue
				}

				subRels[v.ID] = make([]int64, 0)
				for _, member := range v.Members {
					subRels[v.ID] = append(subRels[v.ID], member.ID)
				}

				isValid := false
				line := Busline{OSMID: v.ID}

				overriddenLine := ""
				for k, tv := range v.Tags {
					switch k {
					case "name":
						cleanName := strings.ReplaceAll(tv, "=>", "-")
						cleanName = replaceDiacritics(cleanName)
						line.Name = cleanName
					case "from":
						isValid = true
						line.From = replaceDiacritics(tv)
					case "to":
						line.To = replaceDiacritics(tv)
					case "colour":
						line.Color = tv
					case "ref":
						line.Line = tv
					case "local_ref":
						overriddenLine = tv
					case "website":
						line.Link = tv
					case "description":
						if tv == "Rețea transportului public din Brașov (Metropolitan)" {
							for _, member := range v.Members {
								metropolitans = append(metropolitans, member.ID)
							}
							break
						} else if tv == "Rețea transportului public din Brașov (Urban)" {
							for _, member := range v.Members {
								urbans = append(urbans, member.ID)
							}
							break
						}
					}
				}

				if len(overriddenLine) > 0 {
					line.Line = overriddenLine
				}

				if strings.Contains(line.Link, "-dus") {
					line.Dir = 1
				} else if strings.Contains(line.Link, "-intors") {
					line.Dir = 2
				}

				if _, found := existingBussesMap[v.ID]; found {
					isValid = false
				}

				if !isValid {
					continue
				}

				if _, willKeep := includedBusIDs[v.ID]; !willKeep {
					continue
				}

				_, err = bussesStmt.Exec(line.OSMID, line.Dir, line.Name, line.From, line.To, line.Line, line.Color, line.Link)
				if err != nil {
					if err := tx.Rollback(); err != nil {
						t.Fatalf("error rolling back transaction on insert bus line :%#v", err)
					}
					t.Fatalf("error inserting bus line :%#v", err)
				}

				busses[v.ID] = &line
				// logger.Info("bus line created", "name", line.Name, "link", line.Link)

				stationIndex := 0
				for _, member := range v.Members {
					newMember := Member{Ref: member.ID, Role: member.Role}
					if member.Type == osmpbf.WayType {
						relationWays[v.ID] = append(relationWays[v.ID], newMember)
						continue
					}

					if member.Role == OSMPlatform || member.Role == OSMPlatformEntryOnly || member.Role == OSMPlatformExitOnly {
						_, err = stopsStmt.Exec(v.ID, member.ID, stationIndex)
						if err != nil {
							var sqliteErr sqlite3.Error
							if errors.As(err, &sqliteErr) {
								if !errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
									if err := tx.Rollback(); err != nil {
										t.Fatalf("error rolling back transaction on insert bus stop :%#v", err)
									}
									t.Fatalf("error inserting bus stop :%#v", err)
								}
							}
						}
						station, found := stations[newMember.Ref]
						if found {
							stationIndex++
							_ = station
							// logger.Info("bus stop created", "bus", line.Name, "station", station.Name, "index", stationIndex)
						} else {
							logger.Warn("STATION NOT FOUND", "id", newMember.Ref, "index", stationIndex)
						}
						continue
					}

					if member.Type == osmpbf.NodeType && (member.Role == OSMStop || member.Role == OSMStopExitOnly || member.Role == OSMStopEntryOnly) {
						relationStops[v.ID] = append(relationStops[v.ID], newMember)
						continue
					}

				}
			}
		}
	}

	var collectAllChildren func([]int64) []int64
	collectAllChildren = func(targets []int64) []int64 {
		result := make([]int64, 0)
		for _, targetID := range targets {
			result = append(result, targetID)
			if children, has := subRels[targetID]; has {
				result = append(result, collectAllChildren(children)...)
			}
		}
		return result
	}

	if len(metropolitans) > 0 {
		realMetropolitans := collectAllChildren(metropolitans)

		args := make([]any, len(realMetropolitans))
		for i, id := range realMetropolitans {
			args[i] = id
		}

		sql := `UPDATE busses SET metropolitan = true WHERE id IN (?` + strings.Repeat(`,?`, len(realMetropolitans)-1) + `);`
		_, err = tx.Exec(sql, args...)
		if err != nil {
			t.Fatalf("error updating metropolitan busses:%#v", err)
		}
	}

	if len(urbans) > 0 {
		realUrbans := collectAllChildren(urbans)

		args := make([]any, len(realUrbans))
		for i, id := range realUrbans {
			args[i] = id
		}

		sql := `UPDATE busses SET urban = true WHERE id IN (?` + strings.Repeat(`,?`, len(realUrbans)-1) + `);`
		_, err = tx.Exec(sql, args...)
		if err != nil {
			t.Fatalf("error updating urban busses:%#v", err)
		}
	}

	logger.Info("======= SUMMARY SO FAR ==========")
	logger.Info("busses", "len", len(existingBussesMap))
	logger.Info("stations", "len", len(stations))
	logger.Info("======= END OF SUMMARY ==========")

	for busID, ways := range relationWays {
		if len(ways) <= 0 {
			continue
		}

		pointIndex := 1
		lastNodeID := int64(-1)
		seen := make(map[int64]struct{})
		for _, wayMember := range ways {
			way, hasFoundWay := uniqueWays[wayMember.Ref]
			if !hasFoundWay {
				logger.Error("ERROR FINDING WAY", "id", wayMember.Ref)
				continue
			}

			inReverse := false
			for wayNodeIndex, wayNodeID := range way.Nodes {
				if lastNodeID > 0 {
					if wayNodeIndex == 0 {
						if lastNodeID == wayNodeID {
							continue
						} else {
							inReverse = true
							break
						}
					}
				}

				lastNodeID = wayNodeID
				wayNode, wayNodeFound := uniqueNodes[wayNodeID]
				if !wayNodeFound {
					logger.Error("WAY NODE NOT FOUND", "id", wayNodeID)
					break
				}

				itsAStop := false
				for _, stop := range relationStops[busID] {
					if stop.Ref == wayNodeID {
						itsAStop = true
						break
					}
				}
				_ = itsAStop
				if _, has := seen[wayNode.ID]; !has {
					/**
					_, err = pointsStmt.Exec(wayNode.ID, wayNode.Lat, wayNode.Lon)
					if err != nil {
						var sqliteErr sqlite3.Error
						if errors.As(err, &sqliteErr) {
							if !errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
								if err := tx.Rollback(); err != nil {
									t.Fatalf("error rolling back transaction on insert street point :%#v", err)
								}
								t.Fatalf("error inserting street point :%#v", err)
							}
						}
					}
					*/
				}
				/**
				_, err = relsStmt.Exec(wayNode.ID, pointIndex, busID, itsAStop)
				if err != nil {
					var sqliteErr sqlite3.Error
					if errors.As(err, &sqliteErr) {
						if !errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
							if err := tx.Rollback(); err != nil {
								t.Fatalf("error rolling back transaction on insert relation :%#v", err)
							}
							t.Fatalf("error inserting relation :%#v", err)
						}
					}
				}
				*/
				seen[wayNode.ID] = struct{}{}
				pointIndex++
			}

			if inReverse {
				for wayNodeIndex := len(way.Nodes) - 1; wayNodeIndex >= 0; wayNodeIndex-- {
					wayNodeID := way.Nodes[wayNodeIndex]
					lastNodeID = wayNodeID

					wayNode, wayNodeFound := uniqueNodes[wayNodeID]
					if !wayNodeFound {
						logger.Error("[reverse] WAY NODE NOT FOUND", "id", wayNodeID)
						break
					}

					itsAStop := false
					for _, stop := range relationStops[busID] {
						if stop.Ref == wayNodeID {
							itsAStop = true
							break
						}
					}
					_ = itsAStop
					if _, has := seen[wayNode.ID]; !has {
						/**
						_, err = pointsStmt.Exec(wayNode.ID, wayNode.Lat, wayNode.Lon)
						if err != nil {
							var sqliteErr sqlite3.Error
							if errors.As(err, &sqliteErr) {
								if !errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
									if err := tx.Rollback(); err != nil {
										t.Fatalf("error rolling back transaction on insert street point :%#v", err)
									}
									t.Fatalf("error inserting street point :%#v", err)
								}
							}
						}
						*/
					}
					/**
					_, err = relsStmt.Exec(wayNode.ID, pointIndex, busID, itsAStop)
					if err != nil {
						var sqliteErr sqlite3.Error
						if errors.As(err, &sqliteErr) {
							if !errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
								if err := tx.Rollback(); err != nil {
									t.Fatalf("error rolling back transaction on insert relation:%#v", err)
								}
								t.Fatalf("error inserting relation (reverse):%#v", err)
							}
						}
					}
					*/
					seen[wayNode.ID] = struct{}{}
					pointIndex++
				}
			}
		}

		line, found := existingBussesMap[busID]
		if found {
			_ = line
			// logger.Info("bus-points", "bus", line.Name, "points", pointIndex-1)
		} else {
			logger.Warn("bus line not found???", "id", busID)
		}
	}

	err = tx.Commit()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Fatalf("error rolling back transaction on failed commit :%#v", err)
		}
		t.Fatalf("error commiting transaction:%#v", err)
	}

	t.Log("initial import finished.")
}

func TestCollectTags(t *testing.T) {
	file, err := os.Open("./../../data/brasov.osm.pbf")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := osmpbf.NewDecoder(file)
	err = decoder.Start(runtime.GOMAXPROCS(0) - 1)
	if err != nil {
		panic(err)
	}

	overallTags := make(map[string]struct{})
	uniqueNodesTags := make(map[string]struct{})
	uniqueWaysTags := make(map[string]struct{})
	uniqueRelationsTags := make(map[string]struct{})
	networks := make(map[string]struct{})
	for {
		if v, err := decoder.Decode(); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("error decoding PBF:%#v", err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
				for k, tv := range v.Tags {
					if k == "network" {
						networks[tv] = struct{}{}
					}
					if _, has := uniqueNodesTags[k]; !has {
						uniqueNodesTags[k] = struct{}{}
						overallTags[k] = struct{}{}
					}
				}
			case *osmpbf.Way:
				for k, tv := range v.Tags {
					if k == "network" {
						networks[tv] = struct{}{}
					}
					if _, has := uniqueWaysTags[k]; !has {
						uniqueWaysTags[k] = struct{}{}
						overallTags[k] = struct{}{}
					}
				}
			case *osmpbf.Relation:
				for k, tv := range v.Tags {
					if k == "network" {
						networks[tv] = struct{}{}
					}
					if _, has := uniqueRelationsTags[k]; !has {
						uniqueRelationsTags[k] = struct{}{}
						overallTags[k] = struct{}{}
					}
				}

			}
		}
	}

	tags := make([]string, len(overallTags))
	i := 0
	for k := range overallTags {
		tags[i] = k
		i++
	}

	sort.Strings(tags)
	for _, v := range tags {
		t.Logf("tag : %s", v)
	}

	for net := range networks {
		t.Logf("network: %s", net)
	}
}

func TestMakePolyFile(t *testing.T) {
	file, err := os.ReadFile("./../../data/brasov_boundary.json")
	if err != nil {
		return
	}

	var data Data
	err = json.Unmarshal(file, &data)
	if err != nil {
		t.Fatalf("Error parsing JSON: %s", err)
	}

	nodes := make(map[int64]Node)
	ways := make(map[int64]Node)
	for _, element := range data.Elements {
		switch element.Type {
		default:
		case "node":
			nodes[element.ID] = element
		case "way":
			ways[element.ID] = element
		}
	}

	var sb strings.Builder
	var jsonSB strings.Builder
	sb.WriteString("brasov\n")
	sb.WriteString("brasov_area\n")

	for _, element := range data.Elements {
		switch element.Type {
		default:
		case "relation":
			lastNodeID := int64(-1)
			for _, member := range element.Members {
				if member.Type != "way" {
					continue
				}

				if member.Role != "outer" {
					continue
				}

				theWay, has := ways[member.Ref]
				if !has {
					t.Fatalf("error finding %d in ways", member.Ref)
				}

				inReverse := false
				for wayNodeIndex, wayNodeID := range theWay.Nodes {
					if wayNodeIndex == 0 {
						if lastNodeID > 0 {
							if lastNodeID != wayNodeID {
								inReverse = true
								break
							}
						}
					}

					lastNodeID = wayNodeID
					wayNode, wayNodeFound := nodes[wayNodeID]
					if !wayNodeFound {
						t.Fatalf("error finding %d in nodes", wayNodeID)
					}

					sb.WriteString(fmt.Sprintf("    %.08f    %.08f\n", wayNode.Lon, wayNode.Lat))
					jsonSB.WriteString(fmt.Sprintf("{%s:%d,%s:%.08f,%s:%.08f,%s:%d},\n", "id", lastNodeID, "lat", wayNode.Lat, "lng", wayNode.Lon, "way", theWay.ID))
				}

				if inReverse {
					for wayNodeIndex := len(theWay.Nodes) - 1; wayNodeIndex >= 0; wayNodeIndex-- {
						lastNodeID = theWay.Nodes[wayNodeIndex]
						wayNode, wayNodeFound := nodes[theWay.Nodes[wayNodeIndex]]
						if !wayNodeFound {
							t.Fatalf("error finding %d in nodes", theWay.Nodes[wayNodeIndex])
						}

						sb.WriteString(fmt.Sprintf("    %.08f    %.08f\n", wayNode.Lon, wayNode.Lat))
						jsonSB.WriteString(fmt.Sprintf("{%s:%d,%s:%.08f,%s:%.08f,%s:%d},\n", "id", lastNodeID, "lat", wayNode.Lat, "lng", wayNode.Lon, "way", theWay.ID))
					}
				}

			}
		}
	}
	sb.WriteString(fmt.Sprintf("END\nEND\n"))

	err = os.WriteFile("./../../data/brasov_boundary.poly", []byte(sb.String()), 0644)
	if err != nil {
		t.Fatalf("error writing file : %#v", err)
	}

	t.Logf("const points = [%s];", jsonSB.String())
}

func TestFindBoundary(t *testing.T) {
	file, err := os.ReadFile("./../../data/brasov_boundary.json")
	if err != nil {
		return
	}

	var data Data
	err = json.Unmarshal(file, &data)
	if err != nil {
		t.Fatalf("Error parsing JSON: %s", err)
	}

	nodes := make(map[int64]Node)
	ways := make(map[int64]Node)
	for _, element := range data.Elements {
		switch element.Type {
		default:
		case "node":
			nodes[element.ID] = element
		case "way":
			ways[element.ID] = element
		}
	}

	minLat, minLon := float64(180), float64(180)
	var maxLat, maxLon float64

	for _, element := range data.Elements {
		switch element.Type {
		default:
		case "relation":
			lastNodeID := int64(-1)
			for _, member := range element.Members {
				if member.Type != "way" {
					continue
				}

				if member.Role != "outer" {
					continue
				}

				theWay, has := ways[member.Ref]
				if !has {
					t.Fatalf("error finding %d in ways", member.Ref)
				}

				inReverse := false
				for wayNodeIndex, wayNodeID := range theWay.Nodes {
					if wayNodeIndex == 0 {
						if lastNodeID > 0 {
							if lastNodeID != wayNodeID {
								inReverse = true
								break
							}
						}
					}

					lastNodeID = wayNodeID
					wayNode, wayNodeFound := nodes[wayNodeID]
					if !wayNodeFound {
						t.Fatalf("error finding %d in nodes", wayNodeID)
					}

					if wayNode.Lat < minLat {
						minLat = wayNode.Lat
					}
					if wayNode.Lat > maxLat {
						maxLat = wayNode.Lat
					}
					if wayNode.Lon < minLon {
						minLon = wayNode.Lon
					}
					if wayNode.Lon > maxLon {
						maxLon = wayNode.Lon
					}
				}

				if inReverse {
					for wayNodeIndex := len(theWay.Nodes) - 1; wayNodeIndex >= 0; wayNodeIndex-- {
						lastNodeID = theWay.Nodes[wayNodeIndex]
						wayNode, wayNodeFound := nodes[theWay.Nodes[wayNodeIndex]]
						if !wayNodeFound {
							t.Fatalf("error finding %d in nodes", theWay.Nodes[wayNodeIndex])
						}
						if wayNode.Lat < minLat {
							minLat = wayNode.Lat
						}
						if wayNode.Lat > maxLat {
							maxLat = wayNode.Lat
						}
						if wayNode.Lon < minLon {
							minLon = wayNode.Lon
						}
						if wayNode.Lon > maxLon {
							maxLon = wayNode.Lon
						}
					}
				}

			}
		}
	}

	t.Logf("minLat = %.08f, minLon = %.08f,maxLat = %.08f, maxLon = %.08f", minLat, minLon, maxLat, maxLon)
}

func TestOptimizeLines(t *testing.T) {
	areas := make([]float64, 0)
	areCollinear := func(p1, p2, p3 Node) bool {
		area := p1.Lat*(p2.Lon-p3.Lon) + p2.Lat*(p3.Lon-p1.Lon) + p3.Lat*(p1.Lon-p2.Lon)
		areas = append(areas, area)
		return area == 0
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	repo, err := NewRepository(logger, "./../../data/brasov_busses.sqlite")
	if err != nil {
		t.Fatalf("error creating repository:%#v", err)
	}
	_ = repo
	// TODO : fix me
	var processedWays map[int64][]Node
	for busID, points := range processedWays {
		for i := 0; i < len(points); i++ {
			if i == 0 || i == len(points)-1 {

			} else {
				p1 := points[i-1]
				p2 := points[i]
				p3 := points[i+1]

				if !areCollinear(p1, p2, p3) {

				} else {
					t.Logf("point %d of bus %d can be deleted", p2.ID, busID)
				}
			}
		}
	}

	sort.Float64s(areas)
	for _, area := range areas {
		t.Logf("%.16f", area)
	}
}

func TestSetStreetNames(t *testing.T) {
	type Place struct {
		PlaceID     int64   `json:"place_id"`
		Licence     string  `json:"licence"`
		OsmType     string  `json:"osm_type"`
		OsmID       int64   `json:"osm_id"`
		Lat         string  `json:"lat"`
		Lon         string  `json:"lon"`
		Class       string  `json:"class"`
		Type        string  `json:"type"`
		PlaceRank   int     `json:"place_rank"`
		Importance  float64 `json:"importance"`
		Addresstype string  `json:"addresstype"`
		Name        string  `json:"name"`
		DisplayName string  `json:"display_name"`
		Address     struct {
			Highway      string `json:"highway"`
			Road         string `json:"road"`
			Suburb       string `json:"suburb"`
			City         string `json:"city"`
			Municipality string `json:"municipality"`
			County       string `json:"county"`
			ISO31662Lvl4 string `json:"ISO3166-2-lvl4"`
			Postcode     string `json:"postcode"`
			Country      string `json:"country"`
			CountryCode  string `json:"country_code"`
		} `json:"address"`
		Boundingbox []string `json:"boundingbox"`
	}
	type Places []Place

	file, err := os.ReadFile("./../../data/reverse_geocoding.json")
	if err != nil {
		return
	}

	var places Places
	err = json.Unmarshal(file, &places)
	if err != nil {
		t.Fatalf("Error parsing JSON: %s", err)
	}

	placesMap := make(map[int64]Place)
	for _, place := range places {
		placesMap[place.OsmID] = place
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	repo, err := NewRepository(logger, "./../../data/brasov_busses.sqlite")
	if err != nil {
		t.Fatalf("error creating repository:%#v", err)
	}

	stations, err := repo.GetStations()
	if err != nil {
		t.Fatalf("error reading stations : %#v", err)
	}

	tx, err := repo.DB.Begin()
	if err != nil {
		t.Fatalf("error beginning transaction:%#v", err)
	}

	updateStmt, err := tx.Prepare("UPDATE stations SET street_name = ? WHERE id = ?;")
	if err != nil {
		t.Fatalf("error preparing busses statement:%#v", err)
	}
	defer updateStmt.Close()

	for _, station := range stations {
		if place, has := placesMap[station.OSMID]; has {
			road := replaceDiacritics(place.Address.Road)
			_, err = updateStmt.Exec(road, station.OSMID)
			if err != nil {
				if err := tx.Rollback(); err != nil {
					t.Fatalf("error rolling back transaction on update station :%#v", err)
				}
				t.Fatalf("error updating station :%#v", err)
			}
			t.Logf("updated station %d (%s) with street %s", station.OSMID, station.Name, road)
		} else {
			t.Logf("error %d not found in places map", station.OSMID)
		}

	}

	err = tx.Commit()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Fatalf("error rolling back transaction on failed commit :%#v", err)
		}
		t.Fatalf("error commiting transaction:%#v", err)
	}

	t.Log("update streets finished.")
}

func TestGenerateStationsJS(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	repo, err := NewRepository(logger, "./../../data/brasov_busses.sqlite")
	if err != nil {
		t.Fatalf("error creating repository:%#v", err)
	}

	stations, err := repo.GetStations()
	if err != nil {
		t.Fatalf("error reading stations : %#v", err)
	}

	stationNamesToIDs := make(map[string][]int64)
	for _, station := range stations {
		if _, has := stationNamesToIDs[station.Name]; !has {
			stationNamesToIDs[station.Name] = make([]int64, 0)
		}
		stationNamesToIDs[station.Name] = append(stationNamesToIDs[station.Name], station.OSMID)
	}

	rows, err := repo.Query(`SELECT bus_id, station_id, station_index, stations.name FROM bus_stops INNER JOIN stations ON stations.id = station_id ORDER BY bus_id,station_index;`)
	if err != nil {
		t.Fatalf("error querying bus stops:%#v", err)
	}

	type Stop struct {
		stationID    int64
		stationIndex int
		stationName  string
	}

	stationsToBusses := make(map[int64][]int64)
	busStops := make(map[int64][]Stop)
	for rows.Next() {
		var stop Stop
		var busID int64
		err := rows.Scan(&busID, &stop.stationID, &stop.stationIndex, &stop.stationName)
		if err != nil {
			t.Fatalf("error scanning:%#v", err)
		}
		if _, has := busStops[busID]; !has {
			busStops[busID] = make([]Stop, 0)
		}
		busStops[busID] = append(busStops[busID], stop)

		if _, hasStation := stationsToBusses[stop.stationID]; !hasStation {
			stationsToBusses[stop.stationID] = make([]int64, 0)
		}
		stationsToBusses[stop.stationID] = append(stationsToBusses[stop.stationID], busID)
	}
	rows.Close()

	terminals := make(map[int64]struct{})
	terminalNames := make(map[string]map[int64]struct{})
	for _, stops := range busStops {
		if _, has := terminals[stops[0].stationID]; !has {
			terminals[stops[0].stationID] = struct{}{}
		}

		if _, has := terminalNames[stops[0].stationName]; !has {
			terminalNames[stops[0].stationName] = make(map[int64]struct{})
			terminalNames[stops[0].stationName][stops[0].stationID] = struct{}{}
		} else {
			terminalNames[stops[0].stationName][stops[0].stationID] = struct{}{}
		}

		if _, has := terminals[stops[len(stops)-1].stationID]; !has {
			terminals[stops[len(stops)-1].stationID] = struct{}{}
		}

		if _, has := terminalNames[stops[len(stops)-1].stationName]; !has {
			terminalNames[stops[len(stops)-1].stationName] = make(map[int64]struct{})
			terminalNames[stops[len(stops)-1].stationName][stops[len(stops)-1].stationID] = struct{}{}
		} else {
			terminalNames[stops[len(stops)-1].stationName][stops[len(stops)-1].stationID] = struct{}{}
		}
	}

	var outsideStations strings.Builder
	var insideStations strings.Builder
	insideStations.WriteString("const stations = [")
	outsideStations.WriteString("const metroStations = [")
	for _, station := range stations {
		if station.IsOutsideCity {
			data := fmt.Sprintf("{%q:%d,%q:%q,%q:%q,%q:%.07f,%q:%.07f,%q:true", "i", station.OSMID, "n", station.Name, "s", station.StreetName, "lt", station.Lat, "ln", station.Lon, "o")
			if station.HasBoard {
				data += fmt.Sprintf(",%q:true", "b")
			}

			if _, has := terminals[station.OSMID]; has {
				data += fmt.Sprintf(",%q:true", "t")
			}

			if bussesIDs, has := stationsToBusses[station.OSMID]; has {
				data += fmt.Sprintf(",%q:[", "l")
				for i, busID := range bussesIDs {
					if i > 0 {
						data += ","
					}
					data += fmt.Sprintf("%d", busID)
				}
				data += "]"
			} else {
				data += fmt.Sprintf(",%q:false", "l")
			}

			if stationIDs, has := stationNamesToIDs[station.Name]; has {
				if len(stationIDs) == 2 {
					for _, stationID := range stationIDs {
						if stationID == station.OSMID {
							continue
						}
						data += fmt.Sprintf(",%q:%d", "ss", stationID)
					}
				}
			} else if len(stationIDs) > 0 {
				data += fmt.Sprintf(",%q:[", "sss")
				wroteOne := false
				for _, stationID := range stationIDs {
					if stationID == station.OSMID {
						continue
					}
					if wroteOne {
						data += ","
					}
					data += fmt.Sprintf("%d", stationID)
					if !wroteOne {
						wroteOne = true
					}
				}
				data += "]"
			}

			data += "},\n"
			outsideStations.WriteString(data)
		} else {
			data := fmt.Sprintf("{%q:%d,%q:%q,%q:%q,%q:%.07f,%q:%.07f", "i", station.OSMID, "n", station.Name, "s", station.StreetName, "lt", station.Lat, "ln", station.Lon)
			if station.HasBoard {
				data += fmt.Sprintf(",%q:true", "b")
			}

			if ids, has := terminalNames[station.Name]; has && len(ids) > 2 {
				data += fmt.Sprintf(",%q:true", "t")
			}

			if bussesIDs, has := stationsToBusses[station.OSMID]; has {
				data += fmt.Sprintf(",%q:[", "l")
				for i, busID := range bussesIDs {
					if i > 0 {
						data += ","
					}
					data += fmt.Sprintf("%d", busID)
				}
				data += "]"
			} else {
				data += fmt.Sprintf(",%q:false", "l")
			}

			if stationIDs, has := stationNamesToIDs[station.Name]; has {
				if len(stationIDs) == 2 {
					for _, stationID := range stationIDs {
						if stationID == station.OSMID {
							continue
						}
						data += fmt.Sprintf(",%q:%d", "ss", stationID)
					}
				} else if len(stationIDs) > 1 {
					data += fmt.Sprintf(",%q:[", "sss")
					wroteOne := false
					for _, stationID := range stationIDs {
						if stationID == station.OSMID {
							continue
						}
						if wroteOne {
							data += ","
						}
						data += fmt.Sprintf("%d", stationID)
						if !wroteOne {
							wroteOne = true
						}
					}
					data += "]"
				}
			}

			data += "},\n"
			insideStations.WriteString(data)
		}
	}

	insideStations.WriteString("]\nexport default stations;\n")
	err = os.WriteFile("./../../frontend/web/src/urban_stations.js", []byte(insideStations.String()), 0644)
	if err != nil {
		t.Fatalf("error writing urban_stations.js : %#v", err)
	}

	outsideStations.WriteString("]\nexport default metroStations;\n")
	err = os.WriteFile("./../../frontend/web/src/metro_stations.js", []byte(outsideStations.String()), 0644)
	if err != nil {
		t.Fatalf("error writing urban_stations.js : %#v", err)
	}
}

func TestGenerateBussesJS(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	repo, err := NewRepository(logger, "./../../data/brasov_busses.sqlite")
	if err != nil {
		t.Fatalf("error creating repository:%#v", err)
	}

	busses, err := repo.GetBusses(false)
	if err != nil {
		t.Fatalf("error reading busses : %#v", err)
	}

	siblingsMap := make(map[int64]int64)
	for i, line1 := range busses {
		if _, has := siblingsMap[line1.OSMID]; !has {
			for j, line2 := range busses {
				if line1.Line == line2.Line && line1.OSMID != line2.OSMID {
					busses[i].SiblingOSMID = line2.OSMID
					busses[j].SiblingOSMID = line1.OSMID
					siblingsMap[line1.OSMID] = line2.OSMID
					siblingsMap[line2.OSMID] = line1.OSMID
					break
				}
			}
		}
	}

	var urbanBusses strings.Builder
	var metropolitanBusses strings.Builder
	urbanBusses.WriteString("const busses = [")
	metropolitanBusses.WriteString("const metroBusses = [")
	for _, line := range busses {
		from := replaceDiacritics(line.From)
		to := replaceDiacritics(line.To)
		stationIDs, err := repo.GetStationsForBus(line.OSMID)
		if err != nil {
			t.Fatalf("error getting stations ids for bus: %#v", err)
		}

		ids := ""
		for i, stationID := range stationIDs {
			if i > 0 {
				ids += ","
			}
			ids += strconv.Itoa(int(stationID))
		}

		if line.IsMetropolitan {
			data := fmt.Sprintf("{%q:%d,%q:%d,%q:%q,%q:%q,%q:%q,%q:%q,%q:%q,%q:%d,%q:[%s]},\n", "i", line.OSMID, "si", line.SiblingOSMID, "b", line.Name, "f", from, "t", to, "n", line.Line, "c", line.Color, "d", line.Dir, "s", ids)

			metropolitanBusses.WriteString(data)
		} else {
			data := fmt.Sprintf("{%q:%d,%q:%d,%q:%q,%q:%q,%q:%q,%q:%q,%q:%q,%q:%d,%q:[%s]},\n", "i", line.OSMID, "si", line.SiblingOSMID, "b", line.Name, "f", from, "t", to, "n", line.Line, "c", line.Color, "d", line.Dir, "s", ids)

			urbanBusses.WriteString(data)
		}
	}

	urbanBusses.WriteString("]\nexport default busses;\n")
	err = os.WriteFile("./../../frontend/web/src/urban_busses.js", []byte(urbanBusses.String()), 0644)
	if err != nil {
		t.Fatalf("error writing urban_busses.js : %#v", err)
	}

	metropolitanBusses.WriteString("]\nexport default metroBusses;\n")
	err = os.WriteFile("./../../frontend/web/src/metro_busses.js", []byte(metropolitanBusses.String()), 0644)
	if err != nil {
		t.Fatalf("error writing urban_busses.js : %#v", err)
	}
}

func TestGenerateTiles(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	repo, err := NewRepository(logger, "./../../data/brasov_busses.sqlite")
	if err != nil {
		t.Fatalf("error creating repository:%#v", err)
	}

	data, err := ReadPBFData(logger, "./../../data/brasov.osm.pbf", repo)
	if err != nil {
		logger.Error("error reading PBF data", "err", err)
		os.Exit(1)
	}

	defaultColor, _ := parseHexColor(LightGrey, 1)
	styles := make(map[string]map[string]Style)
	styles[DefaultTag] = make(map[string]Style)
	styles[DefaultTag][DefaultTag] = Style{Color: defaultColor, Width: 2}

	specialColor, _ := parseHexColor(DarkYellow, 1)
	styles[SpecialTag] = make(map[string]Style)
	styles[SpecialTag][SpecialTag] = Style{Color: specialColor, Width: 5}

	boundaryColor, _ := parseHexColor(LightYellow, 1)
	styles["boundary"] = make(map[string]Style)
	styles["boundary"]["administrative"] = Style{Color: boundaryColor, Width: 3}

	var wg sync.WaitGroup
	tasks := make(chan [2]interface{}, 8)

	font_, err := os.ReadFile("./../../frontend/web/public/Roboto-Regular.ttf")
	if err != nil {
		logger.Warn("error reading font file", "err", err)
	}

	font, err := freetype.ParseFont(font_)
	if err != nil {
		logger.Warn("error parsing font file", "err", err)
	}

	dimmedYellow := color.RGBA{R: 0xF5, G: 0xB3, B: 0x01, A: 0xFF}

	t.Logf("start %d workers", runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func(num int) {
			count := 0
			for task := range tasks {
				zoom := task[0].(int)
				xyz := task[1].(XYZ)

				dirPath := fmt.Sprintf("./../../frontend/web/public/%d/%d", zoom, xyz.X)
				if err := os.MkdirAll(dirPath, 0755); err != nil {
					t.Fatalf("error making folders: %#v", err)
				}

				northWestPoint := GetPointByCoords(xyz.X, xyz.Y, zoom)
				southEastPoint := GetPointByCoords(xyz.X+1, xyz.Y+1, zoom)

				img := image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))
				draw.Draw(img, img.Bounds(), image.Transparent, image.ZP, draw.Src)

				result := &Tile{
					image:     img,
					zoom:      zoom,
					tileSize:  tileSize,
					northWest: northWestPoint,
					southEast: southEastPoint,
					p1:        s2.PointFromLatLng(s2.LatLngFromDegrees(northWestPoint.Lat, northWestPoint.Lon)),
					p2:        s2.PointFromLatLng(s2.LatLngFromDegrees(southEastPoint.Lat, northWestPoint.Lon)),
					p3:        s2.PointFromLatLng(s2.LatLngFromDegrees(northWestPoint.Lat, southEastPoint.Lon)),
					p4:        s2.PointFromLatLng(s2.LatLngFromDegrees(southEastPoint.Lat, southEastPoint.Lon)),
					styles:    styles,
				}

				features := result.Draw(data)
				for _, feature := range features {
					if feature.IsWay() {
						continue
					}

					place, ok := feature.(Node)
					if ok {
						placeX, placeY := result.GetRelativeXY(place)
						err = drawText(img, font, 25, dimmedYellow, int(placeX), int(placeY), place.Tags["name"])
						if err != nil {
							logger.Warn("error drawing text", "err", err)
						}
					}
				}

				filePath := fmt.Sprintf("%d/%d/%d.png", zoom, xyz.X, xyz.Y)
				out, err := os.Create("./../../frontend/web/public/" + filePath)
				if err != nil {
					t.Fatalf("error creating png file: %#v", err)
				}

				err = png.Encode(out, result.image)
				if err != nil {
					logger.Error("error encoding PNG tile", "err", err)
					return
				}

				err = out.Close()
				if err != nil {
					t.Fatalf("error closing file:%#v", err)
				}

				wg.Done()
				count++
				if count%1000 == 0 {
					t.Logf("1000 saved.")
				}
			}

			t.Logf("goroutine %d done", num)
		}(i)
	}

	for _, zoom := range []int{13, 14, 15} {
		xyzts, bounds := GetTilesInBBoxForZoom(45.00, 25.00, 46.00, 26.00, zoom)
		t.Logf("saving %d tiles", len(xyzts))
		for _, xyz := range xyzts {
			wg.Add(1)
			tasks <- [2]interface{}{zoom, xyz}
		}
		t.Logf("%d tiles saved for zoom %d X from %d to %d Y from %d to %d", len(xyzts), zoom, bounds.XFrom, bounds.XTo, bounds.YFrom, bounds.YTo)
	}

	wg.Wait()
	close(tasks)
}

func TestGenerateTimeTables(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	repo, err := NewRepository(logger, "./../../data/brasov_busses.sqlite")
	if err != nil {
		t.Fatalf("error creating repository:%#v", err)
	}

	data, err := repo.GetAllTimetables()
	if err != nil {
		t.Fatalf("error getting timetables: %#v", err)
	}

	for _, station := range data {
		var stationTimetable strings.Builder
		stationTimetable.WriteRune('[')
		for j, line := range station.Lines {
			if j > 0 {
				stationTimetable.WriteRune(',')
			}
			stationTimetable.WriteString(fmt.Sprintf("{%q:%d,%q:[", "b", line.BusOSMID, "t"))
			for i, time := range line.Times {
				if i > 0 {
					stationTimetable.WriteRune(',')
				}
				stationTimetable.WriteString(fmt.Sprintf("%d", time))
			}
			stationTimetable.WriteString("]}")
		}
		stationTimetable.WriteRune(']')
		err = os.WriteFile(fmt.Sprintf("./../../frontend/web/public/tt/%d.json", station.OSMID), []byte(stationTimetable.String()), 0644)
		if err != nil {
			t.Fatalf("error writing urban_busses.js : %#v", err)
		}
	}
	t.Logf("%d jsons written", len(data))
}
