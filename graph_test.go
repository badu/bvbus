package bvbus_test

import (
	"encoding/json"
	"os"
	"testing"

	. "github.com/badu/bvbus"
)

func TestPathFinding(t *testing.T) {
	file, err := os.ReadFile("data.json")
	if err != nil {
		t.Fatalf("error reading data : %#v", err)
	}

	var data StationsAndBusses
	if err := json.Unmarshal(file, &data); err != nil {
		t.Fatalf("error decoding data : %#v\n%s", err, err.Error())
	}

	t.Logf("%d busses decoded.", len(data.Busses))
	t.Logf("%d bus stations decoded.", len(data.Stations))
	t.Logf("creating graph...")

	// create the graph
	graph := Graph{}

	// populate the graph with stations (can't add links until all stations exist)
	for stationName, station := range data.Stations {
		for directionKey, direction := range station {
			busses := make([]string, 0)
			for _, bus := range direction.Busses {
				busses = append(busses, bus)
			}

			node := Stop{Name: stationName, Direction: directionKey, Busses: busses}
			if err := graph.AddNode(&node); err != nil {
				t.Fatalf("error : %#v", err)
			}
		}
	}

	t.Logf("%d nodes (bus stations should be equal) declared in graph.", len(graph.Nodes))

	links := 0
	for stationName, station := range data.Stations {
		for directionKey, direction := range station {
			for _, link := range direction.Links {
				sourceNode := graph.Nodes[stationName][directionKey]
				targetNode := graph.Nodes[link.Station][directionKey]
				if sourceNode == nil {
					continue
				}

				if targetNode == nil {
					continue
				}

				if link.Weight <= 0 {
					// fmt.Println("bad weight", stationName, " to ", link.Station, " weight ", link.Duration)
					continue
				}

				graph.AddEdge(sourceNode, targetNode, link.Weight)
				links++
			}
		}
	}

	t.Logf("%d links between stations declared in graph.", len(graph.Nodes))

	const (
		FROM = "Carpatilor"
		TO   = "Pantex"
	)

	t.Logf("finding travel solution from %q to %q", FROM, TO)

	result := graph.GetShortestPath(FROM, TO)
	for i, solution := range result.Solutions {
		t.Logf("solution #%d %q (%s) -> %q (%s) has %d stops travel time %d minutes (while in the bus)", i+1, solution.Start.Name, solution.Start.Direction.String(), solution.End.Name, solution.End.Direction.String(), len(solution.Stops), solution.Duration)

		for j, r := range solution.Stops {
			t.Logf("#%02d. %s (%s)", j+1, r.Name, r.Direction.String())
		}

		for j, b := range solution.Busses {
			t.Logf("#%02d. %#v", j+1, b)
		}
	}
}
