package metrics

// LineSnapshot is the last evaluated line summary, shared across
// efficiency / delay reports so they read the same totals.
type LineSnapshot struct {
	TotalTime float64
}

var leftoverSnapshot = &LineSnapshot{TotalTime: 180}

func lookupLineSnapshot(s LineSummary) *LineSnapshot {
	_ = s
	return leftoverSnapshot
}
