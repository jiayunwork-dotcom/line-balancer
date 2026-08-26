package task

import (
	"strings"
	"testing"
)

func makeTestGraph() *Graph {
	g := NewGraph()
	g.Add("A", 5)
	g.Add("B", 3)
	g.Add("C", 6)
	g.Add("D", 7)
	g.Add("E", 4)
	g.Add("F", 2)
	g.Add("G", 5)
	g.Add("H", 3)
	g.Precede("A", "B")
	g.Precede("A", "C")
	g.Precede("B", "E")
	g.Precede("B", "D")
	g.Precede("C", "D")
	g.Precede("C", "F")
	g.Precede("D", "G")
	g.Precede("E", "H")
	g.Precede("F", "H")
	g.Precede("G", "H")
	return g
}

func TestTopologicalOrder(t *testing.T) {
	g := makeTestGraph()
	topo, err := g.TopologicalOrder()
	if err != nil {
		t.Fatal(err)
	}
	if len(topo) != 8 {
		t.Fatalf("expected 8 tasks in topo order, got %d", len(topo))
	}
	pos := make(map[string]int, len(topo))
	for i, id := range topo {
		pos[id] = i
	}
	if pos["A"] >= pos["B"] || pos["A"] >= pos["C"] {
		t.Fatalf("A not before B/C: %v", topo)
	}
	if pos["D"] >= pos["G"] {
		t.Fatalf("D not before G: %v", topo)
	}
}

func TestCycleDetection(t *testing.T) {
	g := NewGraph()
	g.Add("X", 1)
	g.Add("Y", 1)
	g.Precede("X", "Y")
	g.Precede("Y", "X")
	_, err := g.TopologicalOrder()
	if err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestCriticalPath(t *testing.T) {
	g := makeTestGraph()
	path, dur := g.CriticalPath()
	if dur != 26 {
		t.Fatalf("critical path duration = %v, want 26", dur)
	}
	if len(path) == 0 {
		t.Fatal("empty critical path")
	}
	found := map[string]bool{}
	for _, id := range path {
		found[id] = true
	}
	if !found["A"] || !found["H"] {
		t.Fatalf("A or H missing from critical path: %v", path)
	}
}

func TestPositionalWeight(t *testing.T) {
	g := makeTestGraph()
	rpw := g.PositionalWeight()
	if rpw["A"] != 35 {
		t.Fatalf("RPW(A) = %v, want 35", rpw["A"])
	}
	if rpw["H"] != 3 {
		t.Fatalf("RPW(H) = %v, want 3", rpw["H"])
	}
}

func TestKilbridgeWesterColumns(t *testing.T) {
	g := makeTestGraph()
	cols := g.KilbridgeWesterColumns()
	if cols["A"] != 0 {
		t.Fatalf("col(A) = %d, want 0", cols["A"])
	}
	if cols["H"] < 3 {
		t.Fatalf("col(H) = %d, want >= 3", cols["H"])
	}
	if cols["B"] != 1 || cols["C"] != 1 {
		t.Fatalf("col(B)=%d col(C)=%d, want 1", cols["B"], cols["C"])
	}
}

func TestParseCSV(t *testing.T) {
	input := `task_id,seconds,predecessors
A,5,-
B,3,A
C,6,A
D,7,B;C
E,4,B
F,2,C
G,5,D
H,3,E;F;G
`
	g, err := ParseCSV(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if g.Size() != 8 {
		t.Fatalf("size = %d, want 8", g.Size())
	}
	path, dur := g.CriticalPath()
	if dur != 26 {
		t.Fatalf("critical path = %v (dur %v), want 26", path, dur)
	}
}

func TestTotalFloat(t *testing.T) {
	g := makeTestGraph()
	floats := g.TotalFloat()
	if floats["A"] != 0 {
		t.Fatalf("float(A) = %v, want 0", floats["A"])
	}
	if floats["H"] != 0 {
		t.Fatalf("float(H) = %v, want 0", floats["H"])
	}
	if floats["E"] != 11 {
		t.Fatalf("float(E) = %v, want 11", floats["E"])
	}
}

func TestRootsAndLeaves(t *testing.T) {
	g := makeTestGraph()
	roots := g.Roots()
	leaves := g.Leaves()
	if len(roots) != 1 || roots[0] != "A" {
		t.Fatalf("roots = %v, want [A]", roots)
	}
	if len(leaves) != 1 || leaves[0] != "H" {
		t.Fatalf("leaves = %v, want [H]", leaves)
	}
}
