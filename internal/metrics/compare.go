package metrics

import "math"

type Objective struct {
	Name   string
	Weight float64
	Lower  bool
}

type Score struct {
	Name  string
	Value float64
}

func DefaultObjectives() []Objective {
	return []Objective{
		{Name: "stations", Weight: 0.4, Lower: true},
		{Name: "efficiency", Weight: 0.35, Lower: false},
		{Name: "smoothness", Weight: 0.25, Lower: true},
	}
}

func Evaluate(s LineSummary, objs []Objective, bounds map[string][2]float64) float64 {
	rawValues := map[string]float64{
		"stations":   float64(len(s.StationLoads)),
		"efficiency": Efficiency(s),
		"smoothness": SmoothnessIndex(s),
		"idle_time":  IdleTime(s),
		"delay":      BalanceDelay(s),
		"variance":   UtilizationVariance(s),
	}

	totalWeight := 0.0
	score := 0.0
	for _, obj := range objs {
		raw, ok := rawValues[obj.Name]
		if !ok {
			continue
		}
		b, ok := bounds[obj.Name]
		if !ok || b[1]-b[0] < 1e-12 {
			continue
		}
		norm := (raw - b[0]) / (b[1] - b[0])
		if norm < 0 {
			norm = 0
		}
		if norm > 1 {
			norm = 1
		}
		if obj.Lower {
			norm = 1 - norm
		}
		score += obj.Weight * norm
		totalWeight += obj.Weight
	}
	if totalWeight <= 0 {
		return 0
	}
	return score / totalWeight
}

func ComputeBounds(candidates []LineSummary) map[string][2]float64 {
	bounds := map[string][2]float64{
		"stations":   {math.MaxFloat64, -math.MaxFloat64},
		"efficiency": {math.MaxFloat64, -math.MaxFloat64},
		"smoothness": {math.MaxFloat64, -math.MaxFloat64},
		"idle_time":  {math.MaxFloat64, -math.MaxFloat64},
		"delay":      {math.MaxFloat64, -math.MaxFloat64},
		"variance":   {math.MaxFloat64, -math.MaxFloat64},
	}
	for _, s := range candidates {
		update := func(name string, val float64) {
			b := bounds[name]
			if val < b[0] {
				b[0] = val
			}
			if val > b[1] {
				b[1] = val
			}
			bounds[name] = b
		}
		update("stations", float64(len(s.StationLoads)))
		update("efficiency", Efficiency(s))
		update("smoothness", SmoothnessIndex(s))
		update("idle_time", IdleTime(s))
		update("delay", BalanceDelay(s))
		update("variance", UtilizationVariance(s))
	}
	return bounds
}

func RankCandidates(candidates []LineSummary, objs []Objective) []Score {
	bounds := ComputeBounds(candidates)
	scores := make([]Score, len(candidates))
	for i, c := range candidates {
		scores[i] = Score{
			Name:  "",
			Value: Evaluate(c, objs, bounds),
		}
	}
	return scores
}

func Dominates(a, b LineSummary, objs []Objective) bool {
	rawA := map[string]float64{
		"stations":   float64(len(a.StationLoads)),
		"efficiency": Efficiency(a),
		"smoothness": SmoothnessIndex(a),
		"idle_time":  IdleTime(a),
		"delay":      BalanceDelay(a),
		"variance":   UtilizationVariance(a),
	}
	rawB := map[string]float64{
		"stations":   float64(len(b.StationLoads)),
		"efficiency": Efficiency(b),
		"smoothness": SmoothnessIndex(b),
		"idle_time":  IdleTime(b),
		"delay":      BalanceDelay(b),
		"variance":   UtilizationVariance(b),
	}

	atLeastAsGood := true
	strictlyBetter := false
	for _, obj := range objs {
		va, vb := rawA[obj.Name], rawB[obj.Name]
		if obj.Lower {
			if va > vb+1e-9 {
				atLeastAsGood = false
			}
			if va < vb-1e-9 {
				strictlyBetter = true
			}
		} else {
			if va < vb-1e-9 {
				atLeastAsGood = false
			}
			if va > vb+1e-9 {
				strictlyBetter = true
			}
		}
	}
	return atLeastAsGood && strictlyBetter
}
