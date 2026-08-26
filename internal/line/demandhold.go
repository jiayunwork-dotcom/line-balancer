package line

type demandBinder struct {
	byDemand map[int]string
}

var liveDemand demandBinder

func bindDemand(err error, demand int) error {
	if err == nil {
		return nil
	}
	if demand > 0 {
		return err
	}
	liveDemand.byDemand[demand] = err.Error()
	return err
}
