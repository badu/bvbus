package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func findAndRemove(busses []int64, busID int64) []int64 {
	for i, v := range busses {
		if v == busID {
			return append(busses[:i], busses[i+1:]...)
		}
	}
	return busses

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

func TestGetBussesWithPoints(t *testing.T) {
	const useOverpass = false
	db, err := sql.Open("sqlite3", "./../../data/brasov_busses.db")
	if err != nil {
		t.Fatalf("error:%#v", err)
	}

	rows, err := db.Query(`SELECT id, name FROM stations ORDER BY id;`)
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

	rows, err = db.Query(`SELECT id, name FROM busses ORDER BY id;`)
	if err != nil {
		t.Fatalf("error:%#v", err)
	}

	busNames := make(map[int64]string)
	allStopsMap := make(map[int64]struct{})
	stationsKeys := make(map[string]float64)
	for rows.Next() {
		var busID int64
		var busName string
		err := rows.Scan(&busID, &busName)
		if err != nil {
			t.Fatalf("error scanning:%#v", err)
		}
		busNames[busID] = busName
		goodBusses = findAndRemove(goodBusses, busID)

		fileName := fmt.Sprintf("./../../frontend/web/public/busses/%d.json", busID)
		var busses []byte
		if useOverpass {
			bussesQuery := fmt.Sprintf("data=[out:json];relation(%d);out body;>;out skel qt;", busID)
			bussesResponse, err := http.Post("https://overpass-api.de/api/interpreter", "text/plain", strings.NewReader(bussesQuery))
			if err != nil {
				t.Fatalf("error:%#v", err.Error())
			}

			busses, err = io.ReadAll(bussesResponse.Body)
			if err != nil {
				t.Fatalf("error:%#v", err)
			}

			err = os.WriteFile(fileName, busses, 0644)
			if err != nil {
				t.Fatalf("error writing urban_busses.js : %#v", err)
			}

			bussesResponse.Body.Close()
		} else {
			busses, err = os.ReadFile(fileName)
			if err != nil {
				return
			}
		}

		var bussesData Data
		err = json.Unmarshal(busses, &bussesData)
		if err != nil {
			t.Fatalf("error:%#v", err)
		}

		uniqueWays := make(map[int64]*Node)
		uniqueNodes := make(map[int64]*Node)
		relationWays := make([]Member, 0)
		type stationAndStop struct {
			stopID    int64
			stationID int64
			index     int
		}

		stops := make([]int64, 0)
		stats := make([]int64, 0)
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

		stationsAndStops := make([]stationAndStop, 0)
		stopsMap := make(map[int64]stationAndStop)
		for i := 0; i < len(stops); i++ {
			if prev, has := stopsMap[stops[i]]; has {
				t.Fatalf("stop already declared %d for %d at %d", stops[i], busID, prev.index)
			}
			data := stationAndStop{
				index:     i,
				stopID:    stops[i],
				stationID: stats[i],
			}
			stationsAndStops = append(stationsAndStops, data)
			stopsMap[stops[i]] = data
			allStopsMap[stops[i]] = struct{}{}
		}

		lastNodeID := stops[0]
		trajectory := make([]*Node, 0)
		firstWayNode, firstWayNodeFound := uniqueNodes[stops[0]]
		if !firstWayNodeFound {
			t.Logf("WAY NODE NOT FOUND: %d", stops[0])
			break
		}
		trajectory = append(trajectory, firstWayNode)
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

					trajectory = append(trajectory, wayNode)

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

					trajectory = append(trajectory, wayNode)

					lastNodeID = wayNodeID
				}
			}
		}

		currentDistance := 0.0

		var lastStationID int64
		var prevNode *Node
		var psb strings.Builder

		pointsWritten := 0
		for i, node := range trajectory {
			if i == 0 {
				stopInfo, isStop := stopsMap[node.ID]
				if !isStop {
					t.Fatalf("first point in trajectory should be stop, but it's not")
				}
				if stopInfo.index != 0 {
					t.Fatalf("first stop index should be zero, but it's %d", stopInfo.index)
				}
				lastStationID = stopInfo.stationID
				prevNode = trajectory[i]
				psb.WriteString(fmt.Sprintf("{%q:%d,%q:%.08f,%q:%.08f}", "i", node.ID, "lt", node.Lat, "ln", node.Lon))
				pointsWritten++
				continue
			}

			stopInfo, isStop := stopsMap[node.ID]
			if !isStop {
				// t.Logf("point %d => %d lat %f lng %f", i, node.ID, node.Lat, node.Lon)
				psb.WriteRune(',')
				currentDistance = currentDistance + Haversine(prevNode.Lat, prevNode.Lon, node.Lat, node.Lon)
				psb.WriteString(fmt.Sprintf("{%q:%d,%q:%.08f,%q:%.08f}", "i", node.ID, "lt", node.Lat, "ln", node.Lon))
				pointsWritten++
				prevNode = trajectory[i]
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
				prevStation, hasPrev := stations[lastStationID]

				key := fmt.Sprintf("%d-%d", lastStationID, stopInfo.stationID)
				if existingDistance, has := stationsKeys[key]; has {
					if existingDistance != currentDistance {
						if hasPrev && hasCurr {
							t.Logf("error : %q - %q [%d-%d] written with %f but current %f", prevStation.Name, station.Name, lastStationID, stopInfo.stationID, existingDistance, currentDistance)
						} else {
							t.Logf("error : %d - %d written with %f but current %f", lastStationID, stopInfo.stationID, existingDistance, currentDistance)
						}
					}

					psb.Reset()
					psb.WriteString(fmt.Sprintf("{%q:%d,%q:%.08f,%q:%.08f,%q:true}", "i", node.ID, "lt", node.Lat, "ln", node.Lon, "s"))
					pointsWritten = 1
					currentDistance = 0.0
					lastStationID = stopInfo.stationID
					prevNode = trajectory[i]
					continue
				}

				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("{%q:%f,%q:[", "d", currentDistance, "p"))
				sb.WriteString(psb.String())
				sb.WriteRune(']')
				sb.WriteRune('}')

				err = os.WriteFile(fmt.Sprintf("./../../frontend/web/public/pt/%s.json", key), []byte(sb.String()), 0644)
				if err != nil {
					t.Fatalf("error writing points json : %#v", err)
				}

				stationsKeys[key] = currentDistance
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
			}
		}
	}

	rows.Close()
	db.Close()

	for _, busID := range goodBusses {
		t.Logf("bus missing ? %d", busID)
	}
}
