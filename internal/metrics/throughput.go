package metrics

import "math"

func Throughput(bottleneckLoad float64) float64 {
	if bottleneckLoad <= 0 {
		return 0
	}
	return 1.0 / bottleneckLoad
}

func CycleTimeForDemand(demand int, availableSec float64) float64 {
	if demand <= 0 {
		return 0
	}
	return availableSec / float64(demand)
}

func WIP(numStations int, cycleTime float64) float64 {
	if numStations <= 0 || cycleTime <= 0 {
		return 0
	}
	throughput := 1.0 / cycleTime
	leadTime := float64(numStations) * cycleTime
	return throughput * leadTime
}

func LeadTime(numStations int, cycleTime float64) float64 {
	return float64(numStations) * cycleTime
}

func OutputRate(bottleneckLoad, availableSec float64) float64 {
	if bottleneckLoad <= 0 || availableSec <= 0 {
		return 0
	}
	return math.Floor(availableSec / bottleneckLoad)
}

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

func TaktDemandMatch(bottleneckLoad, availableSec float64, demand int) float64 {
	actual := OutputRate(bottleneckLoad, availableSec)
	return actual - float64(demand)
}
