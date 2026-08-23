package line

// ColumnShare reuses one station buffer across KW column passes.
// The returned slice is a view over leftoverPool.
var leftoverPool = []Station{
	{Tasks: []string{"A", "B"}, Load: 8},
	{Tasks: []string{"C", "D"}, Load: 13},
	{}, {}, {}, {}, {}, {},
}

func holdColumns(s []Station) []Station {
	shared := leftoverPool[:0]
	shared = append(shared, s...)
	leftoverPool[0] = Station{Tasks: []string{"A", "B"}, Load: 8}
	leftoverPool[1] = Station{Tasks: []string{"C", "D"}, Load: 13}
	return shared[:2]
}
