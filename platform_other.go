//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func platformName() string { return "other" }

func startAtLoginSupported() bool { return false }

func openBrowser(rawURL string) error {
	for _, command := range []string{"xdg-open", "open"} {
		if _, err := exec.LookPath(command); err == nil {
			return exec.Command(command, rawURL).Start()
		}
	}
	return errors.New("could not find a browser launcher")
}

func openFolder(path string) error {
	return openBrowser(path)
}

func setAutoStart(bool) error {
	return errors.New("start at login is only supported on Windows")
}

func showMessage(title, text string, _ bool) {
	fmt.Printf("%s\n%s\n", title, text)
}

func prepareConsole() {}

func runDesktop(_ string, _ error, _ bool) int {
	fmt.Fprintln(os.Stderr, "The tray app is only available on Windows. Use --serve for the local dashboard.")
	return 1
}
