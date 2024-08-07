package admin

import (
	"database/sql"
	"fmt"
	"log/slog"
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
	ForBusID      int64
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
	Busses    map[int64]*Busline
	Distances map[string]*DistanceAndMinutes
	Terminals map[string]*Terminal
}

type Terminal struct {
	ID              int64
	Name            string
	Stations        []int64
	StationsMap     map[int64]struct{}
	Arrivals        int
	Departures      int
	ArrivalBusses   []Busline
	DepartureBusses []Busline
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
	siblingBusses := make(map[int64]int64)

	for _, busLine1 := range busses {
		for _, busLine2 := range busses {
			if busLine1.Line == busLine2.Line && busLine1.Dir != busLine2.Dir {
				siblingBusses[busLine1.OSMID] = busLine2.OSMID
				break
			}
		}
	}

	stations := make([]*Station, 0)
	stationsMap := make(map[int64]*Station)
	uniqueStationNames := make(map[string][]int64)
	rows, err = db.Query(`SELECT id, name, lat, lng, board FROM stations ORDER BY id;`)
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

	var currentBus *Busline
	for rows.Next() {
		var stationID, busID int64
		var stationIndex int
		var has bool
		err := rows.Scan(&stationID, &busID, &stationIndex)
		if err != nil {
			logger.Error("error scanning relation", "err", err)
			return nil, err
		}

		if currentBus == nil || currentBus.OSMID != busID {
			currentBus, has = bussesMap[busID]
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

		currentBus.Stations = append(currentBus.Stations, *station)

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

			lineAndTime := &LineNumberAndTime{BusOSMID: busID, No: currentBus.Line, Direction: Direction(currentBus.Dir)}
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

	terminalsMap := make(map[string]*Terminal)
	for _, bus := range busses {
		firstTerminalID := bus.Stations[0].OSMID
		firstTerminalName := bus.Stations[0].Name
		if _, has := terminalsMap[firstTerminalName]; !has {
			terminalsMap[firstTerminalName] = &Terminal{
				ID:              firstTerminalID,
				Name:            firstTerminalName,
				Stations:        make([]int64, 0),
				StationsMap:     make(map[int64]struct{}),
				DepartureBusses: make([]Busline, 0),
			}
		}
		if _, has := terminalsMap[firstTerminalName].StationsMap[firstTerminalID]; !has {
			terminalsMap[firstTerminalName].Stations = append(terminalsMap[firstTerminalName].Stations, firstTerminalID)
			terminalsMap[firstTerminalName].StationsMap[firstTerminalID] = struct{}{}
		}

		if _, found := stationsMap[firstTerminalID]; found {
			stationsMap[firstTerminalID].IsTerminal = true
		} else {
			logger.Error("error looking up station", "id", firstTerminalID)
		}

		terminalsMap[firstTerminalName].Departures++
		terminalsMap[firstTerminalName].DepartureBusses = append(terminalsMap[firstTerminalName].DepartureBusses, Busline{OSMID: bus.OSMID, Line: bus.Line})

		secondTerminalName := bus.Stations[len(bus.Stations)-1].Name
		secondTerminalID := bus.Stations[len(bus.Stations)-1].OSMID
		if _, has := terminalsMap[secondTerminalName]; !has {
			terminalsMap[secondTerminalName] = &Terminal{
				ID:            secondTerminalID,
				Name:          secondTerminalName,
				Stations:      make([]int64, 0),
				StationsMap:   make(map[int64]struct{}),
				ArrivalBusses: make([]Busline, 0),
			}
		}
		if _, has := terminalsMap[secondTerminalName].StationsMap[secondTerminalID]; !has {
			terminalsMap[secondTerminalName].Stations = append(terminalsMap[secondTerminalName].Stations, secondTerminalID)
			terminalsMap[secondTerminalName].StationsMap[secondTerminalID] = struct{}{}
		}

		if _, found := stationsMap[secondTerminalID]; found {
			stationsMap[secondTerminalID].IsTerminal = true
		} else {
			logger.Error("error looking up station", "id", secondTerminalID)
		}

		terminalsMap[secondTerminalName].Arrivals++
		terminalsMap[secondTerminalName].ArrivalBusses = append(terminalsMap[secondTerminalName].ArrivalBusses, Busline{OSMID: bus.OSMID, Line: bus.Line})
	}

	for stationName, terminalInfo := range terminalsMap {
		var sb strings.Builder
		sb.WriteString("arrivals")
		for index, arrival := range terminalInfo.ArrivalBusses {
			if index > 0 {
				sb.WriteRune(',')
			}
			sb.WriteString(fmt.Sprintf("%q", arrival.Line))
		}
		sb.WriteRune(' ')
		sb.WriteString("departures")
		for index, departure := range terminalInfo.DepartureBusses {
			if index > 0 {
				sb.WriteRune(',')
			}
			sb.WriteString(fmt.Sprintf("%q", departure.Line))
		}
		_ = stationName
		// logger.Info("terminal", "name", stationName, "arrivals", terminalInfo.Arrivals, "departures", terminalInfo.Departures, "busses", sb.String())
	}

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

	var prevPoint *Node
	currentBus = nil
	stationIndex := 0
	currentDistance := float64(0)
	mapResult := make(map[string]*DistanceAndMinutes)
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
			logger.Error("ERROR finding point by id", "id", pointID)
			return nil, fmt.Errorf("point not found id = %d", pointID)
		}

		// bus changed
		if currentBus == nil || currentBus.OSMID != busID {
			currentBus, has = bussesMap[busID]
			if !has {
				logger.Error("ERROR current bus not found", "busID", busID)
				return nil, fmt.Errorf("current bus not found id = %d", busID)
			}
			stationIndex = 0
			currentDistance = 0.0
			prevPoint = nil
		}

		// we can calculate Haversine
		if prevPoint != nil {
			currentDistance += Haversine(prevPoint.Lat, prevPoint.Lon, point.Lat, point.Lon)
		}

		// it's a stop and it's not the first one (storing distance "from"-"to")
		if isStop && prevPoint != nil {
			startStationID := currentBus.Stations[stationIndex].OSMID
			stationIndex++
			if stationIndex >= len(currentBus.Stations) {
				logger.Error("ERROR", "", fmt.Sprintf("%s index = %d stations = %d", currentBus.Name, stationIndex, len(currentBus.Stations)))
				continue
			}
			destinationStationID := currentBus.Stations[stationIndex].OSMID
			key := fmt.Sprintf("%d-%d-%d", startStationID, destinationStationID, busID)
			startStation, hasStartStation := stationsMap[startStationID]
			if !hasStartStation {
				logger.Error("ERROR : start station not found in map", "id", startStationID)
				continue
			}

			destinationStation, hasDestinationStation := stationsMap[destinationStationID]
			if !hasDestinationStation {
				logger.Error("ERROR : destination station not found in map", "id", destinationStationID)
				continue
			}

			startTime, validStartTime := startStation.Lines.GetFirstEntry(busID)
			if !validStartTime {
				logger.Error("INVALID START TIME", "", fmt.Sprintf("for %q in %q - has bus ? %t", currentBus.Name, startStation.Name, startStation.Lines.HasBus(busID)))
				continue
			}

			endTime, validEndTime := destinationStation.Lines.GetFirstEntryAfter(busID, startTime)
			if !validEndTime {
				siblingBusID, hasSibling := siblingBusses[busID]
				if !hasSibling {
					logger.Error("NO SIBLING FOUND", "busID", busID)
					continue
				}
				siblingBus, hasSiblingBus := bussesMap[siblingBusID]
				if !hasSiblingBus {
					logger.Error("SIBLING BUS NOT FOUND", "busID", siblingBusID)
					continue
				}
				destinationStationID = siblingBus.Stations[0].OSMID

				siblingDestinationStation, destFound := stationsMap[destinationStationID]
				if !destFound {
					logger.Error("ERROR finding timetable by id while healing", "destStationID", destinationStationID)
					continue
				}

				endTime, validEndTime = siblingDestinationStation.Lines.GetFirstEntryAfter(siblingBus.OSMID, startTime)
				if !validEndTime {
					logger.Error("INVALID END TIME", "endStationID", destinationStationID)
					continue
				}
			}

			if !startTime.After(*endTime) {
				logger.Error("TIME AFTER", "bus", currentBus.Name, "start", startStation.Name, "end", destinationStation.Name, "startTime", startTime, "endTime", endTime)
				continue
			}

			diff := startTime.Diff(*endTime)

			mapResult[key] = &DistanceAndMinutes{
				Key:           key,
				FromStationID: startStation.OSMID,
				ToStationID:   destinationStation.OSMID,
				ForBusID:      busID,
				Meters:        uint16(currentDistance),
				Minutes:       diff,
			}

			currentDistance = 0.0
		}

		prevPoint = &point
	}
	rows.Close()

	return &StationsAndDistances{Stations: stationsMap, Distances: mapResult, Busses: bussesMap, Terminals: terminalsMap}, nil
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
	sb.WriteString("const distances = new Map();\n")

	seen := make(map[string]struct{})

	sortedResult := make([]DistanceAndMinutes, 0)
	for _, measurement := range result.Distances {
		r := *measurement
		sortedResult = append(sortedResult, r)
	}

	sort.Sort(ByDistance(sortedResult))

	for _, measurement := range sortedResult {
		key := fmt.Sprintf("%d-%d", measurement.FromStationID, measurement.ToStationID)
		if _, has := seen[key]; !has {
			sb.WriteString(fmt.Sprintf("distances.set(%q,{%q:%d,%q:%d})\n", key, "d", measurement.Meters, "m", measurement.Minutes))
		}
		seen[key] = struct{}{}
	}
	sb.WriteString("export default distances;")

	err = os.WriteFile("./../../frontend/web/src/distances.js", []byte(sb.String()), 0644)
	if err != nil {
		t.Fatalf("error writing distances.js : %#v", err)
	}
}

func TestPathFinderWithNum(t *testing.T) {
	const writeToDisk = true

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
	graph := simple.NewDirectedGraph()
	// add stations as nodes

	for stationID, station := range result.Stations {
		if _, has := uniqueStationNames[station.Name]; !has {
			uniqueStationNames[station.Name] = make([]int64, 0)
		}

		uniqueStationNames[station.Name] = append(uniqueStationNames[station.Name], stationID)

		if _, isTerminal := result.Terminals[station.Name]; !isTerminal {
			// skip terminals (will be added once)
			// logger.Info("adding STATION node", "name", station.Name, "id", station.OSMID)
			if node := graph.Node(station.OSMID); node == nil {
				graph.AddNode(station)
			} else {
				logger.Warn("node exists", "name", station.Name)
			}
		}
	}

	for _, terminalInfo := range result.Terminals {
		terminalID := terminalInfo.ID
		station, _ := result.Stations[terminalID]
		// logger.Info("adding TERMINAL node", "name", station.Name, "id", terminalID)
		if node := graph.Node(station.OSMID); node == nil {
			graph.AddNode(station)
		} else {
			logger.Warn("node exists", "name", station.Name)
		}
	}

	seen := make(map[string]string)

	for _, bus := range result.Busses {
		if bus.IsMetropolitan {
			continue
		}

		for index := range bus.Stations {
			if index >= len(bus.Stations)-1 {
				continue
			}

			if graph.HasEdgeFromTo(bus.Stations[index].OSMID, bus.Stations[index+1].OSMID) {
				// t.Logf("graph has edge between %d and %d. skipping", bus.Stations[index].OSMID, bus.Stations[index+1].OSMID)
				continue
			}

			startStation, startFound := result.Stations[bus.Stations[index].OSMID]
			if !startFound {
				t.Fatalf("error finding start station %d", bus.Stations[index].OSMID)
			}

			endStation, endFound := result.Stations[bus.Stations[index+1].OSMID]
			if !endFound {
				t.Fatalf("error finding end station %d", bus.Stations[index+1].OSMID)
			}

			// check replacements with terminals, so we have a single node instead of many stations with a lot of edges between
			_, replaceStartWithTerminal := result.Terminals[startStation.Name]
			if replaceStartWithTerminal {
				terminal, terminalFound := result.Terminals[bus.Stations[index].Name]
				if !terminalFound {
					t.Fatalf("error finding terminal for station %q", bus.Stations[index].Name)
				}
				startStation, startFound = result.Stations[terminal.ID]
				if !startFound {
					t.Fatalf("error finding start station %d", bus.Stations[index].OSMID)
				}
			}

			_, replaceEndWithTerminal := result.Terminals[endStation.Name]
			if replaceEndWithTerminal {
				terminal, terminalFound := result.Terminals[bus.Stations[index+1].Name]
				if !terminalFound {
					t.Fatalf("error finding terminal for station %q", bus.Stations[index+1].Name)
				}
				endStation, endFound = result.Stations[terminal.ID]
				if !endFound {
					t.Fatalf("error finding end station %d", bus.Stations[index+1].OSMID)
				}
			}

			seenKey := fmt.Sprintf("%d-%d", startStation.OSMID, endStation.OSMID)
			if _, hasSeen := seen[seenKey]; !hasSeen {
				seen[seenKey] = ""
			}

			if len(seen[seenKey]) > 0 {
				seen[seenKey] += ","
			}

			seen[seenKey] += bus.Line
			graph.SetEdge(simple.Edge{F: startStation, T: endStation})
		}
	}

	// add edges for stations with the same name (crossings)
	/**
		for _, stationsIDs := range uniqueStationNames {
			for i := 0; i < len(stationsIDs); i++ {
				sourceStation, _ := result.Stations[stationsIDs[i]]
				for j := 0; j < len(stationsIDs); j++ {
					if graph.HasEdgeFromTo(stationsIDs[i], stationsIDs[j]) {
						continue
					}

					if i == j {
						continue
					}

					targetStation, _ := result.Stations[stationsIDs[j]]
					if graph.HasEdgeFromTo(sourceStation.OSMID, targetStation.OSMID) {
						logger.Warn("edge exists (crossing)", "first", sourceStation.OSMID, "second", targetStation.OSMID)
						continue
					}

					// logger.Info("adding 100m edge between", "from", sourceStation.Name, "to", targetStation.Name, "fid", sourceStation.OSMID, "tid", targetStation.OSMID)
					graph.SetWeightedEdge(simple.WeightedEdge{F: sourceStation, T: targetStation, W: 100})

				}
			}
		}
	**/
	result.Distances = nil // free up some RAM

	var graphSb strings.Builder
	graphSb.WriteRune('{')
	graphSb.WriteString(fmt.Sprintf("%q:[", "nodes"))
	it := graph.Nodes()
	q := 0
	for it.Next() {
		station, has := result.Stations[it.Node().ID()]
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

		if q > 0 {
			graphSb.WriteRune(',')
		}
		graphSb.WriteString(fmt.Sprintf("{%q:%d,%q:%d,%q:%q}", "f", from.OSMID, "t", to.OSMID, "b", seen[fmt.Sprintf("%d-%d", from.OSMID, to.OSMID)]))
		q++
		// logger.Info("EDGE", "from", from.Name, "to", to.Name, "fromID", from.OSMID, "toID", to.OSMID)
	}
	graphSb.WriteRune(']')
	graphSb.WriteRune('}')

	err = os.WriteFile("./../../frontend/web/public/graph.json", []byte(graphSb.String()), 0644)
	if err != nil {
		t.Fatalf("error writing graph.json : %#v", err)
	}

	count := len(result.Stations)
	for startStationID := range result.Stations {
		startStation, startStationFound := result.Stations[startStationID]
		if !startStationFound {
			t.Fatalf("error looking up start station %d", startStationID)
		}

		_, replaceStartWithTerminal := result.Terminals[startStation.Name]
		if replaceStartWithTerminal {
			terminal, terminalFound := result.Terminals[startStation.Name]
			if !terminalFound {
				t.Fatalf("error finding terminal for station %q", startStation.Name)
			}
			var endFound bool
			startStation, endFound = result.Stations[terminal.ID]
			if !endFound {
				t.Fatalf("error finding station %d", terminal.ID)
			}
		}

		solutions, ok := path.BellmanFordAllFrom(startStation, graph)
		if !ok {
			t.Fatalf("bellman ford error")
		}

		siblings, hasSiblings := uniqueStationNames[startStation.Name]
		var sb strings.Builder
		sb.WriteRune('[')
		wroteOne := false
		for endStationID, endStation := range result.Stations {
			_, replaceEndWithTerminal := result.Terminals[endStation.Name]
			if replaceEndWithTerminal {
				terminal, terminalFound := result.Terminals[endStation.Name]
				if !terminalFound {
					t.Fatalf("error finding terminal for station %q", endStation.Name)
				}
				var endFound bool
				endStation, endFound = result.Stations[terminal.ID]
				if !endFound {
					t.Fatalf("error finding station %d", terminal.ID)
				}

			}

			if startStation.OSMID == endStation.OSMID {
				continue
			}

			willSkip := false
			if !replaceEndWithTerminal && hasSiblings {
				for _, siblingID := range siblings {
					if endStation.OSMID == siblingID {
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

			allPaths, _ := solutions.AllTo(endStation.OSMID)
			if len(allPaths) == 0 {
				// logger.Warn("NO SOLUTION", "from", startStation.Name, "to", endStation.Name, "fid", startStationID, "did", endStationID)
				continue
			}

			if wroteOne {
				sb.WriteRune(',')
			}

			sb.WriteString(fmt.Sprintf("{%q:%d,%q:[", "t", endStationID, "s"))

			for i, solution := range allPaths {
				if i > 5 { // take only first 5 solutions
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

		count--
		// t.Logf("%d written, %d remaining", startStationID, count)
	}

	t.Log("DONE.")
}
