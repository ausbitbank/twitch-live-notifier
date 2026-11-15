package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ChannelStatus struct {
	Live             bool   `json:"live"`
	Title            string `json:"title,omitempty"`
	Game             string `json:"game,omitempty"`
	Viewers          int    `json:"viewers"`
	Thumbnail        string `json:"thumbnail,omitempty"`
	StartedAt        string `json:"started_at,omitempty"`
	LastTitle        string `json:"last_title,omitempty"`
	LastGame         string `json:"last_game,omitempty"`
	LastStartedAt    string `json:"last_started_at,omitempty"`
	LastThumbnail    string `json:"last_thumbnail,omitempty"`
	LastChecked      string `json:"last_checked,omitempty"`
	Error            string `json:"error,omitempty"`
	FilterReason     string `json:"filter_reason,omitempty"`
	FilteredStreamID string `json:"filtered_stream_id,omitempty"`
}

type Activity struct {
	ID      string    `json:"id"`
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

type State struct {
	Version               string                   `json:"version"`
	Config                Config                   `json:"config"`
	Statuses              map[string]ChannelStatus `json:"statuses"`
	Activity              []Activity               `json:"activity"`
	Checking              bool                     `json:"checking"`
	LastCheck             string                   `json:"last_check,omitempty"`
	DataDir               string                   `json:"data_dir"`
	Platform              string                   `json:"platform"`
	StartAtLoginSupported bool                     `json:"start_at_login_supported"`
	Placeholders          []Placeholder            `json:"placeholders"`
	DefaultTemplate       Template                 `json:"default_template"`
	DefaultEndedTemplate  Template                 `json:"default_ended_template"`
	DashboardTokenSet     bool                     `json:"dashboard_token_set"`
	Update                UpdateStatus             `json:"update"`
}

type CheckSummary struct {
	Channels int
	Live     int
	Posted   int
	Errors   int
}

// defaultCheckTimeout prevents a huge live wave or a long rate-limit window
// from leaving a check running forever. A later scheduled check can continue.
const defaultCheckTimeout = 15 * time.Minute

type App struct {
	mu              sync.RWMutex
	cfg             Config
	configPath      string
	statuses        map[string]ChannelStatus
	activity        []Activity
	checking        bool
	lastCheck       time.Time
	startedAt       time.Time
	wake            chan struct{}
	ctx             context.Context
	cancel          context.CancelFunc
	onQuit          func()
	enableAutoCheck bool
	updateStatus    UpdateStatus
}

func NewApp(configPath string) (*App, error) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}
	normalizeConfig(&cfg)
	app := &App{
		cfg:             cfg,
		configPath:      configPath,
		statuses:        make(map[string]ChannelStatus),
		startedAt:       time.Now(),
		wake:            make(chan struct{}, 1),
		enableAutoCheck: true,
	}
	app.activity = append(app.activity, Activity{
		ID:      newID(),
		Time:    time.Now(),
		Level:   "info",
		Message: "Dashboard started.",
	})
	if err := saveConfig(configPath, cfg); err != nil {
		return nil, err
	}
	return app, nil
}

func (a *App) Start(ctx context.Context) {
	a.mu.Lock()
	a.ctx, a.cancel = context.WithCancel(ctx)
	cfg := a.cfg
	a.mu.Unlock()

	if cfg.Settings.StartAtLogin {
		if err := setAutoStart(true); err != nil {
			a.addActivity("error", "Could not update the Windows startup entry: "+err.Error())
		}
	}

	go a.monitorLoop()
	if a.ctx != nil {
		go a.updateLoop(a.ctx)
	}
	go func() {
		select {
		case <-time.After(1500 * time.Millisecond):
			if !a.IsPaused() {
				a.TriggerCheck()
			}
		case <-a.ctx.Done():
		}
	}()
}

func (a *App) Shutdown() {
	a.mu.Lock()
	cancel := a.cancel
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (a *App) monitorLoop() {
	a.mu.RLock()
	ctx := a.ctx
	a.mu.RUnlock()
	if ctx == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.wake:
		case <-time.After(a.pollInterval()):
			if !a.IsPaused() {
				a.TriggerCheck()
			}
		}
	}
}

func (a *App) TriggerCheck() bool {
	a.mu.Lock()
	if a.checking {
		a.mu.Unlock()
		return false
	}
	a.checking = true
	a.mu.Unlock()

	go func() {
		a.mu.RLock()
		parent := a.ctx
		a.mu.RUnlock()
		if parent == nil {
			parent = context.Background()
		}
		ctx, cancel := context.WithTimeout(parent, defaultCheckTimeout)
		defer cancel()
		_, _ = a.runCheck(ctx)
	}()
	return true
}

func (a *App) CheckChannelNow(id string) (bool, error) {
	a.mu.Lock()
	if a.checking {
		a.mu.Unlock()
		return false, errors.New("a check is already running")
	}
	index := a.channelIndexLocked(id)
	if index < 0 {
		a.mu.Unlock()
		return false, errors.New("streamer not found")
	}
	channel := a.cfg.Channels[index]
	webhooks := append([]Webhook(nil), a.cfg.Webhooks...)
	a.checking = true
	a.mu.Unlock()

	go func() {
		a.mu.RLock()
		parent := a.ctx
		a.mu.RUnlock()
		if parent == nil {
			parent = context.Background()
		}
		ctx, cancel := context.WithTimeout(parent, defaultCheckTimeout)
		defer cancel()
		_, _ = a.checkChannel(ctx, channel, webhooks)
		a.finishCheck()
	}()
	return true, nil
}

func (a *App) CheckNow(ctx context.Context) (CheckSummary, error) {
	a.mu.Lock()
	if a.checking {
		a.mu.Unlock()
		return CheckSummary{}, errors.New("a check is already running")
	}
	a.checking = true
	a.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, defaultCheckTimeout)
	defer cancel()
	return a.runCheck(ctx)
}

func (a *App) runCheck(ctx context.Context) (CheckSummary, error) {
	cfg := a.ConfigSnapshot()
	summary := CheckSummary{}
	webhooks := cfg.Webhooks

	for _, channel := range cfg.Channels {
		if !channel.Enabled {
			continue
		}
		summary.Channels++
		select {
		case <-ctx.Done():
			a.finishCheck()
			return summary, ctx.Err()
		default:
		}

		posted, err := a.checkChannel(ctx, channel, webhooks)
		summary.Posted += posted
		if err != nil {
			summary.Errors++
		}
		if a.statusFor(channel.ID).Live {
			summary.Live++
		}
	}

	a.finishCheck()
	return summary, nil
}

func (a *App) finishCheck() {
	a.mu.Lock()
	a.checking = false
	a.lastCheck = time.Now()
	a.mu.Unlock()
}

func (a *App) checkChannel(ctx context.Context, channel Channel, webhooks []Webhook) (int, error) {
	previous := a.statusFor(channel.ID)
	now := time.Now()

	user, err := fetchTwitchChannelContext(ctx, channel.Login)
	if err != nil {
		a.mu.Lock()
		status := previous
		status.LastChecked = now.UTC().Format(time.RFC3339)
		status.Error = err.Error()
		a.statuses[channel.ID] = status
		a.mu.Unlock()
		if previous.Error != err.Error() {
			a.addActivity("error", fmt.Sprintf("Could not check %s: %v", channel.DisplayName, err))
		}
		return 0, err
	}

	status := statusFromUser(user, now)
	a.mu.Lock()
	profileChanged := false
	if index := a.channelIndexLocked(channel.ID); index >= 0 {
		displayName := displayNameFor(user, channel.Login)
		if a.cfg.Channels[index].DisplayName != displayName {
			a.cfg.Channels[index].DisplayName = displayName
			profileChanged = true
		}
		if a.cfg.Channels[index].AvatarURL != user.ProfileImageURL {
			a.cfg.Channels[index].AvatarURL = user.ProfileImageURL
			profileChanged = true
		}
		if a.cfg.Channels[index].BannerURL != user.BannerImageURL {
			a.cfg.Channels[index].BannerURL = user.BannerImageURL
			profileChanged = true
		}
	}
	a.statuses[channel.ID] = status
	if profileChanged {
		a.persistLocked()
	}
	a.mu.Unlock()

	if status.Live && !previous.Live {
		a.addActivity("success", fmt.Sprintf("%s is live on Twitch.", displayNameFor(user, channel.Login)))
	} else if !status.Live && previous.Live {
		a.addActivity("info", fmt.Sprintf("%s went offline.", displayNameFor(user, channel.Login)))
	}

	if !status.Live {
		a.cleanupLiveMessages(ctx, channel, webhooks, streamInfoFromTwitch(user, false), now, "")
		return 0, nil
	}

	info := streamInfoFromTwitch(user, true)
	info.DiscordID = channel.DiscordID

	// A message from an earlier stream may still be present if the app was
	// offline when that stream ended. Clean it up best-effort before posting
	// the new one; a cleanup failure must never block the new alert.
	a.cleanupLiveMessages(ctx, channel, webhooks, info, now, info.StreamID)

	if allowed, reason := evaluateFilters(channel, info, now, time.Local); !allowed {
		a.recordFilterSkip(channel.ID, info.StreamID, reason)
		return 0, nil
	}

	posted := 0
	var postErrors []string
	for _, destination := range channel.Destinations {
		webhook := findWebhook(webhooks, destination.WebhookID)
		if webhook == nil || !a.webhookAvailable(webhook.ID) {
			continue
		}
		if destinationRateLimited(destination, time.Now()) {
			continue
		}
		if destination.LastStreamID != "" && destination.LastStreamID == info.StreamID {
			a.markLiveSnapshot(channel.ID, webhook.ID, info)
			a.maybeUpdateLiveMessage(ctx, channel.ID, destination, *webhook, info, now)
			continue
		}

		result, postErr := postDiscordContextWithOptions(ctx, webhook.URL, buildPayloadForDestination(*webhook, &destination, info, now, false), discordWebhookOptions{ThreadID: webhook.ThreadID})
		if postErr != nil {
			a.recordWebhookFailure(webhook.ID, postErr)
			if delay := rateLimitDelay(postErr); delay > 0 {
				a.markDestinationRateLimited(channel.ID, webhook.ID, delay)
				a.addActivity("error", fmt.Sprintf("Rate limited by %s; retrying after %s.", webhook.Label, formatDuration(delay)))
			} else {
				a.addActivity("error", fmt.Sprintf("Could not post to %s: %v", webhook.Label, postErr))
			}
			postErrors = append(postErrors, fmt.Sprintf("%s: %v", webhook.Label, postErr))
			continue
		}
		a.clearDestinationRateLimit(channel.ID, webhook.ID)
		a.recordWebhookSuccess(webhook.ID)
		a.markPosted(channel.ID, webhook.ID, info, result.MessageID, webhook.URL, webhook.ThreadID, now)
		posted++
		a.addActivity("success", fmt.Sprintf("Posted %q to %s.", info.Title, webhook.Label))
	}

	if len(postErrors) > 0 {
		a.mu.Lock()
		status = a.statuses[channel.ID]
		status.Error = strings.Join(postErrors, " • ")
		a.statuses[channel.ID] = status
		a.mu.Unlock()
		return posted, errors.New(strings.Join(postErrors, " • "))
	}
	return posted, nil
}

// cleanupLiveMessages applies each destination's stream-end policy. When
// keepStreamID is set, records for that stream are left alone (the stream is
// still live). Records for older streams are always cleared after the attempt
// so a failed cleanup cannot hold up the next stream.
func (a *App) cleanupLiveMessages(ctx context.Context, channel Channel, webhooks []Webhook, fallback StreamInfo, now time.Time, keepStreamID string) {
	replacing := keepStreamID != ""
	for _, destination := range channel.Destinations {
		record := destination.LiveMessage
		if record == nil || record.MessageID == "" {
			continue
		}
		if keepStreamID != "" && record.StreamID == keepStreamID {
			continue
		}
		webhook := findWebhook(webhooks, destination.WebhookID)
		if webhook == nil {
			a.clearLiveMessage(channel.ID, destination.WebhookID)
			continue
		}
		if !webhook.Enabled {
			continue
		}
		mode := normalizeCleanupSettings(webhook.Cleanup).Mode
		if mode == cleanupOff && !replacing {
			// Cleanup is off; leave the original message untouched and forget
			// it so it is not retried forever.
			a.clearLiveMessage(channel.ID, destination.WebhookID)
			continue
		}
		if mode != cleanupOff && destinationRateLimited(destination, time.Now()) {
			continue
		}
		clear, err := a.applyCleanup(ctx, *webhook, destination, fallback, now)
		if err != nil {
			a.recordWebhookFailure(webhook.ID, err)
			if delay := rateLimitDelay(err); delay > 0 {
				a.markDestinationRateLimited(channel.ID, webhook.ID, delay)
				a.addActivity("error", fmt.Sprintf("Rate limited while cleaning up the live message in %s; retrying after %s.", webhook.Label, formatDuration(delay)))
			} else {
				a.addActivity("error", fmt.Sprintf("Could not %s the live message in %s: %v", cleanupVerb(mode), webhook.Label, err))
			}
		} else if mode != cleanupOff {
			a.clearDestinationRateLimit(channel.ID, webhook.ID)
			message := "Cleaned up the live message in " + webhook.Label + "."
			switch mode {
			case cleanupDelete:
				message = "Deleted the live message in " + webhook.Label + "."
			case cleanupEdit:
				message = "Updated the live message in " + webhook.Label + "."
			case cleanupPost:
				message = "Posted the ended message in " + webhook.Label + "."
			}
			a.addActivity("info", message)
		}
		if clear || replacing {
			a.clearLiveMessage(channel.ID, destination.WebhookID)
		}
	}
}

func cleanupVerb(mode string) string {
	switch mode {
	case cleanupDelete:
		return "delete"
	case cleanupEdit:
		return "update"
	case cleanupPost:
		return "post the ended message for"
	default:
		return "clean up"
	}
}

// applyCleanup returns whether the stored record can be forgotten. Transient
// failures keep it for the next check; permanent credential/target failures
// drop it so the app does not hammer Discord.
func (a *App) applyCleanup(ctx context.Context, webhook Webhook, destination Destination, fallback StreamInfo, now time.Time) (bool, error) {
	record := destination.LiveMessage
	if record == nil || record.MessageID == "" {
		return true, nil
	}
	cleanup := normalizeCleanupSettings(webhook.Cleanup)
	targetURL := firstNonEmpty(record.WebhookURL, webhook.URL)
	if targetURL == "" {
		return true, errors.New("the original webhook URL is no longer available")
	}
	// An empty stored thread ID means the original post was in the main
	// channel. Do not fall back to the current destination thread setting,
	// which may have changed after the post was created.
	threadID := record.ThreadID
	info := cleanupInfo(record, fallback)

	switch cleanup.Mode {
	case cleanupDelete:
		if err := deleteDiscordContext(ctx, targetURL, record.MessageID, threadID); err != nil {
			return cleanupClearOnError(err), err
		}
		return true, nil
	case cleanupEdit:
		template := routeTemplateFor(webhook, &destination)
		if cleanup.Template != nil {
			template = *cleanup.Template
		}
		payload := buildPayloadWithTemplate(webhook, template, info, now, false)
		payload.Username = ""
		payload.AvatarURL = ""
		if _, err := editDiscordContext(ctx, targetURL, record.MessageID, payload, threadID); err != nil {
			return cleanupClearOnError(err), err
		}
		return true, nil
	case cleanupPost:
		template := routeTemplateFor(webhook, &destination)
		if cleanup.Template != nil {
			template = *cleanup.Template
		}
		payload := buildPayloadWithTemplate(webhook, template, info, now, false)
		if _, err := postDiscordContextWithOptions(ctx, targetURL, payload, discordWebhookOptions{ThreadID: threadID}); err != nil {
			return cleanupClearOnError(err), err
		}
		return true, nil
	default:
		return true, nil
	}
}

func cleanupClearOnError(err error) bool {
	var httpErr *discordHTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 && httpErr.StatusCode != http.StatusTooManyRequests
	}
	return false
}

func cleanupInfo(record *LiveMessage, fallback StreamInfo) StreamInfo {
	info := fallback
	info.Live = false
	if record == nil {
		return info
	}
	if record.StreamID != "" {
		info.StreamID = record.StreamID
	}
	if record.Title != "" {
		info.Title = record.Title
	}
	if record.Game != "" {
		info.Game = record.Game
	}
	if record.StartedAt != "" {
		info.StartedAt = parseTime(record.StartedAt)
	}
	if record.VODURL != "" {
		info.VODURL = record.VODURL
	}
	return info
}

func (a *App) maybeUpdateLiveMessage(ctx context.Context, channelID string, destination Destination, webhook Webhook, info StreamInfo, now time.Time) {
	if webhook.UpdateSeconds <= 0 {
		return
	}
	if destinationRateLimited(destination, time.Now()) {
		return
	}
	record := destination.LiveMessage
	if record == nil || record.MessageID == "" || record.StreamID != info.StreamID {
		return
	}
	interval := time.Duration(normalizeUpdateSeconds(webhook.UpdateSeconds)) * time.Second
	if last := parseTime(record.LastEditedAt); !last.IsZero() && now.Sub(last) < interval {
		return
	}
	targetURL := firstNonEmpty(record.WebhookURL, webhook.URL)
	if targetURL == "" {
		return
	}
	threadID := record.ThreadID
	payload := buildPayloadForDestination(webhook, &destination, info, now, false)
	payload.Username = ""
	payload.AvatarURL = ""
	if _, err := editDiscordContext(ctx, targetURL, record.MessageID, payload, threadID); err != nil {
		a.recordWebhookFailure(webhook.ID, err)
		if delay := rateLimitDelay(err); delay > 0 {
			a.markDestinationRateLimited(channelID, webhook.ID, delay)
			a.addActivity("error", fmt.Sprintf("Rate limited updating the live message in %s; retrying after %s.", webhook.Label, formatDuration(delay)))
		} else {
			if cleanupClearOnError(err) {
				a.clearLiveMessage(channelID, webhook.ID)
			}
			a.addActivity("error", fmt.Sprintf("Could not update the live message in %s: %v", webhook.Label, err))
		}
		return
	}
	a.clearDestinationRateLimit(channelID, webhook.ID)
	a.recordWebhookSuccess(webhook.ID)
	a.markDestinationEdited(channelID, webhook.ID, now)
	a.addActivity("info", fmt.Sprintf("Updated the live message in %s.", webhook.Label))
}

func (a *App) markPosted(channelID, webhookID string, info StreamInfo, messageID, messageURL, threadID string, now time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index := a.channelIndexLocked(channelID); index >= 0 {
		for i := range a.cfg.Channels[index].Destinations {
			destination := &a.cfg.Channels[index].Destinations[i]
			if destination.WebhookID == webhookID {
				timestamp := now.UTC().Format(time.RFC3339)
				destination.LastStreamID = info.StreamID
				destination.LastPostedAt = timestamp
				destination.RetryUntil = ""
				if messageID != "" {
					destination.LiveMessage = &LiveMessage{
						MessageID:    messageID,
						WebhookURL:   messageURL,
						ThreadID:     threadID,
						StreamID:     info.StreamID,
						Title:        info.Title,
						Game:         info.Game,
						VODURL:       info.VODURL,
						StartedAt:    formatOptionalTime(info.StartedAt),
						PostedAt:     timestamp,
						LastEditedAt: timestamp,
					}
				}
				break
			}
		}
		a.cfg.Channels[index].LastAlertAt = now.UTC().Format(time.RFC3339)
		a.cfg.Channels[index].LastAlertDate = now.Local().Format("2006-01-02")
		a.cfg.Channels[index].LastAlertStreamID = info.StreamID
	}
	_ = a.persistLocked()
}

func formatOptionalTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func (a *App) markLiveSnapshot(channelID, webhookID string, info StreamInfo) {
	a.mu.Lock()
	defer a.mu.Unlock()
	index := a.channelIndexLocked(channelID)
	if index < 0 {
		return
	}
	for i := range a.cfg.Channels[index].Destinations {
		destination := &a.cfg.Channels[index].Destinations[i]
		if destination.WebhookID != webhookID || destination.LiveMessage == nil {
			continue
		}
		record := destination.LiveMessage
		if record.StreamID != info.StreamID {
			return
		}
		changed := false
		if info.Title != "" && record.Title != info.Title {
			record.Title = info.Title
			changed = true
		}
		if info.Game != "" && record.Game != info.Game {
			record.Game = info.Game
			changed = true
		}
		if info.VODURL != "" && record.VODURL != info.VODURL {
			record.VODURL = info.VODURL
			changed = true
		}
		if started := formatOptionalTime(info.StartedAt); started != "" && record.StartedAt != started {
			record.StartedAt = started
			changed = true
		}
		if changed {
			_ = a.persistLocked()
		}
		return
	}
}

func (a *App) markDestinationEdited(channelID, webhookID string, now time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	index := a.channelIndexLocked(channelID)
	if index < 0 {
		return
	}
	for i := range a.cfg.Channels[index].Destinations {
		destination := &a.cfg.Channels[index].Destinations[i]
		if destination.WebhookID == webhookID && destination.LiveMessage != nil {
			destination.LiveMessage.LastEditedAt = now.UTC().Format(time.RFC3339)
			destination.RetryUntil = ""
			_ = a.persistLocked()
			return
		}
	}
}

func destinationRateLimited(destination Destination, now time.Time) bool {
	value := strings.TrimSpace(destination.RetryUntil)
	if value == "" {
		return false
	}
	until, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return false
	}
	return until.After(now)
}

func rateLimitDelay(err error) time.Duration {
	var httpErr *discordHTTPError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusTooManyRequests {
		return 0
	}
	delay := httpErr.RetryAfter
	if delay <= 0 {
		delay = discordDefaultRetryIn
	}
	if delay > discordMaxDefer {
		delay = discordMaxDefer
	}
	return delay
}

func (a *App) markDestinationRateLimited(channelID, webhookID string, delay time.Duration) {
	if delay <= 0 {
		return
	}
	if delay > discordMaxDefer {
		delay = discordMaxDefer
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	index := a.channelIndexLocked(channelID)
	if index < 0 {
		return
	}
	for i := range a.cfg.Channels[index].Destinations {
		destination := &a.cfg.Channels[index].Destinations[i]
		if destination.WebhookID == webhookID {
			destination.RetryUntil = time.Now().Add(delay).UTC().Format(time.RFC3339)
			_ = a.persistLocked()
			return
		}
	}
}

func (a *App) clearDestinationRateLimit(channelID, webhookID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	index := a.channelIndexLocked(channelID)
	if index < 0 {
		return
	}
	for i := range a.cfg.Channels[index].Destinations {
		destination := &a.cfg.Channels[index].Destinations[i]
		if destination.WebhookID == webhookID && destination.RetryUntil != "" {
			destination.RetryUntil = ""
			_ = a.persistLocked()
			return
		}
	}
}

func (a *App) clearLiveMessage(channelID, webhookID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	index := a.channelIndexLocked(channelID)
	if index < 0 {
		return
	}
	for i := range a.cfg.Channels[index].Destinations {
		destination := &a.cfg.Channels[index].Destinations[i]
		if destination.WebhookID == webhookID && destination.LiveMessage != nil {
			destination.LiveMessage = nil
			_ = a.persistLocked()
			return
		}
	}
}

func (a *App) webhookAvailable(id string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	index := a.webhookIndexLocked(id)
	if index < 0 {
		return false
	}
	webhook := a.cfg.Webhooks[index]
	return webhook.Enabled && !webhook.Invalid
}

func (a *App) recordWebhookSuccess(id string) {
	now := time.Now().UTC().Format(time.RFC3339)
	a.mu.Lock()
	defer a.mu.Unlock()
	index := a.webhookIndexLocked(id)
	if index < 0 {
		return
	}
	webhook := &a.cfg.Webhooks[index]
	webhook.LastSuccessAt = now
	webhook.LastError = ""
	webhook.LastErrorAt = ""
	webhook.ConsecutiveFailures = 0
	webhook.Invalid = false
	_ = a.persistLocked()
}

func (a *App) recordWebhookFailure(id string, err error) {
	now := time.Now().UTC().Format(time.RFC3339)
	a.mu.Lock()
	defer a.mu.Unlock()
	index := a.webhookIndexLocked(id)
	if index < 0 {
		return
	}
	webhook := &a.cfg.Webhooks[index]
	webhook.LastError = truncateText(err.Error(), 300)
	webhook.LastErrorAt = now
	webhook.ConsecutiveFailures++
	if isPermanentDiscordError(err) {
		webhook.Invalid = true
	}
	_ = a.persistLocked()
}

func (a *App) recordFilterSkip(channelID, streamID, reason string) {
	a.mu.Lock()
	status, ok := a.statuses[channelID]
	changed := !ok || status.FilteredStreamID != streamID
	if ok {
		status.FilterReason = reason
		status.FilteredStreamID = streamID
		a.statuses[channelID] = status
	}
	name := channelID
	if index := a.channelIndexLocked(channelID); index >= 0 {
		name = a.cfg.Channels[index].DisplayName
	}
	a.mu.Unlock()

	if changed {
		a.addActivity("info", fmt.Sprintf("Skipped alert for %s: %s.", name, reason))
	}
}

func (a *App) SetChannelFilters(id string, filters Filters) error {
	validated, err := validateFilters(filters)
	if err != nil {
		return err
	}

	a.mu.Lock()
	index := a.channelIndexLocked(id)
	if index < 0 {
		a.mu.Unlock()
		return errors.New("streamer not found")
	}
	a.cfg.Channels[index].Filters = validated
	name := a.cfg.Channels[index].DisplayName
	err = a.persistLocked()
	a.mu.Unlock()
	if err != nil {
		return err
	}
	a.addActivity("info", "Updated filters for "+name+".")
	if a.enableAutoCheck {
		a.TriggerCheck()
	}
	return nil
}

func (a *App) statusFor(channelID string) ChannelStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.statuses[channelID]
}

func statusFromUser(user *twitchUser, now time.Time) ChannelStatus {
	status := ChannelStatus{
		LastChecked: now.UTC().Format(time.RFC3339),
	}
	if user.Stream != nil {
		status.Live = true
		status.Title = streamTitle(user.Stream)
		status.Game = strings.TrimSpace(user.Stream.Game.Name)
		status.Viewers = user.Stream.ViewersCount
		status.Thumbnail = firstNonEmpty(user.Stream.PreviewImageURL, user.BannerImageURL, user.ProfileImageURL)
		status.StartedAt = user.Stream.CreatedAt
	}
	if user.LastBroadcast != nil {
		status.LastTitle = strings.TrimSpace(user.LastBroadcast.Title)
		status.LastGame = strings.TrimSpace(user.LastBroadcast.Game.Name)
		status.LastStartedAt = user.LastBroadcast.StartedAt
		status.LastThumbnail = firstNonEmpty(user.BannerImageURL, user.LastBroadcast.Game.BoxArtURL, user.ProfileImageURL)
	}
	return status
}

func (a *App) ConfigSnapshot() Config {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return cloneConfig(a.cfg)
}

func (a *App) State() State {
	a.mu.RLock()
	cfg := cloneConfig(a.cfg)
	statuses := make(map[string]ChannelStatus, len(a.statuses))
	for id, status := range a.statuses {
		statuses[id] = status
	}
	activity := append([]Activity(nil), a.activity...)
	checking := a.checking
	lastCheck := a.lastCheck
	dataDir := filepath.Dir(a.configPath)
	update := a.updateStatus
	a.mu.RUnlock()
	if update.CurrentVersion == "" {
		update.CurrentVersion = Version
	}

	lastCheckText := ""
	if !lastCheck.IsZero() {
		lastCheckText = lastCheck.UTC().Format(time.RFC3339)
	}
	dashboardTokenSet := cfg.Settings.DashboardTokenHash != ""
	cfg.Settings.DashboardTokenHash = ""
	return State{
		Version:               Version,
		Config:                cfg,
		Statuses:              statuses,
		Activity:              activity,
		Checking:              checking,
		LastCheck:             lastCheckText,
		DataDir:               dataDir,
		Platform:              platformName(),
		StartAtLoginSupported: startAtLoginSupported(),
		Placeholders:          templatePlaceholders,
		DefaultTemplate:       defaultTemplate(),
		DefaultEndedTemplate:  defaultEndedTemplate(),
		DashboardTokenSet:     dashboardTokenSet,
		Update:                update,
	}
}

func (a *App) AddChannel(loginInput string) (Channel, error) {
	login, err := normalizeTwitchLogin(loginInput)
	if err != nil {
		return Channel{}, err
	}

	a.mu.RLock()
	if a.channelByLoginLocked(login) != nil {
		a.mu.RUnlock()
		return Channel{}, fmt.Errorf("%s is already in your streamer list", login)
	}
	webhooks := append([]Webhook(nil), a.cfg.Webhooks...)
	a.mu.RUnlock()

	user, err := fetchTwitchChannel(login)
	if err != nil {
		return Channel{}, err
	}

	channel := Channel{
		ID:          newID(),
		Login:       login,
		DisplayName: displayNameFor(user, login),
		AvatarURL:   user.ProfileImageURL,
		BannerURL:   user.BannerImageURL,
		Enabled:     true,
	}
	if len(webhooks) == 1 {
		channel.Destinations = []Destination{{WebhookID: webhooks[0].ID}}
	}

	a.mu.Lock()
	if a.channelByLoginLocked(login) != nil {
		a.mu.Unlock()
		return Channel{}, fmt.Errorf("%s is already in your streamer list", login)
	}
	a.cfg.Channels = append(a.cfg.Channels, channel)
	a.statuses[channel.ID] = statusFromUser(user, time.Now())
	a.persistLocked()
	a.mu.Unlock()

	a.addActivity("success", "Added "+channel.DisplayName+".")
	if a.enableAutoCheck {
		a.TriggerCheck()
	}
	return channel, nil
}

func (a *App) RemoveChannel(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	index := a.channelIndexLocked(id)
	if index < 0 {
		return errors.New("streamer not found")
	}
	name := a.cfg.Channels[index].DisplayName
	a.cfg.Channels = append(a.cfg.Channels[:index], a.cfg.Channels[index+1:]...)
	delete(a.statuses, id)
	err := a.persistLocked()
	if err == nil {
		a.activity = append(a.activity, Activity{ID: newID(), Time: time.Now(), Level: "info", Message: "Removed " + name + "."})
	}
	return err
}

func (a *App) SetChannelEnabled(id string, enabled bool) error {
	a.mu.Lock()
	index := a.channelIndexLocked(id)
	if index < 0 {
		a.mu.Unlock()
		return errors.New("streamer not found")
	}
	a.cfg.Channels[index].Enabled = enabled
	name := a.cfg.Channels[index].DisplayName
	err := a.persistLocked()
	a.mu.Unlock()
	if err != nil {
		return err
	}
	message := "Paused " + name + "."
	if enabled {
		message = "Resumed " + name + "."
	}
	a.addActivity("info", message)
	if enabled && a.enableAutoCheck {
		a.TriggerCheck()
	}
	return nil
}

func (a *App) SetChannelDiscordID(id, rawDiscordID string) error {
	discordID, err := normalizeDiscordID(rawDiscordID)
	if err != nil {
		return err
	}

	a.mu.Lock()
	index := a.channelIndexLocked(id)
	if index < 0 {
		a.mu.Unlock()
		return errors.New("streamer not found")
	}
	if a.cfg.Channels[index].DiscordID == discordID {
		a.mu.Unlock()
		return nil
	}
	a.cfg.Channels[index].DiscordID = discordID
	name := a.cfg.Channels[index].DisplayName
	err = a.persistLocked()
	a.mu.Unlock()
	if err != nil {
		return err
	}
	if discordID == "" {
		a.addActivity("info", "Removed Discord ping for "+name+".")
	} else {
		a.addActivity("info", "Updated Discord ping for "+name+".")
	}
	return nil
}

func (a *App) SetDestinations(channelID string, webhookIDs []string) error {
	a.mu.Lock()
	index := a.channelIndexLocked(channelID)
	if index < 0 {
		a.mu.Unlock()
		return errors.New("streamer not found")
	}

	existing := make(map[string]Destination)
	for _, destination := range a.cfg.Channels[index].Destinations {
		existing[destination.WebhookID] = destination
	}
	selected := make(map[string]bool, len(webhookIDs))
	destinations := make([]Destination, 0, len(webhookIDs))
	added := false
	for _, webhookID := range webhookIDs {
		if selected[webhookID] || a.webhookIndexLocked(webhookID) < 0 {
			continue
		}
		selected[webhookID] = true
		if destination, ok := existing[webhookID]; ok {
			destinations = append(destinations, destination)
		} else {
			destinations = append(destinations, Destination{WebhookID: webhookID})
			added = true
		}
	}
	name := a.cfg.Channels[index].DisplayName
	a.cfg.Channels[index].Destinations = destinations
	err := a.persistLocked()
	a.mu.Unlock()
	if err != nil {
		return err
	}
	a.addActivity("info", "Updated destinations for "+name+".")
	if added && a.enableAutoCheck {
		a.TriggerCheck()
	}
	return nil
}

func (a *App) SetRouteTemplate(channelID, webhookID string, template *Template) error {
	var validated *Template
	if template != nil {
		value, err := validateTemplate(*template)
		if err != nil {
			return err
		}
		validated = &value
	}

	a.mu.Lock()
	channelIndex := a.channelIndexLocked(channelID)
	if channelIndex < 0 {
		a.mu.Unlock()
		return errors.New("streamer not found")
	}
	webhookIndex := a.webhookIndexLocked(webhookID)
	if webhookIndex < 0 {
		a.mu.Unlock()
		return errors.New("destination not found")
	}
	found := false
	for i := range a.cfg.Channels[channelIndex].Destinations {
		destination := &a.cfg.Channels[channelIndex].Destinations[i]
		if destination.WebhookID != webhookID {
			continue
		}
		destination.Template = validated
		found = true
		break
	}
	if !found {
		a.mu.Unlock()
		return errors.New("this streamer is not connected to that destination")
	}
	channelName := a.cfg.Channels[channelIndex].DisplayName
	webhookLabel := a.cfg.Webhooks[webhookIndex].Label
	err := a.persistLocked()
	a.mu.Unlock()
	if err != nil {
		return err
	}
	if validated == nil {
		a.addActivity("info", fmt.Sprintf("Using the %s template for %s.", webhookLabel, channelName))
	} else {
		a.addActivity("info", fmt.Sprintf("Saved a custom template for %s → %s.", channelName, webhookLabel))
	}
	return nil
}

func (a *App) AddWebhook(label, rawURL string, roleIDs []string, template *Template) (Webhook, error) {
	input := WebhookInput{Label: label, URL: rawURL, RoleIDs: roleIDs, Enabled: true}
	if template != nil {
		input.Template = *template
	}
	return a.AddWebhookInput(input)
}

func (a *App) AddWebhookInput(input WebhookInput) (Webhook, error) {
	validated, err := validateWebhookInput(input)
	if err != nil {
		return Webhook{}, err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for _, webhook := range a.cfg.Webhooks {
		if strings.EqualFold(webhook.Label, validated.Label) {
			return Webhook{}, fmt.Errorf("a destination called %q already exists", validated.Label)
		}
		if webhook.URL == validated.URL {
			return Webhook{}, errors.New("that Discord webhook is already added")
		}
	}
	webhook := Webhook{
		ID:            newID(),
		Label:         validated.Label,
		URL:           validated.URL,
		Enabled:       true,
		RoleIDs:       validated.RoleIDs,
		Template:      validated.Template,
		Username:      validated.Username,
		AvatarURL:     validated.AvatarURL,
		ThreadID:      validated.ThreadID,
		Buttons:       validated.Buttons,
		UpdateSeconds: validated.UpdateSeconds,
		Cleanup:       validated.Cleanup,
	}
	firstWebhook := len(a.cfg.Webhooks) == 0
	a.cfg.Webhooks = append(a.cfg.Webhooks, webhook)
	if firstWebhook {
		for channelIndex := range a.cfg.Channels {
			if len(a.cfg.Channels[channelIndex].Destinations) == 0 {
				a.cfg.Channels[channelIndex].Destinations = []Destination{{WebhookID: webhook.ID}}
			}
		}
	}
	if err := a.persistLocked(); err != nil {
		return Webhook{}, err
	}
	a.activity = append(a.activity, Activity{ID: newID(), Time: time.Now(), Level: "success", Message: "Added destination " + validated.Label + "."})
	return webhook, nil
}

func (a *App) UpdateWebhook(id, label, rawURL string, roleIDs []string, enabled bool, template Template) (Webhook, error) {
	return a.UpdateWebhookInput(id, WebhookInput{
		Label:    label,
		URL:      rawURL,
		RoleIDs:  roleIDs,
		Enabled:  enabled,
		Template: template,
	})
}

func (a *App) UpdateWebhookInput(id string, input WebhookInput) (Webhook, error) {
	validated, err := validateWebhookInput(input)
	if err != nil {
		return Webhook{}, err
	}

	a.mu.Lock()
	index := a.webhookIndexLocked(id)
	if index < 0 {
		a.mu.Unlock()
		return Webhook{}, errors.New("destination not found")
	}
	for i, webhook := range a.cfg.Webhooks {
		if i == index {
			continue
		}
		if strings.EqualFold(webhook.Label, validated.Label) {
			a.mu.Unlock()
			return Webhook{}, fmt.Errorf("a destination called %q already exists", validated.Label)
		}
		if webhook.URL == validated.URL {
			a.mu.Unlock()
			return Webhook{}, errors.New("that Discord webhook is already added")
		}
	}

	wasEnabled := a.cfg.Webhooks[index].Enabled
	urlChanged := a.cfg.Webhooks[index].URL != validated.URL
	a.cfg.Webhooks[index].Label = validated.Label
	a.cfg.Webhooks[index].URL = validated.URL
	a.cfg.Webhooks[index].Enabled = validated.Enabled
	a.cfg.Webhooks[index].RoleIDs = validated.RoleIDs
	a.cfg.Webhooks[index].Template = validated.Template
	a.cfg.Webhooks[index].Username = validated.Username
	a.cfg.Webhooks[index].AvatarURL = validated.AvatarURL
	a.cfg.Webhooks[index].ThreadID = validated.ThreadID
	a.cfg.Webhooks[index].Buttons = validated.Buttons
	a.cfg.Webhooks[index].UpdateSeconds = validated.UpdateSeconds
	a.cfg.Webhooks[index].Cleanup = validated.Cleanup
	if urlChanged || (validated.Enabled && !wasEnabled) {
		a.cfg.Webhooks[index].Invalid = false
		a.cfg.Webhooks[index].LastError = ""
		a.cfg.Webhooks[index].LastErrorAt = ""
		a.cfg.Webhooks[index].ConsecutiveFailures = 0
	}
	if urlChanged {
		for channelIndex := range a.cfg.Channels {
			for destinationIndex := range a.cfg.Channels[channelIndex].Destinations {
				if a.cfg.Channels[channelIndex].Destinations[destinationIndex].WebhookID == id {
					a.cfg.Channels[channelIndex].Destinations[destinationIndex].LastStreamID = ""
					a.cfg.Channels[channelIndex].Destinations[destinationIndex].LastPostedAt = ""
					a.cfg.Channels[channelIndex].Destinations[destinationIndex].RetryUntil = ""
					a.cfg.Channels[channelIndex].Destinations[destinationIndex].LiveMessage = nil
				}
			}
		}
	}
	if err := a.persistLocked(); err != nil {
		a.mu.Unlock()
		return Webhook{}, err
	}
	webhook := a.cfg.Webhooks[index]
	a.mu.Unlock()
	a.addActivity("info", "Updated destination "+validated.Label+".")
	if validated.Enabled && (!wasEnabled || urlChanged) && a.enableAutoCheck {
		a.TriggerCheck()
	}
	return webhook, nil
}

func validateWebhookInput(input WebhookInput) (WebhookInput, error) {
	input.Label = strings.TrimSpace(input.Label)
	if input.Label == "" {
		return WebhookInput{}, errors.New("give this destination a label")
	}
	if len(input.Label) > 60 {
		return WebhookInput{}, errors.New("destination labels must be 60 characters or fewer")
	}
	webhookURL, err := normalizeWebhookURL(input.URL)
	if err != nil {
		return WebhookInput{}, err
	}
	roles, err := normalizeRoleIDs(input.RoleIDs)
	if err != nil {
		return WebhookInput{}, err
	}
	template, err := validateTemplate(input.Template)
	if err != nil {
		return WebhookInput{}, err
	}
	username, avatar, err := normalizeWebhookIdentity(input.Username, input.AvatarURL)
	if err != nil {
		return WebhookInput{}, err
	}
	threadID, err := normalizeThreadID(input.ThreadID)
	if err != nil {
		return WebhookInput{}, err
	}
	buttons, err := normalizeButtons(input.Buttons)
	if err != nil {
		return WebhookInput{}, err
	}
	cleanup, err := validateCleanupSettings(input.Cleanup)
	if err != nil {
		return WebhookInput{}, err
	}
	return WebhookInput{
		Label:         input.Label,
		URL:           webhookURL,
		RoleIDs:       roles,
		Enabled:       input.Enabled,
		Template:      template,
		Username:      username,
		AvatarURL:     avatar,
		ThreadID:      threadID,
		Buttons:       buttons,
		UpdateSeconds: normalizeUpdateSeconds(input.UpdateSeconds),
		Cleanup:       cleanup,
	}, nil
}

func (a *App) RemoveWebhook(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	index := a.webhookIndexLocked(id)
	if index < 0 {
		return errors.New("destination not found")
	}
	label := a.cfg.Webhooks[index].Label
	a.cfg.Webhooks = append(a.cfg.Webhooks[:index], a.cfg.Webhooks[index+1:]...)
	for channelIndex := range a.cfg.Channels {
		filtered := a.cfg.Channels[channelIndex].Destinations[:0]
		for _, destination := range a.cfg.Channels[channelIndex].Destinations {
			if destination.WebhookID != id {
				filtered = append(filtered, destination)
			}
		}
		a.cfg.Channels[channelIndex].Destinations = filtered
	}
	err := a.persistLocked()
	if err == nil {
		a.activity = append(a.activity, Activity{ID: newID(), Time: time.Now(), Level: "info", Message: "Removed destination " + label + "."})
	}
	return err
}

func (a *App) TestWebhook(id string) (discordPostResult, error) {
	a.mu.RLock()
	index := a.webhookIndexLocked(id)
	if index < 0 {
		a.mu.RUnlock()
		return discordPostResult{}, errors.New("destination not found")
	}
	webhook := a.cfg.Webhooks[index]
	a.mu.RUnlock()

	payload := buildTestWebhookPayload(time.Now())
	payload.Username = webhook.Username
	payload.AvatarURL = webhook.AvatarURL
	result, err := postDiscordContextWithOptions(context.Background(), webhook.URL, payload, discordWebhookOptions{ThreadID: webhook.ThreadID})
	if err != nil {
		a.recordWebhookFailure(webhook.ID, err)
		a.addActivity("error", fmt.Sprintf("Webhook test for %s failed: %v", webhook.Label, err))
		return discordPostResult{}, err
	}
	a.recordWebhookSuccess(webhook.ID)
	a.addActivity("success", "Sent a test message to "+webhook.Label+".")
	return result, nil
}

func (a *App) SendSample(channelID string, webhookIDs []string) (int, error) {
	a.mu.RLock()
	channelIndex := a.channelIndexLocked(channelID)
	if channelIndex < 0 {
		a.mu.RUnlock()
		return 0, errors.New("streamer not found")
	}
	channel := a.cfg.Channels[channelIndex]
	allWebhooks := append([]Webhook(nil), a.cfg.Webhooks...)
	a.mu.RUnlock()

	if len(webhookIDs) == 0 {
		for _, destination := range channel.Destinations {
			webhookIDs = append(webhookIDs, destination.WebhookID)
		}
	}
	if len(webhookIDs) == 0 {
		return 0, errors.New("choose at least one destination")
	}

	user, err := fetchTwitchChannel(channel.Login)
	if err != nil {
		return 0, err
	}
	if user.LastBroadcast == nil {
		return 0, errors.New("this streamer has no previous stream to use as a sample")
	}
	info := streamInfoFromTwitch(user, false)
	info.DiscordID = channel.DiscordID
	now := time.Now()
	count := 0
	for _, webhookID := range webhookIDs {
		webhook := findWebhook(allWebhooks, webhookID)
		if webhook == nil {
			continue
		}
		var destination *Destination
		for i := range channel.Destinations {
			if channel.Destinations[i].WebhookID == webhookID {
				destination = &channel.Destinations[i]
				break
			}
		}
		if _, err := postDiscordContextWithOptions(context.Background(), webhook.URL, buildPayloadForDestination(*webhook, destination, info, now, true), discordWebhookOptions{ThreadID: webhook.ThreadID}); err != nil {
			a.recordWebhookFailure(webhook.ID, err)
			a.addActivity("error", fmt.Sprintf("Sample post to %s failed: %v", webhook.Label, err))
			continue
		}
		a.recordWebhookSuccess(webhook.ID)
		count++
		a.addActivity("success", "Sent a last-stream sample to "+webhook.Label+".")
	}
	if count == 0 {
		return 0, errors.New("the sample could not be posted")
	}
	return count, nil
}

func (a *App) Preview(template Template, channelID, webhookLabel string, roleIDs []string) (discordWebhookPayload, []string) {
	return a.PreviewWithOptions(template, channelID, webhookLabel, roleIDs, false)
}

func (a *App) PreviewWithOptions(template Template, channelID, webhookLabel string, roleIDs []string, ended bool) (discordWebhookPayload, []string) {
	return a.PreviewWebhook(Webhook{
		Label:    webhookLabel,
		RoleIDs:  roleIDs,
		Template: template,
	}, channelID, ended)
}

// PreviewWebhook renders a full destination-shaped webhook, including the
// custom identity, thread-aware button links, and route-level template.
func (a *App) PreviewWebhook(webhook Webhook, channelID string, ended bool) (discordWebhookPayload, []string) {
	if strings.TrimSpace(webhook.Label) == "" {
		webhook.Label = "Discord"
	}
	webhook.RoleIDs = sanitizeRoleIDs(webhook.RoleIDs)
	info := a.previewInfo(channelID)
	if ended {
		info.Live = false
		info.Viewers = 0
	}
	payload := buildPayload(webhook, info, time.Now(), false)
	template := webhook.Template
	unknown := append([]string{}, unknownPlaceholders(template.Content)...)
	unknown = append(unknown, unknownPlaceholders(template.EmbedTitle)...)
	unknown = append(unknown, unknownPlaceholders(template.EmbedDescription)...)
	unknown = append(unknown, unknownPlaceholders(template.Footer)...)
	for _, field := range template.CustomFields {
		unknown = append(unknown, unknownPlaceholders(field.Name)...)
		unknown = append(unknown, unknownPlaceholders(field.Value)...)
	}
	for _, button := range webhook.Buttons {
		unknown = append(unknown, unknownPlaceholders(button.Label)...)
		unknown = append(unknown, unknownPlaceholders(button.URL)...)
	}
	seen := map[string]bool{}
	var unique []string
	for _, token := range unknown {
		if !seen[token] {
			seen[token] = true
			unique = append(unique, token)
		}
	}
	return payload, unique
}

func (a *App) previewInfo(channelID string) StreamInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()
	index := a.channelIndexLocked(channelID)
	if index < 0 {
		return sampleStreamInfo()
	}
	channel := a.cfg.Channels[index]
	status := a.statuses[channelID]
	info := StreamInfo{
		Login:     channel.Login,
		Streamer:  channel.DisplayName,
		URL:       "https://www.twitch.tv/" + channel.Login,
		AvatarURL: channel.AvatarURL,
		DiscordID: channel.DiscordID,
		Live:      status.Live,
	}
	switch {
	case status.Live:
		info.Title = status.Title
		info.Game = status.Game
		info.Thumbnail = status.Thumbnail
		info.Viewers = status.Viewers
		info.StartedAt = parseTime(status.StartedAt)
	case status.LastTitle != "":
		info.Title = status.LastTitle
		info.Game = status.LastGame
		info.Thumbnail = status.LastThumbnail
		info.StartedAt = parseTime(status.LastStartedAt)
	default:
		info.Title = "Your next stream"
		info.Game = "Just Chatting"
		info.Thumbnail = channel.BannerURL
		info.StartedAt = time.Now().Add(-2 * time.Hour)
		info.Viewers = 0
	}
	return info
}

func parseTime(value string) time.Time {
	parsed, _ := time.Parse(time.RFC3339, value)
	return parsed
}

func (a *App) UpdateSettings(settings Settings) error {
	if settings.PollSeconds < 30 {
		settings.PollSeconds = 30
	}
	if settings.PollSeconds > 3600 {
		settings.PollSeconds = 3600
	}
	if settings.Theme != "light" {
		settings.Theme = "dark"
	}

	a.mu.RLock()
	old := a.cfg.Settings
	a.mu.RUnlock()
	if settings.StartAtLogin != old.StartAtLogin {
		if err := setAutoStart(settings.StartAtLogin); err != nil {
			return err
		}
	}

	a.mu.Lock()
	settings.DashboardTokenHash = a.cfg.Settings.DashboardTokenHash
	a.cfg.Settings = settings
	err := a.persistLocked()
	a.mu.Unlock()
	if err != nil {
		return err
	}
	a.wakeMonitor()
	return nil
}

func (a *App) TogglePause() (bool, error) {
	a.mu.Lock()
	paused := !a.cfg.Settings.Paused
	a.cfg.Settings.Paused = paused
	err := a.persistLocked()
	a.mu.Unlock()
	if err != nil {
		return paused, err
	}
	a.wakeMonitor()
	if paused {
		a.addActivity("info", "Monitoring paused.")
	} else {
		a.addActivity("info", "Monitoring resumed.")
		if a.enableAutoCheck {
			a.TriggerCheck()
		}
	}
	return paused, nil
}

func (a *App) IsPaused() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.Settings.Paused
}

func (a *App) pollInterval() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	seconds := a.cfg.Settings.PollSeconds
	if seconds < 30 {
		seconds = 30
	}
	return time.Duration(seconds) * time.Second
}

func (a *App) wakeMonitor() {
	select {
	case a.wake <- struct{}{}:
	default:
	}
}

func (a *App) addActivity(level, message string) {
	a.mu.Lock()
	a.activity = append(a.activity, Activity{ID: newID(), Time: time.Now(), Level: level, Message: message})
	if len(a.activity) > 100 {
		a.activity = a.activity[len(a.activity)-100:]
	}
	a.mu.Unlock()
}

func (a *App) persistLocked() error {
	if err := saveConfig(a.configPath, a.cfg); err != nil {
		a.activity = append(a.activity, Activity{ID: newID(), Time: time.Now(), Level: "error", Message: err.Error()})
		if len(a.activity) > 100 {
			a.activity = a.activity[len(a.activity)-100:]
		}
		return err
	}
	return nil
}

func (a *App) channelIndexLocked(id string) int {
	for i, channel := range a.cfg.Channels {
		if channel.ID == id {
			return i
		}
	}
	return -1
}

func (a *App) channelByLoginLocked(login string) *Channel {
	for i := range a.cfg.Channels {
		if strings.EqualFold(a.cfg.Channels[i].Login, login) {
			return &a.cfg.Channels[i]
		}
	}
	return nil
}

func (a *App) webhookIndexLocked(id string) int {
	for i, webhook := range a.cfg.Webhooks {
		if webhook.ID == id {
			return i
		}
	}
	return -1
}

func findWebhook(webhooks []Webhook, id string) *Webhook {
	for i := range webhooks {
		if webhooks[i].ID == id {
			return &webhooks[i]
		}
	}
	return nil
}

func (a *App) DataDir() string {
	return filepath.Dir(a.configPath)
}

func (a *App) OpenDataFolder() error {
	return openFolder(a.DataDir())
}
