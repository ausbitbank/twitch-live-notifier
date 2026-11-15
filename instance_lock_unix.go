//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

// acquireInstanceLock takes an advisory file lock. The kernel releases it when
// the process exits, so a crash cannot leave a permanent stale lock. The lock
// file itself is intentionally kept: unlinking it while another process holds
// an open descriptor could let a third process acquire a new inode.
func acquireInstanceLock(path string) (func(), bool, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, false, fmt.Errorf("could not create the settings folder: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, false, fmt.Errorf("could not open the instance lock: %w", err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return func() {}, true, nil
		}
		return nil, false, fmt.Errorf("could not lock the settings file: %w", err)
	}

	// Best-effort diagnostics for operators. Failure to write does not
	// invalidate the lock.
	_ = file.Truncate(0)
	_, _ = file.Seek(0, 0)
	_, _ = file.WriteString(strconv.Itoa(os.Getpid()) + "\n")
	_ = file.Sync()

	released := false
	return func() {
		if released {
			return
		}
		released = true
		_ = file.Close()
	}, false, nil
}
