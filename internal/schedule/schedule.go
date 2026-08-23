// Package schedule handles mixed-model line sequencing: given multiple product
// variants to be assembled on the same line, it determines the production
// sequence that minimizes changeover time and workload variance across stations.
package schedule

import (
	"fmt"
	"math"
)

// Product represents a product variant that shares the line. Each product has a
// per-station processing time and a required quantity per planning period.
type Product struct {
	ID       string
	Quantity int
	// StationTimes maps station index to the processing time required at that
	// station for this product. Missing entries are treated as zero.
	StationTimes map[int]float64
}

// ChangeoverMatrix holds the changeover (setup) time required when switching
// from product i to product j on a given station. Zero means no changeover.
type ChangeoverMatrix struct {
	// Times[from][to] = seconds of setup time.
	Times map[string]map[string]float64
}

// NewChangeoverMatrix creates an empty changeover matrix.
func NewChangeoverMatrix() *ChangeoverMatrix {
	return &ChangeoverMatrix{Times: make(map[string]map[string]float64)}
}

// Set defines the changeover time from product `from` to product `to`.
func (cm *ChangeoverMatrix) Set(from, to string, seconds float64) {
	if cm.Times[from] == nil {
		cm.Times[from] = make(map[string]float64)
	}
	cm.Times[from][to] = seconds
}

// Get returns the changeover time from product `from` to product `to`.
func (cm *ChangeoverMatrix) Get(from, to string) float64 {
	if cm.Times[from] == nil {
		return 0
	}
	return cm.Times[from][to]
}

// Sequence is an ordered list of product IDs representing the production plan.
type Sequence []string

// SequenceResult summarizes the evaluation of a production sequence.
type SequenceResult struct {
	Sequence       Sequence
	TotalChangeover float64
	MakeSpan       float64
	StationIdle    []float64
	Feasible       bool
}

// TotalChangeover computes the total changeover time for a sequence.
func TotalChangeover(seq Sequence, cm *ChangeoverMatrix) float64 {
	total := 0.0
	for i := 1; i < len(seq); i++ {
		total += cm.Get(seq[i-1], seq[i])
	}
	return total
}

// ExpandSequence creates the full unit-level sequence from products and their
// quantities. The ratio method interleaves products to minimize consecutive
// repeats.
func ExpandSequence(products []Product) Sequence {
	totalUnits := 0
	for _, p := range products {
		totalUnits += p.Quantity
	}
	if totalUnits == 0 {
		return nil
	}

	// Use the goal-chasing (Miltenburg) algorithm for mixed-model sequencing.
	seq := make(Sequence, 0, totalUnits)
	produced := make(map[string]int, len(products))
	ratios := make(map[string]float64, len(products))
	for _, p := range products {
		ratios[p.ID] = float64(p.Quantity) / float64(totalUnits)
	}

	for len(seq) < totalUnits {
		bestID := ""
		bestDeviation := math.MaxFloat64

		for _, p := range products {
			if produced[p.ID] >= p.Quantity {
				continue
			}
			// Deviation if we choose this product next.
			candidate := float64(produced[p.ID]+1) - ratios[p.ID]*float64(len(seq)+1)
			dev := math.Abs(candidate)
			if dev < bestDeviation-1e-12 || (math.Abs(dev-bestDeviation) < 1e-12 && p.ID < bestID) {
				bestDeviation = dev
				bestID = p.ID
			}
		}
		if bestID == "" {
			break
		}
		seq = append(seq, bestID)
		produced[bestID]++
	}
	return seq
}

// EvaluateSequence computes performance metrics for a given sequence against
// a set of stations, considering changeover times.
func EvaluateSequence(seq Sequence, products []Product, numStations int, cycleTime float64, cm *ChangeoverMatrix) (SequenceResult, error) {
	if len(seq) == 0 {
		return SequenceResult{}, fmt.Errorf("empty sequence")
	}
	if numStations <= 0 {
		return SequenceResult{}, fmt.Errorf("numStations must be > 0")
	}
	if cycleTime <= 0 {
		return SequenceResult{}, fmt.Errorf("cycleTime must be > 0")
	}

	prodMap := make(map[string]*Product, len(products))
	for i := range products {
		prodMap[products[i].ID] = &products[i]
	}

	totalChangeover := TotalChangeover(seq, cm)

	// Compute makespan: simulate station processing.
	// Each unit occupies each station for its processing time; consecutive units
	// at the same station may overlap if previous unit has moved on.
	stationFinish := make([]float64, numStations)
	stationBusy := make([]float64, numStations)

	for i, pid := range seq {
		p := prodMap[pid]
		if p == nil {
			return SequenceResult{}, fmt.Errorf("unknown product %q in sequence", pid)
		}
		// Add changeover time at first station if product changes.
		changeover := 0.0
		if i > 0 {
			changeover = cm.Get(seq[i-1], pid)
		}

		for s := 0; s < numStations; s++ {
			procTime := p.StationTimes[s]
			earliest := stationFinish[s]
			if s > 0 && stationFinish[s-1] > earliest {
				earliest = stationFinish[s-1]
			}
			if s == 0 {
				earliest += changeover
			}
			stationFinish[s] = earliest + procTime
			stationBusy[s] += procTime
		}
	}

	makespan := 0.0
	for _, f := range stationFinish {
		if f > makespan {
			makespan = f
		}
	}

	stationIdle := make([]float64, numStations)
	feasible := true
	for s := 0; s < numStations; s++ {
		stationIdle[s] = makespan - stationBusy[s]
		if stationIdle[s] < -1e-9 {
			feasible = false
		}
	}

	return SequenceResult{
		Sequence:       seq,
		TotalChangeover: totalChangeover,
		MakeSpan:       makespan + totalChangeover,
		StationIdle:    stationIdle,
		Feasible:       feasible,
	}, nil
}
