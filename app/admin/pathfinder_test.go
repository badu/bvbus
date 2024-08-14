package admin

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

/*
*
1. Each transfer incurs a 10 minute penalty (unless it is a timed transfer) in the riders mind. That is to say mentally a trip involving a single bus that takes 40 minutes is roughly equivalent to a 30minute trip that requires a transfer.
2. Maximum distance that most people are willing to walk to a bus stop is 1/4 mile. Train station / Light rail about 1/2 mile.
3. Distance is irrelevant to the public transportation rider. (Only time is important)
4. Frequency matters (if a connection is missed how long until the next bus). Riders will prefer more frequent service options if the alternative is being stranded for an hour for the next express.
5. Rail has a higher preference than bus ( more confidence that the train will come and be going in the right direction)
6. Having to pay a new fare is a big hit. (add about a 15-20min penalty)
7. Total trip time matters as well (with above penalties)
8. How seamless is the connect? Does the rider have to exist a train station cross a busy street? Or is it just step off a train and walk 4 steps to a bus?
9. Crossing busy streets -- another big penalty on transfers -- may miss connection because can't get across street fast enough.
*/
type Edge struct {
	FromStationID int64
	ToStationID   int64
	Meters        float64
	Minutes       uint16
	Walkable      bool
}

type ByDistance []*Edge

func (a ByDistance) Len() int { return len(a) }

func (a ByDistance) Less(i, j int) bool {
	return a[i].Meters > a[j].Meters
}

func (a ByDistance) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type Graph struct {
	Nodes map[int64]*Station
	Edges map[string]*Edge
}

func GetStationsAndDistances(logger *slog.Logger, db *sql.DB) (*Graph, error) {
	rows, err := db.Query(`SELECT from_station, to_station FROM busses WHERE metropolitan = 0 ORDER BY id;`)
	if err != nil {
		logger.Error("error querying bus", "err", err)
		return nil, err
	}

	terminals := make(map[string]struct{})
	for rows.Next() {
		var from, to string
		err := rows.Scan(&from, &to)
		if err != nil {
			logger.Error("error scanning bus", "err", err)
			return nil, err
		}
		from = replaceDiacritics(from)
		to = replaceDiacritics(to)
		if _, has := terminals[from]; !has {
			terminals[from] = struct{}{}
		}
		if _, has := terminals[to]; !has {
			terminals[to] = struct{}{}
		}
	}
	rows.Close()

	rows, err = db.Query(`SELECT id, name, lat, lng FROM stations WHERE outside = 0 ORDER BY id;`)
	if err != nil {
		logger.Error("error querying stations", "err", err)
		return nil, err
	}

	stationz := make([]*Station, 0)
	stations := make(map[int64]*Station)
	for rows.Next() {
		var station Station
		err := rows.Scan(
			&station.OSMID,
			&station.Name,
			&station.Lat,
			&station.Lon,
		)
		if err != nil {
			logger.Error("error scanning station", "err", err)
			return nil, err
		}

		if _, has := stations[station.OSMID]; has {
			logger.Error("error station exists", "id", station.OSMID)
			return nil, fmt.Errorf("station already seen id = %d", station.OSMID)
		}

		if _, isTerminal := terminals[station.Name]; isTerminal {
			station.IsTerminal = true
		}

		stations[station.OSMID] = &station
		stationz = append(stationz, &station)
	}
	rows.Close()

	distances := make(map[string]*Edge)
	rows, err = db.Query(`SELECT from_station_id, to_station_id, distance FROM distances;`)
	if err != nil {
		logger.Error("error querying stations", "err", err)
		return nil, err
	}

	for rows.Next() {
		var fromStationID, toStationID int64
		var distance float64
		err := rows.Scan(&fromStationID, &toStationID, &distance)
		if err != nil {
			logger.Error("error scanning station", "err", err)
			return nil, err
		}

		if _, has := stations[fromStationID]; !has {
			continue
		}

		if _, has := stations[toStationID]; !has {
			continue
		}

		key := fmt.Sprintf("%d-%d", fromStationID, toStationID)
		if _, has := distances[key]; has {
			logger.Error("error: map already contains key", "key", key)
			continue
		}
		distances[key] = &Edge{FromStationID: fromStationID, ToStationID: toStationID, Meters: distance}
	}
	rows.Close()

	for i := 0; i < len(stationz); i++ {
		for j := 0; j < len(stationz); j++ {
			if i == j {
				continue
			}

			key := fmt.Sprintf("%d-%d", stationz[i].OSMID, stationz[j].OSMID)
			if _, has := distances[key]; has {
				continue
			}

			distance := DistanceOnEdges(stationz[i].Lat, stationz[i].Lon, stationz[j].Lat, stationz[j].Lon)
			if distance < 210 { /* 4 minutes walking */
				distances[key] = &Edge{FromStationID: stationz[i].OSMID, ToStationID: stationz[j].OSMID, Meters: distance, Walkable: true}
				logger.Info("distance", "fromID", stationz[i].OSMID, "from", stationz[i].Name, "toID", stationz[j].OSMID, "to", stationz[j].Name, "d", distance)
			}
		}
	}

	return &Graph{Nodes: stations, Edges: distances}, nil
}

func TestPathFinderWithNum(t *testing.T) {
	const writeToDisk = false

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := sql.Open("sqlite3", "./../../data/brasov_busses.db")
	if err != nil {
		t.Fatalf("Error opening SQLite database: %v", err)
	}
	defer db.Close()

	result, err := GetStationsAndDistances(logger, db)
	if err != nil {
		t.Fatalf("error: %#v", err)
	}

	graph := simple.NewDirectedGraph()
	for _, station := range result.Nodes {
		graph.AddNode(station)
	}

	for _, link := range result.Edges {
		startStation, hasStart := result.Nodes[link.FromStationID]
		if !hasStart {
			t.Fatalf("error finding start station %d (%t) %f", link.FromStationID, link.Walkable, link.Meters)
		}
		endStation, hasEnd := result.Nodes[link.ToStationID]
		if !hasEnd {
			t.Fatalf("error finding end station %d", link.FromStationID)
		}

		if graph.HasEdgeFromTo(link.FromStationID, link.ToStationID) {
			t.Logf("SKIPPING EDGE %s-%s because it exists", result.Nodes[link.FromStationID].Name, result.Nodes[link.ToStationID].Name)
		} else {
			edge := simple.WeightedEdge{
				F: startStation,
				T: endStation,
				W: link.Meters,
			}

			sn := graph.Node(link.FromStationID)
			if sn == nil {
				t.Fatalf("start node is not a graph node")
			}
			en := graph.Node(link.ToStationID)
			if en == nil {
				t.Fatalf("end node is not a graph node")
			}
			graph.SetEdge(edge)
		}
	}

	var graphSb strings.Builder
	graphSb.WriteRune('{')
	graphSb.WriteString(fmt.Sprintf("%q:[", "nodes"))
	it := graph.Nodes()
	q := 0
	for it.Next() {
		station, has := result.Nodes[it.Node().ID()]
		if !has {
			logger.Error("node", "NOT FOUND", it.Node().ID())
			continue
		}
		if q > 0 {
			graphSb.WriteRune(',')
		}

		graphSb.WriteString(fmt.Sprintf("{%q:%d,%q:%.08f,%q:%.08f,%q:%q}", "id", station.OSMID, "lt", station.Lat, "ln", station.Lon, "n", station.Name))
		q++
	}
	graphSb.WriteRune(']')
	graphSb.WriteRune(',')

	pairs := make(map[string]struct{})
	ed := graph.Edges()
	graphSb.WriteString(fmt.Sprintf("%q:[", "edges"))
	q = 0
	for ed.Next() {
		edge := ed.Edge()
		from, hasFrom := result.Nodes[edge.From().ID()]
		if !hasFrom {
			logger.Error("edge", "FROM NOT FOUND", edge.From().ID())
			continue
		}

		to, hasTo := result.Nodes[edge.To().ID()]
		if !hasTo {
			logger.Error("edge", "TO NOT FOUND", edge.To().ID())
			continue
		}

		if _, has := pairs[fmt.Sprintf("%d-%d", from.OSMID, to.OSMID)]; has {
			logger.Error("ERROR", "seen", fmt.Sprintf("%d-%d", from.OSMID, to.OSMID))
		} else {
			pairs[fmt.Sprintf("%d-%d", from.OSMID, to.OSMID)] = struct{}{}
		}

		if q > 0 {
			graphSb.WriteRune(',')
		}
		graphSb.WriteString(fmt.Sprintf("{%q:%d,%q:%d}", "f", from.OSMID, "t", to.OSMID))
		q++
		// logger.Info("EDGE", "from", from.Name, "to", to.Name, "fromID", from.OSMID, "toID", to.OSMID)
	}
	graphSb.WriteRune(']')
	graphSb.WriteRune('}')

	err = os.WriteFile("./../../frontend/web/public/graph.json", []byte(graphSb.String()), 0644)
	if err != nil {
		t.Fatalf("error writing graph.json : %#v", err)
	}

	for startStationID := range result.Nodes {
		startStation, startStationFound := result.Nodes[startStationID]
		if !startStationFound {
			t.Fatalf("error looking up start station %d", startStationID)
		}

		solutions, ok := path.BellmanFordAllFrom(startStation, graph)
		if !ok {
			t.Fatalf("bellman ford error")
		}

		var sb strings.Builder
		sb.WriteRune('[')
		wroteOne := false
		for endStationID, endStation := range result.Nodes {
			if startStation.OSMID == endStation.OSMID {
				continue
			}

			allPaths, _ := solutions.AllTo(endStation.OSMID)
			if len(allPaths) == 0 {
				continue
			}

			if wroteOne {
				sb.WriteRune(',')
			}

			sb.WriteString(fmt.Sprintf("{%q:%d,%q:[", "t", endStationID, "s"))

			for i, solution := range allPaths {
				if i > 5 { // take only first 5 solutions
					//t.Logf("%s - %s has %d solutions", startStation.Name, endStation.Name, len(allPaths))
					break
				}

				if i > 0 {
					sb.WriteRune(',')
				}

				sb.WriteString(fmt.Sprintf("{%q:%d,%q:[", "i", i+1, "s"))

				for start := 1; start < len(solution)-1; start++ {
					if start > 1 {
						sb.WriteRune(',')
					}

					station, found := result.Nodes[solution[start].ID()]
					if found {
						sb.WriteString(fmt.Sprintf("%d", station.OSMID))
					} else {
						t.Fatalf("error: station not found %d", solution[start].ID())
					}
				}

				sb.WriteRune(']')
				sb.WriteRune('}')
			}
			wroteOne = true

			sb.WriteRune(']')
			sb.WriteRune('}')

		}
		sb.WriteRune(']')

		if writeToDisk {
			err = os.WriteFile(fmt.Sprintf("./../../frontend/web/public/pf/%d.json", startStationID), []byte(sb.String()), 0644)
			if err != nil {
				t.Fatalf("error writing path finder json : %#v", err)
			}
		}
	}

	t.Log("DONE.")
}
