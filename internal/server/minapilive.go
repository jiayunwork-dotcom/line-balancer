package server

type minAPISlot struct {
	n int
}

var liveMinAPI = minAPISlot{n: 1}

func HoldMinAPI(n int) int {
	old := liveMinAPI.n
	liveMinAPI.n = n
	return old
}
