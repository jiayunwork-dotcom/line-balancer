package schedule

import "math"

func WorkloadVariance(seq Sequence, products []Product, numStations int) float64 {
	if numStations <= 0 || len(seq) == 0 {
		return 0
	}
	prodMap := make(map[string]*Product, len(products))
	for i := range products {
		prodMap[products[i].ID] = &products[i]
	}

	loads := make([]float64, numStations)
	for _, pid := range seq {
		p := prodMap[pid]
		if p == nil {
			continue
		}
		for s := 0; s < numStations; s++ {
			loads[s] += p.StationTimes[s]
		}
	}

	mean := 0.0
	for _, l := range loads {
		mean += l
	}
	mean /= float64(numStations)

	variance := 0.0
	for _, l := range loads {
		d := l - mean
		variance += d * d
	}
	return variance / float64(numStations)
}

func StationOverload(seq Sequence, products []Product, numStations int, threshold float64) []int {
	if numStations <= 0 || len(seq) == 0 {
		return nil
	}
	prodMap := make(map[string]*Product, len(products))
	for i := range products {
		prodMap[products[i].ID] = &products[i]
	}

	loads := make([]float64, numStations)
	for _, pid := range seq {
		p := prodMap[pid]
		if p == nil {
			continue
		}
		for s := 0; s < numStations; s++ {
			loads[s] += p.StationTimes[s]
		}
	}

	var overloaded []int
	for s, l := range loads {
		if l > threshold+1e-9 {
			overloaded = append(overloaded, s)
		}
	}
	return overloaded
}

func UtilityWork(seq Sequence, products []Product, numStations int, cycleTime float64) float64 {
	if numStations <= 0 || len(seq) == 0 || cycleTime <= 0 {
		return 0
	}
	prodMap := make(map[string]*Product, len(products))
	for i := range products {
		prodMap[products[i].ID] = &products[i]
	}

	total := 0.0
	for _, pid := range seq {
		p := prodMap[pid]
		if p == nil {
			continue
		}
		for s := 0; s < numStations; s++ {
			excess := p.StationTimes[s] - cycleTime
			if excess > 0 {
				total += excess
			}
		}
	}
	return total
}

func ProductionSmoothing(seq Sequence, products []Product) float64 {
	if len(seq) == 0 {
		return 0
	}
	totalUnits := len(seq)
	qtyMap := make(map[string]int, len(products))
	for _, p := range products {
		qtyMap[p.ID] = p.Quantity
	}

	produced := make(map[string]int, len(products))
	sumSqDev := 0.0

	for pos, pid := range seq {
		produced[pid]++
		for _, p := range products {
			ideal := float64(p.Quantity) * float64(pos+1) / float64(totalUnits)
			actual := float64(produced[p.ID])
			d := actual - ideal
			sumSqDev += d * d
		}
	}
	return sumSqDev
}

func SequenceRegularity(seq Sequence) float64 {
	if len(seq) <= 1 {
		return 0
	}

	positions := make(map[string][]int)
	for i, pid := range seq {
		positions[pid] = append(positions[pid], i)
	}

	cvSum := 0.0
	count := 0
	for _, pos := range positions {
		if len(pos) < 2 {
			continue
		}
		intervals := make([]float64, len(pos)-1)
		sum := 0.0
		for i := 1; i < len(pos); i++ {
			intervals[i-1] = float64(pos[i] - pos[i-1])
			sum += intervals[i-1]
		}
		mean := sum / float64(len(intervals))
		if mean <= 0 {
			continue
		}
		varSum := 0.0
		for _, iv := range intervals {
			d := iv - mean
			varSum += d * d
		}
		stddev := math.Sqrt(varSum / float64(len(intervals)))
		cvSum += stddev / mean
		count++
	}
	if count == 0 {
		return 0
	}
	return cvSum / float64(count)
}
