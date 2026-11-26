package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// updatePlan describes the already-downloaded, verified payload.
type updatePlan struct {
	Mode      string
	Target    string
	NewBinary string
	NewBundle string
	WorkDir   string
	Restart   bool
	Args      []string
}

// applyResult lets the platform layer tell the manager whether the process
// must exit and whether a detached helper will relaunch the app.
type applyResult struct {
	NeedsExit       bool
	HandlesRelaunch bool
	KeepWorkDir     bool
	Message         string
}

func applyUpdatePlan(plan updatePlan) (applyResult, error) {
	if strings.TrimSpace(plan.Target) == "" {
		return applyResult{}, errors.New("no install target was found")
	}
	return applyPlatformUpdate(plan)
}

// replaceExecutable keeps a copy of the previous binary as <target>.old and
// atomically renames the verified replacement over the target. Unix rename(2)
// replaces the target in one step.
func replaceExecutable(target, newBinary string) error {
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("could not inspect the current version: %w", err)
	}
	backup := target + ".old"
	_ = os.Remove(backup)
	if err := copyFile(target, backup, info.Mode().Perm()|0o600); err != nil {
		return fmt.Errorf("could not keep the previous version for rollback: %w", err)
	}
	if err := os.Chmod(newBinary, 0o755); err != nil {
		return fmt.Errorf("could not make the new version executable: %w", err)
	}
	if err := os.Rename(newBinary, target); err != nil {
		return fmt.Errorf("could not replace the running app (the previous version is at %s): %w", backup, err)
	}
	return nil
}

func copyFile(source, destination string, mode os.FileMode) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	syncErr := output.Sync()
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	if syncErr != nil {
		return syncErr
	}
	return closeErr
}

// writeBundleHelper writes a detached macOS shell helper that waits for the
// current process to exit, replaces the .app bundle, and optionally relaunches.
func writeBundleHelper(workDir, target, newBundle string, parentPID int, relaunch bool) (string, error) {
	scriptPath := filepath.Join(workDir, "replace-bundle.sh")
	relaunchValue := "0"
	if relaunch {
		relaunchValue = "1"
	}
	script := "#!/bin/sh\n" +
		"set -u\n" +
		"parent=" + strconv.Itoa(parentPID) + "\n" +
		"while kill -0 \"$parent\" 2>/dev/null; do sleep 1; done\n" +
		"rm -rf " + shellQuote(target+".old") + "\n" +
		"if [ -e " + shellQuote(target) + " ]; then mv " + shellQuote(target) + " " + shellQuote(target+".old") + " || exit 1; fi\n" +
		"if ! mv " + shellQuote(newBundle) + " " + shellQuote(target) + "; then\n" +
		"  if [ -e " + shellQuote(target+".old") + " ]; then mv " + shellQuote(target+".old") + " " + shellQuote(target) + "; fi\n" +
		"  exit 1\n" +
		"fi\n" +
		"if [ " + relaunchValue + " -eq 1 ]; then open -n " + shellQuote(target) + " || true; fi\n" +
		"rm -rf " + shellQuote(workDir) + "\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		return "", err
	}
	return scriptPath, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// windowsUpdateHelperScript builds a detached batch helper. It waits for the
// running app to exit (which releases the .exe image), keeps the old binary as
// <target>.old, moves the verified download into place, and optionally starts
// it with the original arguments. Retrying the move avoids transient sharing
// violations from antivirus/indexers without ever touching a running image.
func windowsUpdateHelperScript(target, newBinary, workDir string, args []string, parentPID int, relaunch bool) string {
	lines := []string{
		"@echo off",
		"setlocal",
		`powershell -NoProfile -ExecutionPolicy Bypass -Command "Wait-Process -Id ` + strconv.Itoa(parentPID) + ` -ErrorAction SilentlyContinue" >NUL 2>&1`,
		"del /F /Q " + batchQuote(target+".old") + " >NUL 2>&1",
		"set /a tries=0",
		":moveold",
		"move /Y " + batchQuote(target) + " " + batchQuote(target+".old") + " >NUL 2>&1",
		"if not errorlevel 1 goto moved",
		"set /a tries+=1",
		"if %tries% GEQ 60 goto failed",
		"ping -n 2 127.0.0.1 >NUL",
		"goto moveold",
		":moved",
		"set /a tries=0",
		":movenew",
		"move /Y " + batchQuote(newBinary) + " " + batchQuote(target) + " >NUL 2>&1",
		"if not errorlevel 1 goto installed",
		"set /a tries+=1",
		"if %tries% GEQ 20 goto restore",
		"ping -n 2 127.0.0.1 >NUL",
		"goto movenew",
		":restore",
		"move /Y " + batchQuote(target+".old") + " " + batchQuote(target) + " >NUL 2>&1",
		"goto failed",
		":installed",
	}
	start := `start "" ` + batchQuote(target)
	for _, arg := range args {
		start += " " + batchQuote(arg)
	}
	if relaunch {
		lines = append(lines, start)
	}
	lines = append(lines,
		"rmdir /S /Q "+batchQuote(workDir)+" >NUL 2>&1",
		`del "%~f0" >NUL 2>&1`,
		"exit /b 0",
		":failed",
	)
	if relaunch {
		lines = append(lines, "if exist "+batchQuote(target)+" "+start)
	}
	lines = append(lines,
		"rmdir /S /Q "+batchQuote(workDir)+" >NUL 2>&1",
		`del "%~f0" >NUL 2>&1`,
		"exit /b 1",
	)
	return strings.Join(lines, "\r\n") + "\r\n"
}

func batchQuote(value string) string {
	value = strings.ReplaceAll(value, "%", "%%")
	value = strings.ReplaceAll(value, `"`, `\"`)
	return `"` + value + `"`
}
