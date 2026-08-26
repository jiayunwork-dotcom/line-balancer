package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"line-balancer/internal/line"
	"line-balancer/internal/metrics"
)

type Config struct {
	Addr string
}

func New(cfg Config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/balance", handleBalance)
	mux.HandleFunc("/api/takt", handleTakt)
	mux.HandleFunc("/api/metrics", handleMetrics)
	return mux
}

func ListenAndServe(cfg Config) error {
	mux := New(cfg)
	return http.ListenAndServe(cfg.Addr, mux)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type taskInput struct {
	Name string  `json:"name"`
	Time float64 `json:"time"`
}

type balanceRequest struct {
	Tasks     []taskInput `json:"tasks"`
	Demand    int         `json:"demand"`
	Available float64     `json:"available"`
}

type stationOutput struct {
	Tasks []string `json:"tasks"`
	Load  float64  `json:"load"`
}

type balanceResponse struct {
	TaktTime     float64         `json:"takt_time"`
	StationCount int             `json:"station_count"`
	Bottleneck   int             `json:"bottleneck"`
	MaxLoad      float64         `json:"max_load"`
	Efficiency   float64         `json:"efficiency"`
	TotalTime    float64         `json:"total_time"`
	Stations     []stationOutput `json:"stations"`
}

func handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req balanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if len(req.Tasks) == 0 {
		httpError(w, http.StatusBadRequest, "tasks array is empty")
		return
	}
	if req.Demand <= 0 {
		httpError(w, http.StatusBadRequest, "demand must be > 0")
		return
	}
	if req.Available <= 0 {
		req.Available = 28800
	}

	tasks := make([]line.Task, len(req.Tasks))
	for i, ti := range req.Tasks {
		tasks[i] = line.Task{Name: ti.Name, Time: ti.Time}
	}

	res, err := line.Analyze(tasks, req.Demand, req.Available)
	if err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}

	stations := make([]stationOutput, len(res.Stations))
	for i, s := range res.Stations {
		stations[i] = stationOutput{Tasks: s.Tasks, Load: s.Load}
	}

	writeJSON(w, http.StatusOK, balanceResponse{
		TaktTime:     res.TaktTime,
		StationCount: res.StationCount,
		Bottleneck:   res.Bottleneck,
		MaxLoad:      res.MaxLoad,
		Efficiency:   res.Efficiency,
		TotalTime:    res.TotalTime,
		Stations:     stations,
	})
}

type taktRequest struct {
	Demand    int     `json:"demand"`
	Available float64 `json:"available"`
}

type taktResponse struct {
	TaktTime float64 `json:"takt_time"`
	Demand   int     `json:"demand"`
}

func handleTakt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req taktRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Demand <= 0 {
		httpError(w, http.StatusBadRequest, "demand must be > 0")
		return
	}
	if req.Available <= 0 {
		req.Available = 28800
	}
	tt := line.TaktTime(req.Demand, req.Available)
	writeJSON(w, http.StatusOK, taktResponse{TaktTime: tt, Demand: req.Demand})
}

type metricsRequest struct {
	Tasks     []taskInput `json:"tasks"`
	Demand    int         `json:"demand"`
	Available float64     `json:"available"`
}

type metricsResponse struct {
	SmoothnessIndex float64 `json:"smoothness_index"`
	BalanceDelay    float64 `json:"balance_delay_pct"`
	IdleTime        float64 `json:"idle_time"`
	OutputRate      float64 `json:"output_rate"`
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}
	var req metricsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if len(req.Tasks) == 0 {
		httpError(w, http.StatusBadRequest, "tasks array is empty")
		return
	}
	if req.Demand <= 0 {
		httpError(w, http.StatusBadRequest, "demand must be > 0")
		return
	}
	if req.Available <= 0 {
		req.Available = 28800
	}

	tasks := make([]line.Task, len(req.Tasks))
	for i, ti := range req.Tasks {
		tasks[i] = line.Task{Name: ti.Name, Time: ti.Time}
	}

	res, err := line.Analyze(tasks, req.Demand, req.Available)
	if err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}

	summary := metrics.LineSummary{
		StationLoads: make([]float64, len(res.Stations)),
		CycleTime:    res.CycleTime,
		TotalTime:    res.TotalTime,
	}
	for i, s := range res.Stations {
		summary.StationLoads[i] = s.Load
	}

	eff := metrics.Efficiency(summary)
	eff = HoldEffAPI(eff)
	writeJSON(w, http.StatusOK, metricsResponse{
		SmoothnessIndex: metrics.SmoothnessIndex(summary),
		BalanceDelay:    100 - eff,
		IdleTime:        metrics.IdleTime(summary),
		OutputRate:      metrics.OutputRate(res.MaxLoad, req.Available),
	})
}

func httpError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func ParsePort(addr string) int {
	parts := strings.Split(addr, ":")
	if len(parts) < 2 {
		return 0
	}
	p, _ := strconv.Atoi(parts[len(parts)-1])
	return p
}

func FormatAddr(addr string) string {
	port := ParsePort(addr)
	if port == 0 {
		return addr
	}
	return fmt.Sprintf("http://0.0.0.0:%d", port)
}
