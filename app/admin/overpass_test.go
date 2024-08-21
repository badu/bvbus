package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type stationAndStop struct {
	stopID    int64
	stationID int64
	index     int
}

func GetStationFromOverpass(stationID int64) (*Data, error) {
	query := fmt.Sprintf("data=[out:json];node(%d);out body;>;out skel qt;", stationID)
	response, err := http.Post("https://overpass-api.de/api/interpreter", "text/plain", strings.NewReader(query))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var result Data
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	response.Body.Close()
	return &result, nil
}

func TestGetOverpassData(t *testing.T) {
	db, err := sql.Open("sqlite3", "./../../data/brasov_busses.db")
	if err != nil {
		t.Fatalf("error:%#v", err)
	}

	rows, err := db.Query(`SELECT id, name FROM busses ORDER BY id;`)
	if err != nil {
		t.Fatalf("error:%#v", err)
	}
	for rows.Next() {
		var busID int64
		var busName string
		err := rows.Scan(&busID, &busName)
		if err != nil {
			t.Fatalf("error scanning:%#v", err)
		}

		fileName := fmt.Sprintf("./../../data/busses/%d.json", busID)

		bussesQuery := fmt.Sprintf("data=[out:json];relation(%d);out body;>;out skel qt;", busID)
		bussesResponse, err := http.Post("https://overpass-api.de/api/interpreter", "text/plain", strings.NewReader(bussesQuery))
		if err != nil {
			t.Fatalf("error:%#v", err.Error())
		}

		busses, err := io.ReadAll(bussesResponse.Body)
		if err != nil {
			t.Fatalf("error:%#v", err)
		}

		err = os.WriteFile(fileName, busses, 0644)
		if err != nil {
			t.Fatalf("error writing urban_busses.js : %#v", err)
		}

		bussesResponse.Body.Close()
	}

	rows.Close()
	db.Close()
}

func TestMakeTrajectories(t *testing.T) {
	db, err := sql.Open("sqlite3", "./../../data/brasov_busses.db")
	if err != nil {
		t.Fatalf("error:%#v", err)
	}

	rows, err := db.Query(`SELECT id, name FROM busses ORDER BY id;`)
	if err != nil {
		t.Fatalf("error:%#v", err)
	}
	busNames := make(map[int64]string)

	for rows.Next() {
		var busID int64
		var busName string
		err := rows.Scan(&busID, &busName)
		if err != nil {
			t.Fatalf("error scanning:%#v", err)
		}
		busNames[busID] = busName
	}
	rows.Close()

	rows, err = db.Query(`SELECT id, name FROM stations ORDER BY id;`)
	if err != nil {
		t.Fatalf("error:%#v", err)
	}

	stations := make(map[int64]*StationRoute)
	for rows.Next() {
		var station StationRoute
		err := rows.Scan(&station.Id, &station.Name)
		if err != nil {
			t.Fatalf("error:%#v", err)
		}

		if _, has := stations[station.Id]; !has {
			station.Busses = make([]int64, 0)
			station.BussesMap = make(map[int64]struct{})
			stations[station.Id] = &station
		} else {
			t.Fatalf("already seen %q station", station.Name)
		}
	}
	rows.Close()

	trajectories := make(map[int64][]*Node)
	stationsAndStops := make(map[int64][]stationAndStop)
	for busID := range busNames {
		stats := make([]int64, 0)
		stops := make([]int64, 0)
		trajectories[busID] = make([]*Node, 0)
		stationsAndStops[busID] = make([]stationAndStop, 0)
		fileName := fmt.Sprintf("./../../data/busses/%d.json", busID)

		busses, err := os.ReadFile(fileName)
		if err != nil {
			return
		}

		var bussesData Data
		err = json.Unmarshal(busses, &bussesData)
		if err != nil {
			t.Fatalf("error:%#v", err)
		}

		uniqueWays := make(map[int64]*Node)
		uniqueNodes := make(map[int64]*Node)
		relationWays := make([]Member, 0)

		for _, element := range bussesData.Elements {
			switch element.Type {
			case OSMWay:
				if _, has := uniqueWays[element.ID]; has {
					t.Fatalf("error : way exists")
				}
				uniqueWays[element.ID] = &element
			case OSMNode:
				if _, has := uniqueNodes[element.ID]; has {
					t.Fatalf("error : node exists")
				}
				uniqueNodes[element.ID] = &element
			case OSMRelation:
				for _, member := range element.Members {
					switch member.Type {
					case OSMNode:
						switch member.Role {
						case OSMPlatform, OSMPlatformEntryOnly, OSMPlatformExitOnly:
							stats = append(stats, member.Ref)
						case OSMStop, OSMStopEntryOnly, OSMStopExitOnly:

							stops = append(stops, member.Ref)
						}
					case OSMWay:
						relationWays = append(relationWays, member)
					}
				}
			}
		}

		if len(stops) != len(stats) {
			t.Fatalf("error : %d != %d", len(stops), len(stats))
		}

		for i := 0; i < len(stops); i++ {
			data := stationAndStop{
				index:     i,
				stopID:    stops[i],
				stationID: stats[i],
			}
			stationsAndStops[busID] = append(stationsAndStops[busID], data)
		}

		lastNodeID := stops[0]
		firstWayNode, firstWayNodeFound := uniqueNodes[stops[0]]
		if !firstWayNodeFound {
			t.Logf("WAY NODE NOT FOUND: %d", stops[0])
			break
		}

		trajectories[busID] = append(trajectories[busID], firstWayNode)
		for _, wayMember := range relationWays {
			way, hasFoundWay := uniqueWays[wayMember.Ref]
			if !hasFoundWay {
				t.Fatalf("ERROR FINDING WAY: %d", wayMember.Ref)
			}

			if len(way.Nodes) <= 0 {
				t.Fatalf("empty way detected")
			}

			inReverse := false
			if lastNodeID != way.Nodes[0] {
				inReverse = true
			}

			if inReverse {
				for wayNodeIndex := len(way.Nodes) - 1; wayNodeIndex >= 0; wayNodeIndex-- {
					wayNodeID := way.Nodes[wayNodeIndex]
					if lastNodeID == wayNodeID {
						continue
					}

					wayNode, wayNodeFound := uniqueNodes[wayNodeID]
					if !wayNodeFound {
						t.Logf("WAY NODE NOT FOUND: %d", wayNodeID)
						break
					}

					trajectories[busID] = append(trajectories[busID], wayNode)

					lastNodeID = wayNodeID

				}
			} else {
				for wayNodeIndex := 0; wayNodeIndex < len(way.Nodes); wayNodeIndex++ {
					wayNodeID := way.Nodes[wayNodeIndex]
					if lastNodeID == wayNodeID {
						continue
					}

					wayNode, wayNodeFound := uniqueNodes[wayNodeID]
					if !wayNodeFound {
						t.Logf("WAY NODE NOT FOUND: %d", wayNodeID)
						break
					}

					trajectories[busID] = append(trajectories[busID], wayNode)

					lastNodeID = wayNodeID
				}
			}
		}
	}

	stationsToBusMap := make(map[string]map[int64]struct{})
	singleBusKey := make(map[string]int64)
	for busID := range busNames {
		data := stationsAndStops[busID]
		for i := 1; i < len(data); i++ {
			key := fmt.Sprintf("%d-%d", data[i-1].stationID, data[i].stationID)
			if _, has := stationsToBusMap[key]; !has {
				stationsToBusMap[key] = make(map[int64]struct{})
			}
			if _, has := stationsToBusMap[key][busID]; has {
				t.Fatalf("ERROR %q %d exists", key, busID)
			}
			stationsToBusMap[key][busID] = struct{}{}
			if _, has := singleBusKey[key]; !has {
				singleBusKey[key] = busID
			}
		}
	}

	const writeFiles = false
	const writeToDatabase = false
	stationsDistances := make(map[string]float64)
	wroteDistance := make(map[string]struct{})
	for busID, trajectory := range trajectories {
		currentDistance := 0.0

		var lastStationID int64
		var prevNode *Node
		var psb strings.Builder

		pointsWritten := 0
		stopIndex := 0

		nodesKeeper := make([]*Node, 0)
		for i, node := range trajectory {
			if len(stationsAndStops[busID])-1 < stopIndex {
				t.Logf("%d %d of %d", busID, stopIndex, len(stationsAndStops[busID]))
				continue
			}
			stopInfo := stationsAndStops[busID][stopIndex]
			if i == 0 {
				if stopInfo.stopID != node.ID {
					t.Fatalf("first stop index is bad %d should be %d", node.ID, stopInfo.stopID)
				}

				lastStationID = stopInfo.stationID
				prevNode = trajectory[i]
				psb.WriteString(fmt.Sprintf("{%q:%d,%q:%.08f,%q:%.08f}", "i", node.ID, "lt", node.Lat, "ln", node.Lon))
				pointsWritten++
				stopIndex++

				nodesKeeper = append(nodesKeeper, &Node{
					ID:     node.ID,
					Lat:    node.Lat,
					Lon:    node.Lon,
					IsStop: true,
				})
				continue
			}

			if stopInfo.stopID != node.ID {
				// t.Logf("point %d => %d lat %f lng %f", i, node.ID, node.Lat, node.Lon)
				psb.WriteRune(',')
				currentDistance = currentDistance + Haversine(prevNode.Lat, prevNode.Lon, node.Lat, node.Lon)
				psb.WriteString(fmt.Sprintf("{%q:%d,%q:%.08f,%q:%.08f}", "i", node.ID, "lt", node.Lat, "ln", node.Lon))
				pointsWritten++
				prevNode = trajectory[i]
				nodesKeeper = append(nodesKeeper, node)
			} else {
				currentDistance = currentDistance + Haversine(prevNode.Lat, prevNode.Lon, node.Lat, node.Lon)
				psb.WriteRune(',')
				psb.WriteString(fmt.Sprintf("{%q:%d,%q:%.08f,%q:%.08f,%q:true}", "i", node.ID, "lt", node.Lat, "ln", node.Lon, "s"))
				pointsWritten++
				if lastStationID == 0 {
					t.Fatalf("error : previous station is zero and should be set")
				}

				station, hasCurr := stations[stopInfo.stationID]
				if hasCurr {
					//t.Logf("stop %d => %d lat %f lng %f [stop index %d, station %s, stop id %d] %f meters", i, node.ID, node.Lat, node.Lon, stopInfo.index, station.Name, stopInfo.stopID, currentDistance)
				} else {
					//t.Logf("stop %d => %d lat %f lng %f [stop index %d, STATION ID %d, stop id %d] %f meters", i, node.ID, node.Lat, node.Lon, stopInfo.index, stopInfo.stationID, stopInfo.stopID, currentDistance)
				}

				if !hasCurr {
					data, err := GetStationFromOverpass(stopInfo.stationID)
					if err != nil {
						t.Fatalf("error getting station from overpass: %#v", err)
					}

					t.Logf("should create current %d %q lat %f long %f board %t outside %t", data.Elements[0].ID, data.Elements[0].Tags["name"], data.Elements[0].Lat, data.Elements[0].Lon, data.Elements[0].Tags["departures_board"] == "realtime", len(data.Elements[0].Tags["fare_zone"]) > 0)
				}

				prevStation, hasPrev := stations[lastStationID]
				if !hasPrev {
					data, err := GetStationFromOverpass(stopInfo.stationID)
					if err != nil {
						t.Fatalf("error getting station from overpass: %#v", err)
					}

					t.Logf("should create previous %d %q lat %f long %f board %t outside %t", data.Elements[0].ID, data.Elements[0].Tags["name"], data.Elements[0].Lat, data.Elements[0].Lon, data.Elements[0].Tags["departures_board"] == "realtime", len(data.Elements[0].Tags["fare_zone"]) > 0)
				}

				if lastStationID == stopInfo.stationID {
					t.Fatalf("what is going on here? %d => %q %s - %s %d out of %d", busID, busNames[busID], prevStation.Name, station.Name, i, len(trajectory))
				}

				key := fmt.Sprintf("%d-%d", lastStationID, stopInfo.stationID)

				if _, has := stationsToBusMap[key]; !has {
					t.Fatalf("no, really, what is going on? %q on %d", key, busID)
				}

				nodesKeeper = append(nodesKeeper, &Node{
					ID:     node.ID,
					Lat:    node.Lat,
					Lon:    node.Lon,
					IsStop: true,
				})
				_, hasWrote := wroteDistance[key]
				if writeToDatabase && !hasWrote {
					tx, err := db.Begin()
					if err != nil {
						t.Fatalf("error beginning transaction: %#v", err)
					}

					stmt1, err := tx.Prepare(InsertDistanceSQL)
					if err != nil {
						tx.Rollback()
						t.Fatalf("error preparing distance SQL : %#v", err)
					}

					stmt2, err := tx.Prepare(InsertPointSQL)
					if err != nil {
						tx.Rollback()
						t.Fatalf("error preparing point SQL : %#v", err)
					}

					_, err = stmt1.Exec(lastStationID, stopInfo.stationID, currentDistance, 0)
					if err != nil {
						tx.Rollback()
						t.Fatalf("error inserting distance SQL : %#v", err)
					}

					for index, n := range nodesKeeper {
						_, err = stmt2.Exec(n.ID, n.Lat, n.Lon, index+1, n.IsStop, lastStationID, stopInfo.stationID)
						if err != nil {
							//tx.Rollback()
							t.Logf("error inserting point : %#v\n%d %d %d", err, n.ID, lastStationID, stopInfo.stationID)
						}
					}

					tx.Commit()
					wroteDistance[key] = struct{}{}
				} else {
					//t.Logf("%d points into database for %q between %q and %q", len(nodesKeeper), busNames[busID], prevStation.Name, station.Name)
				}

				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("{%q:%f,%q:[", "d", currentDistance, "p"))
				sb.WriteString(psb.String())
				sb.WriteRune(']')
				sb.WriteRune('}')

				if writeFiles {
					err = os.WriteFile(fmt.Sprintf("./../../frontend/web/public/pt/%s.json", key), []byte(sb.String()), 0644)
					if err != nil {
						t.Fatalf("error writing points json : %#v", err)
					}
				}
				stationsDistances[key] = currentDistance
				if hasPrev && hasCurr {
					//t.Logf("%q - %q = %f", prevStation.Name, station.Name, currentDistance)
				} else {
					//t.Logf("%d - %d = %f", lastStationID, stopInfo.stationID, currentDistance)
				}

				psb.Reset()
				psb.WriteString(fmt.Sprintf("{%q:%d,%q:%.08f,%q:%.08f,%q:true}", "i", node.ID, "lt", node.Lat, "ln", node.Lon, "s"))
				pointsWritten = 1
				currentDistance = 0.0
				lastStationID = stopInfo.stationID
				prevNode = trajectory[i]
				stopIndex++
				nodesKeeper = make([]*Node, 0)
			}
		}
	}

	if writeFiles {
		for busID, points := range trajectories {
			var sb strings.Builder
			sb.WriteRune('[')
			for i, point := range points {
				if i > 0 {
					sb.WriteRune(',')
				}
				sb.WriteString(fmt.Sprintf("{%q:%d,%q:%f,%q:%f}", "i", point.ID, "lt", point.Lat, "ln", point.Lon))
			}
			sb.WriteRune(']')
			err = os.WriteFile(fmt.Sprintf("./../../frontend/admin/public/trajectories/%d.json", busID), []byte(sb.String()), 0644)
			if err != nil {
				t.Fatalf("error writing points json : %#v", err)
			}
		}
	}

	type NewLinks struct {
		ID     int64
		Meters float64
	}

	sortedResult := make([]*Edge, 0)
	newDists := make(map[int64][]NewLinks)
	for stationKey, distance := range stationsDistances {
		busID, has := singleBusKey[stationKey]
		if !has {
			t.Fatalf("error finding bus for station key %s", stationKey)
		}

		parts := strings.Split(stationKey, "-")
		fromStationIDInt, _ := strconv.Atoi(parts[0])
		fromStationID := int64(fromStationIDInt)

		if _, hasDist := newDists[fromStationID]; !hasDist {
			newDists[fromStationID] = make([]NewLinks, 0)
		}

		toStationIDInt, _ := strconv.Atoi(parts[1])
		toStationID := int64(toStationIDInt)
		newDists[fromStationID] = append(newDists[fromStationID], NewLinks{ID: toStationID, Meters: distance})

		var firstTime uint16
		err = db.QueryRow(`SELECT enc_time FROM time_tables WHERE station_id = ? AND bus_id = ? LIMIT 1;`, fromStationID, busID).Scan(&firstTime)
		var secondTime uint16
		err = db.QueryRow(`SELECT enc_time FROM time_tables WHERE station_id = ? AND bus_id = ? LIMIT 1;`, toStationID, busID).Scan(&secondTime)

		firstResult := Time{}
		firstResult.Decompress(firstTime)
		secondResult := Time{}
		secondResult.Decompress(secondTime)
		minutes := secondResult.Diff(firstResult)
		normalized := uint16((distance / 600) + 1)
		if minutes > normalized {
			minutes = normalized
		}
		data := Edge{Meters: distance, Minutes: minutes, FromStationID: fromStationID, ToStationID: toStationID}

		//t.Logf("distance %s - %s = %f meters %d minutes", stations[fromStationID].Name, stations[toStationID].Name, distance, minutes)

		if writeToDatabase {
			_, err = db.Exec(`UPDATE distances SET minutes = ? WHERE from_station_id = ? AND to_station_id = ?`, minutes, fromStationID, toStationID)
			if err != nil {
				t.Fatalf("error updating distances: %#v", err)
			}
		}

		sortedResult = append(sortedResult, &data)
	}

	sort.Sort(ByDistance(sortedResult))

	if writeFiles {
		var sb strings.Builder
		sb.WriteString("const distances = [\n")
		c := 0
		for fromStationID, otherStations := range newDists {
			if c > 0 {
				sb.WriteRune(',')
				sb.WriteRune('\n')
			}
			sb.WriteString(fmt.Sprintf("{%q:%d,%q:[", "i", fromStationID, "s"))
			for i, data := range otherStations {
				if i > 0 {
					sb.WriteRune(',')
				}

				sb.WriteString(fmt.Sprintf("{%q:%d,%q:%f}", "t", data.ID, "m", data.Meters))
			}
			sb.WriteRune(']')
			sb.WriteRune('}')
			c++
		}
		sb.WriteString("]\nexport default distances;\n")

		err = os.WriteFile("./../../frontend/web/src/distances.js", []byte(sb.String()), 0644)
		if err != nil {
			t.Fatalf("error writing distances.js : %#v", err)
		}
	}

}
