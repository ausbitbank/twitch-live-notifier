package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDownloadFileProgress(t *testing.T) {
	payload := bytes.Repeat([]byte("update-data-"), 8192)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	destination := filepath.Join(t.TempDir(), "download.bin")
	var lastBytes, lastTotal int64
	events := 0
	err := downloadFile(context.Background(), http.DefaultClient, server.URL, destination, 0, func(written, total int64) {
		events++
		lastBytes = written
		lastTotal = total
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, payload) || lastBytes != int64(len(payload)) || lastTotal != int64(len(payload)) {
		t.Fatalf("download mismatch: bytes=%d total=%d len=%d", lastBytes, lastTotal, len(data))
	}
	if events < 2 {
		t.Fatalf("expected progress events, got %d", events)
	}
	sum := sha256.Sum256(payload)
	if err := verifyFileSHA256(destination, hex.EncodeToString(sum[:])); err != nil {
		t.Fatalf("checksum verify: %v", err)
	}
	if err := verifyFileSHA256(destination, strings.Repeat("0", 64)); err == nil {
		t.Fatal("checksum mismatch should fail")
	}
}

func TestDownloadFileHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	err := downloadFile(context.Background(), http.DefaultClient, server.URL, filepath.Join(t.TempDir(), "x"), 0, nil)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateManagerIdleAndErrors(t *testing.T) {
	app := newTestApp(t)
	manager := app.updates
	if manager == nil {
		t.Fatal("NewApp did not create an update manager")
	}
	if job := manager.Current(); job.Status != updateJobIdle {
		t.Fatalf("idle job = %+v", job)
	}

	app.storeUpdateStatus(UpdateStatus{CanSelfUpdate: false, InstallReason: "manual only"})
	if _, err := manager.Start(false); err == nil || !strings.Contains(err.Error(), "manual only") {
		t.Fatalf("expected manual-only error, got %v", err)
	}

	app.storeUpdateStatus(UpdateStatus{
		CanSelfUpdate: true,
		DownloadURL:   "https://example.test/download",
		AssetSHA256:   strings.Repeat("a", 64),
		AssetName:     "app",
	})
	manager.mu.Lock()
	manager.active = true
	manager.mu.Unlock()
	if _, err := manager.Start(false); err != errUpdateJobActive {
		t.Fatalf("expected active job error, got %v", err)
	}
}

func TestPrepareUpdatePayloadRaw(t *testing.T) {
	workDir := t.TempDir()
	download := filepath.Join(workDir, "download")
	if err := os.WriteFile(download, []byte("binary"), 0o600); err != nil {
		t.Fatal(err)
	}
	prepared, err := prepareUpdatePayload("TwitchLiveNotifier-linux-amd64", download, workDir, installTarget{Mode: installModeInPlace}, "linux", "amd64")
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Binary != download {
		t.Fatalf("prepared = %+v", prepared)
	}
}
