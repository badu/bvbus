package admin

import (
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"os"
	"sort"
	"strings"
	"testing"

	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

type DistanceAndMinutes struct {
	Key           string
	FromStationID int64
	ToStationID   int64
	Meters        uint16
	Minutes       uint16
}

type ByDistance []DistanceAndMinutes

func (a ByDistance) Len() int { return len(a) }

func (a ByDistance) Less(i, j int) bool {
	return a[i].Meters > a[j].Meters
}

func (a ByDistance) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type StationsAndDistances struct {
	Stations  map[int64]*Station
	Distances []DistanceAndMinutes
}

func GetStationsAndDistances(logger *slog.Logger, db *sql.DB) (*StationsAndDistances, error) {
	rows, err := db.Query(`SELECT id, dir, name, from_station, to_station, no, color, website, urban, metropolitan, crawled FROM busses ORDER BY id;`)
	if err != nil {
		logger.Error("error querying bus", "err", err)
		return nil, err
	}

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
			return nil, err
		}
		bussesMap[b.OSMID] = &b
		busses = append(busses, &b)
	}
	rows.Close()

	stations := make([]*Station, 0)
	stationsMap := make(map[int64]*Station)
	uniqueStationNames := make(map[string][]int64)
	rows, err = db.Query(`SELECT id, name, lat, lng, outside, board FROM stations ORDER BY id;`)
	if err != nil {
		logger.Error("error querying stations", "err", err)
		return nil, err
	}

	for rows.Next() {
		var station Station
		err := rows.Scan(
			&station.OSMID,
			&station.Name,
			&station.Lat,
			&station.Lon,
			&station.IsOutsideCity,
			&station.HasBoard,
		)
		if err != nil {
			logger.Error("error scanning station", "err", err)
			return nil, err
		}

		if _, has := stationsMap[station.OSMID]; has {
			logger.Error("error station exists", "id", station.OSMID)
			return nil, fmt.Errorf("station already seen id = %d", station.OSMID)
		}

		stationsMap[station.OSMID] = &station
		stations = append(stations, &station)
		if _, has := uniqueStationNames[station.Name]; !has {
			uniqueStationNames[station.Name] = make([]int64, 0)
		}
		uniqueStationNames[station.Name] = append(uniqueStationNames[station.Name], station.OSMID)
	}
	rows.Close()

	rows, err = db.Query(`SELECT station_id, bus_id, station_index FROM bus_stops ORDER BY bus_id, station_index;`)
	if err != nil {
		logger.Error("error querying relation", "err", err)
		return nil, err
	}

	var curBus *Busline
	for rows.Next() {
		var stationID, busID int64
		var stationIndex int
		var has bool
		err := rows.Scan(&stationID, &busID, &stationIndex)
		if err != nil {
			logger.Error("error scanning relation", "err", err)
			return nil, err
		}

		if curBus == nil || curBus.OSMID != busID {
			curBus, has = bussesMap[busID]
			if !has {
				logger.Error("error current bus not found", "busID", busID)
				return nil, fmt.Errorf("current bus not found id = %d", busID)
			}
		}

		station, hasStation := stationsMap[stationID]
		if !hasStation {
			logger.Error("error station not found", "stationID", stationID)
			return nil, fmt.Errorf("current station not found id = %d", stationID)
		}

		curBus.Stations = append(curBus.Stations, *station)

		hasBus := false
		for _, line := range station.Lines {
			if line.BusOSMID == busID {
				hasBus = true
				break
			}
		}

		if !hasBus {
			trows, err := db.Query(`SELECT enc_time FROM time_tables WHERE station_id = ? AND bus_id = ?;`, stationID, busID)
			if err != nil {
				logger.Error("error loading time tables", "err", err)
				return nil, err
			}

			lineAndTime := &LineNumberAndTime{BusOSMID: busID, No: curBus.Line, Direction: Direction(curBus.Dir)}
			for trows.Next() {
				var encTime uint16
				err := trows.Scan(&encTime)
				if err != nil {
					logger.Error("error scanning timetable", "err", err)
					return nil, err
				}
				lineAndTime.Times = append(lineAndTime.Times, encTime)
			}
			station.Lines = append(station.Lines, lineAndTime)
		}
	}
	rows.Close()

	rows, err = db.Query(`SELECT id, lat, lng FROM street_points ORDER BY id;`)
	if err != nil {
		logger.Error("error querying street points", "err", err)
		return nil, err
	}

	pointsMap := make(map[int64]Node)
	for rows.Next() {
		var node Node
		err := rows.Scan(&node.ID, &node.Lat, &node.Lon)
		if err != nil {
			logger.Error("error scanning", "err", err)
			return nil, err
		}

		pointsMap[node.ID] = node
	}
	rows.Close()

	rows, err = db.Query(`SELECT point_id, bus_id, point_index, is_stop FROM street_rels ORDER BY bus_id,point_index;`)
	if err != nil {
		logger.Error("error querying street relations", "err", err)
		return nil, err
	}

	var prevPoint Node
	currentDistance := float64(0)
	prevBusID := int64(-1)
	calculatedDistances := make(map[int64][]float64)
	firstStop := true
	for rows.Next() {
		var pointID, busID, pointIndex int64
		var isStop bool
		err := rows.Scan(&pointID, &busID, &pointIndex, &isStop)
		if err != nil {
			logger.Error("error scanning", "err", err)
			return nil, err
		}

		point, has := pointsMap[pointID]
		if !has {
			logger.Error("error finding point by id", "id", pointID)
			return nil, fmt.Errorf("point not found id = %d", pointID)
		}

		if prevBusID != busID {
			calculatedDistances[busID] = make([]float64, 0)
			currentDistance = 0.0
			prevBusID = busID
			firstStop = true
		}

		if !firstStop {
			currentDistance += Haversine(prevPoint.Lat, prevPoint.Lon, point.Lat, point.Lon)
			if isStop {
				calculatedDistances[busID] = append(calculatedDistances[busID], currentDistance)
				currentDistance = 0.0
			}
		}

		firstStop = false
		prevPoint = point
	}
	rows.Close()

	mapResult := make(map[string]*DistanceAndMinutes)
	seen := make(map[string]struct{})
	for busID, distances := range calculatedDistances {
		bus, has := bussesMap[busID]
		if !has {
			logger.Error("error finding bus by id", "busID", busID)
			continue
		}

		if len(distances)+1 != len(bus.Stations) {
			var sb strings.Builder
			for index := range distances {
				startStation := bus.Stations[index]
				if index < len(bus.Stations)-1 {
					destinationStation := bus.Stations[index+1]
					sb.WriteString(fmt.Sprintf("[%d] %q -> %q ", index, startStation.Name, destinationStation.Name))
				} else {
					sb.WriteString(fmt.Sprintf("[%d] %q -> [?] ", index, startStation.Name))
				}
			}
			sb.WriteRune('\n')
			for index, station := range bus.Stations {
				sb.WriteString(fmt.Sprintf("[%d] %q ->", index, station.Name))
			}
			sb.WriteRune('\n')
			logger.Error("stations not equal distances", "distances", len(distances), "stations", len(bus.Stations), "busID", busID, "no", bus.Line, "comparison", sb.String())
			continue
		}

		for index, distance := range distances {
			if len(bus.Stations)-1 == index+1 {
				continue // won't have destTimetable timetable
			}

			startStation := bus.Stations[index]
			destinationStation := bus.Stations[index+1]
			key := fmt.Sprintf("%d-%d", startStation.OSMID, destinationStation.OSMID)

			if _, hasSeen := seen[key]; hasSeen {
				continue
			}
			seen[key] = struct{}{}

			if value, hasResult := mapResult[key]; !hasResult {
				startTimetable, startFound := stationsMap[startStation.OSMID]
				if !startFound {
					logger.Error("error finding timetable by id", "startStationID", startStation.OSMID)
					continue
				}

				validStartTime := false
				var startTime Time
				for _, line := range startTimetable.Lines {
					if line.BusOSMID == busID {
						if len(line.Times) > 0 {
							startTime.Decompress(line.Times[0])
							validStartTime = true
						}
						break
					}
				}

				destTimetable, destFound := stationsMap[destinationStation.OSMID]
				if !destFound {
					logger.Error("error finding timetable by id", "destStationID", destinationStation.OSMID)
					continue
				}

				validEndTime := false
				var endTime Time
				for _, line := range destTimetable.Lines {
					if line.BusOSMID == busID {
						if len(line.Times) > 0 {
							endTime.Decompress(line.Times[0])
							validEndTime = true
						}
						break
					}
				}

				if !startTime.After(endTime) {
					logger.Error("TIME AFTER", "bus", bus.Name, "start", startStation.Name, "end", destinationStation.Name, "startTime", startTime, "endTime", endTime)
					continue
				}

				if !validEndTime || !validStartTime {
					logger.Error("INVALID START OR END TIME", "bus", bus.Name, "start", startStation.Name, "end", destinationStation.Name)
					continue
				}

				diff := startTime.Diff(endTime)
				mapResult[key] = &DistanceAndMinutes{Meters: uint16(distance), Minutes: diff, FromStationID: startStation.OSMID, ToStationID: destinationStation.OSMID}
			} else if value.Meters < uint16(distance) {
				mapResult[key].Meters = uint16(distance)
			}
		}
	}

	result := make([]DistanceAndMinutes, 0)
	for key, measurement := range mapResult {
		r := *measurement
		r.Key = key
		result = append(result, r)
	}
	sort.Sort(ByDistance(result))

	return &StationsAndDistances{Stations: stationsMap, Distances: result}, nil
}

func TestGenerateDistancesAndTimesBetweenStations(t *testing.T) {
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

	var sb strings.Builder
	sb.WriteString("const distances = {\n")
	comma := false
	for _, measurement := range result.Distances {
		if comma {
			sb.WriteRune(',')
		} else {
			comma = true
		}
		sb.WriteString(fmt.Sprintf("%q:{%q:%d,%q:%d}", measurement.Key, "d", measurement.Meters, "m", measurement.Minutes))
	}
	sb.WriteString("};\nexport default distances;")

	err = os.WriteFile("./../../frontend/web/src/distances.js", []byte(sb.String()), 0644)
	if err != nil {
		t.Fatalf("error writing urban_busses.js : %#v", err)
	}
}

func TestPathFinderWithNum(t *testing.T) {
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

	uniqueStationNames := make(map[string][]int64)
	graph := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	crossingsGraph := simple.NewWeightedDirectedGraph(0, math.Inf(1))
	for stationID, station := range result.Stations {
		if station.IsOutsideCity {
			continue
		}

		if _, has := uniqueStationNames[station.Name]; !has {
			uniqueStationNames[station.Name] = make([]int64, 0)
		}
		uniqueStationNames[station.Name] = append(uniqueStationNames[station.Name], stationID)

		graph.AddNode(station)
		crossingsGraph.AddNode(station)
	}

	for _, measurement := range result.Distances {
		startStation, startFound := result.Stations[measurement.FromStationID]
		if !startFound {
			t.Fatalf("error finding start station %d", measurement.FromStationID)
		}

		endStation, endFound := result.Stations[measurement.ToStationID]
		if !endFound {
			t.Fatalf("error finding end station %d", measurement.ToStationID)
		}

		graph.SetWeightedEdge(simple.WeightedEdge{F: startStation, T: endStation, W: float64(measurement.Meters)})
		crossingsGraph.SetWeightedEdge(simple.WeightedEdge{F: startStation, T: endStation, W: float64(measurement.Meters)})
	}

	// add edges for stations with the same name (crossings)
	for _, stationsIDs := range uniqueStationNames {
		for i := 0; i < len(stationsIDs); i++ {
			sourceStation, _ := result.Stations[stationsIDs[i]]
			for j := 0; j < len(stationsIDs); j++ {
				if i != j {
					targetStation, _ := result.Stations[stationsIDs[j]]
					crossingsGraph.SetWeightedEdge(simple.WeightedEdge{F: sourceStation, T: targetStation, W: 100})
				}
			}
		}
	}

	result.Distances = nil // free up some RAM

	it := graph.Nodes()
	for it.Next() {
		station, has := result.Stations[it.Node().ID()]
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
		from, hasFrom := result.Stations[edge.From().ID()]
		if !hasFrom {
			logger.Error("edge", "FROM NOT FOUND", edge.From().ID())
			continue
		}

		to, hasTo := result.Stations[edge.To().ID()]
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

	noSolutions := make(map[int64][]int64)
	count := len(result.Stations)
	for startStationID, startStation := range result.Stations {
		if startStation.IsOutsideCity {
			count--
			continue
		}

		startNode, _ := result.Stations[startStationID]
		solutions, ok := path.BellmanFordAllFrom(startNode, graph)
		if !ok {
			t.Fatalf("error finding all solutions")
		}

		siblings, hasSiblings := uniqueStationNames[startStation.Name]
		var sb strings.Builder
		sb.WriteRune('[')
		wroteOne := false
		for endStationID, endStation := range result.Stations {
			if endStation.IsOutsideCity {
				continue
			}

			if startStationID == endStationID {
				continue
			}

			willSkip := false
			if hasSiblings {
				for _, siblingID := range siblings {
					if endStationID == siblingID {
						willSkip = true
						if wroteOne {
							sb.WriteRune(',')
						}
						sibling := fmt.Sprintf("{%q:%d,%q:true}", "t", endStationID, "cross")
						sb.WriteString(sibling)
						wroteOne = true
						break
					}
				}
			}

			if willSkip {
				continue
			}

			allPaths, weight := solutions.AllTo(endStationID)
			if len(allPaths) == 0 {
				if _, has := noSolutions[startStationID]; !has {
					noSolutions[startStationID] = make([]int64, 0)
				}
				hasIt := false
				for _, sID := range noSolutions[startStationID] {
					if sID == endStationID {
						hasIt = true
						break
					}
				}
				if !hasIt {
					noSolutions[startStationID] = append(noSolutions[startStationID], endStationID)
				}
				continue
			}

			if wroteOne {
				sb.WriteRune(',')
			}

			sb.WriteString(fmt.Sprintf("{%q:%d,%q:%.00f,%q:[", "t", endStationID, "d", weight, "s"))
			wroteOne = true
			for i, solution := range allPaths {
				if i > 0 {
					sb.WriteRune(',')
				}

				if i > 5 { // take only first 5 solutions
					break
				}

				sb.WriteString(fmt.Sprintf("{%q:%d,%q:[", "i", i+1, "s"))

				for start := len(solution) - 1; start > 0; start-- {
					if start < len(solution)-1 {
						sb.WriteRune(',')
					}

					station, found := result.Stations[solution[start].ID()]
					if found {
						sb.WriteString(fmt.Sprintf("%d", station.OSMID))
					} else {
						t.Fatalf("error: station not found %d", solution[start].ID())
					}
				}

				sb.WriteRune(']')
				sb.WriteRune('}')
			}

			sb.WriteRune(']')
			sb.WriteRune('}')

		}
		sb.WriteRune(']')

		err = os.WriteFile(fmt.Sprintf("./../../frontend/web/public/pf/%d.json", startStationID), []byte(sb.String()), 0644)
		if err != nil {
			t.Fatalf("error writing urban_busses.js : %#v", err)
		}

		count--
		t.Logf("%d written, %d remaining", startStationID, count)
	}

	count = len(noSolutions)
	for startStationID, endStationIDs := range noSolutions {
		startNode, _ := result.Stations[startStationID]
		solutions, ok := path.BellmanFordAllFrom(startNode, crossingsGraph)
		if !ok {
			t.Fatalf("error finding all solutions")
		}

		var sb strings.Builder
		sb.WriteRune('[')
		wroteOne := false
		for _, endStationID := range endStationIDs {
			if startStationID == endStationID {
				continue
			}

			allPaths, weight := solutions.AllTo(endStationID)
			if len(allPaths) == 0 {
				continue
			}

			if wroteOne {
				sb.WriteRune(',')
			}
			sb.WriteString(fmt.Sprintf("{%q:%d,%q:%.00f,%q:[", "t", endStationID, "d", weight, "s"))
			wroteOne = true

			for i, solution := range allPaths {
				if i > 0 {
					sb.WriteRune(',')
				}

				if i > 5 { // take only first 5 solutions
					break
				}

				sb.WriteString(fmt.Sprintf("{%q:%d,%q:[", "i", i+1, "s"))

				for start := len(solution) - 1; start > 0; start-- {
					if start < len(solution)-1 {
						sb.WriteRune(',')
					}

					sb.WriteString(fmt.Sprintf("%d", solution[start].ID()))
				}

				sb.WriteRune(']')
				sb.WriteRune('}')
			}

			sb.WriteRune(']')
			sb.WriteRune('}')
		}

		sb.WriteRune(']')

		count--
		if wroteOne {
			err = os.WriteFile(fmt.Sprintf("./../../frontend/web/public/pf/%d-cross.json", startStationID), []byte(sb.String()), 0644)
			if err != nil {
				t.Fatalf("error writing urban_busses.js : %#v", err)
			}
			t.Logf("%d written %d, %d remaining", startStationID, len(endStationIDs), count)
		} else {
			t.Logf("%d NOT written (has no solution)", startStationID)
		}
	}
	noSolutions = nil
	t.Log("DONE.")
}
