package line

import (
	"fmt"
	"sort"

	"line-balancer/internal/task"
)

type DualSide int

const (
	SideLeft  DualSide = 0
	SideRight DualSide = 1
)

type DualStation struct {
	Left  StationSide
	Right StationSide
}

type StationSide struct {
	Tasks []string
	Load  float64
}

type DualResult struct {
	TaktTime     float64
	Stations     []DualStation
	StationCount int
	Bottleneck   int
	MaxLoad      float64
	Efficiency   float64
	TotalTime    float64
}

type DualPreference struct {
	TaskID string
	Side   DualSide
}

func DualBalance(g *task.Graph, cycleTime float64, prefs []DualPreference) ([]DualStation, error) {
	if g.Size() == 0 {
		return nil, fmt.Errorf("empty task graph")
	}
	if cycleTime <= 0 {
		return nil, fmt.Errorf("cycle time must be > 0, got %v", cycleTime)
	}

	prefMap := make(map[string]DualSide, len(prefs))
	for _, p := range prefs {
		prefMap[p.TaskID] = p.Side
	}

	rpw := g.PositionalWeight()
	ids := g.IDs()
	sort.Slice(ids, func(i, j int) bool {
		if rpw[ids[i]] != rpw[ids[j]] {
			return rpw[ids[i]] > rpw[ids[j]]
		}
		return ids[i] < ids[j]
	})

	assigned := make(map[string]bool, g.Size())
	var stations []DualStation
	remaining := make([]string, len(ids))
	copy(remaining, ids)

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

			pref, hasPref := prefMap[id]
			placed := false

			for i := range stations {
				if tryDualPlace(&stations[i], id, t.Time, cycleTime, pref, hasPref) {
					assigned[id] = true
					placed = true
					progress = true
					break
				}
			}
			if !placed {
				var ds DualStation
				if hasPref && pref == SideRight {
					ds.Right = StationSide{Tasks: []string{id}, Load: t.Time}
				} else {
					ds.Left = StationSide{Tasks: []string{id}, Load: t.Time}
				}
				stations = append(stations, ds)
				assigned[id] = true
				progress = true
			}
		}
		if !progress {
			for _, id := range next {
				t := g.Get(id)
				stations = append(stations, DualStation{
					Left: StationSide{Tasks: []string{id}, Load: t.Time},
				})
				assigned[id] = true
			}
			break
		}
		remaining = next
	}
	return stations, nil
}

func tryDualPlace(ds *DualStation, id string, time, cycleTime float64, pref DualSide, hasPref bool) bool {
	sides := [2]*StationSide{&ds.Left, &ds.Right}
	order := []int{0, 1}
	if hasPref && pref == SideRight {
		order = []int{1, 0}
	}
	for _, idx := range order {
		if sides[idx].Load+time <= cycleTime+1e-9 {
			sides[idx].Tasks = append(sides[idx].Tasks, id)
			sides[idx].Load += time
			return true
		}
	}
	return false
}

func DualAnalyze(g *task.Graph, demand int, availableSec float64, prefs []DualPreference) (DualResult, error) {
	if demand <= 0 {
		return DualResult{}, fmt.Errorf("demand must be > 0, got %d", demand)
	}
	if availableSec <= 0 {
		return DualResult{}, fmt.Errorf("available time must be > 0, got %v", availableSec)
	}
	takt := TaktTime(demand, availableSec)
	stations, err := DualBalance(g, takt, prefs)
	if err != nil {
		return DualResult{}, err
	}

	total := 0.0
	for _, id := range g.IDs() {
		total += g.Get(id).Time
	}
	maxLoad, bn := 0.0, 0
	for i, ds := range stations {
		load := ds.Left.Load
		if ds.Right.Load > load {
			load = ds.Right.Load
		}
		if load > maxLoad {
			maxLoad, bn = load, i
		}
	}
	eff := 0.0
	if len(stations) > 0 && takt > 0 {
		eff = total / (float64(len(stations)) * 2 * takt) * 100
	}
	return DualResult{
		TaktTime:     takt,
		Stations:     stations,
		StationCount: len(stations),
		Bottleneck:   bn,
		MaxLoad:      maxLoad,
		Efficiency:   eff,
		TotalTime:    total,
	}, nil
}
