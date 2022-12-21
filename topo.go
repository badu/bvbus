package bvbus

import (
	"fmt"
	"strings"
)

type setOrdered[key comparable] struct {
	indexes map[key]int
	items   []key
	length  int
}

func newSetOrdered[key comparable]() *setOrdered[key] {
	return &setOrdered[key]{
		indexes: make(map[key]int),
		length:  0,
	}
}

func (s *setOrdered[key]) index(item key) int {
	index, ok := s.indexes[item]
	if ok {
		return index
	}
	return -1
}

func (s *setOrdered[key]) copy() *setOrdered[key] {
	clone := newSetOrdered[key]()
	for _, item := range s.items {
		clone.add(item)
	}
	return clone
}

func (s *setOrdered[key]) add(item key) bool {
	_, ok := s.indexes[item]
	if !ok {
		s.indexes[item] = s.length
		s.items = append(s.items, item)
		s.length++
	}
	return !ok
}

type TopoGraph[key comparable] struct {
	nodes map[key]concreteNode[key]
}

func (g *TopoGraph[key]) ContainsNode(node key) bool {
	_, ok := g.nodes[node]
	return ok
}

func (g *TopoGraph[key]) getOrAddNode(node key) concreteNode[key] {
	n, ok := g.nodes[node]
	if !ok {
		n = make(concreteNode[key])
		g.nodes[node] = n
	}
	return n
}

func (g *TopoGraph[key]) AddNode(node key) {
	if !g.ContainsNode(node) {
		g.nodes[node] = make(concreteNode[key])
	}
}

func (g *TopoGraph[key]) AddEdge(from key, to key) error {
	f := g.getOrAddNode(from)
	g.AddNode(to)
	f.addEdge(to)
	return nil
}

func (g *TopoGraph[key]) Sort(node key) ([]key, error) {
	results := newSetOrdered[key]()
	err := g.visit(node, results, nil)
	if err != nil {
		return nil, err
	}
	return results.items, nil
}

func (g *TopoGraph[key]) visit(node key, results *setOrdered[key], visited *setOrdered[key]) error {
	if visited == nil {
		visited = newSetOrdered[key]()
	}

	added := visited.add(node)
	if !added {
		index := visited.index(node)
		circular := append(visited.items[index:], node)
		guilty := make([]string, len(circular))
		for i, k := range circular {
			guilty[i] = fmt.Sprintf("%v", k)
		}
		return fmt.Errorf("circular error: %s", strings.Join(guilty, " -> "))
	}

	n := g.nodes[node]
	for _, edge := range n.edges() {
		err := g.visit(edge, results, visited.copy())
		if err != nil {
			return err
		}
	}

	results.add(node)
	return nil
}

type concreteNode[key comparable] map[key]bool

func (n concreteNode[key]) addEdge(node key) {
	n[node] = true
}

func (n concreteNode[key]) edges() []key {
	var keys []key
	for k := range n {
		keys = append(keys, k)
	}
	return keys
}

func NewTopoGraph[key comparable]() *TopoGraph[key] {
	return &TopoGraph[key]{
		nodes: make(map[key]concreteNode[key]),
	}
}
