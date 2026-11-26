//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	windowsProcessSynchronize = 0x00100000
	windowsWaitTimeout        = 0x00000102
	windowsCreateNoWindow     = 0x08000000
	windowsCreateProcessGroup = 0x00000200
)

var (
	procOpenProcess         = kernel32.NewProc("OpenProcess")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
)

// applyPlatformUpdate stages a detached helper instead of renaming the running
// .exe. Windows can report a sharing violation when the image is renamed or
// replaced while it is executing, so the helper waits for this process to exit
// and only then swaps in the verified binary.
func applyPlatformUpdate(plan updatePlan) (applyResult, error) {
	if plan.Mode != installModeInPlace {
		return applyResult{}, fmt.Errorf("automatic updates are not supported in %s mode", plan.Mode)
	}
	if strings.TrimSpace(plan.NewBinary) == "" {
		return applyResult{}, errors.New("the update has no new binary")
	}
	if strings.TrimSpace(plan.WorkDir) == "" {
		return applyResult{}, errors.New("the update is missing its staging folder")
	}
	scriptPath := filepath.Join(filepath.Dir(plan.Target), fmt.Sprintf(".twitch-live-update-%d.cmd", os.Getpid()))
	script := windowsUpdateHelperScript(plan.Target, plan.NewBinary, plan.WorkDir, plan.Args, os.Getpid(), plan.Restart)
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		return applyResult{}, fmt.Errorf("could not stage the Windows update helper: %w", err)
	}
	command := exec.Command("cmd.exe", "/c", scriptPath)
	command.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windowsCreateNoWindow | windowsCreateProcessGroup,
	}
	command.Stdin = nil
	command.Stdout = nil
	command.Stderr = nil
	if err := command.Start(); err != nil {
		_ = os.Remove(scriptPath)
		return applyResult{}, fmt.Errorf("could not start the Windows update helper. Download the installer and update manually: %w", err)
	}
	message := "Update staged. It will be applied when the app quits."
	if plan.Restart {
		message = "Update staged. The app will restart after it quits."
	}
	return applyResult{
		NeedsExit:       true,
		HandlesRelaunch: plan.Restart,
		KeepWorkDir:     true,
		Message:         message,
	}, nil
}

func processRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	handle, _, _ := procOpenProcess.Call(windowsProcessSynchronize, 0, uintptr(pid))
	if handle == 0 {
		return false
	}
	defer procCloseHandle.Call(handle)
	result, _, _ := procWaitForSingleObject.Call(handle, 0)
	return result == windowsWaitTimeout
}
