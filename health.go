package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// Health is the public health payload returned by /healthz.
type Health struct {
	Status              string `json:"status"`
	Version             string `json:"version"`
	UptimeSeconds       int64  `json:"uptime_seconds"`
	Monitoring          bool   `json:"monitoring"`
	Paused              bool   `json:"paused"`
	Checking            bool   `json:"checking"`
	LastCheck           string `json:"last_check,omitempty"`
	LastCheckAgeSeconds int64  `json:"last_check_age_seconds,omitempty"`
	Streamers           int    `json:"streamers"`
	Live                int    `json:"live"`
	Destinations        int    `json:"destinations"`
	LastError           string `json:"last_error,omitempty"`
	Reason              string `json:"reason,omitempty"`
}

const (
	healthStatusHealthy   = "healthy"
	healthStatusUnhealthy = "unhealthy"
)

// Health reports whether the monitor is alive. A paused app is intentionally
// healthy so an operator choosing to pause alerts does not look like a crash.
func (a *App) Health() Health {
	now := time.Now()

	a.mu.RLock()
	startedAt := a.startedAt
	lastCheck := a.lastCheck
	checking := a.checking
	paused := a.cfg.Settings.Paused
	pollSeconds := a.cfg.Settings.PollSeconds
	streamers := len(a.cfg.Channels)
	destinations := len(a.cfg.Webhooks)
	live := 0
	lastError := ""
	lastErrorAt := ""
	for _, status := range a.statuses {
		if status.Live {
			live++
		}
		if status.Error != "" && status.LastChecked >= lastErrorAt {
			lastError = status.Error
			lastErrorAt = status.LastChecked
		}
	}
	a.mu.RUnlock()

	if pollSeconds < 30 {
		pollSeconds = 30
	}
	interval := time.Duration(pollSeconds) * time.Second
	startupGrace := interval + 2*time.Minute
	if startupGrace < 4*time.Minute {
		startupGrace = 4 * time.Minute
	}
	staleAfter := 3 * interval
	if staleAfter < 6*time.Minute {
		staleAfter = 6 * time.Minute
	}

	health := Health{
		Status:       healthStatusHealthy,
		Version:      Version,
		Monitoring:   !paused,
		Paused:       paused,
		Checking:     checking,
		Streamers:    streamers,
		Destinations: destinations,
		Live:         live,
		LastError:    truncateText(lastError, 300),
	}
	if !startedAt.IsZero() {
		health.UptimeSeconds = int64(now.Sub(startedAt).Seconds())
		if health.UptimeSeconds < 0 {
			health.UptimeSeconds = 0
		}
	}
	if !lastCheck.IsZero() {
		health.LastCheck = lastCheck.UTC().Format(time.RFC3339)
		age := int64(now.Sub(lastCheck).Seconds())
		if age < 0 {
			age = 0
		}
		health.LastCheckAgeSeconds = age
	}

	switch {
	case paused:
		health.Reason = "monitoring is paused"
	case checking:
		// A check may legitimately take a while on a slow network.
	case lastCheck.IsZero() && now.Sub(startedAt) > startupGrace:
		health.Status = healthStatusUnhealthy
		health.Reason = "no check has completed yet"
	case !lastCheck.IsZero() && now.Sub(lastCheck) > staleAfter:
		health.Status = healthStatusUnhealthy
		health.Reason = fmt.Sprintf("last check was %s ago", formatDuration(now.Sub(lastCheck)))
	}
	return health
}

// checkHealth probes a local /healthz URL. It is shared by the CLI
// --healthcheck mode and tests.
func checkHealth(rawURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(rawURL)
	if err != nil {
		return fmt.Errorf("could not reach the dashboard: %w", err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	if response.StatusCode != http.StatusOK {
		message := strings.TrimSpace(string(body))
		if message != "" {
			return fmt.Errorf("dashboard returned HTTP %d: %s", response.StatusCode, truncateText(message, 200))
		}
		return fmt.Errorf("dashboard returned HTTP %d", response.StatusCode)
	}
	return nil
}

func runHealthcheck(rawURL string) int {
	if err := checkHealth(rawURL); err != nil {
		fmt.Fprintln(os.Stderr, "unhealthy: "+err.Error())
		return 1
	}
	fmt.Println("healthy")
	return 0
}

// healthcheckURL resolves the URL used by --healthcheck. An explicit flag wins,
// then the environment, then the port from --listen, then the default port.
func healthcheckURL(explicit, listen string) string {
	if value := strings.TrimSpace(explicit); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("TWITCH_LIVE_HEALTH_URL")); value != "" {
		return value
	}
	port := "8787"
	listen = strings.TrimSpace(listen)
	if listen != "" {
		if _, parsedPort, err := net.SplitHostPort(listen); err == nil && parsedPort != "" {
			port = parsedPort
		} else if parsedPort, err := net.LookupPort("tcp", listen); err == nil {
			port = fmt.Sprintf("%d", parsedPort)
		}
	}
	return "http://127.0.0.1:" + port + "/healthz"
}

func truncateText(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 1 {
		return value[:limit]
	}
	return value[:limit-1] + "…"
}
