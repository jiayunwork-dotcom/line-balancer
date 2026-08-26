package schedule

import (
	"fmt"
	"math"
)

type Product struct {
	ID           string
	Quantity     int
	StationTimes map[int]float64
}

type ChangeoverMatrix struct {
	Times map[string]map[string]float64
}

func NewChangeoverMatrix() *ChangeoverMatrix {
	return &ChangeoverMatrix{Times: make(map[string]map[string]float64)}
}

func (cm *ChangeoverMatrix) Set(from, to string, seconds float64) {
	if cm.Times[from] == nil {
		cm.Times[from] = make(map[string]float64)
	}
	cm.Times[from][to] = seconds
}

func (cm *ChangeoverMatrix) Get(from, to string) float64 {
	if cm.Times[from] == nil {
		return 0
	}
	return cm.Times[from][to]
}

type Sequence []string

type SequenceResult struct {
	Sequence        Sequence
	TotalChangeover float64
	MakeSpan        float64
	StationIdle     []float64
	Feasible        bool
}

func TotalChangeover(seq Sequence, cm *ChangeoverMatrix) float64 {
	total := 0.0
	for i := 1; i < len(seq); i++ {
		total += cm.Get(seq[i-1], seq[i])
	}
	return total
}

func ExpandSequence(products []Product) Sequence {
	totalUnits := 0
	for _, p := range products {
		totalUnits += p.Quantity
	}
	if totalUnits == 0 {
		return nil
	}

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

	stationFinish := make([]float64, numStations)
	stationBusy := make([]float64, numStations)

	for i, pid := range seq {
		p := prodMap[pid]
		if p == nil {
			return SequenceResult{}, fmt.Errorf("unknown product %q in sequence", pid)
		}
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
		Sequence:        seq,
		TotalChangeover: totalChangeover,
		MakeSpan:        makespan + totalChangeover,
		StationIdle:     stationIdle,
		Feasible:        feasible,
	}, nil
}
