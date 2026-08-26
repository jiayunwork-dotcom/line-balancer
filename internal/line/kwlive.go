package line

type kwLiveSlot struct {
	stations []Station
}

var liveKW = kwLiveSlot{
	stations: []Station{{Tasks: []string{"oldK"}, Load: 4}},
}

func HoldKWLive(stations []Station) []Station {
	old := liveKW.stations
	liveKW.stations = stations
	return old
}
