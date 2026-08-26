package line

import (
	"fmt"

	"line-balancer/internal/task"
)

type ZoneConstraint struct {
	TaskID string
	Start  int
	End    int
}

type IncompatibilityConstraint struct {
	TaskA string
	TaskB string
}

type ConstrainedConfig struct {
	CycleTime     float64
	Zones         []ZoneConstraint
	Incompatibles []IncompatibilityConstraint
}

func ConstrainedBalance(g *task.Graph, cfg ConstrainedConfig) ([]Station, error) {
	if g.Size() == 0 {
		return nil, fmt.Errorf("empty task graph")
	}
	if cfg.CycleTime <= 0 {
		return nil, fmt.Errorf("cycle time must be > 0")
	}

	zoneMap := make(map[string]ZoneConstraint, len(cfg.Zones))
	for _, z := range cfg.Zones {
		zoneMap[z.TaskID] = z
	}
	incompMap := make(map[string]map[string]bool)
	for _, ic := range cfg.Incompatibles {
		if incompMap[ic.TaskA] == nil {
			incompMap[ic.TaskA] = make(map[string]bool)
		}
		if incompMap[ic.TaskB] == nil {
			incompMap[ic.TaskB] = make(map[string]bool)
		}
		incompMap[ic.TaskA][ic.TaskB] = true
		incompMap[ic.TaskB][ic.TaskA] = true
	}

	rpwSorted := g.RPWSorted()
	assigned := make(map[string]bool, g.Size())
	stationTasks := make(map[int]map[string]bool)
	var stations []Station
	remaining := make([]string, len(rpwSorted))
	copy(remaining, rpwSorted)

	ensureStation := func(idx int) {
		for len(stations) <= idx {
			stations = append(stations, Station{})
		}
		if stationTasks[idx] == nil {
			stationTasks[idx] = make(map[string]bool)
		}
	}

	canPlace := func(id string, stIdx int) bool {
		ensureStation(stIdx)
		if z, ok := zoneMap[id]; ok {
			if stIdx < z.Start || stIdx > z.End {
				return false
			}
		}
		t := g.Get(id)
		if stations[stIdx].Load+t.Time > cfg.CycleTime+1e-9 {
			return false
		}
		if incompat, ok := incompMap[id]; ok {
			for other := range incompat {
				if stationTasks[stIdx][other] {
					return false
				}
			}
		}
		return true
	}

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
			maxIdx := len(stations) + 1
			if z, ok := zoneMap[id]; ok {
				maxIdx = z.End + 1
			}
			for i := 0; i < maxIdx; i++ {
				ensureStation(i)
				if canPlace(id, i) {
					stations[i].Tasks = append(stations[i].Tasks, id)
					stations[i].Load += t.Time
					stationTasks[i][id] = true
					assigned[id] = true
					placed = true
					progress = true
					break
				}
			}
			if !placed {
				idx := len(stations)
				ensureStation(idx)
				stations[idx].Tasks = append(stations[idx].Tasks, id)
				stations[idx].Load += t.Time
				stationTasks[idx] = map[string]bool{id: true}
				assigned[id] = true
				progress = true
			}
		}
		if !progress {
			for _, id := range next {
				t := g.Get(id)
				idx := len(stations)
				ensureStation(idx)
				stations[idx].Tasks = append(stations[idx].Tasks, id)
				stations[idx].Load += t.Time
				stationTasks[idx] = map[string]bool{id: true}
				assigned[id] = true
			}
			break
		}
		remaining = next
	}

	for len(stations) > 0 && len(stations[len(stations)-1].Tasks) == 0 {
		stations = stations[:len(stations)-1]
	}
	return stations, nil
}

func ValidateAssignment(stations []Station, g *task.Graph, cfg ConstrainedConfig) error {
	taskStation := make(map[string]int)
	for i, s := range stations {
		for _, id := range s.Tasks {
			taskStation[id] = i
		}
	}

	for _, id := range g.IDs() {
		t := g.Get(id)
		myStation := taskStation[id]
		for _, p := range t.Predecessors {
			predStation, ok := taskStation[p]
			if !ok {
				return fmt.Errorf("predecessor %q of %q not assigned", p, id)
			}
			if predStation > myStation {
				return fmt.Errorf("precedence violation: %q (station %d) must precede %q (station %d)", p, predStation, id, myStation)
			}
		}
	}

	for _, z := range cfg.Zones {
		st, ok := taskStation[z.TaskID]
		if !ok {
			continue
		}
		if st < z.Start || st > z.End {
			return fmt.Errorf("zone violation: %q at station %d, must be in [%d,%d]", z.TaskID, st, z.Start, z.End)
		}
	}

	for _, ic := range cfg.Incompatibles {
		stA, okA := taskStation[ic.TaskA]
		stB, okB := taskStation[ic.TaskB]
		if okA && okB && stA == stB {
			return fmt.Errorf("incompatibility violation: %q and %q both on station %d", ic.TaskA, ic.TaskB, stA)
		}
	}

	for i, s := range stations {
		if s.Load > cfg.CycleTime+1e-9 {
			return fmt.Errorf("station %d overloaded: %.2f > %.2f", i, s.Load, cfg.CycleTime)
		}
	}
	return nil
}
