package task

import "fmt"

// SubGraph extracts a sub-graph containing only the specified task IDs and the
// edges between them. Returns error if any specified ID does not exist.
func (g *Graph) SubGraph(ids []string) (*Graph, error) {
	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		if _, ok := g.tasks[id]; !ok {
			return nil, fmt.Errorf("unknown task %q", id)
		}
		idSet[id] = true
	}

	sub := NewGraph()
	for _, id := range ids {
		t := g.tasks[id]
		sub.Add(id, t.Time)
	}
	for _, id := range ids {
		t := g.tasks[id]
		for _, s := range t.Successors {
			if idSet[s] {
				sub.Precede(id, s)
			}
		}
	}
	return sub, nil
}

// Merge combines two disjoint graphs into one. Returns error if task IDs
// overlap.
func Merge(a, b *Graph) (*Graph, error) {
	merged := NewGraph()
	for _, id := range a.IDs() {
		t := a.Get(id)
		if err := merged.Add(id, t.Time); err != nil {
			return nil, err
		}
	}
	for _, id := range b.IDs() {
		t := b.Get(id)
		if err := merged.Add(id, t.Time); err != nil {
			return nil, fmt.Errorf("merge conflict: %w", err)
		}
	}
	// Add edges from a.
	for _, id := range a.IDs() {
		t := a.Get(id)
		for _, s := range t.Successors {
			merged.Precede(id, s)
		}
	}
	// Add edges from b.
	for _, id := range b.IDs() {
		t := b.Get(id)
		for _, s := range t.Successors {
			merged.Precede(id, s)
		}
	}
	return merged, nil
}

// Density returns the edge density of the graph: edges / maxPossibleEdges.
// For a DAG with n nodes, max edges = n*(n-1)/2.
func (g *Graph) Density() float64 {
	n := len(g.tasks)
	if n <= 1 {
		return 0
	}
	edges := 0
	for _, t := range g.tasks {
		edges += len(t.Successors)
	}
	maxEdges := n * (n - 1) / 2
	return float64(edges) / float64(maxEdges)
}

// LongestPath returns the length of the longest path (sum of task times) from
// any root to any leaf. This equals the critical path duration.
func (g *Graph) LongestPath() float64 {
	_, dur := g.CriticalPath()
	return dur
}

// ShortestPath returns the minimum total processing time along any path from
// a root to a leaf.
func (g *Graph) ShortestPath() float64 {
	topo, err := g.TopologicalOrder()
	if err != nil {
		return 0
	}
	roots := g.Roots()
	leaves := g.Leaves()

	if len(roots) == 0 || len(leaves) == 0 {
		return 0
	}

	// Forward pass tracking shortest distance from any root.
	dist := make(map[string]float64, len(g.tasks))
	for _, id := range g.order {
		dist[id] = 1e18
	}
	for _, r := range roots {
		dist[r] = g.tasks[r].Time
	}
	for _, id := range topo {
		t := g.tasks[id]
		for _, s := range t.Successors {
			d := dist[id] + g.tasks[s].Time
			if d < dist[s] {
				dist[s] = d
			}
		}
	}

	shortest := 1e18
	leafSet := make(map[string]bool, len(leaves))
	for _, l := range leaves {
		leafSet[l] = true
	}
	for _, id := range g.order {
		if leafSet[id] && dist[id] < shortest {
			shortest = dist[id]
		}
	}
	if shortest >= 1e18 {
		return 0
	}
	return shortest
}
