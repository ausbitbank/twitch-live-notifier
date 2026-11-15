package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		left, right string
		want        int
	}{
		{"2.3.0", "2.3.0", 0},
		{"v2.3.0", "2.3.0", 0},
		{"2.3.1", "2.3.0", 1},
		{"2.2.9", "2.3.0", -1},
		{"2.4", "2.3.9", 1},
		{"2.3.0-rc1", "2.3.0", 0},
		{"3.0.0", "2.99.99", 1},
	}
	for _, test := range tests {
		if got := compareVersions(test.left, test.right); got != test.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", test.left, test.right, got, test.want)
		}
	}
}

func TestCheckForUpdates(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name": "v9.9.9",
			"html_url": "https://example.test/releases/v9.9.9",
			"published_at": "2026-11-16T00:00:00Z",
			"assets": [
				{"name":"TwitchLiveNotifier-v9.9.9-Windows-x64-Setup.exe","browser_download_url":"https://example.test/setup.exe"},
				{"name":"TwitchLiveNotifier-v9.9.9-Linux-x64.AppImage","browser_download_url":"https://example.test/app.AppImage"},
				{"name":"TwitchLiveNotifier-v9.9.9-macOS.tar.gz","browser_download_url":"https://example.test/mac.tar.gz"}
			]
		}`))
	}))
	defer server.Close()

	oldURL := updateAPIURL
	updateAPIURL = server.URL
	defer func() { updateAPIURL = oldURL }()

	app := newTestApp(t)
	status := app.CheckForUpdates(context.Background())
	if status.Error != "" {
		t.Fatalf("status error = %q", status.Error)
	}
	if status.LatestVersion != "9.9.9" || !status.UpdateAvailable {
		t.Fatalf("status = %+v", status)
	}
	if status.ReleaseURL != "https://example.test/releases/v9.9.9" {
		t.Fatalf("release URL = %q", status.ReleaseURL)
	}
	if status.DownloadURL == "" {
		t.Fatal("missing download URL")
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d", calls.Load())
	}
	stateUpdate := app.State().Update
	if stateUpdate.LatestVersion != "9.9.9" || !stateUpdate.UpdateAvailable {
		t.Fatalf("cached state update = %+v", stateUpdate)
	}
}

func TestUpdatesAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v9.9.9","html_url":"https://example.test/release","assets":[]}`))
	}))
	defer server.Close()

	oldURL := updateAPIURL
	updateAPIURL = server.URL
	defer func() { updateAPIURL = oldURL }()

	app := newTestApp(t)
	dashboard := httptest.NewServer(withSecurity(withDashboardAuth(app, newDashboardHandler(app, func() {}))))
	defer dashboard.Close()

	response, err := http.Get(dashboard.URL + "/api/updates")
	if err != nil {
		t.Fatal(err)
	}
	var cached UpdateStatus
	_ = json.NewDecoder(response.Body).Decode(&cached)
	response.Body.Close()
	if cached.CurrentVersion != Version {
		t.Fatalf("cached status = %+v", cached)
	}

	request, _ := http.NewRequest(http.MethodPost, dashboard.URL+"/api/updates/check", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	var checked UpdateStatus
	_ = json.NewDecoder(response.Body).Decode(&checked)
	response.Body.Close()
	if response.StatusCode != http.StatusOK || checked.LatestVersion != "9.9.9" || !checked.UpdateAvailable {
		t.Fatalf("checked status = %+v (%d)", checked, response.StatusCode)
	}
}

func TestCheckForUpdatesDisabled(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
	}))
	defer server.Close()

	oldURL := updateAPIURL
	updateAPIURL = server.URL
	defer func() { updateAPIURL = oldURL }()

	app := newTestApp(t)
	app.mu.Lock()
	app.cfg.Settings.DisableUpdateCheck = true
	app.mu.Unlock()
	status := app.CheckForUpdates(context.Background())
	if !status.Disabled || status.Error != "" {
		t.Fatalf("status = %+v", status)
	}
	if calls.Load() != 0 {
		t.Fatalf("disabled check made %d request(s)", calls.Load())
	}
}

func TestCheckForUpdatesFailureIsVisible(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	oldURL := updateAPIURL
	updateAPIURL = server.URL
	defer func() { updateAPIURL = oldURL }()

	app := newTestApp(t)
	status := app.CheckForUpdates(context.Background())
	if status.Error == "" || !strings.Contains(status.Error, "403") {
		t.Fatalf("status = %+v", status)
	}
	if status.CurrentVersion != Version {
		t.Fatalf("current version = %q", status.CurrentVersion)
	}
}

func TestSelectUpdateDownload(t *testing.T) {
	release := githubRelease{Assets: []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}{
		{Name: "TwitchLiveNotifier-v2.4.0-Windows-x64.zip", BrowserDownloadURL: "https://example.test/windows.zip"},
		{Name: "TwitchLiveNotifier-v2.4.0-Windows-x64-Setup.exe", BrowserDownloadURL: "https://example.test/setup.exe"},
		{Name: "TwitchLiveNotifier-v2.4.0-Linux-x64.tar.gz", BrowserDownloadURL: "https://example.test/linux.tar.gz"},
		{Name: "TwitchLiveNotifier-v2.4.0-Linux-x64.deb", BrowserDownloadURL: "https://example.test/linux.deb"},
		{Name: "TwitchLiveNotifier-v2.4.0-Linux-x64.AppImage", BrowserDownloadURL: "https://example.test/linux.AppImage"},
		{Name: "TwitchLiveNotifier-v2.4.0-macOS.tar.gz", BrowserDownloadURL: "https://example.test/mac.tar.gz"},
	}}
	if got := selectUpdateDownload(release, "windows"); got != "https://example.test/setup.exe" {
		t.Fatalf("windows download = %q", got)
	}
	if got := selectUpdateDownload(release, "linux"); got != "https://example.test/linux.AppImage" {
		t.Fatalf("linux download = %q", got)
	}
	if got := selectUpdateDownload(release, "darwin"); got != "https://example.test/mac.tar.gz" {
		t.Fatalf("macOS download = %q", got)
	}
}
