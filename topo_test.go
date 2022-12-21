package bvbus

import (
	"fmt"
	"strings"
	"testing"
)

const (
	A = "Autogara 3"
	B = "Dramatic"
	C = "Hidro A"
	D = "Livada Postei"
)

func TestTopoSortFork(t *testing.T) {
	graph := NewTopoGraph[string]()

	// Autogara 3 -> Hidro A
	// Autogara 3 -> Dramatic
	// Dramatic -> Hidro A
	graph.AddEdge(A, C)
	graph.AddEdge(A, B)
	graph.AddEdge(B, C)

	results, err := graph.Sort(A)
	if err != nil {
		t.Error(err)
		return
	}
	if results[0] != C || results[1] != B || results[2] != A {
		t.Errorf("wrong sort order: %v", results)
	}
}

func TestTopoSort(t *testing.T) {
	graph := NewTopoGraph[string]()

	// Autogara 3 -> Dramatic -> Hidro A
	graph.AddEdge(A, B)
	graph.AddEdge(B, C)

	results, err := graph.Sort(A)
	if err != nil {
		t.Error(err)
		return
	}
	if results[0] != C || results[1] != B || results[2] != A {
		t.Errorf("wrong sort order: %v", results)
	}
}

func TestTopoSortDiamond(t *testing.T) {
	graph := NewTopoGraph[string]()

	// Autogara 3 -> Dramatic
	// Autogara 3 -> Livada Postei
	// Livada Postei -> Hidro A
	// Hidro A -> Dramatic
	graph.AddEdge(A, B)
	graph.AddEdge(A, D)
	graph.AddEdge(D, C)
	graph.AddEdge(C, B)

	results, err := graph.Sort(A)
	if err != nil {
		t.Error(err)
		return
	}
	if len(results) != 4 {
		t.Errorf("wrong results len: %v", results)
		return
	}
	expected := [4]string{B, C, D, A}
	for i := 0; i < 4; i++ {
		if results[i] != expected[i] {
			t.Errorf("wrong sort order: %v", results)
			break
		}
	}
}

func TestTopoSortWithCycle(t *testing.T) {
	graph := NewTopoGraph[string]()

	// Autogara 3 -> Dramatic
	// Dramatic -> Autogara 3
	graph.AddEdge(A, B)
	graph.AddEdge(B, A)

	_, err := graph.Sort(A)
	if err == nil {
		t.Errorf("Expected cycle error")
		return
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("%s -> %s -> %s", A, B, A)) {
		t.Errorf("Error doesn't print cycle: %q", err)
	}
}

func TestTopoSortCircular(t *testing.T) {
	graph := NewTopoGraph[string]()

	// Autogara 3 -> Dramatic
	// Dramatic -> Hidro A
	// Hidro A -> Dramatic
	graph.AddEdge(A, B)
	graph.AddEdge(B, C)
	graph.AddEdge(C, B)

	_, err := graph.Sort(A)
	if err == nil {
		t.Errorf("Expected cycle error")
		return
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("%s -> %s -> %s", B, C, B)) {
		t.Errorf("Error doesn't print cycle: %q", err)
	}
}

func TestTopoSortCircle(t *testing.T) {
	graph := NewTopoGraph[string]()

	// Autogara 3 -> Dramatic
	// Dramatic -> Hidro A
	// Hidro A -> Autogara 3
	graph.AddEdge(A, B)
	graph.AddEdge(B, C)
	graph.AddEdge(C, A)

	_, err := graph.Sort(A)
	if err == nil {
		t.Errorf("Expected cycle error")
		return
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("%s -> %s -> %s -> %s", A, B, C, A)) {
		t.Errorf("Error doesn't print cycle: %q", err)
	}
}
