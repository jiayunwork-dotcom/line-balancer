package metrics

import "math"

// Throughput computes the maximum achievable production rate (units per time
// unit) given the bottleneck station load.
func Throughput(bottleneckLoad float64) float64 {
	if bottleneckLoad <= 0 {
		return 0
	}
	return 1.0 / bottleneckLoad
}

// CycleTimeForDemand computes the required cycle time to meet a given demand
// within a specified available time.
func CycleTimeForDemand(demand int, availableSec float64) float64 {
	if demand <= 0 {
		return 0
	}
	return availableSec / float64(demand)
}

// WIP estimates the work-in-process inventory using Little's law: WIP = throughput * lead time.
// Lead time is approximated as the number of stations times the cycle time.
func WIP(numStations int, cycleTime float64) float64 {
	if numStations <= 0 || cycleTime <= 0 {
		return 0
	}
	throughput := 1.0 / cycleTime
	leadTime := float64(numStations) * cycleTime
	return throughput * leadTime
}

// LeadTime returns the total time a unit spends in the line (number of stations
// times cycle time, assuming synchronous transfer).
func LeadTime(numStations int, cycleTime float64) float64 {
	return float64(numStations) * cycleTime
}

// OutputRate returns the number of units produced per shift given the
// bottleneck time and available seconds.
func OutputRate(bottleneckLoad, availableSec float64) float64 {
	if bottleneckLoad <= 0 || availableSec <= 0 {
		return 0
	}
	return math.Floor(availableSec / bottleneckLoad)
}

// LaborProductivity returns units produced per labor-hour. It accounts for the
// number of workers (one per station) and the shift duration.
func LaborProductivity(unitsPerShift float64, numStations int, shiftSec float64) float64 {
	if numStations <= 0 || shiftSec <= 0 {
		return 0
	}
	laborHours := float64(numStations) * (shiftSec / 3600.0)
	if laborHours <= 0 {
		return 0
	}
	return unitsPerShift / laborHours
}

// TaktDemandMatch checks whether the line can meet demand: returns the
// surplus capacity (positive = meets demand, negative = shortfall) in units.
func TaktDemandMatch(bottleneckLoad, availableSec float64, demand int) float64 {
	actual := OutputRate(bottleneckLoad, availableSec)
	return actual - float64(demand)
}
