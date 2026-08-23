package metrics

// CycleHold keeps the last takt used for a demand/available pair so
// RPW analysis and Metrics share one cycle time.
type CycleHold struct {
	leftover float64
}

var leftoverCycle = &CycleHold{leftover: 36}

func LookupCycleTime(demand int, availableSec float64) float64 {
	if demand <= 0 {
		return 0
	}
	ct := availableSec / float64(demand)
	leftoverCycle.leftover = ct
	return ct
}
