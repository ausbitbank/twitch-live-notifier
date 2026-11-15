package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	configExportVersion = 1
	redactedURLMessage  = "The webhook URL was redacted in this backup. Re-enter it on the Destinations page."
)

// resolveConfigSecrets fills in webhook URLs that live in the secret store.
// A missing or unreadable secret does not make the whole config unreadable; the
// destination is marked invalid so the user can re-enter the URL.
func resolveConfigSecrets(cfg *Config, dir string) {
	store, err := openSecretStore(dir)
	if err != nil {
		markSecretUnavailable(cfg)
		return
	}
	for i := range cfg.Webhooks {
		webhook := &cfg.Webhooks[i]
		if strings.TrimSpace(webhook.URL) != "" || strings.TrimSpace(webhook.URLSecretID) == "" {
			continue
		}
		value, err := store.Get(webhook.URLSecretID)
		if err != nil {
			webhook.Invalid = true
			webhook.LastError = redactedURLMessage
			webhook.LastErrorAt = time.Now().UTC().Format(time.RFC3339)
			continue
		}
		webhook.URL = value
	}
	for channelIndex := range cfg.Channels {
		for destinationIndex := range cfg.Channels[channelIndex].Destinations {
			record := cfg.Channels[channelIndex].Destinations[destinationIndex].LiveMessage
			if record == nil || strings.TrimSpace(record.WebhookURL) != "" || strings.TrimSpace(record.WebhookURLSecretID) == "" {
				continue
			}
			if value, err := store.Get(record.WebhookURLSecretID); err == nil {
				record.WebhookURL = value
			}
		}
	}
}

func markSecretUnavailable(cfg *Config) {
	now := time.Now().UTC().Format(time.RFC3339)
	for i := range cfg.Webhooks {
		webhook := &cfg.Webhooks[i]
		if strings.TrimSpace(webhook.URLSecretID) == "" || strings.TrimSpace(webhook.URL) != "" {
			continue
		}
		webhook.Invalid = true
		webhook.LastError = redactedURLMessage
		webhook.LastErrorAt = now
	}
}

// protectConfigSecrets returns a copy of the config where webhook URLs have
// been moved into the secret store. If anything fails, the original config is
// returned unchanged so a keyring problem cannot silently blank the URLs.
func protectConfigSecrets(cfg Config, dir string) Config {
	store, err := openSecretStore(dir)
	if err != nil {
		return cfg
	}
	protected := cloneConfig(cfg)
	for i := range protected.Webhooks {
		webhook := &protected.Webhooks[i]
		if strings.TrimSpace(webhook.URL) == "" {
			continue
		}
		if strings.TrimSpace(webhook.URLSecretID) == "" {
			webhook.URLSecretID = "url-" + webhook.ID
		}
		if err := store.Set(webhook.URLSecretID, webhook.URL); err != nil {
			return cfg
		}
		webhook.URL = ""
		webhook.URLRedacted = false
	}
	for channelIndex := range protected.Channels {
		for destinationIndex := range protected.Channels[channelIndex].Destinations {
			destination := &protected.Channels[channelIndex].Destinations[destinationIndex]
			record := destination.LiveMessage
			if record == nil || strings.TrimSpace(record.WebhookURL) == "" {
				continue
			}
			if strings.TrimSpace(record.WebhookURLSecretID) == "" {
				record.WebhookURLSecretID = "msg-" + protected.Channels[channelIndex].ID + "-" + destination.WebhookID
			}
			if err := store.Set(record.WebhookURLSecretID, record.WebhookURL); err != nil {
				return cfg
			}
			record.WebhookURL = ""
		}
	}
	allowed := map[string]bool{}
	for _, webhook := range protected.Webhooks {
		if webhook.URLSecretID != "" {
			allowed[webhook.URLSecretID] = true
		}
	}
	for _, channel := range protected.Channels {
		for _, destination := range channel.Destinations {
			if destination.LiveMessage != nil && destination.LiveMessage.WebhookURLSecretID != "" {
				allowed[destination.LiveMessage.WebhookURLSecretID] = true
			}
		}
	}
	for id := range store.entries {
		if !allowed[id] {
			store.Delete(id)
		}
	}
	if err := store.Save(); err != nil {
		return cfg
	}
	return protected
}

type configExport struct {
	Version    int       `json:"version"`
	ExportedAt string    `json:"exported_at"`
	Redacted   bool      `json:"redacted"`
	Channels   []Channel `json:"channels"`
	Webhooks   []Webhook `json:"webhooks"`
	Settings   Settings  `json:"settings"`
}

// ExportConfig returns a backup. Secrets are redacted by default; callers must
// explicitly ask to include live webhook URLs.
func (a *App) ExportConfig(includeSecrets bool) configExport {
	cfg := a.ConfigSnapshot()
	exported := configExport{
		Version:    configExportVersion,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Redacted:   !includeSecrets,
		Channels:   cfg.Channels,
		Webhooks:   cfg.Webhooks,
		Settings:   cfg.Settings,
	}
	exported.Settings.DashboardTokenHash = ""
	for i := range exported.Webhooks {
		webhook := &exported.Webhooks[i]
		webhook.URLSecretID = ""
		if includeSecrets {
			webhook.URLRedacted = false
		} else {
			webhook.URL = ""
			webhook.URLRedacted = true
		}
	}
	for channelIndex := range exported.Channels {
		for destinationIndex := range exported.Channels[channelIndex].Destinations {
			record := exported.Channels[channelIndex].Destinations[destinationIndex].LiveMessage
			if record == nil {
				continue
			}
			record.WebhookURLSecretID = ""
			if !includeSecrets {
				record.WebhookURL = ""
			}
		}
	}
	return exported
}

func (a *App) ExportConfigJSON(includeSecrets bool) ([]byte, error) {
	data, err := json.MarshalIndent(a.ExportConfig(includeSecrets), "", "  ")
	if err != nil {
		return nil, errors.New("could not encode the backup")
	}
	return append(data, '\n'), nil
}

// ImportConfig replaces the active config with a validated backup. Redacted
// webhook URLs are accepted but marked invalid until the user re-enters them.
func (a *App) ImportConfig(data []byte) error {
	if len(data) == 0 {
		return errors.New("the backup file is empty")
	}
	var exported configExport
	if err := json.Unmarshal(data, &exported); err != nil {
		return errors.New("the backup file is not valid JSON")
	}
	if exported.Version > configVersion {
		return fmt.Errorf("this backup was made by a newer version (schema %d)", exported.Version)
	}
	cfg := Config{
		Version:  exported.Version,
		Channels: exported.Channels,
		Webhooks: exported.Webhooks,
		Settings: exported.Settings,
	}
	if cfg.Version == 0 {
		cfg.Version = configVersion
	}
	cfg.Settings.DashboardTokenHash = ""
	normalizeConfig(&cfg)
	for i := range cfg.Webhooks {
		webhook := &cfg.Webhooks[i]
		webhook.URLSecretID = ""
		if strings.TrimSpace(webhook.URL) == "" {
			webhook.Invalid = true
			webhook.LastError = redactedURLMessage
			webhook.LastErrorAt = time.Now().UTC().Format(time.RFC3339)
			continue
		}
		normalized, urlErr := normalizeWebhookURL(webhook.URL)
		if urlErr != nil {
			webhook.Invalid = true
			webhook.LastError = "The backup contains an invalid webhook URL. Re-enter it on the Destinations page."
			webhook.LastErrorAt = time.Now().UTC().Format(time.RFC3339)
			continue
		}
		webhook.URL = normalized
		webhook.Invalid = false
		webhook.LastError = ""
		webhook.LastErrorAt = ""
	}
	if len(cfg.Webhooks) == 0 && len(cfg.Channels) == 0 {
		return errors.New("the backup has no streamers or destinations")
	}

	a.mu.Lock()
	statuses := make(map[string]ChannelStatus, len(cfg.Channels))
	for _, channel := range cfg.Channels {
		if status, ok := a.statuses[channel.ID]; ok {
			statuses[channel.ID] = status
		}
	}
	a.cfg = cfg
	a.statuses = statuses
	err := a.persistLocked()
	a.mu.Unlock()
	if err != nil {
		return err
	}
	a.addActivity("success", "Restored settings from a backup.")
	a.wakeMonitor()
	return nil
}

func configExportFilename(includeSecrets bool) string {
	suffix := "redacted"
	if includeSecrets {
		suffix = "with-secrets"
	}
	return "twitch-live-config-" + time.Now().Format("20060102-150405") + "-" + suffix + ".json"
}
