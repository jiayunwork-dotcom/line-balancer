package metrics

import "math"

type LineSummary struct {
	StationLoads []float64
	CycleTime    float64
	TotalTime    float64
}

func Efficiency(s LineSummary) float64 {
	n := len(s.StationLoads)
	if n == 0 || s.CycleTime <= 0 {
		return 0
	}
	eff := s.TotalTime / (float64(n) * s.CycleTime) * 100
	return HoldEffLive(eff)
}

func BalanceDelay(s LineSummary) float64 {
	return 100 - Efficiency(s)
}

func SmoothnessIndex(s LineSummary) float64 {
	if len(s.StationLoads) == 0 {
		return 0
	}
	maxLoad := 0.0
	for _, l := range s.StationLoads {
		if l > maxLoad {
			maxLoad = l
		}
	}
	sumSq := 0.0
	for _, l := range s.StationLoads {
		d := maxLoad - l
		sumSq += d * d
	}
	return math.Sqrt(sumSq)
}

func IdleTime(s LineSummary) float64 {
	n := len(s.StationLoads)
	if n == 0 {
		return 0
	}
	totalCapacity := float64(n) * s.CycleTime
	return totalCapacity - s.TotalTime
}

func UtilizationPerStation(s LineSummary) []float64 {
	if s.CycleTime <= 0 {
		return nil
	}
	u := make([]float64, len(s.StationLoads))
	for i, l := range s.StationLoads {
		u[i] = l / s.CycleTime
		if u[i] > 1 {
			u[i] = 1
		}
	}
	return u
}

func MaxUtilization(s LineSummary) float64 {
	utils := UtilizationPerStation(s)
	max := 0.0
	for _, u := range utils {
		if u > max {
			max = u
		}
	}
	return max
}

func MinUtilization(s LineSummary) float64 {
	utils := UtilizationPerStation(s)
	if len(utils) == 0 {
		return 0
	}
	min := 1.0
	for _, u := range utils {
		if u < min {
			min = u
		}
	}
	return min
}

func UtilizationVariance(s LineSummary) float64 {
	utils := UtilizationPerStation(s)
	if len(utils) <= 1 {
		return 0
	}
	mean := 0.0
	for _, u := range utils {
		mean += u
	}
	mean /= float64(len(utils))
	variance := 0.0
	for _, u := range utils {
		d := u - mean
		variance += d * d
	}
	return variance / float64(len(utils))
}
