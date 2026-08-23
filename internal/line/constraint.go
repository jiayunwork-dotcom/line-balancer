package line

import (
	"fmt"

	"line-balancer/internal/task"
)

// ZoneConstraint restricts tasks to specific station zones. A zone is a
// contiguous range of station indices [Start, End] (inclusive).
type ZoneConstraint struct {
	TaskID string
	Start  int
	End    int
}

// IncompatibilityConstraint declares that two tasks cannot be assigned to
// the same station.
type IncompatibilityConstraint struct {
	TaskA string
	TaskB string
}

// ConstrainedConfig holds all constraints for a balance problem.
type ConstrainedConfig struct {
	CycleTime       float64
	Zones           []ZoneConstraint
	Incompatibles   []IncompatibilityConstraint
}

// ConstrainedBalance assigns tasks using RPW order while respecting zone and
// incompatibility constraints. Tasks that cannot be placed due to constraints
// are assigned to new stations beyond the constrained range.
func ConstrainedBalance(g *task.Graph, cfg ConstrainedConfig) ([]Station, error) {
	if g.Size() == 0 {
		return nil, fmt.Errorf("empty task graph")
	}
	if cfg.CycleTime <= 0 {
		return nil, fmt.Errorf("cycle time must be > 0")
	}

	// Build constraint maps.
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
	stationTasks := make(map[int]map[string]bool) // station index -> task IDs on it
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
		// Check zone constraint.
		if z, ok := zoneMap[id]; ok {
			if stIdx < z.Start || stIdx > z.End {
				return false
			}
		}
		// Check capacity.
		t := g.Get(id)
		if stations[stIdx].Load+t.Time > cfg.CycleTime+1e-9 {
			return false
		}
		// Check incompatibility.
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
				// Open a new station beyond constraints.
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

	// Remove empty trailing stations.
	for len(stations) > 0 && len(stations[len(stations)-1].Tasks) == 0 {
		stations = stations[:len(stations)-1]
	}
	return stations, nil
}

// ValidateAssignment checks whether a station assignment satisfies all
// constraints. Returns nil if valid, or an error describing the first
// violation found.
func ValidateAssignment(stations []Station, g *task.Graph, cfg ConstrainedConfig) error {
	taskStation := make(map[string]int)
	for i, s := range stations {
		for _, id := range s.Tasks {
			taskStation[id] = i
		}
	}

	// Check precedence.
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

	// Check zones.
	for _, z := range cfg.Zones {
		st, ok := taskStation[z.TaskID]
		if !ok {
			continue
		}
		if st < z.Start || st > z.End {
			return fmt.Errorf("zone violation: %q at station %d, must be in [%d,%d]", z.TaskID, st, z.Start, z.End)
		}
	}

	// Check incompatibility.
	for _, ic := range cfg.Incompatibles {
		stA, okA := taskStation[ic.TaskA]
		stB, okB := taskStation[ic.TaskB]
		if okA && okB && stA == stB {
			return fmt.Errorf("incompatibility violation: %q and %q both on station %d", ic.TaskA, ic.TaskB, stA)
		}
	}

	// Check cycle time.
	for i, s := range stations {
		if s.Load > cfg.CycleTime+1e-9 {
			return fmt.Errorf("station %d overloaded: %.2f > %.2f", i, s.Load, cfg.CycleTime)
		}
	}
	return nil
}
