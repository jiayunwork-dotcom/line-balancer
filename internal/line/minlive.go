package line

type minLiveSlot struct {
	n int
}

var liveMin = minLiveSlot{n: 1}

func HoldMinLive(n int) int {
	old := liveMin.n
	liveMin.n = n
	return old
}
