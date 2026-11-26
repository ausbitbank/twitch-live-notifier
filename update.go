package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const updateCheckInterval = 24 * time.Hour

const updateManifestAssetName = "update-manifest.json"
const updateManifestSignatureName = "update-manifest.json.sig"
const updateChecksumsAssetName = "SHA256SUMS.txt"

var (
	updateAPIURL     = "https://api.github.com/repos/ausbitbank/twitch-live-notifier/releases/latest"
	updateHTTPClient = &http.Client{Timeout: 10 * time.Second}
)

const (
	installModeInPlace     = "in_place"
	installModeAppBundle   = "app_bundle"
	installModeAppImage    = "appimage"
	installModeGuided      = "guided"
	installModeUnsupported = "unsupported"
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

	AssetName     string     `json:"asset_name,omitempty"`
	AssetSize     int64      `json:"asset_size,omitempty"`
	AssetSHA256   string     `json:"asset_sha256,omitempty"`
	CanSelfUpdate bool       `json:"can_self_update"`
	InstallMode   string     `json:"install_mode,omitempty"`
	InstallReason string     `json:"install_reason,omitempty"`
	InstallJob    *UpdateJob `json:"install_job,omitempty"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type githubRelease struct {
	TagName     string        `json:"tag_name"`
	HTMLURL     string        `json:"html_url"`
	PublishedAt string        `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

// updateManifest is the optional release asset that describes every supported
// platform so the updater does not have to guess from release asset names.
type updateManifest struct {
	Version     string                         `json:"version"`
	PublishedAt string                         `json:"published_at"`
	Assets      map[string]updateManifestAsset `json:"assets"`
}

type updateManifestAsset struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// updateAsset is the resolved download chosen for this machine.
type updateAsset struct {
	Name   string
	URL    string
	SHA256 string
	Size   int64
}

// installTarget describes where and how this copy of the app can be updated.
type installTarget struct {
	Executable string
	Target     string
	Mode       string
	Reason     string
	Writable   bool
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

	target := inspectInstallTarget()
	status.InstallMode = target.Mode
	status.InstallReason = target.Reason
	asset, assetErr := resolveReleaseUpdateAsset(ctx, release, target, runtime.GOOS, runtime.GOARCH)
	if assetErr != nil {
		status.InstallMode = target.Mode
		status.InstallReason = "The update could not be verified: " + assetErr.Error()
		status.DownloadURL = status.ReleaseURL
		status.CanSelfUpdate = false
		status.CheckedAt = time.Now().UTC().Format(time.RFC3339)
		a.storeUpdateStatus(status)
		return status
	}
	applyUpdateAssetStatus(&status, target, asset)
	if status.DownloadURL == "" {
		status.DownloadURL = status.ReleaseURL
	}
	status.CheckedAt = time.Now().UTC().Format(time.RFC3339)
	a.storeUpdateStatus(status)
	return status
}

// resolveReleaseUpdateAsset prefers update-manifest.json and falls back to the
// release assets plus SHA256SUMS.txt.
func resolveReleaseUpdateAsset(ctx context.Context, release githubRelease, target installTarget, goos, goarch string) (updateAsset, error) {
	// When a signing key is pinned and this install can actually self-update,
	// trust only the signed manifest. Falling back to unsigned checksums would
	// let an attacker replace both the asset and SHA256SUMS.txt while bypassing
	// the pinned key.
	if updateSignatureRequired() && target.Mode != installModeGuided && target.Mode != installModeUnsupported {
		manifestAsset, ok := findReleaseAsset(release, updateManifestAssetName)
		if !ok {
			return updateAsset{}, errors.New("this release has no signed update manifest")
		}
		manifest, err := fetchUpdateManifest(ctx, release, manifestAsset)
		if err != nil {
			return updateAsset{}, err
		}
		candidate, ok := manifestAssetForMode(manifest, goos, goarch, target.Mode)
		if !ok {
			return updateAsset{}, errors.New("the signed update manifest has no asset for this platform")
		}
		if !assetCompatible(target.Mode, candidate.Name) {
			return updateAsset{}, errors.New("the signed update manifest asset cannot be installed automatically")
		}
		asset := updateAsset{
			Name:   candidate.Name,
			URL:    candidate.URL,
			SHA256: strings.ToLower(strings.TrimSpace(candidate.SHA256)),
			Size:   candidate.Size,
		}
		if asset.SHA256 == "" {
			return updateAsset{}, errors.New("the signed update manifest asset has no SHA-256 checksum")
		}
		if asset.Size == 0 {
			if match, ok := findReleaseAsset(release, asset.Name); ok {
				asset.Size = match.Size
			}
		}
		return asset, nil
	}

	var manifest *updateManifest
	if asset, ok := findReleaseAsset(release, updateManifestAssetName); ok {
		if parsed, err := fetchUpdateManifest(ctx, release, asset); err == nil {
			manifest = parsed
		}
	}
	var sums map[string]string
	if manifest == nil || manifestNeedsChecksums(manifest, goos, goarch, target.Mode) {
		sums = fetchReleaseChecksums(ctx, release)
	}
	asset := resolveUpdateAsset(release, manifest, goos, goarch, target.Mode, sums)
	if asset.Size == 0 {
		if match, ok := findReleaseAsset(release, asset.Name); ok {
			asset.Size = match.Size
		}
	}
	return asset, nil
}

func manifestNeedsChecksums(manifest *updateManifest, goos, goarch, mode string) bool {
	if manifest == nil {
		return true
	}
	candidate, ok := manifestAssetForMode(manifest, goos, goarch, mode)
	return !ok || strings.TrimSpace(candidate.SHA256) == ""
}

// manifestAssetForMode picks a manifest entry only when its asset can actually
// be applied in this install mode. A manifest that points linux-amd64 at a raw
// binary does not prevent AppImage installs from falling back to the published
// AppImage asset.
func manifestAssetForMode(manifest *updateManifest, goos, goarch, mode string) (updateManifestAsset, bool) {
	if manifest == nil {
		return updateManifestAsset{}, false
	}
	for _, key := range manifestAssetKeys(goos, goarch, mode) {
		candidate, ok := manifest.Assets[key]
		if !ok || strings.TrimSpace(candidate.URL) == "" {
			continue
		}
		if !assetCompatible(mode, candidate.Name) {
			continue
		}
		return candidate, true
	}
	return updateManifestAsset{}, false
}

func fetchUpdateManifest(ctx context.Context, release githubRelease, asset githubAsset) (*updateManifest, error) {
	if strings.TrimSpace(asset.BrowserDownloadURL) == "" {
		return nil, errors.New("manifest has no download URL")
	}
	body, err := fetchUpdateAsset(ctx, updateHTTPClient, asset.BrowserDownloadURL, 2<<20)
	if err != nil {
		return nil, err
	}
	manifest, err := parseUpdateManifest(body)
	if err != nil {
		return nil, err
	}
	if updateSignatureRequired() {
		signatureAsset, ok := findReleaseAsset(release, updateManifestSignatureName)
		if !ok || strings.TrimSpace(signatureAsset.BrowserDownloadURL) == "" {
			return nil, errors.New("update-manifest.json is signed but no signature asset was published")
		}
		signature, err := fetchUpdateAsset(ctx, updateHTTPClient, signatureAsset.BrowserDownloadURL, 64<<10)
		if err != nil {
			return nil, err
		}
		if err := verifyUpdateManifestSignature(body, signature); err != nil {
			return nil, err
		}
	}
	return &manifest, nil
}

func fetchReleaseChecksums(ctx context.Context, release githubRelease) map[string]string {
	asset, ok := findReleaseAsset(release, updateChecksumsAssetName)
	if !ok {
		return nil
	}
	body, err := fetchUpdateAsset(ctx, updateHTTPClient, asset.BrowserDownloadURL, 2<<20)
	if err != nil {
		return nil
	}
	return parseSHA256SUMS(body)
}

func fetchUpdateAsset(ctx context.Context, client *http.Client, rawURL string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", appName+"/"+Version)
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", response.StatusCode)
	}
	return io.ReadAll(io.LimitReader(response.Body, limit))
}

func findReleaseAsset(release githubRelease, name string) (githubAsset, bool) {
	for _, asset := range release.Assets {
		if strings.EqualFold(strings.TrimSpace(asset.Name), strings.TrimSpace(name)) {
			return asset, true
		}
	}
	return githubAsset{}, false
}

func parseUpdateManifest(data []byte) (updateManifest, error) {
	var manifest updateManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return updateManifest{}, errors.New("update-manifest.json is not valid JSON")
	}
	if len(manifest.Assets) == 0 {
		return updateManifest{}, errors.New("update-manifest.json has no assets")
	}
	for key, asset := range manifest.Assets {
		if strings.TrimSpace(asset.URL) == "" {
			return updateManifest{}, fmt.Errorf("update-manifest.json asset %q has no URL", key)
		}
		if strings.TrimSpace(asset.SHA256) != "" {
			decoded, err := hex.DecodeString(strings.TrimSpace(asset.SHA256))
			if err != nil || len(decoded) != 32 {
				return updateManifest{}, fmt.Errorf("update-manifest.json asset %q has an invalid SHA-256", key)
			}
		}
	}
	return manifest, nil
}

func parseSHA256SUMS(data []byte) map[string]string {
	sums := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 {
			continue
		}
		hash := strings.ToLower(fields[0])
		if len(hash) != 64 {
			continue
		}
		if _, err := hex.DecodeString(hash); err != nil {
			continue
		}
		name := strings.TrimPrefix(fields[len(fields)-1], "*")
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		sums[name] = hash
		sums[filepathBase(name)] = hash
	}
	return sums
}

func filepathBase(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if index := strings.LastIndex(path, "/"); index >= 0 {
		return path[index+1:]
	}
	return path
}

func findSHA256(sums map[string]string, name string) string {
	if len(sums) == 0 || strings.TrimSpace(name) == "" {
		return ""
	}
	if sum := sums[name]; sum != "" {
		return sum
	}
	return sums[filepathBase(name)]
}

// manifestAssetKeys lists the manifest keys to try in preference order for a
// platform and install mode. Extra mode-specific keys are accepted so releases
// can publish both a raw binary and an AppImage.
func manifestAssetKeys(goos, goarch, mode string) []string {
	base := goos + "-" + goarch
	switch {
	case mode == installModeAppImage:
		return []string{base + "-appimage", base}
	case mode == installModeAppBundle:
		return []string{base, goos + "-universal"}
	default:
		return []string{base}
	}
}

// resolveUpdateAsset is the pure selection step shared by tests and the network
// path. It never fetches anything.
func resolveUpdateAsset(release githubRelease, manifest *updateManifest, goos, goarch, mode string, sums map[string]string) updateAsset {
	asset := updateAsset{}
	if candidate, ok := manifestAssetForMode(manifest, goos, goarch, mode); ok {
		asset = updateAsset{
			Name:   candidate.Name,
			URL:    candidate.URL,
			SHA256: strings.ToLower(strings.TrimSpace(candidate.SHA256)),
			Size:   candidate.Size,
		}
	}
	if asset.URL == "" {
		if candidate, ok := selectReleaseAsset(release, goos, goarch, mode); ok {
			asset = candidate
		}
	}
	if asset.SHA256 == "" {
		asset.SHA256 = findSHA256(sums, asset.Name)
	}
	return asset
}

// selectUpdateDownload keeps the historical helper available for callers that
// only know the OS. New code should use resolveUpdateAsset for OS+arch+mode.
func selectUpdateDownload(release githubRelease, goos string) string {
	asset, _ := selectReleaseAsset(release, goos, runtime.GOARCH, "")
	return asset.URL
}

func selectReleaseAsset(release githubRelease, goos, goarch, mode string) (updateAsset, bool) {
	preferences := assetPreferences(goos, goarch, mode)
	for _, preference := range preferences {
		for _, asset := range release.Assets {
			name := strings.TrimSpace(asset.Name)
			if name == "" || strings.TrimSpace(asset.BrowserDownloadURL) == "" {
				continue
			}
			if preference(name) {
				return updateAsset{Name: name, URL: asset.BrowserDownloadURL, Size: asset.Size}, true
			}
		}
	}
	return updateAsset{}, false
}

type assetPreference func(name string) bool

func assetPreferences(goos, goarch, mode string) []assetPreference {
	lower := func(name string) string { return strings.ToLower(name) }
	archOK := func(name string) bool { return assetArchMatches(name, goarch) }

	isRawExe := func(name string) bool {
		value := lower(name)
		return strings.HasSuffix(value, ".exe") && !strings.Contains(value, "setup") && !strings.Contains(value, "unins")
	}
	isWindowsZip := func(name string) bool {
		value := lower(name)
		return strings.Contains(value, "windows") && strings.HasSuffix(value, ".zip")
	}
	isSetup := func(name string) bool { return strings.Contains(lower(name), "setup") }
	isAppImage := func(name string) bool { return strings.HasSuffix(lower(name), ".appimage") }
	isLinuxArchive := func(name string) bool {
		value := lower(name)
		return strings.Contains(value, "linux") && (strings.HasSuffix(value, ".tar.gz") || strings.HasSuffix(value, ".tgz") || strings.HasSuffix(value, ".zip"))
	}
	isLinuxRaw := func(name string) bool {
		value := lower(name)
		if !strings.Contains(value, "linux") || strings.Contains(value, ".tar.gz") || strings.Contains(value, ".tgz") || strings.Contains(value, ".zip") || strings.Contains(value, ".deb") || strings.HasSuffix(value, ".appimage") {
			return false
		}
		return archOK(name)
	}
	isMacArchive := func(name string) bool {
		value := lower(name)
		return (strings.Contains(value, "macos") || strings.Contains(value, "darwin") || strings.Contains(value, "osx")) &&
			(strings.HasSuffix(value, ".tar.gz") || strings.HasSuffix(value, ".tgz") || strings.HasSuffix(value, ".zip"))
	}
	isMacRaw := func(name string) bool {
		value := lower(name)
		return (strings.Contains(value, "macos") || strings.Contains(value, "darwin")) && !strings.Contains(value, ".tar.gz") && !strings.Contains(value, ".zip") && archOK(name)
	}

	switch goos {
	case "windows":
		switch mode {
		case installModeInPlace, "":
			return []assetPreference{
				func(name string) bool { return lower(name) == "twitchlivenotifier.exe" },
				func(name string) bool { return isRawExe(name) && archOK(name) },
				isWindowsZip,
				isSetup,
			}
		}
		return []assetPreference{isSetup, isWindowsZip, isRawExe}
	case "linux":
		switch mode {
		case installModeAppImage:
			return []assetPreference{func(name string) bool { return isAppImage(name) && archOK(name) }}
		case installModeInPlace:
			return []assetPreference{isLinuxRaw, isLinuxArchive}
		default:
			return []assetPreference{
				func(name string) bool { return isAppImage(name) && archOK(name) },
				isLinuxRaw,
				isLinuxArchive,
				func(name string) bool { return strings.HasSuffix(lower(name), ".deb") && archOK(name) },
			}
		}
	case "darwin":
		return []assetPreference{isMacArchive, isMacRaw}
	default:
		if mode == installModeInPlace {
			return []assetPreference{isLinuxRaw, isLinuxArchive}
		}
		return []assetPreference{isLinuxArchive, isLinuxRaw}
	}
}

func assetArchMatches(name, goarch string) bool {
	value := strings.ToLower(name)
	hasARM := strings.Contains(value, "arm64") || strings.Contains(value, "aarch64")
	switch strings.ToLower(goarch) {
	case "arm64":
		return hasARM
	case "amd64", "x86_64":
		if hasARM {
			return false
		}
		return true
	default:
		return true
	}
}

// assetCompatible rejects installers that cannot be swapped in place.
func assetCompatible(mode, name string) bool {
	value := strings.ToLower(strings.TrimSpace(name))
	switch mode {
	case installModeAppImage:
		return strings.HasSuffix(value, ".appimage")
	case installModeAppBundle:
		return strings.HasSuffix(value, ".tar.gz") || strings.HasSuffix(value, ".tgz") || strings.HasSuffix(value, ".zip")
	case installModeInPlace:
		if strings.Contains(value, "setup") || strings.HasSuffix(value, ".deb") || strings.HasSuffix(value, ".msi") {
			return false
		}
		return value != ""
	default:
		return false
	}
}

func applyUpdateAssetStatus(status *UpdateStatus, target installTarget, asset updateAsset) {
	status.AssetName = asset.Name
	status.AssetSize = asset.Size
	status.AssetSHA256 = strings.ToLower(strings.TrimSpace(asset.SHA256))
	status.CanSelfUpdate = false
	status.InstallMode = target.Mode
	status.InstallReason = target.Reason

	if target.Mode == installModeGuided || target.Mode == installModeUnsupported {
		status.InstallMode = target.Mode
		status.InstallReason = firstNonEmpty(target.Reason, "This installation cannot update itself.")
		if asset.URL != "" {
			status.DownloadURL = asset.URL
		}
		return
	}
	if asset.URL == "" {
		status.InstallMode = target.Mode
		status.InstallReason = "No download is available for this platform yet. Open the release page to update manually."
		return
	}
	if !assetCompatible(target.Mode, asset.Name) {
		status.InstallMode = installModeGuided
		status.InstallReason = "The release only provides an installer or package that must be run manually."
		status.DownloadURL = asset.URL
		return
	}
	if status.AssetSHA256 == "" {
		status.InstallMode = installModeGuided
		status.InstallReason = "The release has no SHA-256 checksum for this download, so automatic installation is disabled."
		status.DownloadURL = asset.URL
		return
	}
	if !target.Writable {
		status.InstallMode = installModeGuided
		status.InstallReason = "The application folder is not writable. Move the app to a user-writable folder or update manually."
		status.DownloadURL = asset.URL
		return
	}
	status.CanSelfUpdate = true
	status.DownloadURL = asset.URL
}

func (a *App) storeUpdateStatus(status UpdateStatus) {
	a.mu.Lock()
	if status.InstallJob == nil {
		status.InstallJob = a.updateStatus.InstallJob
	}
	a.updateStatus = status
	a.mu.Unlock()
}

func (a *App) storeInstallJob(job UpdateJob) {
	a.mu.Lock()
	copy := job
	a.updateStatus.InstallJob = &copy
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
