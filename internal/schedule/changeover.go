package schedule

import "sort"

type ChangeoverOptimizer struct {
	Products []Product
	Matrix   *ChangeoverMatrix
}

func (co *ChangeoverOptimizer) NearestNeighbor() Sequence {
	if len(co.Products) == 0 {
		return nil
	}

	remaining := make(map[string]int, len(co.Products))
	totalUnits := 0
	for _, p := range co.Products {
		remaining[p.ID] = p.Quantity
		totalUnits += p.Quantity
	}

	sorted := make([]Product, len(co.Products))
	copy(sorted, co.Products)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Quantity != sorted[j].Quantity {
			return sorted[i].Quantity > sorted[j].Quantity
		}
		return sorted[i].ID < sorted[j].ID
	})

	seq := make(Sequence, 0, totalUnits)
	current := sorted[0].ID
	seq = append(seq, current)
	remaining[current]--

	for len(seq) < totalUnits {
		if remaining[current] > 0 {
			seq = append(seq, current)
			remaining[current]--
			continue
		}
		bestID := ""
		bestCO := -1.0
		for id, qty := range remaining {
			if qty <= 0 {
				continue
			}
			co := co.Matrix.Get(current, id)
			if bestID == "" || co < bestCO || (co == bestCO && id < bestID) {
				bestCO = co
				bestID = id
			}
		}
		if bestID == "" {
			break
		}
		current = bestID
		seq = append(seq, current)
		remaining[current]--
	}
	return seq
}

func (co *ChangeoverOptimizer) TwoOptImprove(seq Sequence, maxIter int) Sequence {
	if len(seq) <= 3 || maxIter <= 0 {
		return seq
	}

	var transitions []int
	for i := 1; i < len(seq); i++ {
		if seq[i] != seq[i-1] {
			transitions = append(transitions, i)
		}
	}
	if len(transitions) <= 2 {
		return seq
	}

	improved := true
	for iter := 0; iter < maxIter && improved; iter++ {
		improved = false
		for i := 0; i < len(transitions)-1; i++ {
			for j := i + 1; j < len(transitions); j++ {
				start := transitions[i] - 1
				end := transitions[j]
				if end >= len(seq) {
					end = len(seq) - 1
				}
				if start < 0 {
					start = 0
				}

				costBefore := co.Matrix.Get(seq[start], seq[start+1])
				if end+1 < len(seq) {
					costBefore += co.Matrix.Get(seq[end], seq[end+1])
				}

				costAfter := co.Matrix.Get(seq[start], seq[end])
				if end+1 < len(seq) {
					costAfter += co.Matrix.Get(seq[start+1], seq[end+1])
				}

				if costAfter < costBefore-1e-9 {
					for l, r := start+1, end; l < r; l, r = l+1, r-1 {
						seq[l], seq[r] = seq[r], seq[l]
					}
					improved = true
				}
			}
		}
	}
	return seq
}

func ChangeoverCount(seq Sequence) int {
	if len(seq) <= 1 {
		return 0
	}
	count := 0
	for i := 1; i < len(seq); i++ {
		if seq[i] != seq[i-1] {
			count++
		}
	}
	return count
}
