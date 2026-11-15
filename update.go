package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const updateCheckInterval = 24 * time.Hour

var (
	updateAPIURL     = "https://api.github.com/repos/ausbitbank/twitch-live-notifier/releases/latest"
	updateHTTPClient = &http.Client{Timeout: 10 * time.Second}
)

// UpdateStatus is the cached result of the latest GitHub release check.
type UpdateStatus struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version,omitempty"`
	UpdateAvailable bool   `json:"update_available"`
	ReleaseURL      string `json:"release_url,omitempty"`
	DownloadURL     string `json:"download_url,omitempty"`
	PublishedAt     string `json:"published_at,omitempty"`
	CheckedAt       string `json:"checked_at,omitempty"`
	Error           string `json:"error,omitempty"`
	Disabled        bool   `json:"disabled,omitempty"`
}

type githubRelease struct {
	TagName     string `json:"tag_name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (a *App) updateChecksEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return !a.cfg.Settings.DisableUpdateCheck
}

func (a *App) cachedUpdateStatus() UpdateStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	status := a.updateStatus
	if status.CurrentVersion == "" {
		status.CurrentVersion = Version
	}
	return status
}

// CheckForUpdates asks GitHub for the latest published release. It is safe to
// call from the monitor loop or a manual Settings action.
func (a *App) CheckForUpdates(ctx context.Context) UpdateStatus {
	status := UpdateStatus{CurrentVersion: Version}
	if !a.updateChecksEnabled() {
		status.Disabled = true
		status.CheckedAt = time.Now().UTC().Format(time.RFC3339)
		a.storeUpdateStatus(status)
		return status
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, updateAPIURL, nil)
	if err != nil {
		status.Error = "could not build the update request"
		status.CheckedAt = time.Now().UTC().Format(time.RFC3339)
		a.storeUpdateStatus(status)
		return status
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", appName+"/"+Version)

	response, err := updateHTTPClient.Do(req)
	if err != nil {
		status.Error = "could not reach GitHub: " + networkErrorText(err)
		status.CheckedAt = time.Now().UTC().Format(time.RFC3339)
		a.storeUpdateStatus(status)
		return status
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if response.StatusCode != http.StatusOK {
		status.Error = fmt.Sprintf("GitHub returned HTTP %d", response.StatusCode)
		status.CheckedAt = time.Now().UTC().Format(time.RFC3339)
		a.storeUpdateStatus(status)
		return status
	}

	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		status.Error = "GitHub returned an invalid release response"
		status.CheckedAt = time.Now().UTC().Format(time.RFC3339)
		a.storeUpdateStatus(status)
		return status
	}
	latest := strings.TrimSpace(release.TagName)
	if latest == "" {
		status.Error = "the latest GitHub release has no tag"
		status.CheckedAt = time.Now().UTC().Format(time.RFC3339)
		a.storeUpdateStatus(status)
		return status
	}

	status.LatestVersion = strings.TrimPrefix(latest, "v")
	status.ReleaseURL = release.HTMLURL
	status.PublishedAt = release.PublishedAt
	status.UpdateAvailable = compareVersions(status.LatestVersion, Version) > 0
	status.DownloadURL = selectUpdateDownload(release, runtime.GOOS)
	if status.DownloadURL == "" {
		status.DownloadURL = status.ReleaseURL
	}
	status.CheckedAt = time.Now().UTC().Format(time.RFC3339)
	a.storeUpdateStatus(status)
	return status
}

func (a *App) storeUpdateStatus(status UpdateStatus) {
	a.mu.Lock()
	a.updateStatus = status
	a.mu.Unlock()
}

func (a *App) updateLoop(ctx context.Context) {
	timer := time.NewTimer(8 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if a.updateChecksEnabled() {
				_ = a.CheckForUpdates(ctx)
			}
			timer.Reset(updateCheckInterval)
		}
	}
}

func selectUpdateDownload(release githubRelease, goos string) string {
	preferences := []string{}
	switch goos {
	case "windows":
		preferences = []string{"Setup.exe", "Windows-x64.zip", ".exe"}
	case "darwin":
		preferences = []string{"macOS.tar.gz", "macOS.zip"}
	case "linux":
		preferences = []string{"Linux-x64.AppImage", "Linux-x64.deb", "Linux-x64.tar.gz"}
	default:
		preferences = []string{"Linux-x64.tar.gz"}
	}
	for _, preference := range preferences {
		for _, asset := range release.Assets {
			if strings.Contains(asset.Name, preference) && asset.BrowserDownloadURL != "" {
				return asset.BrowserDownloadURL
			}
		}
	}
	return ""
}

// compareVersions compares dotted numeric versions, ignoring a leading v and
// any pre-release suffix. It returns -1, 0, or 1.
func compareVersions(left, right string) int {
	leftParts := versionParts(left)
	rightParts := versionParts(right)
	count := len(leftParts)
	if len(rightParts) > count {
		count = len(rightParts)
	}
	for i := 0; i < count; i++ {
		leftValue := 0
		if i < len(leftParts) {
			leftValue = leftParts[i]
		}
		rightValue := 0
		if i < len(rightParts) {
			rightValue = rightParts[i]
		}
		if leftValue < rightValue {
			return -1
		}
		if leftValue > rightValue {
			return 1
		}
	}
	return 0
}

func versionParts(value string) []int {
	value = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(value), "v"))
	if index := strings.IndexAny(value, "-+"); index >= 0 {
		value = value[:index]
	}
	rawParts := strings.Split(value, ".")
	parts := make([]int, 0, len(rawParts))
	for _, raw := range rawParts {
		number := 0
		for _, character := range raw {
			if character < '0' || character > '9' {
				break
			}
			number = number*10 + int(character-'0')
		}
		parts = append(parts, number)
	}
	return parts
}
