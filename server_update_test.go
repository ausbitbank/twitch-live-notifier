package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newUpdateTestServer(t *testing.T, app *App) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	registerUpdateRoutes(mux, app)
	return httptest.NewServer(mux)
}

func TestUpdateInstallAPIIdleAndValidation(t *testing.T) {
	app := newTestApp(t)
	server := newUpdateTestServer(t, app)
	defer server.Close()

	response, err := http.Get(server.URL + "/api/updates/install")
	if err != nil {
		t.Fatal(err)
	}
	var job UpdateJob
	_ = json.NewDecoder(response.Body).Decode(&job)
	response.Body.Close()
	if response.StatusCode != http.StatusOK || job.Status != updateJobIdle {
		t.Fatalf("idle job = %+v (%d)", job, response.StatusCode)
	}

	request, _ := http.NewRequest(http.MethodPost, server.URL+"/api/updates/install", strings.NewReader(`{"restart":true}`))
	request.Header.Set("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("manual-only install status = %d", response.StatusCode)
	}

	request, _ = http.NewRequest(http.MethodPost, server.URL+"/api/updates/install", strings.NewReader(`not json`))
	request.Header.Set("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid JSON status = %d", response.StatusCode)
	}
}

func TestUpdateInstallAPIConflict(t *testing.T) {
	app := newTestApp(t)
	app.storeUpdateStatus(UpdateStatus{
		CanSelfUpdate: true,
		DownloadURL:   "https://example.test/download",
		AssetSHA256:   strings.Repeat("a", 64),
		AssetName:     "TwitchLiveNotifier-linux-amd64",
	})
	app.updates.mu.Lock()
	app.updates.active = true
	app.updates.mu.Unlock()

	server := newUpdateTestServer(t, app)
	defer server.Close()

	request, _ := http.NewRequest(http.MethodPost, server.URL+"/api/updates/install", strings.NewReader(`{"restart":false}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("conflict status = %d", response.StatusCode)
	}
}
