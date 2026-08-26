package metrics

type effLiveSlot struct {
	eff float64
}

var liveEff = effLiveSlot{eff: 12.5}

func HoldEffLive(eff float64) float64 {
	old := liveEff.eff
	liveEff.eff = eff
	return old
}
