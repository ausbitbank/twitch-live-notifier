package main

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestApplyInPlaceToTempFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the Windows apply path renames the running executable")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "twitch-live-notifier")
	replacement := filepath.Join(dir, "new-binary")
	if err := os.WriteFile(target, []byte("old-version"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(replacement, []byte("new-version"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := applyUpdatePlan(updatePlan{
		Mode:      installModeInPlace,
		Target:    target,
		NewBinary: replacement,
		Restart:   false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.NeedsExit {
		t.Fatalf("unexpected exit result: %+v", result)
	}
	newData, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(newData) != "new-version" {
		t.Fatalf("target = %q", newData)
	}
	backup, err := os.ReadFile(target + ".old")
	if err != nil {
		t.Fatal(err)
	}
	if string(backup) != "old-version" {
		t.Fatalf("backup = %q", backup)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Fatalf("target is not executable: %v", info.Mode())
	}
}

func TestApplyGuidedModeRefuses(t *testing.T) {
	if _, err := applyUpdatePlan(updatePlan{Mode: installModeGuided, Target: "/tmp/x"}); err == nil {
		t.Fatal("guided mode should never be applied")
	}
}

func TestExtractTarGzUpdatePayload(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "Linux-amd64.tar.gz")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)
	contents := []byte("#!/bin/sh\necho test\n")
	if err := tarWriter.WriteHeader(&tar.Header{Name: "pkg/TwitchLiveNotifier-linux-amd64", Mode: 0o755, Size: int64(len(contents))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.WriteHeader(&tar.Header{Name: "pkg/README.md", Mode: 0o644, Size: 3}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write([]byte("doc")); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	prepared, err := prepareUpdatePayload("TwitchLiveNotifier-Linux-amd64.tar.gz", archive, filepath.Join(dir, "work"), installTarget{Mode: installModeInPlace}, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(prepared.Binary, "TwitchLiveNotifier-linux-amd64") {
		t.Fatalf("prepared binary = %q", prepared.Binary)
	}
	data, err := os.ReadFile(prepared.Binary)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(contents) {
		t.Fatalf("binary contents = %q", data)
	}
}

func TestWindowsUpdateHelperScript(t *testing.T) {
	script := windowsUpdateHelperScript(
		`C:\Users\A\AppData\Local\Programs\TwitchLiveNotifier\TwitchLiveNotifier.exe`,
		`C:\Users\A\AppData\Local\Programs\TwitchLiveNotifier\.twitch-live-update-1\download`,
		`C:\Users\A\AppData\Local\Programs\TwitchLiveNotifier\.twitch-live-update-1`,
		[]string{"--background", "--config", `C:\Users\A\100% config.json`},
		4242,
		true,
	)
	checks := []string{
		"Wait-Process -Id 4242",
		`move /Y "C:\Users\A\AppData\Local\Programs\TwitchLiveNotifier\TwitchLiveNotifier.exe"`,
		"TwitchLiveNotifier.exe.old",
		`start "" "C:\Users\A\AppData\Local\Programs\TwitchLiveNotifier\TwitchLiveNotifier.exe"`,
		`"--background"`,
		`C:\Users\A\100%% config.json`,
	}
	for _, check := range checks {
		if !strings.Contains(script, check) {
			t.Errorf("Windows helper script missing %q:\n%s", check, script)
		}
	}
	if strings.Contains(script, "start \"\"") == false {
		t.Fatal("Windows helper should relaunch when requested")
	}
	noRelaunch := windowsUpdateHelperScript("C:\\a.exe", "C:\\new.exe", "C:\\work", nil, 1, false)
	if strings.Contains(noRelaunch, "start \"\"") {
		t.Fatal("Windows helper should not relaunch when restart is disabled")
	}
}

func TestWriteBundleHelper(t *testing.T) {
	dir := t.TempDir()
	script, err := writeBundleHelper(dir, "/Applications/Twitch Live to Discord.app", filepath.Join(dir, "new.app"), 123, true)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "parent=123") || !strings.Contains(text, "open -n") {
		t.Fatalf("script = %s", text)
	}
	info, err := os.Stat(script)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Fatalf("helper is not executable: %v", info.Mode())
	}
}
