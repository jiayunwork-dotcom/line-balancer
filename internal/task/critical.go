package task

// CriticalPath computes the longest path through the task graph (the critical
// path) and returns the ordered task IDs along that path together with its
// total duration. If the graph is empty, returns nil and 0.
func (g *Graph) CriticalPath() ([]string, float64) {
	if len(g.tasks) == 0 {
		return nil, 0
	}
	topo, err := g.TopologicalOrder()
	if err != nil {
		return nil, 0
	}

	// Forward pass: earliest start / earliest finish.
	es := make(map[string]float64, len(g.tasks))
	ef := make(map[string]float64, len(g.tasks))
	for _, id := range topo {
		t := g.tasks[id]
		ef[id] = es[id] + t.Time
		for _, s := range t.Successors {
			if ef[id] > es[s] {
				es[s] = ef[id]
			}
		}
	}

	// Total project duration = max EF among leaves.
	projectDur := 0.0
	for _, id := range g.Leaves() {
		if ef[id] > projectDur {
			projectDur = ef[id]
		}
	}

	// Backward pass: latest start / latest finish.
	lf := make(map[string]float64, len(g.tasks))
	ls := make(map[string]float64, len(g.tasks))
	for _, id := range g.order {
		lf[id] = projectDur
	}
	for i := len(topo) - 1; i >= 0; i-- {
		id := topo[i]
		t := g.tasks[id]
		ls[id] = lf[id] - t.Time
		for _, p := range t.Predecessors {
			if ls[id] < lf[p] {
				lf[p] = ls[id]
			}
		}
	}

	// Tasks on the critical path have zero total float (LS - ES = 0).
	var path []string
	for _, id := range topo {
		slack := ls[id] - es[id]
		if slack < 1e-9 && slack > -1e-9 {
			path = append(path, id)
		}
	}
	return path, projectDur
}

// EarliestTimes returns the earliest start and earliest finish for each task.
func (g *Graph) EarliestTimes() (es, ef map[string]float64) {
	topo, err := g.TopologicalOrder()
	if err != nil {
		return nil, nil
	}
	es = make(map[string]float64, len(g.tasks))
	ef = make(map[string]float64, len(g.tasks))
	for _, id := range topo {
		t := g.tasks[id]
		ef[id] = es[id] + t.Time
		for _, s := range t.Successors {
			if ef[id] > es[s] {
				es[s] = ef[id]
			}
		}
	}
	return es, ef
}

// LatestTimes returns the latest start and latest finish for each task given
// the project deadline (typically the critical path duration).
func (g *Graph) LatestTimes(deadline float64) (ls, lf map[string]float64) {
	topo, err := g.TopologicalOrder()
	if err != nil {
		return nil, nil
	}
	lf = make(map[string]float64, len(g.tasks))
	ls = make(map[string]float64, len(g.tasks))
	for _, id := range g.order {
		lf[id] = deadline
	}
	for i := len(topo) - 1; i >= 0; i-- {
		id := topo[i]
		t := g.tasks[id]
		ls[id] = lf[id] - t.Time
		for _, p := range t.Predecessors {
			if ls[id] < lf[p] {
				lf[p] = ls[id]
			}
		}
	}
	return ls, lf
}

// TotalFloat returns the total float (slack) for each task. A task on the
// critical path has zero float.
func (g *Graph) TotalFloat() map[string]float64 {
	topo, err := g.TopologicalOrder()
	if err != nil {
		return nil
	}
	es := make(map[string]float64, len(g.tasks))
	ef := make(map[string]float64, len(g.tasks))
	for _, id := range topo {
		t := g.tasks[id]
		ef[id] = es[id] + t.Time
		for _, s := range t.Successors {
			if ef[id] > es[s] {
				es[s] = ef[id]
			}
		}
	}
	projectDur := 0.0
	for _, id := range g.Leaves() {
		if ef[id] > projectDur {
			projectDur = ef[id]
		}
	}
	lf := make(map[string]float64, len(g.tasks))
	ls := make(map[string]float64, len(g.tasks))
	for _, id := range g.order {
		lf[id] = projectDur
	}
	for i := len(topo) - 1; i >= 0; i-- {
		id := topo[i]
		t := g.tasks[id]
		ls[id] = lf[id] - t.Time
		for _, p := range t.Predecessors {
			if ls[id] < lf[p] {
				lf[p] = ls[id]
			}
		}
	}
	floats := make(map[string]float64, len(g.tasks))
	for _, id := range g.order {
		floats[id] = ls[id] - es[id]
	}
	return floats
}
