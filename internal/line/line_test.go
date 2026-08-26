package line

import (
	"strings"
	"testing"

	"line-balancer/internal/task"
)

func TestTaktTime(t *testing.T) {
	if got := TaktTime(100, 28800); got != 288 {
		t.Fatalf("takt = %v, want 288", got)
	}
}

func TestTheoreticalMinStations(t *testing.T) {
	tasks := []Task{
		{Name: "A", Time: 10},
		{Name: "B", Time: 15},
		{Name: "C", Time: 20},
		{Name: "D", Time: 12},
	}
	min := TheoreticalMinStations(tasks, 20)
	if min != 3 {
		t.Fatalf("min stations = %d, want 3", min)
	}
}

func TestAnalyzePlacesAllTasks(t *testing.T) {
	tasks := []Task{
		{Name: "weld", Time: 45},
		{Name: "paint", Time: 30},
		{Name: "assemble", Time: 60},
		{Name: "inspect", Time: 20},
		{Name: "pack", Time: 25},
		{Name: "cut", Time: 40},
		{Name: "polish", Time: 35},
		{Name: "test", Time: 15},
	}
	res, err := Analyze(tasks, 400, 28800)
	if err != nil {
		t.Fatal(err)
	}
	if res.TaktTime != 72 {
		t.Fatalf("takt = %v, want 72", res.TaktTime)
	}
	if res.StationCount < 1 {
		t.Fatalf("expected at least 1 station, got %d", res.StationCount)
	}
	assigned := 0
	for _, s := range res.Stations {
		assigned += len(s.Tasks)
	}
	if assigned != len(tasks) {
		t.Fatalf("assigned %d tasks, want %d", assigned, len(tasks))
	}
	if res.MaxLoad <= 0 {
		t.Fatalf("max load = %v", res.MaxLoad)
	}
	if res.Efficiency <= 0 || res.Efficiency > 100 {
		t.Fatalf("efficiency = %v, want (0,100]", res.Efficiency)
	}
}

func TestAnalyzeBadDemand(t *testing.T) {
	if _, err := Analyze([]Task{{Name: "a", Time: 1}}, 0, 100); err == nil {
		t.Fatal("expected error for zero demand")
	}
}

func TestParseTasks(t *testing.T) {
	csv := "task,seconds\nweld,45\npaint,30\n"
	tasks, err := ParseTasks(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 || tasks[0].Name != "weld" || tasks[0].Time != 45 {
		t.Fatalf("parsed %+v", tasks)
	}
}

func makeTestTaskGraph() *task.Graph {
	g := task.NewGraph()
	g.Add("A", 5)
	g.Add("B", 3)
	g.Add("C", 6)
	g.Add("D", 7)
	g.Add("E", 4)
	g.Add("F", 2)
	g.Add("G", 5)
	g.Add("H", 3)
	g.Precede("A", "B")
	g.Precede("A", "C")
	g.Precede("B", "E")
	g.Precede("B", "D")
	g.Precede("C", "D")
	g.Precede("C", "F")
	g.Precede("D", "G")
	g.Precede("E", "H")
	g.Precede("F", "H")
	g.Precede("G", "H")
	return g
}

func TestRPWBalance(t *testing.T) {
	g := makeTestTaskGraph()
	stations, err := RPWBalance(g, 18)
	if err != nil {
		t.Fatal(err)
	}
	if len(stations) < 2 {
		t.Fatalf("expected >= 2 stations, got %d", len(stations))
	}
	assigned := 0
	for _, s := range stations {
		assigned += len(s.Tasks)
	}
	if assigned != 8 {
		t.Fatalf("assigned %d, want 8", assigned)
	}
	for i, s := range stations {
		if s.Load > 18+0.01 {
			t.Fatalf("station %d overloaded: %v > 18", i, s.Load)
		}
	}
}

func TestKWBalance(t *testing.T) {
	g := makeTestTaskGraph()
	stations, err := KWBalance(g, 18)
	if err != nil {
		t.Fatal(err)
	}
	assigned := 0
	for _, s := range stations {
		assigned += len(s.Tasks)
	}
	if assigned != 8 {
		t.Fatalf("assigned %d, want 8", assigned)
	}
}

func TestMoodieYoungPhase1(t *testing.T) {
	g := makeTestTaskGraph()
	stations, err := MoodieYoungPhase1(g, 18)
	if err != nil {
		t.Fatal(err)
	}
	assigned := 0
	for _, s := range stations {
		assigned += len(s.Tasks)
	}
	if assigned != 8 {
		t.Fatalf("assigned %d, want 8", assigned)
	}
}

func TestMoodieYoungPhase2Improves(t *testing.T) {
	g := makeTestTaskGraph()
	stations, _ := MoodieYoungPhase1(g, 18)
	maxBefore := 0.0
	for _, s := range stations {
		if s.Load > maxBefore {
			maxBefore = s.Load
		}
	}
	improved := MoodieYoungPhase2(g, stations, 18, 50)
	maxAfter := 0.0
	for _, s := range improved {
		if s.Load > maxAfter {
			maxAfter = s.Load
		}
	}
	if maxAfter > maxBefore+0.01 {
		t.Fatalf("phase 2 worsened: before max=%v, after max=%v", maxBefore, maxAfter)
	}
}

func TestDualBalance(t *testing.T) {
	g := makeTestTaskGraph()
	prefs := []DualPreference{
		{TaskID: "C", Side: SideRight},
		{TaskID: "F", Side: SideRight},
	}
	stations, err := DualBalance(g, 18, prefs)
	if err != nil {
		t.Fatal(err)
	}
	assigned := 0
	for _, ds := range stations {
		assigned += len(ds.Left.Tasks) + len(ds.Right.Tasks)
	}
	if assigned != 8 {
		t.Fatalf("assigned %d, want 8", assigned)
	}
}

func TestRPWAnalyze(t *testing.T) {
	g := makeTestTaskGraph()
	res, err := RPWAnalyze(g, 400, 28800)
	if err != nil {
		t.Fatal(err)
	}
	if res.TaktTime != 72 {
		t.Fatalf("takt = %v, want 72", res.TaktTime)
	}
	if res.StationCount != 1 {
		t.Fatalf("stations = %d, want 1 (total 35 < takt 72)", res.StationCount)
	}
}
