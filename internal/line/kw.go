package line

import (
	"fmt"
	"sort"

	"line-balancer/internal/task"
)

// KWBalance assigns tasks to stations using the Kilbridge-Wester method.
// Tasks are processed column by column (column 0 = root tasks, etc.), and
// within each column sorted by descending task time. Assignment respects the
// cycle time constraint.
func KWBalance(g *task.Graph, cycleTime float64) ([]Station, error) {
	if g.Size() == 0 {
		return nil, fmt.Errorf("empty task graph")
	}
	if cycleTime <= 0 {
		return nil, fmt.Errorf("cycle time must be > 0, got %v", cycleTime)
	}

	cols := g.KilbridgeWesterColumns()
	ids := g.IDs()

	// Sort by column ascending, then by descending time, then by ID.
	sort.Slice(ids, func(i, j int) bool {
		ci, cj := cols[ids[i]], cols[ids[j]]
		if ci != cj {
			return ci < cj
		}
		ti, tj := g.Get(ids[i]).Time, g.Get(ids[j]).Time
		if ti != tj {
			return ti > tj
		}
		return ids[i] < ids[j]
	})

	return assignWithPrecedence(g, ids, cycleTime), nil
}

// KWAnalyze runs KW balance and returns a full Result.
func KWAnalyze(g *task.Graph, demand int, availableSec float64) (Result, error) {
	if demand <= 0 {
		return Result{}, fmt.Errorf("demand must be > 0, got %d", demand)
	}
	if availableSec <= 0 {
		return Result{}, fmt.Errorf("available time must be > 0, got %v", availableSec)
	}
	takt := TaktTime(demand, availableSec)
	stations, err := KWBalance(g, takt)
	if err != nil {
		return Result{}, err
	}
	return buildResult(g, stations, takt, demand, availableSec), nil
}

// RPWAnalyze runs RPW balance and returns a full Result.
func RPWAnalyze(g *task.Graph, demand int, availableSec float64) (Result, error) {
	if demand <= 0 {
		return Result{}, fmt.Errorf("demand must be > 0, got %d", demand)
	}
	if availableSec <= 0 {
		return Result{}, fmt.Errorf("available time must be > 0, got %v", availableSec)
	}
	takt := TaktTime(demand, availableSec)
	stations, err := RPWBalance(g, takt)
	if err != nil {
		return Result{}, err
	}
	return buildResult(g, stations, takt, demand, availableSec), nil
}

// buildResult constructs a Result from a set of stations.
func buildResult(g *task.Graph, stations []Station, takt float64, demand int, avail float64) Result {
	total := 0.0
	for _, id := range g.IDs() {
		total += g.Get(id).Time
	}
	maxLoad, bn := 0.0, 0
	for i, s := range stations {
		if s.Load > maxLoad {
			maxLoad, bn = s.Load, i
		}
	}
	eff := 0.0
	if len(stations) > 0 && takt > 0 {
		eff = total / (float64(len(stations)) * takt) * 100
	}
	return Result{
		TaktTime:     takt,
		CycleTime:    takt,
		Stations:     stations,
		StationCount: len(stations),
		Bottleneck:   bn,
		MaxLoad:      maxLoad,
		Efficiency:   eff,
		TotalTime:    total,
	}
}
