package admin

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/mattn/go-sqlite3"
	"github.com/qedus/osmpbf"
)

func TestFirstImport(t *testing.T) {
	file, err := os.Open("./../../data/romania.osm.pbf")
	if err != nil {
		t.Fatalf("error opening PBF file:%#v", err)
	}
	defer file.Close()

	decoder := osmpbf.NewDecoder(file)
	err = decoder.Start(runtime.GOMAXPROCS(0) - 1)
	if err != nil {
		t.Fatalf("error starting decoder:%#v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	repo, err := NewRepository(logger, "./../../data/brasov_busses.db")
	if err != nil {
		t.Fatalf("error creating repository:%#v", err)
	}

	_, err = repo.DB.Exec("DELETE FROM busses WHERE 1=1;")
	if err != nil {
		t.Fatalf("error deleting busses:%#v", err)
	}
	_, err = repo.DB.Exec("DELETE FROM stations WHERE 1=1;")
	if err != nil {
		t.Fatalf("error deleting stations:%#v", err)
	}
	_, err = repo.DB.Exec("DELETE FROM bus_stops WHERE 1=1;")
	if err != nil {
		t.Fatalf("error deleting bus stops:%#v", err)
	}
	_, err = repo.DB.Exec("DELETE FROM street_points WHERE 1=1;")
	if err != nil {
		t.Fatalf("error deleting street points:%#v", err)
	}
	_, err = repo.DB.Exec("DELETE FROM street_rels WHERE 1=1;")
	if err != nil {
		t.Fatalf("error deleting street rels:%#v", err)
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

	pointsStmt, err := tx.Prepare("INSERT INTO street_points(id, lat, lng) VALUES (?, ?, ?);")
	if err != nil {
		t.Fatalf("error preparing street points statement:%#v", err)
	}
	defer pointsStmt.Close()

	relsStmt, err := tx.Prepare("INSERT INTO street_rels(point_id, point_index, bus_id, is_stop) VALUES (?, ?, ?, ?);")
	if err != nil {
		t.Fatalf("error preparing street rels statement:%#v", err)
	}
	defer relsStmt.Close()

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
					if err := tx.Rollback(); err != nil {
						t.Fatalf("error rolling back transaction on insert station :%#v", err)
					}
					t.Fatalf("error inserting station :%#v", err)
				}

				stations[station.OSMID] = &station
				// logger.Info("station created", "name", station.Name, "metropolitan", station.IsOutsideCity)
			case *osmpbf.Way:
				uniqueWays[v.ID] = Node{Nodes: v.NodeIDs, ID: v.ID, Tags: v.Tags}
			case *osmpbf.Relation:
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
						if tv == "Rețea transportului public din Brașov (Mteropolitan)" {
							for _, memberID := range v.Members {
								metropolitans = append(metropolitans, memberID.ID)
							}
							break
						} else if tv == "Rețea transportului public din Brașov (Urban)" {
							for _, memberID := range v.Members {
								urbans = append(urbans, memberID.ID)
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

				if !isValid {
					continue
				}

				if _, willSkip := excludedBusses[v.ID]; willSkip {
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

					if member.Role == "platform" || member.Role == "platform_entry_only" || member.Role == "platform_exit_only" {
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

					if member.Type == osmpbf.NodeType && (member.Role == "stop" || member.Role == "stop_exit_only" || member.Role == "stop_entry_only") {
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
	logger.Info("busses", "len", len(busses))
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

				if _, has := seen[wayNode.ID]; !has {
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
				}

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

					if _, has := seen[wayNode.ID]; !has {
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
					}

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

					seen[wayNode.ID] = struct{}{}
					pointIndex++
				}
			}
		}

		line, found := busses[busID]
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
