//go:build darwin

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func applyPlatformUpdate(plan updatePlan) (applyResult, error) {
	switch plan.Mode {
	case installModeInPlace:
		if strings.TrimSpace(plan.NewBinary) == "" {
			return applyResult{}, errors.New("the update has no new binary")
		}
		if err := replaceExecutable(plan.Target, plan.NewBinary); err != nil {
			return applyResult{}, err
		}
		if plan.Restart {
			return applyResult{NeedsExit: true, Message: "Update installed. Restarting the app."}, nil
		}
		return applyResult{Message: "Update installed. Restart the app to use the new version."}, nil
	case installModeAppBundle:
		if strings.TrimSpace(plan.NewBundle) == "" {
			return applyResult{}, errors.New("the update has no new app bundle")
		}
		script, err := writeBundleHelper(plan.WorkDir, plan.Target, plan.NewBundle, os.Getpid(), plan.Restart)
		if err != nil {
			return applyResult{}, fmt.Errorf("could not prepare the app bundle helper: %w", err)
		}
		command := exec.Command("/bin/sh", script)
		command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if err := command.Start(); err != nil {
			return applyResult{}, fmt.Errorf("could not start the app bundle helper: %w", err)
		}
		return applyResult{
			NeedsExit:       true,
			HandlesRelaunch: plan.Restart,
			KeepWorkDir:     true,
			Message:         "The new version will replace the app after it quits.",
		}, nil
	default:
		return applyResult{}, fmt.Errorf("automatic updates are not supported in %s mode", plan.Mode)
	}
}

func processRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
