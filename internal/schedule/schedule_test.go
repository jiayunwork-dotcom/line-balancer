package schedule

import (
	"math"
	"testing"
)

func TestExpandSequenceMiltenburg(t *testing.T) {
	products := []Product{
		{ID: "A", Quantity: 4},
		{ID: "B", Quantity: 2},
		{ID: "C", Quantity: 1},
	}
	seq := ExpandSequence(products)
	if len(seq) != 7 {
		t.Fatalf("len(seq) = %d, want 7", len(seq))
	}
	counts := make(map[string]int)
	for _, id := range seq {
		counts[id]++
	}
	if counts["A"] != 4 || counts["B"] != 2 || counts["C"] != 1 {
		t.Fatalf("counts = %v, want A=4 B=2 C=1", counts)
	}
}

func TestTotalChangeover(t *testing.T) {
	cm := NewChangeoverMatrix()
	cm.Set("A", "B", 10)
	cm.Set("B", "A", 8)
	cm.Set("A", "C", 15)

	seq := Sequence{"A", "A", "B", "A", "C"}
	total := TotalChangeover(seq, cm)
	if math.Abs(total-33) > 0.01 {
		t.Fatalf("total changeover = %v, want 33", total)
	}
}

func TestCreateBatchPlan(t *testing.T) {
	cm := NewChangeoverMatrix()
	cm.Set("A", "B", 5)
	cm.Set("B", "A", 5)

	seq := Sequence{"A", "A", "A", "B", "B", "A", "A"}
	cfg := BatchConfig{MinBatchSize: 1, MaxBatchSize: 4}
	plan, err := CreateBatchPlan(seq, cfg, cm)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Batches) < 2 {
		t.Fatalf("expected at least 2 batches, got %d", len(plan.Batches))
	}
	if plan.NumChangeovers != 2 {
		t.Fatalf("num changeovers = %d, want 2", plan.NumChangeovers)
	}
}

func TestOptimalBatchSize(t *testing.T) {
	ebs := OptimalBatchSize(1000, 50, 2)
	if ebs < 200 || ebs > 250 {
		t.Fatalf("EBS = %d, want ~224", ebs)
	}
}

func TestNearestNeighbor(t *testing.T) {
	cm := NewChangeoverMatrix()
	cm.Set("A", "B", 10)
	cm.Set("A", "C", 5)
	cm.Set("B", "C", 3)
	cm.Set("B", "A", 10)
	cm.Set("C", "A", 5)
	cm.Set("C", "B", 3)

	products := []Product{
		{ID: "A", Quantity: 3},
		{ID: "B", Quantity: 2},
		{ID: "C", Quantity: 2},
	}
	opt := &ChangeoverOptimizer{Products: products, Matrix: cm}
	seq := opt.NearestNeighbor()
	if len(seq) != 7 {
		t.Fatalf("len(seq) = %d, want 7", len(seq))
	}
	if seq[0] != "A" {
		t.Fatalf("first product = %q, want A", seq[0])
	}
}

func TestLevelSchedule(t *testing.T) {
	products := []Product{
		{ID: "A", Quantity: 6},
		{ID: "B", Quantity: 3},
		{ID: "C", Quantity: 1},
	}
	batches := LevelSchedule(products, 10)
	total := 0
	for _, b := range batches {
		total += b.Size
	}
	if total != 10 {
		t.Fatalf("total units = %d, want 10", total)
	}
}

func TestChangeoverCount(t *testing.T) {
	seq := Sequence{"A", "A", "B", "B", "B", "C", "A"}
	count := ChangeoverCount(seq)
	if count != 3 {
		t.Fatalf("changeover count = %d, want 3", count)
	}
}

func TestEvaluateSequence(t *testing.T) {
	products := []Product{
		{ID: "A", Quantity: 2, StationTimes: map[int]float64{0: 10, 1: 8, 2: 12}},
		{ID: "B", Quantity: 1, StationTimes: map[int]float64{0: 15, 1: 10, 2: 9}},
	}
	cm := NewChangeoverMatrix()
	cm.Set("A", "B", 5)
	cm.Set("B", "A", 5)

	seq := Sequence{"A", "B", "A"}
	res, err := EvaluateSequence(seq, products, 3, 20, cm)
	if err != nil {
		t.Fatal(err)
	}
	if res.MakeSpan <= 0 {
		t.Fatalf("makespan = %v, want > 0", res.MakeSpan)
	}
}
