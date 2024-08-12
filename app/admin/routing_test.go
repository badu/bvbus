package admin

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"

	"gonum.org/v1/gonum/graph"
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
type StationRoute struct {
	Id        int64
	Name      string
	Busses    []int64
	BussesMap map[int64]struct{}
}

func (s StationRoute) ID() int64 {
	return s.Id
}

type ByBusNumbers []StationRoute

func (a ByBusNumbers) Len() int { return len(a) }

func (a ByBusNumbers) Less(i, j int) bool {
	return len(a[i].Busses) > len(a[j].Busses)
}

func (a ByBusNumbers) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type BusRoute struct {
	ID            int64
	Dir           uint
	Number        string
	Stations      []int64
	StationsIndex map[int64]int
}

type RouteEdge struct {
	Src *StationRoute
	Tgt *StationRoute
}

func NewEdge(from, to *StationRoute) *RouteEdge {
	return &RouteEdge{
		Src: from,
		Tgt: to,
	}
}

func (r RouteEdge) From() graph.Node {
	return r.Src
}

func (r RouteEdge) To() graph.Node {
	return r.Tgt
}

func (r RouteEdge) ReversedEdge() graph.Edge {
	return &RouteEdge{
		Src: r.Src,
		Tgt: r.Tgt,
	}
}

type StationsAndBusses struct {
	stations map[int64]*StationRoute
	busses   map[int64]*BusRoute
	points   map[int64]*Node
	ways     map[int64][]Node
}

func GetStationsAndBusses() (*StationsAndBusses, error) {
	result := StationsAndBusses{
		stations: make(map[int64]*StationRoute),
		busses:   make(map[int64]*BusRoute),
		points:   make(map[int64]*Node),
		ways:     make(map[int64][]Node),
	}

	db, err := sql.Open("sqlite3", "./../../data/brasov_busses.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, name FROM stations WHERE outside = 0 ORDER BY id;`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var station StationRoute
		err := rows.Scan(&station.Id, &station.Name)
		if err != nil {
			return nil, err
		}

		if _, has := result.stations[station.Id]; !has {
			station.Busses = make([]int64, 0)
			station.BussesMap = make(map[int64]struct{})
			result.stations[station.Id] = &station
		} else {
			return nil, fmt.Errorf("already seen %q station", station.Name)
		}
	}
	rows.Close()

	rows, err = db.Query(`SELECT id, dir, no, station_id, station_index FROM busses INNER JOIN bus_stops on busses.id = bus_stops.bus_id WHERE busses.metropolitan = 0 ORDER BY busses.id,station_index;`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var b BusRoute
		var stationID int64
		var stationIndex int
		err := rows.Scan(
			&b.ID,
			&b.Dir,
			&b.Number,
			&stationID,
			&stationIndex,
		)
		if err != nil {
			return nil, err
		}

		if _, has := result.busses[b.ID]; !has {
			b.Stations = make([]int64, 0)
			b.Stations = append(b.Stations, stationID)
			b.StationsIndex = make(map[int64]int)
			b.StationsIndex[stationID] = stationIndex
			result.busses[b.ID] = &b
		} else {
			result.busses[b.ID].Stations = append(result.busses[b.ID].Stations, stationID)
			result.busses[b.ID].StationsIndex[stationID] = stationIndex
		}

		if _, hasStation := result.stations[stationID]; hasStation {
			if _, hasBus := result.stations[stationID].BussesMap[b.ID]; !hasBus {
				result.stations[stationID].Busses = append(result.stations[stationID].Busses, b.ID)
				result.stations[stationID].BussesMap[b.ID] = struct{}{}
			}
		} else {
			fmt.Printf("%s stops in %d, but it's probably outside town\n", b.Number, stationID)
		}
	}
	rows.Close()

	rows, err = db.Query(`SELECT id, lat, lng FROM street_points ORDER BY id;`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var node Node
		err := rows.Scan(&node.ID, &node.Lat, &node.Lon)
		if err != nil {
			return nil, err
		}

		if _, has := result.points[node.ID]; !has {
			result.points[node.ID] = &node
		} else {
			return nil, fmt.Errorf("already seen %d node", node.ID)
		}
	}
	rows.Close()

	rows, err = db.Query(`SELECT point_id, bus_id, is_stop FROM street_rels ORDER BY bus_id,point_index;`)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var node Node
		var busID int64
		err := rows.Scan(&node.ID, &busID, &node.IsStop)
		if err != nil {
			return nil, err
		}

		if _, has := result.ways[busID]; !has {
			result.ways[busID] = make([]Node, 0)
		}

		result.ways[busID] = append(result.ways[busID], node)
	}
	rows.Close()

	return &result, nil
}

func TestFindAllPossibleRoutes(t *testing.T) {
	result, err := GetStationsAndBusses()
	if err != nil {
		t.Fatalf("error : %#v", err)
	}

	graph := simple.NewDirectedGraph()
	for _, station := range result.stations {
		if node := graph.Node(station.Id); node == nil {
			graph.AddNode(station)
		}
	}

	for _, bus := range result.busses {
		for i := 1; i < len(bus.Stations); i++ {
			start, hasStart := result.stations[bus.Stations[i-1]]
			if !hasStart {
				continue
			}

			end, hasEnd := result.stations[bus.Stations[i]]
			if !hasEnd {
				continue
			}

			if !graph.HasEdgeFromTo(start.Id, end.Id) {
				edge := NewEdge(start, end)
				graph.SetEdge(edge)
			}
		}
	}

	for _, start := range result.stations {
		startStation := result.stations[start.Id]
		solutions, ok := path.BellmanFordAllFrom(startStation, graph)
		if !ok {
			t.Fatalf("bellman ford error")
		}

		var sb strings.Builder
		sb.WriteRune('[')
		for _, end := range result.stations {
			if start.Id == end.Id {
				continue
			}

			allPaths, _ := solutions.AllTo(end.Id)
			if len(allPaths) == 0 {
				//t.Logf("no solution for %s -> %s", start.Name, end.Name)
				continue
			}

			if sb.Len() > 1 {
				sb.WriteRune(',')
			}

			sb.WriteString(fmt.Sprintf("{%q:%d,%q:[", "t", end.Id, "s"))
			for i, solution := range allPaths {
				if i > 5 { // take only first 5 solutions
					break
				}

				if i > 0 {
					sb.WriteRune(',')
				}

				sb.WriteString(fmt.Sprintf("{%q:%d,%q:[", "i", i+1, "s"))
				for s := 1; s < len(solution)-1; s++ {
					if s > 1 {
						sb.WriteRune(',')
					}
					station, hasStation := result.stations[solution[s].ID()]
					if hasStation {
						sb.WriteString(fmt.Sprintf("%d", station.Id))
					}
				}
				sb.WriteRune(']')
				sb.WriteRune('}')
			}

			sb.WriteRune(']')
			sb.WriteRune('}')
		}
		sb.WriteRune(']')
		err = os.WriteFile(fmt.Sprintf("./../../frontend/web/public/pf/%d.json", start.Id), []byte(sb.String()), 0644)
		if err != nil {
			t.Fatalf("error writing path finder json : %#v", err)
		}
	}
}
