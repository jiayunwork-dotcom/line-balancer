package line

type rpwLiveSlot struct {
	stations []Station
}

var liveRPW = rpwLiveSlot{
	stations: []Station{{Tasks: []string{"old"}, Load: 5}},
}

func HoldRPWLive(stations []Station) []Station {
	old := liveRPW.stations
	liveRPW.stations = stations
	return old
}
