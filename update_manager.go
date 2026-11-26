package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	updateJobDownloading = "downloading"
	updateJobVerifying   = "verifying"
	updateJobApplying    = "applying"
	updateJobRestarting  = "restarting"
	updateJobDone        = "done"
	updateJobError       = "error"
	updateJobIdle        = "idle"
)

var errUpdateJobActive = errors.New("an update is already in progress")

// UpdateJob is the single in-flight update operation exposed to the dashboard.
type UpdateJob struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
	Bytes      int64  `json:"bytes"`
	Total      int64  `json:"total"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
	StartedAt  string `json:"started_at,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
}

// UpdateManager downloads and applies one update at a time. It is deliberately
// conservative: the cached status decides whether self-update is allowed, and
// every downloaded asset is checksum-verified before it can be applied.
type UpdateManager struct {
	app        *App
	httpClient *http.Client
	exitFunc   func(int)

	mu     sync.Mutex
	active bool
	job    UpdateJob
}

func NewUpdateManager(app *App) *UpdateManager {
	return &UpdateManager{
		app:        app,
		httpClient: &http.Client{Timeout: 30 * time.Minute},
		exitFunc:   os.Exit,
	}
}

// Current returns the latest job. A manager that has never run reports idle.
func (m *UpdateManager) Current() UpdateJob {
	m.mu.Lock()
	defer m.mu.Unlock()
	job := m.job
	if strings.TrimSpace(job.Status) == "" {
		job.Status = updateJobIdle
	}
	return job
}

func (m *UpdateManager) IsActive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active
}

// Start begins a download/install job. It returns errUpdateJobActive if any
// job is already running.
func (m *UpdateManager) Start(restart bool) (UpdateJob, error) {
	m.mu.Lock()
	if m.active {
		m.mu.Unlock()
		return UpdateJob{Status: updateJobIdle}, errUpdateJobActive
	}
	m.mu.Unlock()

	status := m.app.cachedUpdateStatus()
	if !status.CanSelfUpdate || strings.TrimSpace(status.DownloadURL) == "" {
		reason := strings.TrimSpace(status.InstallReason)
		if reason == "" {
			reason = "This installation cannot update itself."
		}
		return UpdateJob{Status: updateJobIdle}, errors.New(reason)
	}
	if strings.TrimSpace(status.AssetSHA256) == "" {
		return UpdateJob{Status: updateJobIdle}, errors.New("the update has no SHA-256 checksum")
	}
	target := inspectInstallTarget()
	if target.Mode != installModeInPlace && target.Mode != installModeAppImage && target.Mode != installModeAppBundle {
		reason := strings.TrimSpace(target.Reason)
		if reason == "" {
			reason = "This installation cannot update itself."
		}
		return UpdateJob{Status: updateJobIdle}, errors.New(reason)
	}

	m.mu.Lock()
	if m.active {
		m.mu.Unlock()
		return UpdateJob{Status: updateJobIdle}, errUpdateJobActive
	}
	now := time.Now().UTC().Format(time.RFC3339)
	job := UpdateJob{
		ID:        newID(),
		Status:    updateJobDownloading,
		Message:   "Downloading " + status.AssetName,
		StartedAt: now,
		Total:     status.AssetSize,
	}
	m.active = true
	m.job = job
	m.mu.Unlock()
	m.app.storeInstallJob(job)

	go m.run(job.ID, status, target, restart)
	return job, nil
}

func (m *UpdateManager) run(id string, status UpdateStatus, target installTarget, restart bool) {
	workDir := ""
	keepWorkDir := false
	defer func() {
		if workDir != "" && !keepWorkDir {
			_ = os.RemoveAll(workDir)
		}
	}()

	workDir, err := os.MkdirTemp(filepath.Dir(target.Target), ".twitch-live-update-")
	if err != nil {
		m.fail(id, "Could not create a temporary folder next to the app: "+err.Error())
		return
	}

	downloadPath := filepath.Join(workDir, "download")
	m.setJob(id, func(job *UpdateJob) {
		job.Status = updateJobDownloading
		job.Message = "Downloading " + status.AssetName
		job.Bytes = 0
		job.Total = status.AssetSize
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	err = downloadFile(ctx, m.httpClient, status.DownloadURL, downloadPath, status.AssetSize, func(bytes, total int64) {
		m.setJob(id, func(job *UpdateJob) {
			job.Bytes = bytes
			job.Total = total
			job.Progress = progressPercent(bytes, total)
		})
	})
	if err != nil {
		m.fail(id, "Download failed: "+err.Error())
		return
	}

	m.setJob(id, func(job *UpdateJob) {
		job.Status = updateJobVerifying
		job.Progress = 100
		job.Message = "Verifying the download"
	})
	if err := verifyFileSHA256(downloadPath, status.AssetSHA256); err != nil {
		m.fail(id, err.Error())
		return
	}

	prepared, err := prepareUpdatePayload(status.AssetName, downloadPath, workDir, target, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		m.fail(id, "Could not prepare the update: "+err.Error())
		return
	}
	if prepared.Binary != "" {
		if err := os.Chmod(prepared.Binary, 0o755); err != nil {
			m.fail(id, "Could not make the new version executable: "+err.Error())
			return
		}
		m.setJob(id, func(job *UpdateJob) { job.Message = "Testing the new version" })
		if err := preflightBinary(prepared.Binary); err != nil {
			m.fail(id, err.Error())
			return
		}
	}

	m.setJob(id, func(job *UpdateJob) {
		job.Status = updateJobApplying
		job.Message = "Installing the update"
	})
	result, err := applyUpdatePlan(updatePlan{
		Mode:      target.Mode,
		Target:    target.Target,
		NewBinary: prepared.Binary,
		NewBundle: prepared.Bundle,
		WorkDir:   workDir,
		Restart:   restart,
		Args:      os.Args[1:],
	})
	if err != nil {
		m.fail(id, "Could not install the update: "+err.Error())
		return
	}
	keepWorkDir = result.KeepWorkDir

	if result.NeedsExit && restart {
		m.setJob(id, func(job *UpdateJob) {
			job.Status = updateJobRestarting
			job.Message = firstNonEmpty(result.Message, "Restarting the app")
		})
		if !result.HandlesRelaunch {
			restartPath := target.Executable
			if target.Mode == installModeAppImage {
				restartPath = target.Target
			}
			if err := startRestartedProcess(restartPath, os.Args[1:]); err != nil {
				m.fail(id, "The update was installed, but the app could not restart: "+err.Error())
				return
			}
		}
		time.Sleep(250 * time.Millisecond)
		m.finish(id, updateJobDone, firstNonEmpty(result.Message, "Restarting the app"))
		m.exitFunc(0)
		return
	}

	message := result.Message
	if message == "" {
		message = "Update installed. Restart the app to use the new version."
	}
	m.finish(id, updateJobDone, message)
}

func (m *UpdateManager) setJob(id string, mutate func(*UpdateJob)) {
	m.mu.Lock()
	if m.job.ID != id {
		m.mu.Unlock()
		return
	}
	mutate(&m.job)
	job := m.job
	m.mu.Unlock()
	m.app.storeInstallJob(job)
}

func (m *UpdateManager) fail(id, message string) {
	m.mu.Lock()
	if m.job.ID != id {
		m.mu.Unlock()
		return
	}
	m.job.Status = updateJobError
	m.job.Error = message
	m.job.Message = "Update failed"
	m.job.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	m.active = false
	job := m.job
	m.mu.Unlock()
	m.app.storeInstallJob(job)
}

func (m *UpdateManager) finish(id, state, message string) {
	m.mu.Lock()
	if m.job.ID != id {
		m.mu.Unlock()
		return
	}
	m.job.Status = state
	m.job.Message = message
	m.job.Progress = 100
	m.job.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	m.active = false
	job := m.job
	m.mu.Unlock()
	m.app.storeInstallJob(job)
}

// inspectInstallTarget discovers how the running executable is installed.
func inspectInstallTarget() installTarget {
	executable, err := os.Executable()
	if err != nil {
		return installTarget{Mode: installModeUnsupported, Reason: "Could not determine the application path."}
	}
	if resolved, resolveErr := filepath.EvalSymlinks(executable); resolveErr == nil {
		executable = resolved
	}
	appImage := os.Getenv("APPIMAGE")
	if strings.TrimSpace(appImage) != "" {
		if resolved, resolveErr := filepath.EvalSymlinks(appImage); resolveErr == nil {
			appImage = resolved
		}
	}
	target := installTargetFor(executable, runtime.GOOS, runtime.GOARCH, appImage, runningInDocker())
	target.Executable = executable
	writablePath := target.Target
	if target.Mode == installModeAppBundle {
		writablePath = filepath.Dir(target.Target)
	}
	target.Writable = targetWritable(writablePath)
	if (target.Mode == installModeInPlace || target.Mode == installModeAppImage || target.Mode == installModeAppBundle) && !target.Writable {
		target.Mode = installModeGuided
		target.Reason = "The application folder is not writable. Move the app to a user-writable folder or update manually."
	}
	return target
}

func installTargetFor(executable, goos, goarch, appImage string, docker bool) installTarget {
	target := installTarget{Executable: executable}
	if docker {
		target.Mode = installModeGuided
		target.Reason = "Docker installations update with `docker compose pull` and `docker compose up -d`."
		return target
	}
	lower := strings.ToLower(strings.ReplaceAll(filepath.ToSlash(executable), "\\", "/"))
	switch goos {
	case "windows":
		target.Target = executable
		if strings.Contains(lower, "/program files") || strings.Contains(lower, "/windows/") {
			target.Mode = installModeGuided
			target.Reason = "This copy is installed system-wide. Download the installer to update it."
			return target
		}
		target.Mode = installModeInPlace
	case "linux":
		if strings.TrimSpace(appImage) != "" {
			target.Target = appImage
			target.Mode = installModeAppImage
			return target
		}
		if strings.HasSuffix(lower, ".appimage") {
			target.Target = executable
			target.Mode = installModeAppImage
			return target
		}
		target.Target = executable
		if strings.HasPrefix(lower, "/usr/") || strings.HasPrefix(lower, "/opt/") || strings.HasPrefix(lower, "/snap/") {
			target.Mode = installModeGuided
			target.Reason = "This copy is installed as a system package. Update it with your package manager or download the release manually."
			return target
		}
		target.Mode = installModeInPlace
	case "darwin":
		if index := strings.Index(executable, ".app/Contents/MacOS/"); index >= 0 {
			target.Target = executable[:index+len(".app")]
			target.Mode = installModeAppBundle
			return target
		}
		target.Target = executable
		target.Mode = installModeInPlace
	default:
		target.Mode = installModeUnsupported
		target.Reason = "Automatic updates are not supported on " + goos + "/" + goarch + "."
	}
	return target
}

func runningInDocker() bool {
	if value := strings.ToLower(strings.TrimSpace(os.Getenv("TWITCH_LIVE_DOCKER"))); value != "" {
		return value == "1" || value == "true" || value == "yes"
	}
	if value := strings.ToLower(strings.TrimSpace(os.Getenv("container"))); value == "docker" {
		return true
	}
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func targetWritable(target string) bool {
	if strings.TrimSpace(target) == "" {
		return false
	}
	directory := target
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		directory = filepath.Dir(target)
	} else if err != nil {
		directory = filepath.Dir(target)
	}
	file, err := os.CreateTemp(directory, ".twitch-live-write-test-*")
	if err != nil {
		return false
	}
	name := file.Name()
	_ = file.Close()
	_ = os.Remove(name)
	return true
}

type preparedUpdate struct {
	Binary string
	Bundle string
}

// prepareUpdatePayload turns a downloaded file into the binary or bundle to
// install. Archive extraction stays inside workDir.
func prepareUpdatePayload(assetName, downloadPath, workDir string, target installTarget, goos, goarch string) (preparedUpdate, error) {
	name := strings.ToLower(assetName)
	if target.Mode == installModeAppBundle {
		if !isArchiveName(name) {
			return preparedUpdate{}, errors.New("the macOS release asset is not an app bundle archive")
		}
		extracted := filepath.Join(workDir, "extracted")
		if err := extractUpdateArchive(downloadPath, extracted, name); err != nil {
			return preparedUpdate{}, err
		}
		bundle, err := findAppBundle(extracted, goarch)
		if err != nil {
			return preparedUpdate{}, err
		}
		return preparedUpdate{
			Binary: filepath.Join(bundle, "Contents", "MacOS", "TwitchLiveNotifier"),
			Bundle: bundle,
		}, nil
	}
	if isArchiveName(name) {
		extracted := filepath.Join(workDir, "extracted")
		if err := extractUpdateArchive(downloadPath, extracted, name); err != nil {
			return preparedUpdate{}, err
		}
		binary, err := findExtractedBinary(extracted, goarch)
		if err != nil {
			return preparedUpdate{}, err
		}
		return preparedUpdate{Binary: binary}, nil
	}
	return preparedUpdate{Binary: downloadPath}, nil
}

func isArchiveName(name string) bool {
	return strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz") || strings.HasSuffix(name, ".zip")
}

func findAppBundle(root, goarch string) (string, error) {
	var candidates []string
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".app") {
			candidates = append(candidates, path)
			return filepath.SkipDir
		}
		return nil
	})
	if len(candidates) == 0 {
		return "", errors.New("no .app bundle was found in the archive")
	}
	for _, candidate := range candidates {
		if strings.Contains(strings.ToLower(candidate), strings.ToLower(goarch)) {
			return candidate, nil
		}
	}
	return candidates[0], nil
}

func findExtractedBinary(root, goarch string) (string, error) {
	best := ""
	bestScore := -1
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			lowerName := strings.ToLower(entry.Name())
			if strings.HasSuffix(lowerName, ".app") || lowerName == "docs" {
				return filepath.SkipDir
			}
			return nil
		}
		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".png") ||
			strings.HasSuffix(name, ".rb") || strings.HasSuffix(name, ".deb") || strings.HasSuffix(name, ".appimage") ||
			strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".gz") || strings.HasSuffix(name, ".json") {
			return nil
		}
		if !binaryArchMatches(name, goarch) {
			return nil
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return nil
		}
		score := 0
		if strings.Contains(name, "twitchlivenotifier") || strings.Contains(name, "twitch-live-notifier") {
			score += 10
		}
		if info.Mode()&0o111 != 0 {
			score += 3
		}
		if info.Size() > 1<<20 {
			score += 2
		}
		if score > bestScore {
			bestScore = score
			best = path
		}
		return nil
	})
	if best == "" {
		return "", errors.New("no application binary was found in the archive")
	}
	return best, nil
}

func binaryArchMatches(name, goarch string) bool {
	lower := strings.ToLower(name)
	hasARM := strings.Contains(lower, "arm64") || strings.Contains(lower, "aarch64")
	switch strings.ToLower(goarch) {
	case "arm64":
		return !strings.Contains(lower, "amd64") && !strings.Contains(lower, "x64")
	case "amd64":
		return !hasARM
	default:
		return true
	}
}

func extractUpdateArchive(path, destination, name string) error {
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	if strings.HasSuffix(name, ".zip") {
		return extractZip(path, destination)
	}
	return extractTarGz(path, destination)
}

func extractTarGz(path, destination string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return errors.New("the update archive is not valid gzip")
	}
	defer gzipReader.Close()
	reader := tar.NewReader(gzipReader)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return errors.New("the update archive is not valid tar")
		}
		target, err := safeArchivePath(destination, header.Name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			mode := os.FileMode(header.Mode) & 0o777
			if mode == 0 {
				mode = 0o644
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, reader); err != nil {
				out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractZip(path, destination string) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return errors.New("the update archive is not a valid zip file")
	}
	defer reader.Close()
	for _, file := range reader.File {
		target, err := safeArchivePath(destination, file.Name)
		if err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		input, err := file.Open()
		if err != nil {
			return err
		}
		mode := file.Mode() & 0o777
		if mode == 0 {
			mode = 0o644
		}
		output, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err != nil {
			input.Close()
			return err
		}
		_, copyErr := io.Copy(output, input)
		input.Close()
		output.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

func safeArchivePath(destination, name string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", errors.New("the update archive contains an unsafe path")
	}
	target := filepath.Join(destination, clean)
	root := filepath.Clean(destination) + string(filepath.Separator)
	if !strings.HasPrefix(filepath.Clean(target)+string(filepath.Separator), root) {
		return "", errors.New("the update archive contains an unsafe path")
	}
	return target, nil
}

type progressWriter struct {
	writer  io.Writer
	written int64
	total   int64
	onWrite func(written, total int64)
}

func (p *progressWriter) Write(data []byte) (int, error) {
	written, err := p.writer.Write(data)
	p.written += int64(written)
	if p.onWrite != nil {
		p.onWrite(p.written, p.total)
	}
	return written, err
}

func progressPercent(written, total int64) int {
	if total <= 0 {
		return 0
	}
	if written >= total {
		return 100
	}
	percent := int(written * 100 / total)
	if percent < 0 {
		return 0
	}
	if percent > 99 {
		return 99
	}
	return percent
}

func downloadFile(ctx context.Context, client *http.Client, rawURL, destination string, totalHint int64, onProgress func(bytes, total int64)) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", appName+"/"+Version)
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", response.StatusCode)
	}
	total := response.ContentLength
	if total <= 0 {
		total = totalHint
	}
	if onProgress != nil {
		onProgress(0, total)
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	writer := &progressWriter{writer: file, total: total, onWrite: onProgress}
	_, copyErr := io.Copy(writer, response.Body)
	syncErr := file.Sync()
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(destination)
		return copyErr
	}
	if syncErr != nil {
		_ = os.Remove(destination)
		return syncErr
	}
	if closeErr != nil {
		_ = os.Remove(destination)
		return closeErr
	}
	if onProgress != nil {
		onProgress(writer.written, total)
	}
	return nil
}

func verifyFileSHA256(path, expected string) error {
	expected = strings.ToLower(strings.TrimSpace(expected))
	if expected == "" {
		return errors.New("the update has no SHA-256 checksum")
	}
	file, err := os.Open(path)
	if err != nil {
		return errors.New("could not read the downloaded update")
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return errors.New("could not read the downloaded update")
	}
	actual := hex.EncodeToString(hasher.Sum(nil))
	if actual != expected {
		return errors.New("the downloaded update failed its SHA-256 checksum")
	}
	return nil
}

// preflightBinary runs the new executable with --version. Windows GUI
// subsystem builds sometimes exit without output, so an empty but successful
// run is accepted.
func preflightBinary(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, path, "--version")
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	err := command.Run()
	if ctx.Err() != nil {
		return errors.New("the new version did not respond to --version")
	}
	if err != nil {
		if runtime.GOOS == "windows" && output.Len() == 0 {
			return nil
		}
		return fmt.Errorf("the new version failed its startup check: %v", err)
	}
	return nil
}
