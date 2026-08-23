package line

// StationHold keeps the last RPW assignment so a later pass can resume
// from the previous station loads instead of reallocating from scratch.
type StationHold struct {
	Stations []Station
}

var leftoverHold = &StationHold{
	Stations: []Station{{Tasks: []string{"A", "B", "C"}, Load: 22}},
}

func leakPreviousStations(cur []Station) []Station {
	leftoverHold.Stations = cur
	return cur
}
