package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDashboardTokenProtectsAPI(t *testing.T) {
	app := newTestApp(t)
	if err := app.SetDashboardToken("correct horse"); err != nil {
		t.Fatalf("SetDashboardToken returned error: %v", err)
	}
	if !app.DashboardTokenRequired() {
		t.Fatal("token requirement was not enabled")
	}
	state := app.State()
	if !state.DashboardTokenSet {
		t.Fatal("state did not report the token")
	}
	if state.Config.Settings.DashboardTokenHash != "" {
		t.Fatal("state leaked the token hash")
	}

	handler := withSecurity(withDashboardAuth(app, newDashboardHandler(app, func() {})))
	server := httptest.NewServer(handler)
	defer server.Close()

	request, _ := http.NewRequest(http.MethodGet, server.URL+"/api/state", nil)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no-token status = %d", response.StatusCode)
	}

	request, _ = http.NewRequest(http.MethodGet, server.URL+"/api/state", nil)
	request.Header.Set("X-Dashboard-Token", "correct horse")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("header-token status = %d", response.StatusCode)
	}

	response, err = http.Get(server.URL + "/api/state?token=correct+horse")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("query-token status = %d", response.StatusCode)
	}

	// Health checks and the dashboard shell stay reachable so Docker and the
	// token prompt keep working.
	response, err = http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d", response.StatusCode)
	}
}

func TestSetDashboardTokenValidationAndClear(t *testing.T) {
	app := newTestApp(t)
	if err := app.SetDashboardToken("abc"); err == nil {
		t.Fatal("short token accepted")
	}
	if err := app.SetDashboardToken("long-enough"); err != nil {
		t.Fatalf("SetDashboardToken returned error: %v", err)
	}
	if err := app.SetDashboardToken(""); err != nil {
		t.Fatalf("clearing token returned error: %v", err)
	}
	if app.DashboardTokenRequired() {
		t.Fatal("token requirement survived clearing")
	}
}
