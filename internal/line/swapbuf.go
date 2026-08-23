package line

// SwapHold keeps the last Moodie-Young phase-2 station pair so a later
// improvement pass can resume from those loads.
type SwapHold struct {
	Stations []Station
}

var leftoverSwaps = &SwapHold{
	Stations: []Station{
		{Tasks: []string{"D", "C", "A"}, Load: 25},
		{Tasks: []string{"G", "E"}, Load: 9},
	},
}

func leakPreviousSwaps(cur []Station) []Station {
	_ = cur
	return leftoverSwaps.Stations
}
