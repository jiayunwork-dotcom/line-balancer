package line

import (
	"fmt"
	"sort"

	"line-balancer/internal/task"
)

func KWBalance(g *task.Graph, cycleTime float64) ([]Station, error) {
	if g.Size() == 0 {
		return nil, fmt.Errorf("empty task graph")
	}
	if cycleTime <= 0 {
		return nil, fmt.Errorf("cycle time must be > 0, got %v", cycleTime)
	}

	cols := g.KilbridgeWesterColumns()
	ids := g.IDs()

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

	stations := assignWithPrecedence(g, ids, cycleTime)
	return HoldKWLive(stations), nil
}

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
