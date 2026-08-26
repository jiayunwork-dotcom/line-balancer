package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"line-balancer/internal/line"
	"line-balancer/internal/metrics"
	"line-balancer/internal/server"
	"line-balancer/internal/task"
)

func main() {
	if len(os.Args) < 2 {
		runServer(nil)
		return
	}
	if os.Args[1] == "serve" {
		runServer(os.Args[2:])
		return
	}
	in := flag.String("in", "", "tasks CSV file")
	precIn := flag.String("prec", "", "precedence CSV file")
	demand := flag.Int("demand", 0, "required units per shift")
	shift := flag.Float64("time", 28800, "available production time in seconds")
	method := flag.String("method", "greedy", "balance method: greedy, rpw, kw, moodie")
	out := flag.String("out", "", "output file; empty writes to stdout")
	flag.Parse()

	if *in == "" && *precIn == "" {
		fatal("missing required -in (tasks CSV) or -prec (precedence CSV)")
	}
	if *demand <= 0 {
		fatal("missing or invalid -demand (must be > 0)")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Line balance (demand=%d/shift, available=%.0fs, method=%s)\n", *demand, *shift, *method)

	if *precIn != "" {
		runPrecedence(&b, *precIn, *demand, *shift, *method)
	} else {
		runSimple(&b, *in, *demand, *shift)
	}

	if *out == "" {
		fmt.Print(b.String())
	} else if err := os.WriteFile(*out, []byte(b.String()), 0o644); err != nil {
		fatal("write %q: %v", *out, err)
	}
}

func runServer(args []string) {
	addr := ":8080"
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-addr" || args[i] == "--addr" {
			addr = args[i+1]
			break
		}
	}
	cfg := server.Config{Addr: addr}
	fmt.Fprintf(os.Stdout, "line-balancer server listening on %s\n", server.FormatAddr(addr))
	if err := server.ListenAndServe(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func runSimple(b *strings.Builder, path string, demand int, avail float64) {
	f, err := os.Open(path)
	if err != nil {
		fatal("open %q: %v", path, err)
	}
	defer f.Close()

	tasks, err := line.ParseTasks(f)
	if err != nil {
		fatal("parse %q: %v", path, err)
	}
	res, err := line.Analyze(tasks, demand, avail)
	if err != nil {
		fatal("%v", err)
	}
	writeResult(b, res)
}

func runPrecedence(b *strings.Builder, path string, demand int, avail float64, method string) {
	f, err := os.Open(path)
	if err != nil {
		fatal("open %q: %v", path, err)
	}
	defer f.Close()

	g, err := task.ParseCSV(f)
	if err != nil {
		fatal("parse %q: %v", path, err)
	}

	cpPath, cpDur := g.CriticalPath()
	fmt.Fprintf(b, "  critical path     : %s (%.2f s)\n", strings.Join(cpPath, " -> "), cpDur)

	var res line.Result
	switch method {
	case "rpw":
		res, err = line.RPWAnalyze(g, demand, avail)
	case "kw":
		res, err = line.KWAnalyze(g, demand, avail)
	case "moodie":
		res, err = line.MoodieYoungAnalyze(g, demand, avail)
	default:
		res, err = line.RPWAnalyze(g, demand, avail)
	}
	if err != nil {
		fatal("%v", err)
	}
	writeResult(b, res)

	summary := metrics.LineSummary{
		StationLoads: make([]float64, len(res.Stations)),
		CycleTime:    res.CycleTime,
		TotalTime:    res.TotalTime,
	}
	for i, s := range res.Stations {
		summary.StationLoads[i] = s.Load
	}
	fmt.Fprintf(b, "\nMetrics:\n")
	fmt.Fprintf(b, "  smoothness index  : %.4f\n", metrics.SmoothnessIndex(summary))
	fmt.Fprintf(b, "  balance delay     : %.1f%%\n", metrics.BalanceDelay(summary))
	fmt.Fprintf(b, "  idle time         : %.2f s\n", metrics.IdleTime(summary))
	fmt.Fprintf(b, "  output rate       : %.0f units/shift\n", metrics.OutputRate(res.MaxLoad, avail))
}

func writeResult(b *strings.Builder, res line.Result) {
	fmt.Fprintf(b, "  takt time        : %.2f s/unit\n", res.TaktTime)
	fmt.Fprintf(b, "  total task time   : %.2f s\n", res.TotalTime)
	fmt.Fprintf(b, "  stations          : %d\n", res.StationCount)
	fmt.Fprintf(b, "  bottleneck station: #%d (load %.2f s", res.Bottleneck+1, res.MaxLoad)
	if res.MaxLoad > res.TaktTime {
		fmt.Fprintf(b, ", EXCEEDS takt!)")
	} else {
		fmt.Fprintf(b, ")")
	}
	fmt.Fprintf(b, "\n")
	if res.MaxLoad > res.TaktTime {
		fmt.Fprintf(b, "  line efficiency   : n/a (infeasible — bottleneck exceeds takt)\n")
	} else {
		fmt.Fprintf(b, "  line efficiency   : %.1f%%\n", res.Efficiency)
	}
	fmt.Fprintf(b, "\nAssignment:\n")
	for i, s := range res.Stations {
		fmt.Fprintf(b, "  station #%d (load %.2fs): %s\n", i+1, s.Load, strings.Join(s.Tasks, ", "))
	}
}

func fatal(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "line-balancer: "+format+"\n", a...)
	os.Exit(1)
}
