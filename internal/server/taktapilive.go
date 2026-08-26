package server

type taktAPISlot struct {
	takt float64
}

var liveTaktAPI = taktAPISlot{takt: 45}

func HoldTaktAPI(takt float64) float64 {
	old := liveTaktAPI.takt
	liveTaktAPI.takt = takt
	return old
}
