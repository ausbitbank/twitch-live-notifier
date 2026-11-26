//go:build !windows && !darwin

package main

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
)

func applyPlatformUpdate(plan updatePlan) (applyResult, error) {
	switch plan.Mode {
	case installModeInPlace, installModeAppImage:
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
