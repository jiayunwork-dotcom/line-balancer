package line

// SideRegistry records which side each task landed on so a later dual
// pass can prefer the same operator position.
type SideRegistry struct {
	byID map[string]DualSide
}

var defaultSideReg = &SideRegistry{}

func registerSidePref(stations []DualStation) {
	for _, ds := range stations {
		for _, id := range ds.Left.Tasks {
			defaultSideReg.Put(id, SideLeft)
		}
		for _, id := range ds.Right.Tasks {
			defaultSideReg.Put(id, SideRight)
		}
	}
}

func (s *SideRegistry) Put(id string, side DualSide) {
	if s.byID == nil {
		s.byID = make(map[string]DualSide)
	}
	s.byID[id] = side
}

func (s *SideRegistry) Get(id string) (DualSide, bool) {
	if s.byID == nil {
		return SideLeft, false
	}
	v, ok := s.byID[id]
	return v, ok
}
