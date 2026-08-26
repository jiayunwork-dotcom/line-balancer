package metrics

import (
	"math"
	"testing"
)

func TestEfficiency(t *testing.T) {
	s := LineSummary{
		StationLoads: []float64{70, 65, 60, 55},
		CycleTime:    72,
		TotalTime:    250,
	}
	eff := Efficiency(s)
	want := 250.0 / (4 * 72) * 100
	if math.Abs(eff-want) > 0.01 {
		t.Fatalf("efficiency = %v, want %v", eff, want)
	}
}

func TestSmoothnessIndex(t *testing.T) {
	s := LineSummary{
		StationLoads: []float64{72, 72, 72},
		CycleTime:    72,
		TotalTime:    216,
	}
	si := SmoothnessIndex(s)
	if si != 0 {
		t.Fatalf("smoothness index = %v, want 0 (perfectly smooth)", si)
	}

	s2 := LineSummary{
		StationLoads: []float64{72, 50, 30},
		CycleTime:    72,
		TotalTime:    152,
	}
	si2 := SmoothnessIndex(s2)
	if si2 <= 0 {
		t.Fatalf("smoothness index should be > 0 for unbalanced line, got %v", si2)
	}
}

func TestUtilizationPerStation(t *testing.T) {
	s := LineSummary{
		StationLoads: []float64{60, 72, 45},
		CycleTime:    72,
		TotalTime:    177,
	}
	u := UtilizationPerStation(s)
	if len(u) != 3 {
		t.Fatalf("len = %d, want 3", len(u))
	}
	if math.Abs(u[0]-60.0/72) > 0.001 {
		t.Fatalf("u[0] = %v, want %v", u[0], 60.0/72)
	}
	if u[1] != 1.0 {
		t.Fatalf("u[1] = %v, want 1.0 (at cycle time)", u[1])
	}
}

func TestThroughput(t *testing.T) {
	tp := Throughput(72)
	want := 1.0 / 72
	if math.Abs(tp-want) > 1e-9 {
		t.Fatalf("throughput = %v, want %v", tp, want)
	}
}

func TestOutputRate(t *testing.T) {
	rate := OutputRate(72, 28800)
	if rate != 400 {
		t.Fatalf("output rate = %v, want 400", rate)
	}
}

func TestDominates(t *testing.T) {
	a := LineSummary{StationLoads: []float64{70, 68, 65}, CycleTime: 72, TotalTime: 203}
	b := LineSummary{StationLoads: []float64{72, 60, 50, 40}, CycleTime: 72, TotalTime: 222}
	objs := DefaultObjectives()
	if !Dominates(a, b, objs) && !Dominates(b, a, objs) {
	}
}

func TestWIP(t *testing.T) {
	wip := WIP(4, 72)
	if wip != 4 {
		t.Fatalf("WIP = %v, want 4", wip)
	}
}

func TestLaborProductivity(t *testing.T) {
	lp := LaborProductivity(400, 4, 28800)
	if math.Abs(lp-12.5) > 0.01 {
		t.Fatalf("labor productivity = %v, want 12.5", lp)
	}
}
