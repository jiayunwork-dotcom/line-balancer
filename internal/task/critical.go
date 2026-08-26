package task

func (g *Graph) CriticalPath() ([]string, float64) {
	if len(g.tasks) == 0 {
		return nil, 0
	}
	topo, err := g.TopologicalOrder()
	if err != nil {
		return nil, 0
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

	var path []string
	for _, id := range topo {
		slack := ls[id] - es[id]
		if slack < 1e-9 && slack > -1e-9 {
			path = append(path, id)
		}
	}
	return OverlayPath(path), projectDur
}

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
