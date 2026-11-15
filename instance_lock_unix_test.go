//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestAcquireInstanceLockIsExclusive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "instance.lock")
	release, already, err := acquireInstanceLock(path)
	if err != nil {
		t.Fatalf("first acquire returned error: %v", err)
	}
	if already {
		t.Fatal("first acquire reported an existing instance")
	}

	_, already, err = acquireInstanceLock(path)
	if err != nil {
		t.Fatalf("second acquire returned error: %v", err)
	}
	if !already {
		t.Fatal("second acquire did not detect the running instance")
	}

	release()
	releaseAgain, already, err := acquireInstanceLock(path)
	if err != nil {
		t.Fatalf("reacquire returned error: %v", err)
	}
	if already {
		t.Fatal("reacquire still reported an existing instance after release")
	}
	releaseAgain()
}

func TestInstanceLockRecoversFromStaleFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "instance.lock")
	if err := os.WriteFile(path, []byte("999999\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	release, already, err := acquireInstanceLock(path)
	if err != nil {
		t.Fatalf("acquire returned error: %v", err)
	}
	if already {
		t.Fatal("stale lock file blocked acquisition")
	}
	release()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read lock file: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid != os.Getpid() {
		t.Fatalf("lock file = %q, want current pid %d", string(data), os.Getpid())
	}
}

func TestInstanceLockPathUsesConfigDirectory(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "data", "config.json")
	want := filepath.Join(filepath.Dir(configPath), "instance.lock")
	if got := instanceLockPath(configPath); got != want {
		t.Fatalf("instanceLockPath = %q, want %q", got, want)
	}
}
