package schedule

import "sort"

// ChangeoverOptimizer finds a good production sequence that minimizes total
// changeover time using a nearest-neighbor heuristic.
type ChangeoverOptimizer struct {
	Products []Product
	Matrix   *ChangeoverMatrix
}

// NearestNeighbor builds a sequence starting from the product with the highest
// quantity, always choosing the next product with the lowest changeover time
// from the current product.
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

	// Start with the product that has the highest quantity.
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
		// If current product still has quantity, continue with it (zero changeover).
		if remaining[current] > 0 {
			seq = append(seq, current)
			remaining[current]--
			continue
		}
		// Find nearest neighbor.
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

// TwoOptImprove performs a 2-opt local search on a batch-level sequence to
// reduce total changeover time. It operates on the batch boundaries (product
// transitions) and attempts to reverse segments that reduce changeover cost.
func (co *ChangeoverOptimizer) TwoOptImprove(seq Sequence, maxIter int) Sequence {
	if len(seq) <= 3 || maxIter <= 0 {
		return seq
	}

	// Extract transition points (indices where product changes).
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
				// Evaluate cost before and after reversing segment between transitions[i] and transitions[j].
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

				// Cost after reversing [start+1, end].
				costAfter := co.Matrix.Get(seq[start], seq[end])
				if end+1 < len(seq) {
					costAfter += co.Matrix.Get(seq[start+1], seq[end+1])
				}

				if costAfter < costBefore-1e-9 {
					// Reverse the segment.
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

// ChangeoverCount returns the number of product transitions in a sequence.
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
