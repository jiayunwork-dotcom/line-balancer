package line

type countLiveSlot struct {
	n int
}

var liveCount = countLiveSlot{n: 0}

func HoldCountLive(n int) int {
	old := liveCount.n
	liveCount.n = n
	return old
}
