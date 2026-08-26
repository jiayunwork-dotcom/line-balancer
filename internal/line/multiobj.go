package line

import (
	"line-balancer/internal/task"
)

type CompareResult struct {
	Method     string
	Result     Result
	MinTheory  int
	ExtraCount int
}

func CompareAll(g *task.Graph, demand int, availableSec float64) ([]CompareResult, error) {
	takt := TaktTime(demand, availableSec)
	totalTime := 0.0
	for _, id := range g.IDs() {
		totalTime += g.Get(id).Time
	}
	minStations := int(totalTime / takt)
	if totalTime-float64(minStations)*takt > 1e-9 {
		minStations++
	}

	methods := []struct {
		name string
		fn   func(*task.Graph, int, float64) (Result, error)
	}{
		{"rpw", RPWAnalyze},
		{"kw", KWAnalyze},
		{"moodie", MoodieYoungAnalyze},
	}

	results := make([]CompareResult, 0, len(methods))
	for _, m := range methods {
		res, err := m.fn(g, demand, availableSec)
		if err != nil {
			return nil, err
		}
		results = append(results, CompareResult{
			Method:     m.name,
			Result:     res,
			MinTheory:  minStations,
			ExtraCount: res.StationCount - minStations,
		})
	}
	return results, nil
}

func BestResult(results []CompareResult) CompareResult {
	if len(results) == 0 {
		return CompareResult{}
	}
	best := results[0]
	for _, r := range results[1:] {
		rFeasible := r.Result.MaxLoad <= r.Result.TaktTime+1e-9
		bFeasible := best.Result.MaxLoad <= best.Result.TaktTime+1e-9
		if rFeasible && !bFeasible {
			best = r
			continue
		}
		if !rFeasible && bFeasible {
			continue
		}
		if r.Result.Efficiency > best.Result.Efficiency {
			best = r
		}
	}
	return best
}
