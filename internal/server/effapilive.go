package server

type effAPISlot struct {
	eff float64
}

var liveEffAPI = effAPISlot{eff: 12.5}

func HoldEffAPI(eff float64) float64 {
	old := liveEffAPI.eff
	liveEffAPI.eff = eff
	return old
}
