package admin

import (
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"os"
	"testing"

	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

func TestPathFinderWithNum(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := sql.Open("sqlite3", "./../data/rat_brasov.db")
	if err != nil {
		t.Fatalf("Error opening SQLite database: %v", err)
	}
	defer db.Close()

	const selectBusses = `SELECT id, dir, name, from_station, to_station, no, color, website, urban, metropolitan, crawled FROM busses;`

	rows, err := db.Query(selectBusses)
	if err != nil {
		logger.Error("error querying bus", "err", err)
		return
	}
	defer rows.Close()

	busses := make([]*Busline, 0)
	bussesMap := make(map[int64]*Busline)
	for rows.Next() {
		var b Busline
		err := rows.Scan(
			&b.OSMID,
			&b.Dir,
			&b.Name,
			&b.From,
			&b.To,
			&b.Line,
			&b.Color,
			&b.Link,
			&b.IsUrban,
			&b.IsMetropolitan,
			&b.WasCrawled,
		)
		if err != nil {
			logger.Error("error scanning bus", "err", err)
			return
		}
		bussesMap[b.OSMID] = &b
		busses = append(busses, &b)
	}

	const selectStations = `SELECT id, name, lat, lng, outside, board FROM stations;`

	logger.Info("busses", "len", len(busses))

	stations := make([]*Station, 0)
	stationsMap := make(map[int64]*Station)
	uniqueStationNames := make(map[string][]int64)
	srows, err := db.Query(selectStations)
	if err != nil {
		logger.Error("error querying stations", "err", err)
	}
	defer srows.Close()

	graph := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	for srows.Next() {
		var station Station
		err := srows.Scan(
			&station.OSMID,
			&station.Name,
			&station.Lat,
			&station.Lon,
			&station.IsOutsideCity,
			&station.HasBoard,
		)
		if err != nil {
			logger.Error("error scanning station", "err", err)
		}
		stationsMap[station.OSMID] = &station
		stations = append(stations, &station)
		if _, has := uniqueStationNames[station.Name]; !has {
			uniqueStationNames[station.Name] = make([]int64, 0)
		}
		uniqueStationNames[station.Name] = append(uniqueStationNames[station.Name], station.OSMID)
		// add station nodes to graph
		graph.AddNode(station)
	}

	logger.Info("stations", "len", len(stations))

	const selectRelations = `SELECT station_id, bus_id, station_index FROM bus_stops ORDER BY bus_id, station_index;`

	rrows, err := db.Query(selectRelations)
	if err != nil {
		logger.Error("error querying relation", "err", err)
	}
	defer rrows.Close()

	var curBus *Busline
	for rrows.Next() {
		var stationID, busID int64
		var stationIndex int
		err := rrows.Scan(
			&stationID,
			&busID,
			&stationIndex,
		)
		if err != nil {
			logger.Error("error scanning relation", "err", err)
		}

		if curBus == nil || curBus.OSMID != busID {
			curBus, _ = bussesMap[busID]
		}

		station, _ := stationsMap[stationID]
		curBus.Stations = append(curBus.Stations, *station)

		hasBus := false
		for _, line := range station.Lines {
			if line.BusOSMID == busID {
				hasBus = true
				break
			}
		}

		if !hasBus {
			station.Lines = append(station.Lines, &LineNumberAndTime{BusOSMID: busID, No: curBus.Line, Direction: Direction(curBus.Dir)})
		}
	}

	// set edges for all stations
	for _, bus := range busses {
		for i := 0; i < len(bus.Stations)-1; i++ {
			graph.SetWeightedEdge(simple.WeightedEdge{F: bus.Stations[i], T: bus.Stations[i+1], W: 1.0})
		}
	}

	// TODO : take links from a DB table
	// set edges for stations with the same name
	for stationName, stationsIDs := range uniqueStationNames {
		if stationName == "Facultativă" || stationName == "New Bus Stop" { // 22 pieces
			continue
		}

		for i := 0; i < len(stationsIDs); i++ {
			sourceStation, _ := stationsMap[stationsIDs[i]]
			for j := 0; j < len(stationsIDs); j++ {
				if i != j {
					targetStation, _ := stationsMap[stationsIDs[j]]
					graph.SetWeightedEdge(simple.WeightedEdge{F: sourceStation, T: targetStation, W: 0})
				}
			}
		}
	}

	it := graph.Nodes()
	for it.Next() {
		station, has := stationsMap[it.Node().ID()]
		if !has {
			logger.Error("node", "NOT FOUND", it.Node().ID())
			continue
		}
		_ = station
		// logger.Info("NODE", "named", station.Name)
	}

	pairs := make(map[string]struct{})
	ed := graph.Edges()
	for ed.Next() {
		edge := ed.Edge()
		from, hasFrom := stationsMap[edge.From().ID()]
		if !hasFrom {
			logger.Error("edge", "FROM NOT FOUND", edge.From().ID())
			continue
		}
		to, hasTo := stationsMap[edge.To().ID()]
		if !hasTo {
			logger.Error("edge", "TO NOT FOUND", edge.To().ID())
			continue
		}

		if _, has := pairs[fmt.Sprintf("%d-%d", from.OSMID, to.OSMID)]; has {
			logger.Error("ERROR", "seen", fmt.Sprintf("%d-%d", from.OSMID, to.OSMID))
		} else {
			pairs[fmt.Sprintf("%d-%d", from.OSMID, to.OSMID)] = struct{}{}
		}

		// logger.Info("EDGE", "from", from.Name, "to", to.Name, "fromID", from.OSMID, "toID", to.OSMID)
	}

	startNode, _ := stationsMap[3709600108]
	solutions, ok := path.BellmanFordAllFrom(startNode, graph)
	if !ok {
		t.Fatalf("error finding all solutions")
	}

	allPaths, weight := solutions.AllTo(273437289)
	t.Logf("w = %f => %d solutions", weight, len(allPaths))
	for i, solution := range allPaths {
		if i > 5 {
			break
		}
		t.Logf("solution %d", i)
		for start := len(solution) - 1; start > 0; start-- {
			station, _ := stationsMap[solution[start].ID()]
			t.Logf("[%d] %q", station.OSMID, station.Name)
		}
		t.Logf("==========")
	}
}
