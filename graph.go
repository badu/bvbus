package main

import (
	"math"
	"sort"
	"sync"
)

type LatLong struct {
	Lat float64 // The field value must be a valid WGS 84 latitude.
	Lon float64 // The field value must be a valid WGS 84 longitude value from -180 to 180.
}

const (
	EarthRadius = 6371000.0
	ToRadians   = math.Pi / 180.0
)

func (l *LatLong) DistanceToCoordinate(lat, lon float64) float64 {
	return math.Acos(math.Sin(l.Lat*ToRadians)*math.Sin(lat*ToRadians)+
		math.Cos(l.Lat*ToRadians)*math.Cos(lat*ToRadians)*
			math.Cos((lon-l.Lon)*ToRadians)) * EarthRadius
}

type Stop struct {
	Name      string
	Direction Direction
	Busses    []string
	Links     []Path
}

type Path struct {
	Stop     *Stop
	Duration int
}

type Graph struct {
	mu    sync.RWMutex
	Nodes map[string]map[Direction]*Stop
}

type StopQueue struct {
	mu    sync.RWMutex
	Paths []Path
}

func (s *StopQueue) Enqueue(p Path) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Paths) == 0 {
		s.Paths = append(s.Paths, p)
		return
	}

	insertFlag := false
	for k, v := range s.Paths {
		if p.Duration >= v.Duration {
			continue
		}
		// add if distance less than travers's distance
		s.Paths = append(s.Paths[:k+1], s.Paths[k:]...)
		s.Paths[k] = p
		insertFlag = true
		break
	}

	if !insertFlag {
		s.Paths = append(s.Paths, p)
	}
}

func (s *StopQueue) Dequeue() *Path {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := s.Paths[0]
	s.Paths = s.Paths[1:len(s.Paths)]
	return &item
}

func (s *StopQueue) NewQueue() *StopQueue {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Paths = []Path{}
	return s
}

func (s *StopQueue) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Paths) == 0
}

func (s *StopQueue) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Paths)
}

func (g *Graph) AddNode(s *Stop) error {
	g.mu.Lock()
	if g.Nodes == nil {
		g.Nodes = make(map[string]map[Direction]*Stop)
	}
	if g.Nodes[s.Name] == nil {
		g.Nodes[s.Name] = make(map[Direction]*Stop)
	}
	sort.Strings(s.Busses)
	g.Nodes[s.Name][s.Direction] = s
	g.mu.Unlock()
	return nil
}

func (g *Graph) AddEdge(s1, s2 *Stop, duration int) {
	g.mu.Lock()
	if s1.Links == nil {
		s1.Links = make([]Path, 0)
	}
	s1.Links = append(s1.Links, Path{Stop: s2, Duration: duration})
	g.mu.Unlock()
}

func (g *Graph) getShortestPath(startNode, endNode *Stop) ([]*Stop, int) {
	visited := make(map[string]struct{})
	durations := make(map[string]int)
	routes := make(map[string]*Stop)

	queue := StopQueue{}
	pQueue := queue.NewQueue()
	start := Path{
		Stop:     startNode,
		Duration: 0,
	}

	for stationName, directions := range g.Nodes {
		for direction := range directions {
			durationKey := stationName + direction.String()
			durations[durationKey] = math.MaxInt64
		}
	}

	stopKey := startNode.Name + startNode.Direction.String()
	durations[stopKey] = start.Duration
	pQueue.Enqueue(start)

	for !pQueue.IsEmpty() {
		v := pQueue.Dequeue()
		stopKey = v.Stop.Name + v.Stop.Direction.String()

		if _, has := visited[stopKey]; has {
			continue
		}

		visited[stopKey] = struct{}{}

		for _, edge := range v.Stop.Links {
			edgeKey := edge.Stop.Name + edge.Stop.Direction.String()
			if _, has := visited[edgeKey]; has {
				continue
			}

			if durations[stopKey]+edge.Duration < durations[edgeKey] {
				pathNode := Path{
					Stop:     edge.Stop,
					Duration: durations[v.Stop.Name] + edge.Duration,
				}
				durations[edgeKey] = durations[stopKey] + edge.Duration
				routes[edge.Stop.Name] = v.Stop

				pQueue.Enqueue(pathNode)
			}

		}
	}

	stopKey = endNode.Name + endNode.Direction.String()
	if durations[stopKey] == math.MaxInt64 {
		return nil, -1
	}

	pathValue := routes[endNode.Name]
	if pathValue == nil {
		return nil, -1
	}

	result := make([]*Stop, 0)
	result = append(result, endNode)

	for pathValue != startNode {
		result = append(result, pathValue)
		pathValue = routes[pathValue.Name]
	}

	result = append(result, pathValue)

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, durations[stopKey]

}

type Solution struct {
	Start    *Stop
	End      *Stop
	Stops    []*Stop `json:"stops"`
	Duration int     `json:"durations"`
}

type Result struct {
	Solutions []Solution
}

func (g *Graph) GetShortestPath(from, to string) *Result {
	result := Result{Solutions: make([]Solution, 0)}
	for _, destination := range g.Nodes[to] {
		for _, source := range g.Nodes[from] {
			path, distance := g.getShortestPath(source, destination)
			if path == nil { // no solution
				continue
			}

			result.Solutions = append(result.Solutions, Solution{
				Start:    source,
				End:      destination,
				Stops:    path,
				Duration: distance,
			})
		}
	}

	return &result
}
