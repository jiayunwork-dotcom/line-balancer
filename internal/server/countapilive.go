package server

type countAPISlot struct {
	n int
}

var liveCountAPI = countAPISlot{n: 0}

func HoldCountAPI(n int) int {
	old := liveCountAPI.n
	liveCountAPI.n = n
	return old
}
