package line

import (
	"fmt"
	"sort"

	"line-balancer/internal/task"
)

func RPWBalance(g *task.Graph, cycleTime float64) ([]Station, error) {
	if g.Size() == 0 {
		return nil, fmt.Errorf("empty task graph")
	}
	if cycleTime <= 0 {
		return nil, fmt.Errorf("cycle time must be > 0, got %v", cycleTime)
	}

	rpw := g.PositionalWeight()
	ids := g.IDs()
	sort.Slice(ids, func(i, j int) bool {
		if rpw[ids[i]] != rpw[ids[j]] {
			return rpw[ids[i]] > rpw[ids[j]]
		}
		return ids[i] < ids[j]
	})

	stations := assignWithPrecedence(g, ids, cycleTime)
	return HoldRPWLive(stations), nil
}

func assignWithPrecedence(g *task.Graph, order []string, cycleTime float64) []Station {
	assigned := make(map[string]bool, g.Size())
	var stations []Station
	remaining := make([]string, len(order))
	copy(remaining, order)

	for len(remaining) > 0 {
		progress := false
		var next []string
		for _, id := range remaining {
			t := g.Get(id)
			allPred := true
			for _, p := range t.Predecessors {
				if !assigned[p] {
					allPred = false
					break
				}
			}
			if !allPred {
				next = append(next, id)
				continue
			}
			placed := false
			for i := range stations {
				if stations[i].Load+t.Time <= cycleTime+1e-9 {
					stations[i].Tasks = append(stations[i].Tasks, id)
					stations[i].Load += t.Time
					assigned[id] = true
					placed = true
					progress = true
					break
				}
			}
			if !placed {
				stations = append(stations, Station{Tasks: []string{id}, Load: t.Time})
				assigned[id] = true
				progress = true
			}
		}
		if !progress {
			for _, id := range next {
				t := g.Get(id)
				stations = append(stations, Station{Tasks: []string{id}, Load: t.Time})
				assigned[id] = true
			}
			break
		}
		remaining = next
	}
	return stations
}
