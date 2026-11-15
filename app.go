package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ChannelStatus struct {
	Live          bool   `json:"live"`
	Title         string `json:"title,omitempty"`
	Game          string `json:"game,omitempty"`
	Viewers       int    `json:"viewers"`
	Thumbnail     string `json:"thumbnail,omitempty"`
	StartedAt     string `json:"started_at,omitempty"`
	LastTitle     string `json:"last_title,omitempty"`
	LastGame      string `json:"last_game,omitempty"`
	LastStartedAt string `json:"last_started_at,omitempty"`
	LastThumbnail string `json:"last_thumbnail,omitempty"`
	LastChecked   string `json:"last_checked,omitempty"`
	Error         string `json:"error,omitempty"`
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
}

type CheckSummary struct {
	Channels int
	Live     int
	Posted   int
	Errors   int
}

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
		_, _ = a.runCheck(context.Background())
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
		_, _ = a.checkChannel(channel, webhooks)
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

		posted, err := a.checkChannel(channel, webhooks)
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

func (a *App) checkChannel(channel Channel, webhooks []Webhook) (int, error) {
	previous := a.statusFor(channel.ID)
	now := time.Now()

	user, err := fetchTwitchChannel(channel.Login)
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
		return 0, nil
	}

	info := streamInfoFromTwitch(user, true)
	info.DiscordID = channel.DiscordID
	posted := 0
	var postErrors []string
	for _, destination := range channel.Destinations {
		webhook := findWebhook(webhooks, destination.WebhookID)
		if webhook == nil || !webhook.Enabled {
			continue
		}
		if destination.LastStreamID != "" && destination.LastStreamID == info.StreamID {
			continue
		}

		result, postErr := postDiscord(webhook.URL, buildPayload(*webhook, info, now, false))
		if postErr != nil {
			postErrors = append(postErrors, fmt.Sprintf("%s: %v", webhook.Label, postErr))
			a.addActivity("error", fmt.Sprintf("Could not post to %s: %v", webhook.Label, postErr))
			continue
		}
		a.markPosted(channel.ID, webhook.ID, info.StreamID, result.MessageID, now)
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

func (a *App) markPosted(channelID, webhookID, streamID, messageID string, now time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if index := a.channelIndexLocked(channelID); index >= 0 {
		for i := range a.cfg.Channels[index].Destinations {
			destination := &a.cfg.Channels[index].Destinations[i]
			if destination.WebhookID == webhookID {
				destination.LastStreamID = streamID
				destination.LastPostedAt = now.UTC().Format(time.RFC3339)
				break
			}
		}
	}
	a.persistLocked()
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
	a.mu.RUnlock()

	lastCheckText := ""
	if !lastCheck.IsZero() {
		lastCheckText = lastCheck.UTC().Format(time.RFC3339)
	}
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

func (a *App) AddWebhook(label, rawURL string, template *Template) (Webhook, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return Webhook{}, errors.New("give this destination a label")
	}
	if len(label) > 60 {
		return Webhook{}, errors.New("destination labels must be 60 characters or fewer")
	}
	webhookURL, err := normalizeWebhookURL(rawURL)
	if err != nil {
		return Webhook{}, err
	}
	chosen := defaultTemplate()
	if template != nil {
		chosen, err = validateTemplate(*template)
		if err != nil {
			return Webhook{}, err
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for _, webhook := range a.cfg.Webhooks {
		if strings.EqualFold(webhook.Label, label) {
			return Webhook{}, fmt.Errorf("a destination called %q already exists", label)
		}
		if webhook.URL == webhookURL {
			return Webhook{}, errors.New("that Discord webhook is already added")
		}
	}
	webhook := Webhook{
		ID:       newID(),
		Label:    label,
		URL:      webhookURL,
		Enabled:  true,
		Template: chosen,
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
	a.activity = append(a.activity, Activity{ID: newID(), Time: time.Now(), Level: "success", Message: "Added destination " + label + "."})
	return webhook, nil
}

func (a *App) UpdateWebhook(id, label, rawURL string, enabled bool, template Template) (Webhook, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return Webhook{}, errors.New("give this destination a label")
	}
	if len(label) > 60 {
		return Webhook{}, errors.New("destination labels must be 60 characters or fewer")
	}
	webhookURL, err := normalizeWebhookURL(rawURL)
	if err != nil {
		return Webhook{}, err
	}
	template, err = validateTemplate(template)
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
		if strings.EqualFold(webhook.Label, label) {
			a.mu.Unlock()
			return Webhook{}, fmt.Errorf("a destination called %q already exists", label)
		}
		if webhook.URL == webhookURL {
			a.mu.Unlock()
			return Webhook{}, errors.New("that Discord webhook is already added")
		}
	}

	wasEnabled := a.cfg.Webhooks[index].Enabled
	urlChanged := a.cfg.Webhooks[index].URL != webhookURL
	a.cfg.Webhooks[index].Label = label
	a.cfg.Webhooks[index].URL = webhookURL
	a.cfg.Webhooks[index].Enabled = enabled
	a.cfg.Webhooks[index].Template = template
	if urlChanged {
		for channelIndex := range a.cfg.Channels {
			for destinationIndex := range a.cfg.Channels[channelIndex].Destinations {
				if a.cfg.Channels[channelIndex].Destinations[destinationIndex].WebhookID == id {
					a.cfg.Channels[channelIndex].Destinations[destinationIndex].LastStreamID = ""
					a.cfg.Channels[channelIndex].Destinations[destinationIndex].LastPostedAt = ""
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
	a.addActivity("info", "Updated destination "+label+".")
	if enabled && (!wasEnabled || urlChanged) && a.enableAutoCheck {
		a.TriggerCheck()
	}
	return webhook, nil
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

	result, err := postDiscord(webhook.URL, buildTestWebhookPayload(time.Now()))
	if err != nil {
		a.addActivity("error", fmt.Sprintf("Webhook test for %s failed: %v", webhook.Label, err))
		return discordPostResult{}, err
	}
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
		if _, err := postDiscord(webhook.URL, buildPayload(*webhook, info, now, true)); err != nil {
			a.addActivity("error", fmt.Sprintf("Sample post to %s failed: %v", webhook.Label, err))
			continue
		}
		count++
		a.addActivity("success", "Sent a last-stream sample to "+webhook.Label+".")
	}
	if count == 0 {
		return 0, errors.New("the sample could not be posted")
	}
	return count, nil
}

func (a *App) Preview(template Template, channelID, webhookLabel string) (discordWebhookPayload, []string) {
	if strings.TrimSpace(webhookLabel) == "" {
		webhookLabel = "Discord"
	}
	info := a.previewInfo(channelID)
	webhook := Webhook{Label: webhookLabel, Template: template}
	payload := buildPayload(webhook, info, time.Now(), false)
	unknown := append([]string{}, unknownPlaceholders(template.Content)...)
	unknown = append(unknown, unknownPlaceholders(template.EmbedTitle)...)
	unknown = append(unknown, unknownPlaceholders(template.EmbedDescription)...)
	unknown = append(unknown, unknownPlaceholders(template.Footer)...)
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
