package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func sampleTasks() []taskInput {
	return []taskInput{
		{Name: "cut", Time: 10},
		{Name: "drill", Time: 15},
		{Name: "weld", Time: 20},
		{Name: "paint", Time: 12},
		{Name: "inspect", Time: 8},
		{Name: "pack", Time: 5},
	}
}

func TestHealthEndpoint(t *testing.T) {
	mux := New(Config{Addr: ":8080"})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBalanceEndpoint(t *testing.T) {
	mux := New(Config{Addr: ":8080"})
	payload := balanceRequest{Tasks: sampleTasks(), Demand: 400, Available: 28800}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/balance", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp balanceResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.StationCount <= 0 {
		t.Error("expected station count > 0")
	}
	if resp.TaktTime <= 0 {
		t.Error("expected takt time > 0")
	}
}

func TestBalanceEndpoint_EmptyTasks(t *testing.T) {
	mux := New(Config{Addr: ":8080"})
	body := []byte(`{"tasks":[],"demand":100}`)
	req := httptest.NewRequest(http.MethodPost, "/api/balance", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBalanceEndpoint_ZeroDemand(t *testing.T) {
	mux := New(Config{Addr: ":8080"})
	payload := balanceRequest{Tasks: sampleTasks(), Demand: 0}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/balance", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestTaktEndpoint(t *testing.T) {
	mux := New(Config{Addr: ":8080"})
	payload := taktRequest{Demand: 400, Available: 28800}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/takt", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp taktResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.TaktTime != 72 {
		t.Errorf("expected takt=72, got %f", resp.TaktTime)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	mux := New(Config{Addr: ":8080"})
	payload := metricsRequest{Tasks: sampleTasks(), Demand: 400, Available: 28800}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/metrics", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp metricsResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.OutputRate <= 0 {
		t.Error("expected output rate > 0")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	mux := New(Config{Addr: ":8080"})
	endpoints := []string{"/api/balance", "/api/takt", "/api/metrics"}
	for _, ep := range endpoints {
		req := httptest.NewRequest(http.MethodGet, ep, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", ep, rec.Code)
		}
	}
}
