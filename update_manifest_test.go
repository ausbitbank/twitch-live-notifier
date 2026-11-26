package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
)

func TestParseUpdateManifest(t *testing.T) {
	manifest, err := parseUpdateManifest([]byte(`{
		"version":"2.6.0",
		"published_at":"2026-11-26T00:00:00Z",
		"assets":{
			"windows-amd64":{"name":"TwitchLiveNotifier.exe","url":"https://example.test/app.exe","sha256":"` + strings.Repeat("a", 64) + `","size":123},
			"linux-amd64":{"name":"TwitchLiveNotifier-linux-amd64","url":"https://example.test/linux","sha256":"` + strings.Repeat("b", 64) + `","size":456}
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Version != "2.6.0" || len(manifest.Assets) != 2 {
		t.Fatalf("manifest = %+v", manifest)
	}
	if manifest.Assets["windows-amd64"].Size != 123 {
		t.Fatalf("windows asset = %+v", manifest.Assets["windows-amd64"])
	}
	if _, err := parseUpdateManifest([]byte(`{}`)); err == nil {
		t.Fatal("empty manifest should fail")
	}
	if _, err := parseUpdateManifest([]byte(`{"assets":{"bad":{"url":"x","sha256":"not-hex"}}}`)); err == nil {
		t.Fatal("invalid checksum should fail")
	}
}

func TestResolveUpdateAssetAllPlatforms(t *testing.T) {
	sums := map[string]string{
		"TwitchLiveNotifier-v9.9.9-Linux-x64.AppImage": strings.Repeat("c", 64),
		"TwitchLiveNotifier-v9.9.9-Linux-arm64.tar.gz": strings.Repeat("d", 64),
		"TwitchLiveNotifier-v9.9.9-macOS.tar.gz":       strings.Repeat("e", 64),
		"TwitchLiveNotifier.exe":                       strings.Repeat("f", 64),
	}
	release := githubRelease{Assets: []githubAsset{
		{Name: "TwitchLiveNotifier.exe", BrowserDownloadURL: "https://example.test/app.exe", Size: 10},
		{Name: "TwitchLiveNotifier-v9.9.9-Windows-x64-Setup.exe", BrowserDownloadURL: "https://example.test/setup.exe", Size: 20},
		{Name: "TwitchLiveNotifier-linux-amd64", BrowserDownloadURL: "https://example.test/linux-amd64", Size: 30},
		{Name: "TwitchLiveNotifier-v9.9.9-Linux-arm64.tar.gz", BrowserDownloadURL: "https://example.test/linux-arm64.tar.gz", Size: 40},
		{Name: "TwitchLiveNotifier-v9.9.9-Linux-x64.AppImage", BrowserDownloadURL: "https://example.test/app.AppImage", Size: 50},
		{Name: "TwitchLiveNotifier-v9.9.9-macOS.tar.gz", BrowserDownloadURL: "https://example.test/mac.tar.gz", Size: 60},
	}}
	manifest := &updateManifest{Assets: map[string]updateManifestAsset{
		"windows-amd64": {Name: "TwitchLiveNotifier.exe", URL: "https://example.test/manifest.exe", SHA256: strings.Repeat("1", 64), Size: 11},
		"linux-amd64":   {Name: "TwitchLiveNotifier-linux-amd64", URL: "https://example.test/manifest-linux", SHA256: strings.Repeat("2", 64), Size: 31},
		"linux-arm64":   {Name: "TwitchLiveNotifier-v9.9.9-Linux-arm64.tar.gz", URL: "https://example.test/manifest-arm64.tar.gz", SHA256: strings.Repeat("3", 64), Size: 41},
		"darwin-amd64":  {Name: "TwitchLiveNotifier-v9.9.9-macOS.tar.gz", URL: "https://example.test/manifest-mac.tar.gz", SHA256: strings.Repeat("4", 64), Size: 61},
		"darwin-arm64":  {Name: "TwitchLiveNotifier-v9.9.9-macOS.tar.gz", URL: "https://example.test/manifest-mac.tar.gz", SHA256: strings.Repeat("4", 64), Size: 61},
	}}

	tests := []struct {
		goos, goarch, mode string
		name, url, sha     string
	}{
		{"windows", "amd64", installModeInPlace, "TwitchLiveNotifier.exe", "https://example.test/manifest.exe", strings.Repeat("1", 64)},
		{"linux", "amd64", installModeInPlace, "TwitchLiveNotifier-linux-amd64", "https://example.test/manifest-linux", strings.Repeat("2", 64)},
		{"linux", "arm64", installModeInPlace, "TwitchLiveNotifier-v9.9.9-Linux-arm64.tar.gz", "https://example.test/manifest-arm64.tar.gz", strings.Repeat("3", 64)},
		{"linux", "amd64", installModeAppImage, "TwitchLiveNotifier-v9.9.9-Linux-x64.AppImage", "https://example.test/app.AppImage", strings.Repeat("c", 64)},
		{"darwin", "amd64", installModeAppBundle, "TwitchLiveNotifier-v9.9.9-macOS.tar.gz", "https://example.test/manifest-mac.tar.gz", strings.Repeat("4", 64)},
		{"darwin", "arm64", installModeAppBundle, "TwitchLiveNotifier-v9.9.9-macOS.tar.gz", "https://example.test/manifest-mac.tar.gz", strings.Repeat("4", 64)},
	}
	for _, test := range tests {
		asset := resolveUpdateAsset(release, manifest, test.goos, test.goarch, test.mode, sums)
		if asset.Name != test.name || asset.URL != test.url || asset.SHA256 != test.sha {
			t.Errorf("%s/%s %s = %+v", test.goos, test.goarch, test.mode, asset)
		}
	}
}

func TestResolveUpdateAssetFallback(t *testing.T) {
	release := githubRelease{Assets: []githubAsset{
		{Name: "TwitchLiveNotifier-v9.9.9-Windows-x64-Setup.exe", BrowserDownloadURL: "https://example.test/setup.exe"},
		{Name: "TwitchLiveNotifier.exe", BrowserDownloadURL: "https://example.test/app.exe"},
		{Name: "TwitchLiveNotifier-linux-amd64", BrowserDownloadURL: "https://example.test/linux-amd64"},
		{Name: "TwitchLiveNotifier-v9.9.9-Linux-arm64.tar.gz", BrowserDownloadURL: "https://example.test/linux-arm64.tar.gz"},
		{Name: "TwitchLiveNotifier-v9.9.9-Linux-x64.AppImage", BrowserDownloadURL: "https://example.test/app.AppImage"},
		{Name: "TwitchLiveNotifier-v9.9.9-macOS.tar.gz", BrowserDownloadURL: "https://example.test/mac.tar.gz"},
	}}
	sums := map[string]string{"TwitchLiveNotifier.exe": strings.Repeat("a", 64)}

	asset := resolveUpdateAsset(release, nil, "windows", "amd64", installModeInPlace, sums)
	if asset.Name != "TwitchLiveNotifier.exe" || asset.SHA256 != strings.Repeat("a", 64) {
		t.Fatalf("windows fallback = %+v", asset)
	}
	asset = resolveUpdateAsset(release, nil, "linux", "arm64", installModeInPlace, nil)
	if asset.Name != "TwitchLiveNotifier-v9.9.9-Linux-arm64.tar.gz" {
		t.Fatalf("linux arm64 fallback = %+v", asset)
	}
	asset = resolveUpdateAsset(release, nil, "linux", "amd64", installModeAppImage, nil)
	if asset.Name != "TwitchLiveNotifier-v9.9.9-Linux-x64.AppImage" {
		t.Fatalf("appimage fallback = %+v", asset)
	}
	asset = resolveUpdateAsset(release, nil, "darwin", "arm64", installModeAppBundle, nil)
	if asset.Name != "TwitchLiveNotifier-v9.9.9-macOS.tar.gz" {
		t.Fatalf("mac fallback = %+v", asset)
	}
}

func TestParseSHA256SUMS(t *testing.T) {
	sums := parseSHA256SUMS([]byte(strings.Repeat("a", 64) + "  TwitchLiveNotifier.exe\n" + strings.Repeat("b", 64) + " *docs/file.tar.gz\nnonsense\n"))
	if sums["TwitchLiveNotifier.exe"] != strings.Repeat("a", 64) {
		t.Fatalf("sums = %+v", sums)
	}
	if sums["file.tar.gz"] != strings.Repeat("b", 64) {
		t.Fatalf("sums = %+v", sums)
	}
}

func TestApplyUpdateAssetStatus(t *testing.T) {
	target := installTarget{Mode: installModeInPlace, Writable: true}
	status := UpdateStatus{}
	applyUpdateAssetStatus(&status, target, updateAsset{Name: "TwitchLiveNotifier-linux-amd64", URL: "https://example.test/app", SHA256: strings.Repeat("a", 64), Size: 10})
	if !status.CanSelfUpdate || status.InstallMode != installModeInPlace || status.AssetSize != 10 {
		t.Fatalf("status = %+v", status)
	}

	status = UpdateStatus{}
	applyUpdateAssetStatus(&status, target, updateAsset{Name: "TwitchLiveNotifier-linux-amd64", URL: "https://example.test/app"})
	if status.CanSelfUpdate || !strings.Contains(status.InstallReason, "SHA-256") {
		t.Fatalf("missing checksum status = %+v", status)
	}

	status = UpdateStatus{}
	applyUpdateAssetStatus(&status, installTarget{Mode: installModeInPlace, Writable: false}, updateAsset{Name: "TwitchLiveNotifier-linux-amd64", URL: "https://example.test/app", SHA256: strings.Repeat("a", 64)})
	if status.CanSelfUpdate || !strings.Contains(status.InstallReason, "writable") {
		t.Fatalf("unwritable status = %+v", status)
	}

	status = UpdateStatus{}
	applyUpdateAssetStatus(&status, installTarget{Mode: installModeGuided, Reason: "Docker"}, updateAsset{URL: "https://example.test/app", Name: "x", SHA256: strings.Repeat("a", 64)})
	if status.CanSelfUpdate || status.InstallReason != "Docker" {
		t.Fatalf("guided status = %+v", status)
	}
}

func TestInstallTargetFor(t *testing.T) {
	tests := []struct {
		path, goos, arch, appImage string
		docker                     bool
		mode                       string
	}{
		{`C:\Users\A\AppData\Local\Programs\TwitchLiveNotifier.exe`, "windows", "amd64", "", false, installModeInPlace},
		{`C:\Program Files\TwitchLiveNotifier\TwitchLiveNotifier.exe`, "windows", "amd64", "", false, installModeGuided},
		{"/home/a/.local/bin/twitch-live-notifier", "linux", "amd64", "", false, installModeInPlace},
		{"/usr/bin/twitch-live-notifier", "linux", "amd64", "", false, installModeGuided},
		{"/tmp/.mount_abc/AppRun", "linux", "amd64", "/home/a/Twitch.AppImage", false, installModeAppImage},
		{"/Applications/Twitch Live to Discord.app/Contents/MacOS/TwitchLiveNotifier", "darwin", "arm64", "", false, installModeAppBundle},
		{"/home/a/bin/twitch-live-notifier", "linux", "amd64", "", true, installModeGuided},
	}
	for _, test := range tests {
		target := installTargetFor(test.path, test.goos, test.arch, test.appImage, test.docker)
		if target.Mode != test.mode {
			t.Errorf("installTargetFor(%q, %s) mode = %q, want %q", test.path, test.goos, target.Mode, test.mode)
		}
	}
}

func TestCheckForUpdatesWithManifest(t *testing.T) {
	t.Setenv("TWITCH_LIVE_DOCKER", "0")
	key := runtime.GOOS + "-" + runtime.GOARCH
	name := "TwitchLiveNotifier-" + runtime.GOOS + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		name = "TwitchLiveNotifier.exe"
	} else if runtime.GOOS == "darwin" {
		name = "TwitchLiveNotifier-v9.9.9-macOS.tar.gz"
	}
	manifest := map[string]any{
		"version": "9.9.9",
		"assets": map[string]any{
			key: map[string]any{
				"name":   name,
				"url":    "https://example.test/download",
				"sha256": strings.Repeat("a", 64),
				"size":   1234,
			},
		},
	}
	manifestBody, _ := json.Marshal(manifest)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(manifestBody)
		default:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"tag_name":"v9.9.9",
				"html_url":"https://example.test/release",
				"assets":[
					{"name":"update-manifest.json","browser_download_url":"` + "http://" + r.Host + `/manifest","size":` + "123" + `}
				]
			}`))
		}
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
	if status.AssetName != name || status.AssetSHA256 != strings.Repeat("a", 64) || status.AssetSize != 1234 {
		t.Fatalf("asset status = %+v", status)
	}
	if !status.CanSelfUpdate {
		t.Fatalf("expected self-update to be allowed: %+v", status)
	}
}
