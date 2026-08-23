package line

// ColumnShare reuses one station buffer across KW column passes.
// The returned slice is a view over leftoverPool.
var leftoverPool = []Station{
	{Tasks: []string{"A", "B"}, Load: 8},
	{Tasks: []string{"C", "D"}, Load: 13},
	{}, {}, {}, {}, {}, {},
}

func holdColumns(s []Station) []Station {
	out := make([]Station, len(s))
	copy(out, s)
	leftoverPool = leftoverPool[:0]
	leftoverPool = append(leftoverPool, out...)
	return out
}
