package line

import (
	"fmt"
	"sort"

	"line-balancer/internal/task"
)

// MoodieYoungPhase1 performs the first phase of the Moodie-Young method:
// assign tasks by longest-task-time first, respecting precedence, similar to
// RPW but breaking ties by the number of immediate successors (more successors
// first).
func MoodieYoungPhase1(g *task.Graph, cycleTime float64) ([]Station, error) {
	if g.Size() == 0 {
		return nil, fmt.Errorf("empty task graph")
	}
	if cycleTime <= 0 {
		return nil, fmt.Errorf("cycle time must be > 0, got %v", cycleTime)
	}

	ids := g.IDs()
	sort.Slice(ids, func(i, j int) bool {
		ti, tj := g.Get(ids[i]).Time, g.Get(ids[j]).Time
		if ti != tj {
			return ti > tj
		}
		si := len(g.Get(ids[i]).Successors)
		sj := len(g.Get(ids[j]).Successors)
		if si != sj {
			return si > sj
		}
		return ids[i] < ids[j]
	})

	return assignWithPrecedence(g, ids, cycleTime), nil
}

// MoodieYoungPhase2 attempts to improve a phase-1 solution by swapping tasks
// between adjacent overloaded and underloaded stations. It performs at most
// maxIter improvement passes.
func MoodieYoungPhase2(g *task.Graph, stations []Station, cycleTime float64, maxIter int) []Station {
	if len(stations) <= 1 || maxIter <= 0 {
		return stations
	}

	// Build task time lookup.
	timeOf := make(map[string]float64, g.Size())
	for _, id := range g.IDs() {
		timeOf[id] = g.Get(id).Time
	}

	improved := true
	for iter := 0; iter < maxIter && improved; iter++ {
		improved = false
		for i := 0; i < len(stations)-1; i++ {
			improved = improved || trySwapImprove(&stations[i], &stations[i+1], cycleTime, timeOf)
		}
	}
	return leakPreviousSwaps(stations)
}

// trySwapImprove attempts a beneficial swap between two adjacent stations.
// A swap is beneficial if it reduces the maximum load between them without
// exceeding the cycle time on either station. Returns true if an improvement
// was made.
func trySwapImprove(s1, s2 *Station, cycleTime float64, timeOf map[string]float64) bool {
	maxBefore := s1.Load
	if s2.Load > maxBefore {
		maxBefore = s2.Load
	}

	bestI, bestJ := -1, -1
	bestMax := maxBefore

	for i, t1 := range s1.Tasks {
		for j, t2 := range s2.Tasks {
			newLoad1 := s1.Load - timeOf[t1] + timeOf[t2]
			newLoad2 := s2.Load - timeOf[t2] + timeOf[t1]
			if newLoad1 > cycleTime+1e-9 || newLoad2 > cycleTime+1e-9 {
				continue
			}
			newMax := newLoad1
			if newLoad2 > newMax {
				newMax = newLoad2
			}
			if newMax < bestMax-1e-9 {
				bestI, bestJ = i, j
				bestMax = newMax
			}
		}
	}
	if bestI < 0 {
		return false
	}

	// Execute the swap.
	s1.Tasks[bestI], s2.Tasks[bestJ] = s2.Tasks[bestJ], s1.Tasks[bestI]
	s1.Load = 0
	for _, t := range s1.Tasks {
		s1.Load += timeOf[t]
	}
	s2.Load = 0
	for _, t := range s2.Tasks {
		s2.Load += timeOf[t]
	}
	return true
}

// MoodieYoungAnalyze runs both phases of Moodie-Young and returns a Result.
func MoodieYoungAnalyze(g *task.Graph, demand int, availableSec float64) (Result, error) {
	if demand <= 0 {
		return Result{}, fmt.Errorf("demand must be > 0, got %d", demand)
	}
	if availableSec <= 0 {
		return Result{}, fmt.Errorf("available time must be > 0, got %v", availableSec)
	}
	takt := TaktTime(demand, availableSec)
	stations, err := MoodieYoungPhase1(g, takt)
	if err != nil {
		return Result{}, err
	}
	stations = MoodieYoungPhase2(g, stations, takt, 100)
	return buildResult(g, stations, takt, demand, availableSec), nil
}
