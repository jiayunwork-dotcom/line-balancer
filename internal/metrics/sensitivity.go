package metrics

import "math"

type SensitivityResult struct {
	Baseline   float64
	HighDemand float64
	LowDemand  float64
	Elasticity float64
}

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
