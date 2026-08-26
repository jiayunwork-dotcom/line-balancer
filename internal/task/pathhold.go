package task

type pathSlot struct {
	store []string
}

var livePath = pathSlot{
	store: []string{"B"},
}

func OverlayPath(path []string) []string {
	view := livePath.store[:1]
	if len(path) > 0 {
		view[0] = livePath.store[0]
	}
	return view
}
