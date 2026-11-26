package main

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// Diagnostics is the redacted support bundle returned by GET /api/diagnostics
// and printed by --doctor. It never contains webhook URLs, webhook tokens,
// secret-store IDs, or raw config secrets.
type Diagnostics struct {
	App           string               `json:"app"`
	Version       string               `json:"version"`
	GoVersion     string               `json:"go_version"`
	Platform      string               `json:"platform"`
	DataDir       string               `json:"data_dir"`
	ConfigPath    string               `json:"config_path"`
	UptimeSeconds int64                `json:"uptime_seconds"`
	LastCheck     string               `json:"last_check,omitempty"`
	Channels      DiagnosticsChannels  `json:"channels"`
	Webhooks      []DiagnosticsWebhook `json:"webhooks"`
	RecentErrors  []string             `json:"recent_errors"`
	RecentHistory []DeliveryRecord     `json:"recent_history"`
	Update        UpdateStatus         `json:"update"`
}

type DiagnosticsChannels struct {
	Total   int `json:"total"`
	Enabled int `json:"enabled"`
	Live    int `json:"live"`
	Errors  int `json:"errors"`
}

// DiagnosticsWebhook is the label-only destination health shape. The URL and
// secret ID are deliberately omitted.
type DiagnosticsWebhook struct {
	Label     string `json:"label"`
	Enabled   bool   `json:"enabled"`
	Invalid   bool   `json:"invalid"`
	LastError string `json:"last_error"`
}

var (
	diagnosticWebhookURLPattern  = regexp.MustCompile(`(?i)https?://[^\s"'<>]*api/webhooks/[^\s"'<>]*`)
	diagnosticTokenPattern       = regexp.MustCompile(`[A-Za-z0-9_-]{24,}\.[A-Za-z0-9_-]{6,}\.[A-Za-z0-9_-]{27,}`)
	diagnosticSecretParamPattern = regexp.MustCompile(`(?i)\b(token|secret|api[_-]?key|signature)=([^&\s]+)`)
)

// buildDiagnostics assembles the support bundle from the live App and the
// persisted delivery history. history may be nil.
func buildDiagnostics(app *App, history *deliveryLog) Diagnostics {
	now := time.Now()
	diagnostics := Diagnostics{
		App:           appName,
		Version:       Version,
		GoVersion:     runtime.Version(),
		Platform:      runtime.GOOS,
		Webhooks:      []DiagnosticsWebhook{},
		RecentErrors:  []string{},
		RecentHistory: []DeliveryRecord{},
		Update:        UpdateStatus{CurrentVersion: Version},
	}
	if app == nil {
		if history != nil {
			diagnostics.RecentHistory = sanitizedHistory(history, 20)
		}
		return diagnostics
	}

	var (
		enabled   int
		live      int
		errors    int
		statusErr []string
		webhookEr []string
	)
	app.mu.RLock()
	startedAt := app.startedAt
	lastCheck := app.lastCheck
	diagnostics.DataDir = filepath.Dir(app.configPath)
	diagnostics.ConfigPath = app.configPath
	for _, channel := range app.cfg.Channels {
		diagnostics.Channels.Total++
		if channel.Enabled {
			enabled++
		}
		status := app.statuses[channel.ID]
		if status.Live {
			live++
		}
		if status.Error != "" {
			errors++
			statusErr = append(statusErr, status.Error)
		}
	}
	for _, webhook := range app.cfg.Webhooks {
		diagnostics.Webhooks = append(diagnostics.Webhooks, DiagnosticsWebhook{
			Label:     sanitizeDiagnosticText(webhook.Label, 120),
			Enabled:   webhook.Enabled,
			Invalid:   webhook.Invalid,
			LastError: sanitizeDiagnosticText(webhook.LastError, 300),
		})
		if webhook.LastError != "" {
			webhookEr = append(webhookEr, webhook.LastError)
		}
	}
	activity := append([]Activity(nil), app.activity...)
	update := app.updateStatus
	app.mu.RUnlock()

	diagnostics.Channels.Enabled = enabled
	diagnostics.Channels.Live = live
	diagnostics.Channels.Errors = errors

	if !startedAt.IsZero() {
		diagnostics.UptimeSeconds = int64(now.Sub(startedAt).Seconds())
		if diagnostics.UptimeSeconds < 0 {
			diagnostics.UptimeSeconds = 0
		}
	}
	if !lastCheck.IsZero() {
		diagnostics.LastCheck = lastCheck.UTC().Format(time.RFC3339)
	}

	if update.CurrentVersion == "" {
		update.CurrentVersion = Version
	}
	// Round-trip the update status so every string it grows over time (release
	// URLs, install-job messages, errors) is redacted without this package
	// needing to know the fields. If the round-trip ever fails, keep only the
	// version rather than risk emitting an unredacted string.
	update = redactUpdateStatus(update)
	diagnostics.Update = update

	var activityErr []string
	for _, entry := range activity {
		if strings.EqualFold(strings.TrimSpace(entry.Level), deliveryLevelError) && strings.TrimSpace(entry.Message) != "" {
			activityErr = append(activityErr, entry.Message)
		}
	}

	diagnostics.RecentHistory = sanitizedHistory(history, 20)
	var historyErr []string
	for _, record := range diagnostics.RecentHistory {
		if strings.EqualFold(record.Level, deliveryLevelError) && record.Reason != "" {
			historyErr = append(historyErr, record.Reason)
		}
	}

	diagnostics.RecentErrors = recentErrors(activityErr, statusErr, webhookEr, historyErr)
	return diagnostics
}

// sanitizedHistory returns at most limit redacted records, newest first.
func sanitizedHistory(history *deliveryLog, limit int) []DeliveryRecord {
	if history == nil {
		return []DeliveryRecord{}
	}
	records, _ := history.List(limit, DeliveryFilters{})
	if records == nil {
		return []DeliveryRecord{}
	}
	sanitized := make([]DeliveryRecord, 0, len(records))
	for _, record := range records {
		record.ChannelLogin = sanitizeDiagnosticText(record.ChannelLogin, 100)
		record.ChannelName = sanitizeDiagnosticText(record.ChannelName, 120)
		record.WebhookLabel = sanitizeDiagnosticText(record.WebhookLabel, 120)
		record.StreamID = sanitizeDiagnosticText(record.StreamID, 100)
		record.Title = sanitizeDiagnosticText(record.Title, 300)
		record.Game = sanitizeDiagnosticText(record.Game, 120)
		record.MessageID = sanitizeDiagnosticText(record.MessageID, 100)
		record.MessageURL = sanitizeDiagnosticText(record.MessageURL, 300)
		record.Reason = sanitizeDiagnosticText(record.Reason, 500)
		sanitized = append(sanitized, record)
	}
	return sanitized
}

func redactUpdateStatus(update UpdateStatus) UpdateStatus {
	data, err := json.Marshal(update)
	if err != nil {
		return UpdateStatus{CurrentVersion: update.CurrentVersion}
	}
	data = []byte(redactDiagnosticText(string(data)))
	var redacted UpdateStatus
	if err := json.Unmarshal(data, &redacted); err != nil {
		return UpdateStatus{CurrentVersion: update.CurrentVersion}
	}
	return redacted
}

// recentErrors merges the supplied error sources in chronological-ish order,
// removes exact duplicates, and keeps the newest ten.
func recentErrors(sources ...[]string) []string {
	const limit = 10
	errors := make([]string, 0, limit)
	seen := make(map[string]bool)
	for _, source := range sources {
		for _, raw := range source {
			value := sanitizeDiagnosticText(raw, 300)
			if value == "" || seen[value] {
				continue
			}
			seen[value] = true
			errors = append(errors, value)
		}
	}
	if len(errors) > limit {
		errors = errors[len(errors)-limit:]
	}
	return errors
}

// sanitizeDiagnosticText removes anything that could be a Discord webhook URL,
// bot token, or secret query parameter, then truncates the result.
func sanitizeDiagnosticText(value string, limit int) string {
	return truncateText(redactDiagnosticText(strings.TrimSpace(value)), limit)
}

// redactDiagnosticText removes secret-looking values without truncating, so it
// can also be applied to an encoded JSON document.
func redactDiagnosticText(value string) string {
	if value == "" {
		return ""
	}
	value = diagnosticWebhookURLPattern.ReplaceAllString(value, "[redacted webhook]")
	value = diagnosticTokenPattern.ReplaceAllString(value, "[redacted token]")
	value = diagnosticSecretParamPattern.ReplaceAllString(value, "$1=[redacted]")
	return value
}

// DiagnosticsJSON renders the same redacted bundle that --doctor prints.
func (a *App) DiagnosticsJSON() ([]byte, error) {
	diagnostics := buildDiagnostics(a, appHistoryResolver(a))
	data, err := json.MarshalIndent(diagnostics, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
