package metrics

import "math"

// SensitivityResult captures how a metric changes when demand or cycle time is
// perturbed by a given fraction.
type SensitivityResult struct {
	Baseline float64
	HighDemand float64  // metric value when demand increased by fraction
	LowDemand  float64  // metric value when demand decreased by fraction
	Elasticity float64  // approximate elasticity: (deltaMetric/metric) / (deltaDemand/demand)
}

// DemandSensitivity analyzes how efficiency changes when demand is perturbed by
// +/- fraction (e.g., 0.1 = 10%). It computes a simple finite-difference
// elasticity estimate.
func DemandSensitivity(totalTime float64, numStations int, baseDemand int, availableSec, fraction float64) SensitivityResult {
	baseCycle := CycleTimeForDemand(baseDemand, availableSec)
	baseEff := Efficiency(LineSummary{
		StationLoads: uniformLoads(numStations, totalTime),
		CycleTime:    baseCycle,
		TotalTime:    totalTime,
	})

	highDemand := int(math.Ceil(float64(baseDemand) * (1 + fraction)))
	highCycle := CycleTimeForDemand(highDemand, availableSec)
	highEff := Efficiency(LineSummary{
		StationLoads: uniformLoads(numStations, totalTime),
		CycleTime:    highCycle,
		TotalTime:    totalTime,
	})

	lowDemand := int(math.Floor(float64(baseDemand) * (1 - fraction)))
	if lowDemand < 1 {
		lowDemand = 1
	}
	lowCycle := CycleTimeForDemand(lowDemand, availableSec)
	lowEff := Efficiency(LineSummary{
		StationLoads: uniformLoads(numStations, totalTime),
		CycleTime:    lowCycle,
		TotalTime:    totalTime,
	})

	elasticity := 0.0
	if baseEff > 0 && fraction > 0 {
		deltaMetric := highEff - baseEff
		elasticity = (deltaMetric / baseEff) / fraction
	}

	return SensitivityResult{
		Baseline:   baseEff,
		HighDemand: highEff,
		LowDemand:  lowEff,
		Elasticity: elasticity,
	}
}

// uniformLoads creates a uniform load distribution across n stations.
func uniformLoads(n int, totalTime float64) []float64 {
	if n <= 0 {
		return nil
	}
	perStation := totalTime / float64(n)
	loads := make([]float64, n)
	for i := range loads {
		loads[i] = perStation
	}
	return loads
}

// BottleneckImpact estimates the effect of reducing the bottleneck by a given
// amount. Returns the new efficiency if the bottleneck is reduced.
func BottleneckImpact(s LineSummary, reduction float64) float64 {
	if len(s.StationLoads) == 0 || reduction < 0 {
		return 0
	}
	maxIdx := 0
	for i, l := range s.StationLoads {
		if l > s.StationLoads[maxIdx] {
			maxIdx = i
		}
	}
	modified := make([]float64, len(s.StationLoads))
	copy(modified, s.StationLoads)
	modified[maxIdx] -= reduction
	if modified[maxIdx] < 0 {
		modified[maxIdx] = 0
	}
	newTotal := 0.0
	for _, l := range modified {
		newTotal += l
	}
	return Efficiency(LineSummary{
		StationLoads: modified,
		CycleTime:    s.CycleTime,
		TotalTime:    newTotal,
	})
}

// CycleTimeSweep computes efficiency for a range of cycle times from min to max
// in the given number of steps. Useful for understanding the trade-off between
// cycle time and number of stations needed.
func CycleTimeSweep(totalTime float64, numStations int, minCycle, maxCycle float64, steps int) []float64 {
	if steps <= 0 || minCycle >= maxCycle {
		return nil
	}
	effs := make([]float64, steps)
	step := (maxCycle - minCycle) / float64(steps-1)
	for i := 0; i < steps; i++ {
		ct := minCycle + float64(i)*step
		effs[i] = Efficiency(LineSummary{
			StationLoads: uniformLoads(numStations, totalTime),
			CycleTime:    ct,
			TotalTime:    totalTime,
		})
	}
	return effs
}
