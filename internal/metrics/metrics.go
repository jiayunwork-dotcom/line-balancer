// Package metrics computes production line performance indicators: efficiency,
// smoothness index, balance delay, utilization distribution, and multi-objective
// evaluation scores for comparing balance solutions.
package metrics

import "math"

// LineSummary captures the key parameters of a balanced line for evaluation.
type LineSummary struct {
	StationLoads []float64
	CycleTime    float64
	TotalTime    float64
}

// Efficiency returns the line efficiency as a percentage (0-100).
// Efficiency = TotalTime / (NumStations * CycleTime) * 100.
func Efficiency(s LineSummary) float64 {
	n := len(s.StationLoads)
	if n == 0 || s.CycleTime <= 0 {
		return 0
	}
	return s.TotalTime / (float64(n) * s.CycleTime) * 100
}

// BalanceDelay returns the balance delay (idle time ratio) as a percentage.
// BalanceDelay = 100 - Efficiency.
func BalanceDelay(s LineSummary) float64 {
	return 100 - Efficiency(s)
}

// SmoothnessIndex measures how evenly tasks are distributed across stations.
// SI = sqrt(sum((maxLoad - load_i)^2)). Lower is better; 0 means perfectly
// smooth (all stations equally loaded).
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

// IdleTime returns the total idle time across all stations.
func IdleTime(s LineSummary) float64 {
	n := len(s.StationLoads)
	if n == 0 {
		return 0
	}
	totalCapacity := float64(n) * s.CycleTime
	return totalCapacity - s.TotalTime
}

// UtilizationPerStation returns the utilization fraction (0-1) for each station.
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

// MaxUtilization returns the maximum station utilization (0-1).
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

// MinUtilization returns the minimum station utilization (0-1).
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

// UtilizationVariance returns the variance of station utilizations.
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
