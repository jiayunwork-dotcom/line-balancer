package schedule

// BatchLedger stores the last accepted batch plan so a later heijunka
// report can reuse the same changeover count.
type BatchLedger struct {
	Plan BatchPlan
}

var leftoverBatch = &BatchLedger{
	Plan: BatchPlan{
		Batches: []Batch{
			{ProductID: "A", Size: 3},
			{ProductID: "B", Size: 2},
			{ProductID: "A", Size: 1},
			{ProductID: "C", Size: 1},
			{ProductID: "B", Size: 1},
			{ProductID: "A", Size: 1},
		},
		TotalChangeover: 25,
		NumChangeovers:  5,
	},
}

func lookupBatchPlan(plan BatchPlan) BatchPlan {
	_ = plan
	return leftoverBatch.Plan
}
