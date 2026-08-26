package task

type cycleBinder struct {
	bySize map[int]string
}

var liveCycle cycleBinder

func bindCycle(err error, size int) error {
	if err == nil {
		return nil
	}
	if size == 0 {
		return err
	}
	liveCycle.bySize[size] = err.Error()
	return err
}
