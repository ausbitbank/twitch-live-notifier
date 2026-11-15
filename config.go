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
	"time"
	"unicode/utf8"
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
	WebhookID    string       `json:"webhook_id"`
	LastStreamID string       `json:"last_stream_id,omitempty"`
	LastPostedAt string       `json:"last_posted_at,omitempty"`
	RetryUntil   string       `json:"retry_until,omitempty"`
	Template     *Template    `json:"template,omitempty"`
	LiveMessage  *LiveMessage `json:"live_message,omitempty"`
}

// LiveMessage records the Discord message posted for a stream on one route so
// later checks can edit it, or delete/replace it when the stream ends. The
// webhook URL is stored with the message because a destination's URL can be
// changed later, and edits must still target the original post.
type LiveMessage struct {
	MessageID          string `json:"message_id"`
	WebhookURL         string `json:"webhook_url,omitempty"`
	WebhookURLSecretID string `json:"webhook_url_secret_id,omitempty"`
	ThreadID           string `json:"thread_id,omitempty"`
	StreamID           string `json:"stream_id,omitempty"`
	Title              string `json:"title,omitempty"`
	Game               string `json:"game,omitempty"`
	VODURL             string `json:"vod_url,omitempty"`
	StartedAt          string `json:"started_at,omitempty"`
	PostedAt           string `json:"posted_at,omitempty"`
	LastEditedAt       string `json:"last_edited_at,omitempty"`
}

// CleanupSettings controls what happens to the stored live message when a
// stream ends. Mode is one of "", "delete", "edit", or "post".
type CleanupSettings struct {
	Mode     string    `json:"mode,omitempty"`
	Template *Template `json:"template,omitempty"`
}

const (
	cleanupOff    = ""
	cleanupDelete = "delete"
	cleanupEdit   = "edit"
	cleanupPost   = "post"

	minLiveUpdateSeconds = 60
	maxLiveUpdateSeconds = 3600
	maxWebhookButtons    = 5
)

// TemplateButton is a link button attached to a webhook message. Labels and
// URLs may contain the normal message placeholders.
type TemplateButton struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type Webhook struct {
	ID          string           `json:"id"`
	Label       string           `json:"label"`
	URL         string           `json:"url"`
	URLSecretID string           `json:"url_secret_id,omitempty"`
	URLRedacted bool             `json:"url_redacted,omitempty"`
	Enabled     bool             `json:"enabled"`
	RoleIDs     []string         `json:"role_ids,omitempty"`
	Template    Template         `json:"template"`
	Username    string           `json:"username,omitempty"`
	AvatarURL   string           `json:"avatar_url,omitempty"`
	ThreadID    string           `json:"thread_id,omitempty"`
	Buttons     []TemplateButton `json:"buttons,omitempty"`

	UpdateSeconds int             `json:"update_seconds,omitempty"`
	Cleanup       CleanupSettings `json:"cleanup,omitempty"`

	LastError           string `json:"last_error,omitempty"`
	LastErrorAt         string `json:"last_error_at,omitempty"`
	LastSuccessAt       string `json:"last_success_at,omitempty"`
	ConsecutiveFailures int    `json:"consecutive_failures,omitempty"`
	Invalid             bool   `json:"invalid,omitempty"`
}

type Template struct {
	Content          string          `json:"content"`
	EmbedTitle       string          `json:"embed_title"`
	EmbedDescription string          `json:"embed_description"`
	Footer           string          `json:"footer"`
	Color            string          `json:"color"`
	ShowAuthor       bool            `json:"show_author"`
	ShowImage        bool            `json:"show_image"`
	ShowGame         bool            `json:"show_game"`
	ShowViewers      bool            `json:"show_viewers"`
	ShowLiveSince    bool            `json:"show_live_since"`
	PingStreamer     *bool           `json:"ping_streamer,omitempty"`
	PingRoles        *bool           `json:"ping_roles,omitempty"`
	AllowEveryone    bool            `json:"allow_everyone,omitempty"`
	CustomFields     []TemplateField `json:"custom_fields,omitempty"`
	FieldOrder       []string        `json:"field_order,omitempty"`
}

type TemplateField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type Settings struct {
	PollSeconds        int    `json:"poll_seconds"`
	StartAtLogin       bool   `json:"start_at_login"`
	Paused             bool   `json:"paused"`
	Theme              string `json:"theme"`
	DisableUpdateCheck bool   `json:"disable_update_check,omitempty"`
	DashboardTokenHash string `json:"dashboard_token_hash,omitempty"`
}

// WebhookInput is the validated user-editable shape of a destination. It keeps
// the HTTP handlers free of the growing list of individual webhook fields.
type WebhookInput struct {
	Label         string
	URL           string
	RoleIDs       []string
	Enabled       bool
	Template      Template
	Username      string
	AvatarURL     string
	ThreadID      string
	Buttons       []TemplateButton
	UpdateSeconds int
	Cleanup       CleanupSettings
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

func defaultEndedTemplate() Template {
	pingStreamer := false
	pingRoles := false
	return Template{
		Content:          "📴 **{streamer}**'s stream has ended.",
		EmbedTitle:       "{title}",
		EmbedDescription: "Thanks for watching! Last playing **{game}**.\n\n{vod}",
		Footer:           appName,
		Color:            "#2C2F33",
		ShowAuthor:       true,
		ShowImage:        false,
		ShowGame:         false,
		ShowViewers:      false,
		ShowLiveSince:    true,
		PingStreamer:     &pingStreamer,
		PingRoles:        &pingRoles,
	}
}

func normalizeCleanupSettings(cleanup CleanupSettings) CleanupSettings {
	cleanup.Mode = strings.ToLower(strings.TrimSpace(cleanup.Mode))
	switch cleanup.Mode {
	case cleanupDelete, cleanupEdit, cleanupPost:
	case cleanupOff, "off", "none":
		cleanup.Mode = cleanupOff
	default:
		cleanup.Mode = cleanupOff
	}
	if cleanup.Mode == cleanupEdit || cleanup.Mode == cleanupPost {
		if cleanup.Template == nil {
			template := defaultEndedTemplate()
			cleanup.Template = &template
		} else {
			cleanup.Template = normalizeEndedTemplate(*cleanup.Template)
		}
	} else {
		cleanup.Template = nil
	}
	return cleanup
}

// normalizeEndedTemplate keeps the same limits as a live template but defaults
// pings to off so closing a stream never pings a whole community by accident.
func normalizeEndedTemplate(t Template) *Template {
	if t.Color == "" {
		t.Color = "#2C2F33"
	}
	if t.Content == "" && t.EmbedTitle == "" && t.EmbedDescription == "" &&
		t.Footer == "" && !t.ShowAuthor && !t.ShowImage && !t.ShowGame &&
		!t.ShowViewers && !t.ShowLiveSince && len(t.CustomFields) == 0 {
		t = defaultEndedTemplate()
		return &t
	}
	if t.PingStreamer == nil {
		value := false
		t.PingStreamer = &value
	}
	if t.PingRoles == nil {
		value := false
		t.PingRoles = &value
	}
	t.AllowEveryone = false
	t.CustomFields = normalizeCustomFields(t.CustomFields)
	t.FieldOrder = normalizeFieldOrder(t.FieldOrder, len(t.CustomFields))
	return &t
}

func normalizeUpdateSeconds(seconds int) int {
	if seconds <= 0 {
		return 0
	}
	if seconds < minLiveUpdateSeconds {
		return minLiveUpdateSeconds
	}
	if seconds > maxLiveUpdateSeconds {
		return maxLiveUpdateSeconds
	}
	return seconds
}

// normalizeLiveMessage drops a record that cannot possibly be managed. Other
// fields are retained so cleanup can still use the original title/game when
// Twitch has already moved on to another broadcast.
func normalizeRetryUntil(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil || !parsed.After(time.Now()) {
		return ""
	}
	return value
}

func normalizeLiveMessage(message *LiveMessage) *LiveMessage {
	if message == nil {
		return nil
	}
	message.MessageID = strings.TrimSpace(message.MessageID)
	if message.MessageID == "" {
		return nil
	}
	message.WebhookURL = strings.TrimSpace(message.WebhookURL)
	message.ThreadID = strings.TrimSpace(message.ThreadID)
	message.StreamID = strings.TrimSpace(message.StreamID)
	message.Title = strings.TrimSpace(message.Title)
	message.Game = strings.TrimSpace(message.Game)
	message.VODURL = strings.TrimSpace(message.VODURL)
	message.StartedAt = strings.TrimSpace(message.StartedAt)
	message.PostedAt = strings.TrimSpace(message.PostedAt)
	message.LastEditedAt = strings.TrimSpace(message.LastEditedAt)
	return message
}

func cloneLiveMessage(message *LiveMessage) *LiveMessage {
	if message == nil {
		return nil
	}
	clone := *message
	return &clone
}

func cloneCleanupSettings(cleanup CleanupSettings) CleanupSettings {
	clone := cleanup
	if cleanup.Template != nil {
		template := cloneTemplate(*cleanup.Template)
		clone.Template = &template
	}
	return clone
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
	resolveConfigSecrets(&cfg, filepath.Dir(path))
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
		if username, avatar, err := normalizeWebhookIdentity(webhook.Username, webhook.AvatarURL); err == nil {
			webhook.Username, webhook.AvatarURL = username, avatar
		} else {
			webhook.Username, webhook.AvatarURL = "", ""
		}
		if threadID, err := normalizeThreadID(webhook.ThreadID); err == nil {
			webhook.ThreadID = threadID
		} else {
			webhook.ThreadID = ""
		}
		if buttons, err := normalizeButtons(webhook.Buttons); err == nil {
			webhook.Buttons = buttons
		} else {
			webhook.Buttons = nil
		}
		webhook.UpdateSeconds = normalizeUpdateSeconds(webhook.UpdateSeconds)
		webhook.Cleanup = normalizeCleanupSettings(webhook.Cleanup)
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
			destination.RetryUntil = normalizeRetryUntil(destination.RetryUntil)
			if destination.Template != nil {
				normalized := normalizeTemplate(*destination.Template)
				destination.Template = &normalized
			}
			destination.LiveMessage = normalizeLiveMessage(destination.LiveMessage)
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
	cfg = protectConfigSecrets(cfg, filepath.Dir(path))

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
		clone.Channels[i].Destinations = make([]Destination, len(channel.Destinations))
		for j, destination := range channel.Destinations {
			clone.Channels[i].Destinations[j] = destination
			if destination.Template != nil {
				template := cloneTemplate(*destination.Template)
				clone.Channels[i].Destinations[j].Template = &template
			}
			clone.Channels[i].Destinations[j].LiveMessage = cloneLiveMessage(destination.LiveMessage)
		}
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
		clone.Webhooks[i].Buttons = append([]TemplateButton(nil), webhook.Buttons...)
		clone.Webhooks[i].Template = cloneTemplate(webhook.Template)
		clone.Webhooks[i].Cleanup = cloneCleanupSettings(webhook.Cleanup)
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
	clone.CustomFields = cloneTemplateFields(template.CustomFields)
	clone.FieldOrder = cloneStringSlice(template.FieldOrder)
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

func normalizeWebhookIdentity(username, avatarURL string) (string, string, error) {
	username = strings.TrimSpace(username)
	if len([]rune(username)) > 80 {
		return "", "", errors.New("webhook names must be 80 characters or fewer")
	}
	avatarURL = strings.TrimSpace(avatarURL)
	if avatarURL == "" {
		return username, "", nil
	}
	parsed, err := url.Parse(avatarURL)
	if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", "", errors.New("avatar URLs must be a normal http(s) URL")
	}
	if len(avatarURL) > 2048 {
		return "", "", errors.New("avatar URLs must be 2,048 characters or fewer")
	}
	return username, avatarURL, nil
}

func normalizeButtons(input []TemplateButton) ([]TemplateButton, error) {
	if len(input) == 0 {
		return nil, nil
	}
	if len(input) > maxWebhookButtons {
		return nil, fmt.Errorf("add at most %d link buttons per destination", maxWebhookButtons)
	}
	buttons := make([]TemplateButton, 0, len(input))
	for _, raw := range input {
		label := strings.TrimSpace(raw.Label)
		buttonURL := strings.TrimSpace(raw.URL)
		if label == "" && buttonURL == "" {
			continue
		}
		if label == "" {
			return nil, errors.New("every link button needs a label")
		}
		if utf8.RuneCountInString(label) > 80 {
			return nil, errors.New("button labels must be 80 characters or fewer")
		}
		if buttonURL == "" {
			return nil, errors.New("every link button needs a URL")
		}
		candidate := placeholderPattern.ReplaceAllString(buttonURL, "https://example.com/x")
		parsed, err := url.Parse(candidate)
		if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			return nil, errors.New("button URLs must be normal http(s) links")
		}
		buttons = append(buttons, TemplateButton{Label: label, URL: buttonURL})
	}
	if len(buttons) == 0 {
		return nil, nil
	}
	return buttons, nil
}

func normalizeThreadID(input string) (string, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return "", nil
	}
	if !discordIDPattern.MatchString(value) {
		return "", errors.New("thread IDs are 17–20 digit numbers; paste the thread ID")
	}
	return value, nil
}

func validateCleanupSettings(cleanup CleanupSettings) (CleanupSettings, error) {
	normalized := normalizeCleanupSettings(cleanup)
	rawMode := strings.ToLower(strings.TrimSpace(cleanup.Mode))
	if rawMode != cleanupOff && rawMode != "off" && rawMode != "none" &&
		rawMode != cleanupDelete && rawMode != cleanupEdit && rawMode != cleanupPost {
		if rawMode != "" {
			return CleanupSettings{}, errors.New("cleanup must be off, delete, edit, or post")
		}
	}
	if normalized.Mode == cleanupEdit || normalized.Mode == cleanupPost {
		if normalized.Template == nil {
			return CleanupSettings{}, errors.New("add ended-message text for this cleanup mode")
		}
		validated, err := validateTemplate(*normalized.Template)
		if err != nil {
			return CleanupSettings{}, fmt.Errorf("ended message: %w", err)
		}
		normalized.Template = normalizeEndedTemplate(validated)
	}
	return normalized, nil
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
