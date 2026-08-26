package task

import "fmt"

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
	for _, id := range a.IDs() {
		t := a.Get(id)
		for _, s := range t.Successors {
			merged.Precede(id, s)
		}
	}
	for _, id := range b.IDs() {
		t := b.Get(id)
		for _, s := range t.Successors {
			merged.Precede(id, s)
		}
	}
	return merged, nil
}

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

func (g *Graph) LongestPath() float64 {
	_, dur := g.CriticalPath()
	return dur
}

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
