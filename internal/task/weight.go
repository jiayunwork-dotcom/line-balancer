package task

import "sort"

func (g *Graph) PositionalWeight() map[string]float64 {
	rpw := make(map[string]float64, len(g.tasks))
	for _, id := range g.order {
		reachable := g.reachableFrom(id)
		sum := g.tasks[id].Time
		for r := range reachable {
			if r != id {
				sum += g.tasks[r].Time
			}
		}
		rpw[id] = sum
	}
	return rpw
}

func (g *Graph) reachableFrom(id string) map[string]bool {
	visited := make(map[string]bool)
	var dfs func(string)
	dfs = func(cur string) {
		if visited[cur] {
			return
		}
		visited[cur] = true
		for _, s := range g.tasks[cur].Successors {
			dfs(s)
		}
	}
	dfs(id)
	return visited
}

func (g *Graph) ReversePositionalWeight() map[string]float64 {
	topo, err := g.TopologicalOrder()
	if err != nil {
		return nil
	}
	rrpw := make(map[string]float64, len(g.tasks))
	for _, id := range topo {
		t := g.tasks[id]
		rrpw[id] = t.Time
		for _, p := range t.Predecessors {
			rrpw[id] += rrpw[p]
		}
	}
	return rrpw
}

func (g *Graph) RPWSorted() []string {
	rpw := g.PositionalWeight()
	if rpw == nil {
		return nil
	}
	ids := g.IDs()
	sort.Slice(ids, func(i, j int) bool {
		if rpw[ids[i]] != rpw[ids[j]] {
			return rpw[ids[i]] > rpw[ids[j]]
		}
		return ids[i] < ids[j]
	})
	return ids
}

func (g *Graph) KilbridgeWesterColumns() map[string]int {
	topo, err := g.TopologicalOrder()
	if err != nil {
		return nil
	}
	col := make(map[string]int, len(g.tasks))
	for _, id := range topo {
		t := g.tasks[id]
		maxPredCol := -1
		for _, p := range t.Predecessors {
			if col[p] > maxPredCol {
				maxPredCol = col[p]
			}
		}
		col[id] = maxPredCol + 1
	}
	return col
}

func (g *Graph) KWSorted() []string {
	cols := g.KilbridgeWesterColumns()
	if cols == nil {
		return nil
	}
	ids := g.IDs()
	sort.Slice(ids, func(i, j int) bool {
		ci, cj := cols[ids[i]], cols[ids[j]]
		if ci != cj {
			return ci < cj
		}
		ti, tj := g.tasks[ids[i]].Time, g.tasks[ids[j]].Time
		if ti != tj {
			return ti > tj
		}
		return ids[i] < ids[j]
	})
	return ids
}

func (g *Graph) LevelCount() int {
	cols := g.KilbridgeWesterColumns()
	if cols == nil {
		return 0
	}
	maxCol := 0
	for _, c := range cols {
		if c > maxCol {
			maxCol = c
		}
	}
	return maxCol + 1
}
