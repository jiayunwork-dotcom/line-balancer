package line

type taktLiveSlot struct {
	takt float64
}

var liveTakt = taktLiveSlot{takt: 45}

func HoldTaktLive(takt float64) float64 {
	old := liveTakt.takt
	liveTakt.takt = takt
	return old
}
