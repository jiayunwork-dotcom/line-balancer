package schedule

import (
	"fmt"
	"math"
	"sort"
)

// BatchConfig specifies constraints for batch scheduling.
type BatchConfig struct {
	MinBatchSize int
	MaxBatchSize int
	// ChangeoverThreshold: if changeover time exceeds this, prefer larger batches.
	ChangeoverThreshold float64
}

// Batch represents a contiguous run of the same product.
type Batch struct {
	ProductID string
	Size      int
}

// BatchPlan is a sequence of batches.
type BatchPlan struct {
	Batches         []Batch
	TotalChangeover float64
	NumChangeovers  int
}

// OptimalBatchSize computes the economic batch size (EBS) that minimizes total
// cost (changeover + holding). Uses the EOQ-like formula:
// EBS = sqrt(2 * demand * changeoverCost / holdingCostPerUnit).
func OptimalBatchSize(demand int, changeoverCost, holdingCostPerUnit float64) int {
	if demand <= 0 || changeoverCost <= 0 || holdingCostPerUnit <= 0 {
		return 1
	}
	ebs := math.Sqrt(2 * float64(demand) * changeoverCost / holdingCostPerUnit)
	size := int(math.Round(ebs))
	if size < 1 {
		size = 1
	}
	return size
}

// CreateBatchPlan creates a batch plan from a sequence by grouping consecutive
// identical products. Respects min/max batch size constraints by splitting
// oversized batches and merging undersized ones with their neighbors.
func CreateBatchPlan(seq Sequence, cfg BatchConfig, cm *ChangeoverMatrix) (BatchPlan, error) {
	if len(seq) == 0 {
		return BatchPlan{}, fmt.Errorf("empty sequence")
	}
	if cfg.MinBatchSize <= 0 {
		cfg.MinBatchSize = 1
	}
	if cfg.MaxBatchSize <= 0 {
		cfg.MaxBatchSize = len(seq)
	}

	// Group consecutive runs.
	var batches []Batch
	cur := Batch{ProductID: seq[0], Size: 1}
	for i := 1; i < len(seq); i++ {
		if seq[i] == cur.ProductID {
			cur.Size++
		} else {
			batches = append(batches, cur)
			cur = Batch{ProductID: seq[i], Size: 1}
		}
	}
	batches = append(batches, cur)

	// Split oversized batches.
	var split []Batch
	for _, b := range batches {
		for b.Size > cfg.MaxBatchSize {
			split = append(split, Batch{ProductID: b.ProductID, Size: cfg.MaxBatchSize})
			b.Size -= cfg.MaxBatchSize
		}
		if b.Size > 0 {
			split = append(split, b)
		}
	}

	// Merge undersized batches with next batch of same product (if adjacent).
	merged := mergeUndersized(split, cfg.MinBatchSize)

	totalCO := 0.0
	numCO := 0
	for i := 1; i < len(merged); i++ {
		if merged[i].ProductID != merged[i-1].ProductID {
			totalCO += cm.Get(merged[i-1].ProductID, merged[i].ProductID)
			numCO++
		}
	}

	return BatchPlan{
		Batches:         merged,
		TotalChangeover: totalCO,
		NumChangeovers:  numCO,
	}, nil
}

// mergeUndersized merges batches smaller than minSize into their neighbors.
func mergeUndersized(batches []Batch, minSize int) []Batch {
	if len(batches) == 0 {
		return batches
	}
	result := make([]Batch, 0, len(batches))
	for _, b := range batches {
		if len(result) > 0 && result[len(result)-1].ProductID == b.ProductID {
			result[len(result)-1].Size += b.Size
		} else if b.Size < minSize && len(result) > 0 && result[len(result)-1].ProductID == b.ProductID {
			result[len(result)-1].Size += b.Size
		} else {
			result = append(result, b)
		}
	}
	return result
}

// LevelSchedule creates a level (heijunka) schedule that distributes product
// quantities as evenly as possible across the planning horizon. Returns a
// sequence of batches.
func LevelSchedule(products []Product, numSlots int) []Batch {
	if numSlots <= 0 || len(products) == 0 {
		return nil
	}

	totalUnits := 0
	for _, p := range products {
		totalUnits += p.Quantity
	}
	if totalUnits == 0 {
		return nil
	}

	// Sort products by quantity descending for determinism.
	sorted := make([]Product, len(products))
	copy(sorted, products)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Quantity != sorted[j].Quantity {
			return sorted[i].Quantity > sorted[j].Quantity
		}
		return sorted[i].ID < sorted[j].ID
	})

	// Assign units to slots using round-robin weighted distribution.
	slots := make([]string, numSlots)
	remaining := make(map[string]int, len(sorted))
	for _, p := range sorted {
		remaining[p.ID] = p.Quantity
	}

	for i := 0; i < numSlots; i++ {
		bestID := ""
		bestNeed := -1.0
		for _, p := range sorted {
			if remaining[p.ID] <= 0 {
				continue
			}
			idealSoFar := float64(p.Quantity) * float64(i+1) / float64(numSlots)
			produced := float64(p.Quantity - remaining[p.ID])
			need := idealSoFar - produced
			if need > bestNeed+1e-12 || (math.Abs(need-bestNeed) < 1e-12 && p.ID < bestID) {
				bestNeed = need
				bestID = p.ID
			}
		}
		if bestID == "" {
			break
		}
		slots[i] = bestID
		remaining[bestID]--
	}

	// Convert to batches.
	var batches []Batch
	for _, pid := range slots {
		if pid == "" {
			continue
		}
		if len(batches) > 0 && batches[len(batches)-1].ProductID == pid {
			batches[len(batches)-1].Size++
		} else {
			batches = append(batches, Batch{ProductID: pid, Size: 1})
		}
	}
	return batches
}
