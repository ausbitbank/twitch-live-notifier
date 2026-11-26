package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	// updateRestartWaitEnv marks a child process that was launched by the
	// updater. Startup should call waitForUpdateRestart before taking the
	// single-instance lock so the child can retry after its parent exits.
	updateRestartWaitEnv   = "TWITCH_LIVE_RESTART_WAIT"
	updateRestartParentEnv = "TWITCH_LIVE_RESTART_PARENT_PID"
	updateRestartTimeout   = 30 * time.Second
)

// updateRestartWaitRequested reports whether this process is a restart child.
func updateRestartWaitRequested() bool {
	value := strings.TrimSpace(os.Getenv(updateRestartWaitEnv))
	return value != "" && value != "0"
}

func updateRestartParentPID() int {
	pid, err := strconv.Atoi(strings.TrimSpace(os.Getenv(updateRestartParentEnv)))
	if err != nil || pid <= 0 {
		return 0
	}
	return pid
}

// waitForUpdateRestart waits for the old process to release the single-instance
// lock, then clears the restart environment so future launches are normal.
// It is a no-op for ordinary launches.
func waitForUpdateRestart() {
	if !updateRestartWaitRequested() {
		return
	}
	parent := updateRestartParentPID()
	deadline := time.Now().Add(updateRestartTimeout)
	for parent > 0 && time.Now().Before(deadline) && processRunning(parent) {
		time.Sleep(200 * time.Millisecond)
	}
	_ = os.Unsetenv(updateRestartWaitEnv)
	_ = os.Unsetenv(updateRestartParentEnv)
}

// acquireInstanceLockWithRestartWait is a drop-in startup helper for main.go:
// it lets a restarted child wait for its parent to exit before locking.
func acquireInstanceLockWithRestartWait(path string) (func(), bool, error) {
	waitForUpdateRestart()
	return acquireInstanceLock(path)
}

// startRestartedProcess launches the updated app with the same arguments and a
// short-lived environment marker used by waitForUpdateRestart.
func startRestartedProcess(executable string, args []string) error {
	command := exec.Command(executable, args...)
	command.Env = append(os.Environ(),
		updateRestartWaitEnv+"=1",
		updateRestartParentEnv+"="+strconv.Itoa(os.Getpid()),
	)
	command.Stdin = nil
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Start()
}
