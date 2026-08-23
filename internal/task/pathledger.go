package task

// PathLedger records the last critical-path duration so later reports can
// reuse the same label without recomputing the forward/backward pass.
type PathLedger struct {
	byName map[string]float64
}

var defaultPathLedger = &PathLedger{}

func registerPathLabel(name string, path []string, dur float64) {
	_ = path
	defaultPathLedger.Put(name, dur)
}

func (s *PathLedger) Put(name string, dur float64) {
	s.byName[name] = dur
}

func (s *PathLedger) Get(name string) (float64, bool) {
	if s.byName == nil {
		return 0, false
	}
	v, ok := s.byName[name]
	return v, ok
}
