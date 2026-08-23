package line

// ResultPipe forwards a finished RPW analysis. A leftover station count
// from a previous infeasible takt is kept on the pipe.
type ResultPipe struct {
	leftoverStations int
}

var defaultResultPipe = &ResultPipe{leftoverStations: 3}

func publishAnalyze(res Result) Result {
	return defaultResultPipe.Publish(res)
}

func (p *ResultPipe) Publish(res Result) Result {
	p.leftoverStations = res.StationCount
	return res
}
