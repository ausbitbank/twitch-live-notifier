package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
	Version  int       `json:"version"`
	Channels []Channel `json:"channels"`
	Webhooks []Webhook `json:"webhooks"`
	Settings Settings  `json:"settings"`
}

type Channel struct {
	ID           string        `json:"id"`
	Login        string        `json:"login"`
	DisplayName  string        `json:"display_name"`
	AvatarURL    string        `json:"avatar_url,omitempty"`
	BannerURL    string        `json:"banner_url,omitempty"`
	DiscordID    string        `json:"discord_id,omitempty"`
	Enabled      bool          `json:"enabled"`
	Destinations []Destination `json:"destinations"`
	Filters      Filters       `json:"filters"`

	LastAlertAt       string `json:"last_alert_at,omitempty"`
	LastAlertDate     string `json:"last_alert_date,omitempty"`
	LastAlertStreamID string `json:"last_alert_stream_id,omitempty"`
}

// Filters controls whether a live stream becomes a Discord alert. Rules are
// evaluated in the user's local timezone. Excludes always win over includes.
type Filters struct {
	IncludeGames     []string `json:"include_games,omitempty"`
	ExcludeGames     []string `json:"exclude_games,omitempty"`
	IncludeTitles    []string `json:"include_titles,omitempty"`
	ExcludeTitles    []string `json:"exclude_titles,omitempty"`
	MinViewers       int      `json:"min_viewers,omitempty"`
	QuietStart       string   `json:"quiet_start,omitempty"`
	QuietEnd         string   `json:"quiet_end,omitempty"`
	AllowedDays      []int    `json:"allowed_days,omitempty"`
	CooldownMinutes  int      `json:"cooldown_minutes,omitempty"`
	FirstStreamOfDay bool     `json:"first_stream_of_day,omitempty"`
	DelayMinutes     int      `json:"delay_minutes,omitempty"`
}

type Destination struct {
	WebhookID    string `json:"webhook_id"`
	LastStreamID string `json:"last_stream_id,omitempty"`
	LastPostedAt string `json:"last_posted_at,omitempty"`
}

type Webhook struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	URL      string   `json:"url"`
	Enabled  bool     `json:"enabled"`
	RoleIDs  []string `json:"role_ids,omitempty"`
	Template Template `json:"template"`

	LastError           string `json:"last_error,omitempty"`
	LastErrorAt         string `json:"last_error_at,omitempty"`
	LastSuccessAt       string `json:"last_success_at,omitempty"`
	ConsecutiveFailures int    `json:"consecutive_failures,omitempty"`
	Invalid             bool   `json:"invalid,omitempty"`
}

type Template struct {
	Content          string `json:"content"`
	EmbedTitle       string `json:"embed_title"`
	EmbedDescription string `json:"embed_description"`
	Footer           string `json:"footer"`
	Color            string `json:"color"`
	ShowAuthor       bool   `json:"show_author"`
	ShowImage        bool   `json:"show_image"`
	ShowGame         bool   `json:"show_game"`
	ShowViewers      bool   `json:"show_viewers"`
	ShowLiveSince    bool   `json:"show_live_since"`
	PingStreamer     *bool  `json:"ping_streamer,omitempty"`
	PingRoles        *bool  `json:"ping_roles,omitempty"`
	AllowEveryone    bool   `json:"allow_everyone,omitempty"`
}

type Settings struct {
	PollSeconds  int    `json:"poll_seconds"`
	StartAtLogin bool   `json:"start_at_login"`
	Paused       bool   `json:"paused"`
	Theme        string `json:"theme"`
}

var (
	twitchLoginPattern   = regexp.MustCompile(`^[A-Za-z0-9_]{1,25}$`)
	webhookTokenPattern  = regexp.MustCompile(`^[A-Za-z0-9._-]{10,}$`)
	discordIDPattern     = regexp.MustCompile(`^\d{17,20}$`)
	discordMentionRE     = regexp.MustCompile(`^<@!?(\d{17,20})>$`)
	discordRoleMentionRE = regexp.MustCompile(`^<@&(\d{17,20})>$`)
	discordUserURLRE     = regexp.MustCompile(`(?i)^https?://(?:www\.)?discord(?:app)?\.com/users/(\d{17,20})/?$`)
	discordWebhookHosts  = map[string]bool{
		"discord.com":        true,
		"discordapp.com":     true,
		"ptb.discord.com":    true,
		"canary.discord.com": true,
	}
)

func defaultTemplate() Template {
	pingStreamer := true
	pingRoles := true
	return Template{
		Content:          "🔴 **{streamer}** is live on Twitch!",
		EmbedTitle:       "{title}",
		EmbedDescription: "**{streamer}** is live now — come hang out!",
		Footer:           appName,
		Color:            "#9146FF",
		ShowAuthor:       true,
		ShowImage:        true,
		ShowGame:         true,
		ShowViewers:      true,
		ShowLiveSince:    true,
		PingStreamer:     &pingStreamer,
		PingRoles:        &pingRoles,
	}
}

func defaultSettings() Settings {
	return Settings{
		PollSeconds: 120,
		Theme:       "dark",
	}
}

func defaultConfig() Config {
	return Config{
		Version:  configVersion,
		Channels: []Channel{},
		Webhooks: []Webhook{},
		Settings: defaultSettings(),
	}
}

func defaultConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not find your user settings folder: %w", err)
	}
	return filepath.Join(dir, configDirName, "config.json"), nil
}

type rawConfig struct {
	Version  int       `json:"version"`
	Channels []Channel `json:"channels"`
	Webhooks []Webhook `json:"webhooks"`
	Settings *Settings `json:"settings"`

	// Version 1 fields.
	TwitchLogin          string `json:"twitch_login"`
	DiscordWebhook       string `json:"discord_webhook"`
	LastNotifiedStreamID string `json:"last_notified_stream_id"`
	LastNotifiedAt       string `json:"last_notified_at"`
}

func loadConfig(path string) (Config, error) {
	if strings.TrimSpace(path) == "" {
		return Config{}, errors.New("settings path is empty")
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaultConfig(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("could not read settings: %w", err)
	}

	var raw rawConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return Config{}, errors.New("the settings file is not valid JSON")
	}

	cfg := Config{
		Version:  raw.Version,
		Channels: raw.Channels,
		Webhooks: raw.Webhooks,
	}
	if raw.Settings != nil {
		cfg.Settings = *raw.Settings
	}
	migrateLegacyConfig(&cfg, raw)
	normalizeConfig(&cfg)
	return cfg, nil
}

func migrateLegacyConfig(cfg *Config, raw rawConfig) {
	if len(cfg.Channels) > 0 || len(cfg.Webhooks) > 0 {
		return
	}
	if strings.TrimSpace(raw.TwitchLogin) == "" && strings.TrimSpace(raw.DiscordWebhook) == "" {
		return
	}

	channelID := newID()
	webhookID := newID()
	login := strings.ToLower(strings.TrimSpace(raw.TwitchLogin))
	if normalized, err := normalizeTwitchLogin(raw.TwitchLogin); err == nil {
		login = normalized
	}
	webhookURL := strings.TrimSpace(raw.DiscordWebhook)
	if normalized, err := normalizeWebhookURL(webhookURL); err == nil {
		webhookURL = normalized
	}

	cfg.Channels = []Channel{{
		ID:          channelID,
		Login:       login,
		DisplayName: login,
		Enabled:     true,
		Destinations: []Destination{{
			WebhookID:    webhookID,
			LastStreamID: raw.LastNotifiedStreamID,
			LastPostedAt: raw.LastNotifiedAt,
		}},
	}}
	cfg.Webhooks = []Webhook{{
		ID:       webhookID,
		Label:    "Discord",
		URL:      webhookURL,
		Enabled:  true,
		Template: defaultTemplate(),
	}}
	cfg.Version = configVersion
	cfg.Settings = defaultSettings()
}

func normalizeConfig(cfg *Config) {
	if cfg.Version < configVersion {
		cfg.Version = configVersion
	}
	if cfg.Channels == nil {
		cfg.Channels = []Channel{}
	}
	if cfg.Webhooks == nil {
		cfg.Webhooks = []Webhook{}
	}

	settings := cfg.Settings
	if settings.PollSeconds == 0 {
		settings.PollSeconds = 120
	}
	if settings.PollSeconds < 30 {
		settings.PollSeconds = 30
	}
	if settings.PollSeconds > 3600 {
		settings.PollSeconds = 3600
	}
	if settings.Theme != "light" {
		settings.Theme = "dark"
	}
	cfg.Settings = settings

	webhookIDs := make(map[string]bool, len(cfg.Webhooks))
	for i := range cfg.Webhooks {
		webhook := &cfg.Webhooks[i]
		if webhook.ID == "" {
			webhook.ID = newID()
		}
		webhookIDs[webhook.ID] = true
		if webhook.Label == "" {
			webhook.Label = "Discord"
		}
		webhook.RoleIDs = sanitizeRoleIDs(webhook.RoleIDs)
		webhook.Template = normalizeTemplate(webhook.Template)
	}

	for i := range cfg.Channels {
		channel := &cfg.Channels[i]
		if channel.ID == "" {
			channel.ID = newID()
		}
		if channel.DisplayName == "" {
			channel.DisplayName = channel.Login
		}
		if channel.DiscordID != "" {
			if id, err := normalizeDiscordID(channel.DiscordID); err == nil {
				channel.DiscordID = id
			} else {
				channel.DiscordID = ""
			}
		}
		channel.Filters = normalizeFilters(channel.Filters)
		seen := make(map[string]bool, len(channel.Destinations))
		destinations := make([]Destination, 0, len(channel.Destinations))
		for _, destination := range channel.Destinations {
			if destination.WebhookID == "" || !webhookIDs[destination.WebhookID] || seen[destination.WebhookID] {
				continue
			}
			seen[destination.WebhookID] = true
			destinations = append(destinations, destination)
		}
		channel.Destinations = destinations
	}
}

func saveConfig(path string, cfg Config) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("settings path is empty")
	}
	normalizeConfig(&cfg)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return errors.New("could not encode settings")
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return errors.New("could not create the settings folder")
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return errors.New("could not write the settings file")
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return errors.New("could not save the settings file")
	}
	return nil
}

func cloneConfig(cfg Config) Config {
	clone := cfg
	clone.Channels = make([]Channel, len(cfg.Channels))
	for i, channel := range cfg.Channels {
		clone.Channels[i] = channel
		clone.Channels[i].Destinations = append([]Destination(nil), channel.Destinations...)
		clone.Channels[i].Filters.IncludeGames = append([]string(nil), channel.Filters.IncludeGames...)
		clone.Channels[i].Filters.ExcludeGames = append([]string(nil), channel.Filters.ExcludeGames...)
		clone.Channels[i].Filters.IncludeTitles = append([]string(nil), channel.Filters.IncludeTitles...)
		clone.Channels[i].Filters.ExcludeTitles = append([]string(nil), channel.Filters.ExcludeTitles...)
		clone.Channels[i].Filters.AllowedDays = append([]int(nil), channel.Filters.AllowedDays...)
	}
	clone.Webhooks = make([]Webhook, len(cfg.Webhooks))
	for i, webhook := range cfg.Webhooks {
		clone.Webhooks[i] = webhook
		clone.Webhooks[i].RoleIDs = append([]string(nil), webhook.RoleIDs...)
		clone.Webhooks[i].Template = cloneTemplate(webhook.Template)
	}
	return clone
}

func cloneTemplate(template Template) Template {
	clone := template
	if template.PingStreamer != nil {
		value := *template.PingStreamer
		clone.PingStreamer = &value
	}
	if template.PingRoles != nil {
		value := *template.PingRoles
		clone.PingRoles = &value
	}
	return clone
}

func newID() string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%d", os.Getpid())
	}
	return hex.EncodeToString(bytes[:])
}

func normalizeDiscordID(input string) (string, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return "", nil
	}
	if match := discordMentionRE.FindStringSubmatch(value); match != nil {
		value = match[1]
	} else if match := discordUserURLRE.FindStringSubmatch(value); match != nil {
		value = match[1]
	}
	if !discordIDPattern.MatchString(value) {
		return "", errors.New("Discord user IDs are 17–20 digit numbers; paste the ID or profile URL")
	}
	return value, nil
}

func normalizeRoleID(input string) (string, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return "", nil
	}
	if match := discordRoleMentionRE.FindStringSubmatch(value); match != nil {
		value = match[1]
	}
	if !discordIDPattern.MatchString(value) {
		return "", errors.New("Discord role IDs are 17–20 digit numbers; paste the ID or <@&role> mention")
	}
	return value, nil
}

// normalizeRoleIDs validates user input and returns a stable, deduplicated
// list. Discord accepts up to 100 allowed roles; ten keeps the UI manageable.
func normalizeRoleIDs(input []string) ([]string, error) {
	seen := make(map[string]bool, len(input))
	roles := make([]string, 0, len(input))
	for _, raw := range input {
		id, err := normalizeRoleID(raw)
		if err != nil {
			return nil, err
		}
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		roles = append(roles, id)
		if len(roles) > 10 {
			return nil, errors.New("add at most 10 role pings per destination")
		}
	}
	if len(roles) == 0 {
		return nil, nil
	}
	return roles, nil
}

// sanitizeRoleIDs is the non-failing config-loading variant.
func sanitizeRoleIDs(input []string) []string {
	seen := make(map[string]bool, len(input))
	roles := make([]string, 0, len(input))
	for _, raw := range input {
		id, err := normalizeRoleID(raw)
		if err != nil || id == "" || seen[id] {
			continue
		}
		seen[id] = true
		roles = append(roles, id)
		if len(roles) == 10 {
			break
		}
	}
	if len(roles) == 0 {
		return nil
	}
	return roles
}

func boolPointerValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func normalizeTwitchLogin(input string) (string, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return "", errors.New("enter your Twitch username")
	}

	value = strings.TrimPrefix(value, "@")
	lower := strings.ToLower(value)

	if strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "www.") ||
		strings.Contains(lower, "twitch.tv/") {
		if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
			value = "https://" + value
		}
		u, err := url.Parse(value)
		if err != nil || u.Host == "" {
			return "", errors.New("that does not look like a Twitch username or URL")
		}
		host := strings.ToLower(u.Hostname())
		if host != "twitch.tv" && !strings.HasSuffix(host, ".twitch.tv") {
			return "", errors.New("the URL must be on twitch.tv")
		}
		value = strings.Trim(u.Path, "/")
		if i := strings.IndexByte(value, '/'); i >= 0 {
			value = value[:i]
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return "", errors.New("the Twitch URL is missing a username")
		}
	}

	value = strings.TrimPrefix(value, "@")
	if !twitchLoginPattern.MatchString(value) {
		return "", errors.New("Twitch usernames use 1–25 letters, numbers, or underscores")
	}
	return strings.ToLower(value), nil
}

func normalizeWebhookURL(input string) (string, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return "", errors.New("enter your Discord webhook URL")
	}

	u, err := url.Parse(value)
	if err != nil || u.Host == "" || u.User != nil {
		return "", errors.New("that does not look like a Discord webhook URL")
	}
	if !strings.EqualFold(u.Scheme, "https") {
		return "", errors.New("the Discord webhook URL must start with https://")
	}
	host := strings.ToLower(u.Hostname())
	if !discordWebhookHosts[host] {
		return "", errors.New("the URL must be a Discord webhook URL")
	}
	if port := u.Port(); port != "" && port != "443" {
		return "", errors.New("the Discord webhook URL has an unexpected port")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "api" || parts[1] != "webhooks" {
		return "", errors.New("the URL must look like https://discord.com/api/webhooks/…")
	}
	if !allDigits(parts[2]) {
		return "", errors.New("the Discord webhook ID is invalid")
	}
	if !webhookTokenPattern.MatchString(parts[3]) {
		return "", errors.New("the Discord webhook token is invalid")
	}

	u.Scheme = "https"
	u.Host = host
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

func allDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
