package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	configPath, pathErr := defaultConfigPath()
	if len(os.Args) == 1 {
		os.Exit(runDesktop(configPath, pathErr, true))
	}
	os.Exit(runCLI(os.Args[1:], configPath, pathErr))
}

func runCLI(args []string, configPath string, pathErr error) int {
	fs := flag.NewFlagSet("TwitchLiveNotifier", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var (
		background         = fs.Bool("background", false, "start in the tray without opening the dashboard")
		check              = fs.Bool("check", false, "check all streamers once and exit")
		serve              = fs.Bool("serve", false, "run the dashboard without the tray icon")
		healthcheck        = fs.Bool("healthcheck", false, "check the local /healthz endpoint and exit")
		healthcheckURLFlag = fs.String("healthcheck-url", "", "health endpoint URL used by --healthcheck")
		configFlag         = fs.String("config", "", "use a custom settings file")
		listenFlag         = fs.String("listen", "", "dashboard listen address (default 127.0.0.1:8787)")
		showVersion        = fs.Bool("version", false, "print the version")
		showHelp           = fs.Bool("help", false, "show this help")
	)
	fs.BoolVar(showHelp, "h", false, "show this help")

	if err := fs.Parse(args); err != nil {
		prepareConsole()
		fmt.Fprintln(os.Stderr, "Invalid arguments.\n\n"+usageText())
		return 2
	}
	if fs.NArg() > 0 {
		prepareConsole()
		fmt.Fprintln(os.Stderr, "Unexpected argument: "+fs.Arg(0)+"\n\n"+usageText())
		return 2
	}
	if *showHelp {
		prepareConsole()
		fmt.Print(usageText())
		return 0
	}
	if *showVersion {
		prepareConsole()
		fmt.Printf("%s v%s\n", appName, Version)
		return 0
	}

	actions := 0
	for _, enabled := range []bool{*background, *check, *serve, *healthcheck} {
		if enabled {
			actions++
		}
	}
	if actions > 1 {
		prepareConsole()
		fmt.Fprintln(os.Stderr, "Choose only one of --background, --check, --serve, or --healthcheck.")
		return 2
	}
	if *healthcheck {
		return runHealthcheck(healthcheckURL(*healthcheckURLFlag, *listenFlag))
	}

	if *configFlag != "" {
		configPath = *configFlag
		pathErr = nil
	}
	if pathErr != nil {
		return cliError(pathErr)
	}
	if strings.TrimSpace(configPath) == "" {
		return cliError(fmt.Errorf("could not determine the settings path"))
	}

	switch {
	case *background:
		return runDesktop(configPath, nil, false)
	case *serve:
		return runServerOnly(configPath, *listenFlag)
	case *check:
		return runCheckOnce(configPath)
	default:
		prepareConsole()
		fmt.Print(usageText())
		return 2
	}
}

func runCheckOnce(configPath string) int {
	release, alreadyRunning, err := acquireInstanceLock(instanceLockPath(configPath))
	if err != nil {
		return cliError(err)
	}
	defer release()
	if alreadyRunning {
		prepareConsole()
		fmt.Println("Another instance is already running; skipping this check to avoid duplicate posts.")
		return 0
	}

	app, err := NewApp(configPath)
	if err != nil {
		return cliError(err)
	}
	summary, err := app.CheckNow(context.Background())
	if err != nil {
		return cliError(err)
	}
	prepareConsole()
	fmt.Printf("Checked %d streamer(s); posted %d alert(s).\n", summary.Channels, summary.Posted)
	if summary.Errors > 0 {
		fmt.Printf("%d check(s) failed.\n", summary.Errors)
		return 1
	}
	return 0
}

func cliError(err error) int {
	prepareConsole()
	fmt.Fprintln(os.Stderr, "❌ "+err.Error())
	return 1
}

func usageText() string {
	exe := filepath.Base(os.Args[0])
	if exe == "" || exe == "." {
		exe = "TwitchLiveNotifier"
	}
	settingsPath := "your user configuration folder"
	if path, err := defaultConfigPath(); err == nil {
		settingsPath = path
	}
	if platformName() == "windows" {
		return fmt.Sprintf(`%s v%s

Usage:
  %s
      Open the dashboard and keep monitoring in the system tray.

  %s --background
      Start in the tray without opening the browser. Used at login.

  %s --check
      Check every streamer once, post any pending alerts, and exit.

  %s --serve
      Run the dashboard without the tray icon.

  %s --healthcheck
      Probe the local /healthz endpoint and exit. Used by Docker.

Options:
  --config PATH          Use a custom settings file.
  --listen ADDR          Dashboard address (default 127.0.0.1:8787).
  --healthcheck-url URL  Health URL for --healthcheck.
  --version              Print the version.
  --help, -h             Show this help.

Settings are stored in:
  %s`, appName, Version, exe, exe, exe, exe, exe, settingsPath)
	}
	return fmt.Sprintf(`%s v%s

Usage:
  %s --serve
      Run the dashboard in the foreground. Open the printed URL.

  %s --check
      Check every streamer once, post any pending alerts, and exit.

  %s --healthcheck
      Probe the local /healthz endpoint and exit. Used by Docker.

Options:
  --config PATH          Use a custom settings file.
  --listen ADDR          Dashboard address (default 127.0.0.1:8787).
  --healthcheck-url URL  Health URL for --healthcheck.
  --version              Print the version.
  --help, -h             Show this help.

Settings are stored in:
  %s`, appName, Version, exe, exe, exe, settingsPath)
}
