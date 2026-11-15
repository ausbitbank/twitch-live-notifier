package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthEndpointHealthy(t *testing.T) {
	app := newTestApp(t)
	server := httptest.NewServer(withSecurity(newDashboardHandler(app, func() {})))
	defer server.Close()

	response, err := http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	var health Health
	if err := json.NewDecoder(response.Body).Decode(&health); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if health.Status != healthStatusHealthy || health.Version != Version {
		t.Fatalf("health = %+v", health)
	}
	if health.Streamers != 0 || health.Destinations != 0 || health.Monitoring == false {
		t.Fatalf("health counts = %+v", health)
	}
}

func TestHealthEndpointStaleAndPaused(t *testing.T) {
	app := newTestApp(t)
	app.mu.Lock()
	app.cfg.Settings.PollSeconds = 30
	app.startedAt = time.Now().Add(-time.Hour)
	app.lastCheck = time.Now().Add(-time.Hour)
	app.mu.Unlock()

	handler := withSecurity(newDashboardHandler(app, func() {}))
	server := httptest.NewServer(handler)
	defer server.Close()

	response, err := http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("stale status = %d", response.StatusCode)
	}

	app.mu.Lock()
	app.cfg.Settings.Paused = true
	app.mu.Unlock()

	response, err = http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("paused status = %d", response.StatusCode)
	}
	var health Health
	if err := json.NewDecoder(response.Body).Decode(&health); err != nil {
		t.Fatal(err)
	}
	if health.Status != healthStatusHealthy || !health.Paused {
		t.Fatalf("paused health = %+v", health)
	}
}

func TestCheckHealth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	}))
	defer server.Close()
	if err := checkHealth(server.URL + "/healthz"); err != nil {
		t.Fatalf("checkHealth returned error: %v", err)
	}

	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"unhealthy","reason":"stale"}`))
	}))
	defer unhealthy.Close()
	err := checkHealth(unhealthy.URL + "/healthz")
	if err == nil || !strings.Contains(err.Error(), "503") || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("checkHealth error = %v", err)
	}
}

func TestHealthcheckURL(t *testing.T) {
	tests := []struct {
		explicit string
		listen   string
		want     string
	}{
		{"http://example.test/healthz", "", "http://example.test/healthz"},
		{"", "0.0.0.0:9000", "http://127.0.0.1:9000/healthz"},
		{"", ":9001", "http://127.0.0.1:9001/healthz"},
		{"", "9002", "http://127.0.0.1:9002/healthz"},
		{"", "", "http://127.0.0.1:8787/healthz"},
	}
	for _, test := range tests {
		if got := healthcheckURL(test.explicit, test.listen); got != test.want {
			t.Errorf("healthcheckURL(%q, %q) = %q, want %q", test.explicit, test.listen, got, test.want)
		}
	}
}

func TestStartDashboardExplicitListen(t *testing.T) {
	app := newTestApp(t)
	dashboard, err := startDashboard(app, func() {}, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startDashboard returned error: %v", err)
	}
	defer dashboard.Shutdown(context.Background())
	if !strings.HasPrefix(dashboard.URL(), "http://127.0.0.1:") {
		t.Fatalf("URL = %q", dashboard.URL())
	}
	response, err := http.Get(dashboard.URL() + "healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d", response.StatusCode)
	}
}
