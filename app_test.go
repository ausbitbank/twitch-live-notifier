package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type capturedDiscord struct {
	mu       sync.Mutex
	payloads []discordWebhookPayload
}

func (c *capturedDiscord) add(payload discordWebhookPayload) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.payloads = append(c.payloads, payload)
}

func (c *capturedDiscord) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.payloads)
}

func (c *capturedDiscord) last() discordWebhookPayload {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.payloads) == 0 {
		return discordWebhookPayload{}
	}
	return c.payloads[len(c.payloads)-1]
}

func newDiscordCapture(t *testing.T) (*httptest.Server, *capturedDiscord) {
	t.Helper()
	capture := &capturedDiscord{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload discordWebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode Discord payload: %v", err)
		}
		capture.add(payload)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"message123"}`)
	}))
	t.Cleanup(server.Close)
	return server, capture
}

func newTestApp(t *testing.T) *App {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	app, err := NewApp(path)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	app.enableAutoCheck = false
	return app
}

func injectWebhook(t *testing.T, app *App, label, rawURL string) Webhook {
	t.Helper()
	webhook := Webhook{
		ID:       newID(),
		Label:    label,
		URL:      rawURL,
		Enabled:  true,
		Template: defaultTemplate(),
	}
	app.mu.Lock()
	app.cfg.Webhooks = append(app.cfg.Webhooks, webhook)
	if err := app.persistLocked(); err != nil {
		app.mu.Unlock()
		t.Fatalf("persist webhook: %v", err)
	}
	app.mu.Unlock()
	return webhook
}

func TestCheckNowPostsAndDeduplicates(t *testing.T) {
	twitch := newTwitchTestServer(t, liveTwitchJSON)
	setTwitchTestURL(t, twitch)
	discord, capture := newDiscordCapture(t)

	app := newTestApp(t)
	webhook := injectWebhook(t, app, "Community", discord.URL+"/api/webhooks/1/token")
	channel, err := app.AddChannel("teststreamer")
	if err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if len(channel.Destinations) != 1 || channel.Destinations[0].WebhookID != webhook.ID {
		t.Fatalf("channel destinations = %+v", channel.Destinations)
	}

	summary, err := app.CheckNow(context.Background())
	if err != nil {
		t.Fatalf("CheckNow returned error: %v", err)
	}
	if summary.Channels != 1 || summary.Posted != 1 || summary.Live != 1 {
		t.Fatalf("summary = %+v", summary)
	}
	if capture.count() != 1 {
		t.Fatalf("Discord calls = %d", capture.count())
	}
	payload := capture.last()
	if len(payload.Embeds) != 1 || payload.Embeds[0].Title != "Testing the live notifier" {
		t.Fatalf("payload = %+v", payload)
	}
	cfg := app.ConfigSnapshot()
	if cfg.Channels[0].Destinations[0].LastStreamID != "319913347671" {
		t.Fatalf("last stream = %+v", cfg.Channels[0].Destinations[0])
	}

	summary, err = app.CheckNow(context.Background())
	if err != nil {
		t.Fatalf("second CheckNow returned error: %v", err)
	}
	if summary.Posted != 0 || capture.count() != 1 {
		t.Fatalf("second check summary = %+v, calls = %d", summary, capture.count())
	}
}

func TestCheckNowPostsToMultipleWebhooks(t *testing.T) {
	twitch := newTwitchTestServer(t, liveTwitchJSON)
	setTwitchTestURL(t, twitch)
	discord, capture := newDiscordCapture(t)

	app := newTestApp(t)
	first := injectWebhook(t, app, "First", discord.URL+"/api/webhooks/1/token")
	second := injectWebhook(t, app, "Second", discord.URL+"/api/webhooks/2/token")
	channel, err := app.AddChannel("teststreamer")
	if err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if err := app.SetDestinations(channel.ID, []string{first.ID, second.ID}); err != nil {
		t.Fatalf("SetDestinations returned error: %v", err)
	}
	summary, err := app.CheckNow(context.Background())
	if err != nil {
		t.Fatalf("CheckNow returned error: %v", err)
	}
	if summary.Posted != 2 || capture.count() != 2 {
		t.Fatalf("summary = %+v, calls = %d", summary, capture.count())
	}
}

func TestRemoveWebhookClearsDestinations(t *testing.T) {
	app := newTestApp(t)
	webhook := injectWebhook(t, app, "Community", "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz")
	app.mu.Lock()
	app.cfg.Channels = []Channel{{
		ID: "channel1", Login: "teststreamer", DisplayName: "TestStreamer", Enabled: true,
		Destinations: []Destination{{WebhookID: webhook.ID}},
	}}
	app.mu.Unlock()
	if err := app.RemoveWebhook(webhook.ID); err != nil {
		t.Fatalf("RemoveWebhook returned error: %v", err)
	}
	cfg := app.ConfigSnapshot()
	if len(cfg.Webhooks) != 0 || len(cfg.Channels[0].Destinations) != 0 {
		t.Fatalf("config = %+v", cfg)
	}
}

func TestUpdateWebhookURLResetsLastStream(t *testing.T) {
	app := newTestApp(t)
	webhook := injectWebhook(t, app, "Community", "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz")
	app.mu.Lock()
	app.cfg.Channels = []Channel{{
		ID: "channel1", Login: "teststreamer", DisplayName: "TestStreamer", Enabled: true,
		Destinations: []Destination{{WebhookID: webhook.ID, LastStreamID: "old-stream"}},
	}}
	app.mu.Unlock()

	_, err := app.UpdateWebhook(webhook.ID, "Community", "https://discord.com/api/webhooks/1234567890/zyxwvutsrqponmlkjihgfedcba", nil, true, defaultTemplate())
	if err != nil {
		t.Fatalf("UpdateWebhook returned error: %v", err)
	}
	cfg := app.ConfigSnapshot()
	if cfg.Channels[0].Destinations[0].LastStreamID != "" {
		t.Fatalf("last stream was not reset: %+v", cfg.Channels[0].Destinations[0])
	}
}

func TestSendSample(t *testing.T) {
	twitch := newTwitchTestServer(t, offlineTwitchJSON)
	setTwitchTestURL(t, twitch)
	discord, capture := newDiscordCapture(t)

	app := newTestApp(t)
	webhook := injectWebhook(t, app, "Community", discord.URL+"/api/webhooks/1/token")
	channel, err := app.AddChannel("teststreamer")
	if err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	count, err := app.SendSample(channel.ID, []string{webhook.ID})
	if err != nil {
		t.Fatalf("SendSample returned error: %v", err)
	}
	if count != 1 || capture.count() != 1 {
		t.Fatalf("count = %d, calls = %d", count, capture.count())
	}
	payload := capture.last()
	if !strings.Contains(payload.Content, "Example") {
		t.Fatalf("content = %q", payload.Content)
	}
}

func TestPreviewAndUnknownPlaceholders(t *testing.T) {
	twitch := newTwitchTestServer(t, liveTwitchJSON)
	setTwitchTestURL(t, twitch)
	app := newTestApp(t)
	channel, err := app.AddChannel("teststreamer")
	if err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	template := defaultTemplate()
	template.Content = "{streamer} {mystery}"
	payload, unknown := app.Preview(template, channel.ID, "Community", nil)
	if len(payload.Embeds) != 1 || payload.Embeds[0].Title != "Testing the live notifier" {
		t.Fatalf("payload = %+v", payload)
	}
	if len(unknown) != 1 || unknown[0] != "{mystery}" {
		t.Fatalf("unknown = %#v", unknown)
	}

	const roleID = "123456789012345678"
	template.Content = "{role_mentions}"
	payload, unknown = app.Preview(template, channel.ID, "Community", []string{"<@&" + roleID + ">"})
	if len(unknown) != 0 {
		t.Fatalf("role placeholder reported as unknown: %#v", unknown)
	}
	if !strings.Contains(payload.Content, "<@&"+roleID+">") {
		t.Fatalf("preview content = %q", payload.Content)
	}
	if payload.AllowedMentions == nil || len(payload.AllowedMentions.Roles) != 1 || payload.AllowedMentions.Roles[0] != roleID {
		t.Fatalf("preview allowed mentions = %+v", payload.AllowedMentions)
	}
}

func TestSetRouteTemplate(t *testing.T) {
	app := newTestApp(t)
	webhook := injectWebhook(t, app, "Community", "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz")
	app.mu.Lock()
	app.cfg.Channels = []Channel{{
		ID: "channel1", Login: "teststreamer", DisplayName: "TestStreamer", Enabled: true,
		Destinations: []Destination{{WebhookID: webhook.ID}},
	}}
	app.mu.Unlock()

	template := defaultTemplate()
	template.Content = "Route {streamer}"
	if err := app.SetRouteTemplate("channel1", webhook.ID, &template); err != nil {
		t.Fatalf("SetRouteTemplate returned error: %v", err)
	}
	cfg := app.ConfigSnapshot()
	if cfg.Channels[0].Destinations[0].Template == nil || cfg.Channels[0].Destinations[0].Template.Content != "Route {streamer}" {
		t.Fatalf("route template = %+v", cfg.Channels[0].Destinations[0].Template)
	}
	if err := app.SetRouteTemplate("channel1", webhook.ID, nil); err != nil {
		t.Fatalf("clearing route template returned error: %v", err)
	}
	if app.ConfigSnapshot().Channels[0].Destinations[0].Template != nil {
		t.Fatal("route template was not cleared")
	}
	if err := app.SetRouteTemplate("missing", webhook.ID, &template); err == nil {
		t.Fatal("unknown streamer accepted")
	}
}

func TestAddWebhookInputStoresLifecycleIdentityAndTemplatePower(t *testing.T) {
	app := newTestApp(t)
	ended := defaultEndedTemplate()
	input := WebhookInput{
		Label:         "Community",
		URL:           "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
		Enabled:       true,
		Template:      defaultTemplate(),
		Username:      "Community Alerts",
		AvatarURL:     "https://example.com/avatar.png",
		ThreadID:      "987654321012345678",
		Buttons:       []TemplateButton{{Label: "Watch", URL: "{url}"}},
		UpdateSeconds: 60,
		Cleanup:       CleanupSettings{Mode: cleanupEdit, Template: &ended},
	}
	webhook, err := app.AddWebhookInput(input)
	if err != nil {
		t.Fatalf("AddWebhookInput returned error: %v", err)
	}
	cfg := app.ConfigSnapshot()
	if len(cfg.Webhooks) != 1 {
		t.Fatalf("webhooks = %+v", cfg.Webhooks)
	}
	stored := cfg.Webhooks[0]
	if stored.ID != webhook.ID || stored.Username != input.Username || stored.AvatarURL != input.AvatarURL {
		t.Fatalf("identity = %+v", stored)
	}
	if stored.ThreadID != input.ThreadID || stored.UpdateSeconds != 60 || stored.Cleanup.Mode != cleanupEdit {
		t.Fatalf("lifecycle = %+v", stored)
	}
	if len(stored.Buttons) != 1 || stored.Buttons[0].URL != "{url}" {
		t.Fatalf("buttons = %+v", stored.Buttons)
	}
	if _, err := app.AddWebhookInput(WebhookInput{
		Label:    "Bad thread",
		URL:      "https://discord.com/api/webhooks/1234567891/abcdefghijklmnopqrstuvwxyz",
		Template: defaultTemplate(),
		ThreadID: "not-a-thread",
	}); err == nil {
		t.Fatal("invalid thread ID accepted")
	}
}

func TestAddWebhookValidation(t *testing.T) {
	app := newTestApp(t)
	_, err := app.AddWebhook("Community", "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz", nil, nil)
	if err != nil {
		t.Fatalf("AddWebhook returned error: %v", err)
	}
	if _, err := app.AddWebhook("Community", "https://discord.com/api/webhooks/1234567891/abcdefghijklmnopqrstuvwxyz", nil, nil); err == nil {
		t.Fatal("duplicate label accepted")
	}
	if _, err := app.AddWebhook("Other", "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz", nil, nil); err == nil {
		t.Fatal("duplicate URL accepted")
	}
}

func TestUpdateSettings(t *testing.T) {
	app := newTestApp(t)
	if err := app.UpdateSettings(Settings{PollSeconds: 300, Theme: "light", Paused: true}); err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}
	state := app.State()
	if state.Config.Settings.PollSeconds != 300 || state.Config.Settings.Theme != "light" || !state.Config.Settings.Paused {
		t.Fatalf("settings = %+v", state.Config.Settings)
	}
}

func TestDashboardAPIStateAndGuards(t *testing.T) {
	app := newTestApp(t)
	handler := withSecurity(newDashboardHandler(app, func() {}))
	server := httptest.NewServer(handler)
	defer server.Close()

	response, err := http.Get(server.URL + "/api/state")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("state status = %d", response.StatusCode)
	}
	var state State
	if err := json.NewDecoder(response.Body).Decode(&state); err != nil {
		t.Fatalf("decode state: %v", err)
	}
	if state.Version != Version {
		t.Fatalf("version = %q", state.Version)
	}

	request, _ := http.NewRequest(http.MethodPost, server.URL+"/api/check", strings.NewReader(`{}`))
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("missing content type status = %d", response.StatusCode)
	}

	request, _ = http.NewRequest(http.MethodPost, server.URL+"/api/check", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://evil.example")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin status = %d", response.StatusCode)
	}

	request, _ = http.NewRequest(http.MethodPost, server.URL+"/api/check", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", server.URL)
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("same-origin status = %d", response.StatusCode)
	}
}

func TestNewAppCreatesConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	if _, err := NewApp(path); err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config was not created: %v", err)
	}
}

func TestFirstWebhookConnectsExistingStreamers(t *testing.T) {
	app := newTestApp(t)
	app.mu.Lock()
	app.cfg.Channels = []Channel{{
		ID: "channel1", Login: "teststreamer", DisplayName: "TestStreamer", Enabled: true,
	}}
	app.mu.Unlock()

	webhook, err := app.AddWebhook("Community", "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz", nil, nil)
	if err != nil {
		t.Fatalf("AddWebhook returned error: %v", err)
	}
	cfg := app.ConfigSnapshot()
	if len(cfg.Channels[0].Destinations) != 1 || cfg.Channels[0].Destinations[0].WebhookID != webhook.ID {
		t.Fatalf("first webhook did not auto-connect: %+v", cfg.Channels[0].Destinations)
	}
}

func TestCheckChannelNow(t *testing.T) {
	twitch := newTwitchTestServer(t, liveTwitchJSON)
	setTwitchTestURL(t, twitch)
	discord, capture := newDiscordCapture(t)

	app := newTestApp(t)
	injectWebhook(t, app, "Community", discord.URL+"/api/webhooks/1/token")
	channel, err := app.AddChannel("teststreamer")
	if err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	started, err := app.CheckChannelNow(channel.ID)
	if err != nil || !started {
		t.Fatalf("CheckChannelNow = %v, %v", started, err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for app.State().Checking && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if app.State().Checking {
		t.Fatal("channel check did not finish")
	}
	if capture.count() != 1 {
		t.Fatalf("Discord calls = %d", capture.count())
	}
}

func TestSetChannelDiscordID(t *testing.T) {
	app := newTestApp(t)
	const channelID = "channel1"
	const discordID = "123456789012345678"
	app.mu.Lock()
	app.cfg.Channels = []Channel{{
		ID: channelID, Login: "teststreamer", DisplayName: "TestStreamer", Enabled: true,
	}}
	app.mu.Unlock()

	if err := app.SetChannelDiscordID(channelID, "<@"+discordID+">"); err != nil {
		t.Fatalf("SetChannelDiscordID returned error: %v", err)
	}
	cfg := app.ConfigSnapshot()
	if cfg.Channels[0].DiscordID != discordID {
		t.Fatalf("discord id = %q", cfg.Channels[0].DiscordID)
	}
	if err := app.SetChannelDiscordID(channelID, "@not-an-id"); err == nil {
		t.Fatal("invalid Discord ID accepted")
	}
	if err := app.SetChannelDiscordID(channelID, ""); err != nil {
		t.Fatalf("clearing Discord ID returned error: %v", err)
	}
	if app.ConfigSnapshot().Channels[0].DiscordID != "" {
		t.Fatal("Discord ID was not cleared")
	}
}

func TestFilteredStreamStillUpdatesStatus(t *testing.T) {
	twitch := newTwitchTestServer(t, liveTwitchJSON)
	setTwitchTestURL(t, twitch)
	discord, capture := newDiscordCapture(t)

	app := newTestApp(t)
	injectWebhook(t, app, "Community", discord.URL+"/api/webhooks/1/token")
	channel, err := app.AddChannel("teststreamer")
	if err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if err := app.SetChannelFilters(channel.ID, Filters{ExcludeGames: []string{"Just Chatting"}}); err != nil {
		t.Fatalf("SetChannelFilters returned error: %v", err)
	}

	summary, err := app.CheckNow(context.Background())
	if err != nil {
		t.Fatalf("CheckNow returned error: %v", err)
	}
	if summary.Posted != 0 || summary.Live != 1 || capture.count() != 0 {
		t.Fatalf("summary = %+v, calls = %d", summary, capture.count())
	}
	status := app.State().Statuses[channel.ID]
	if !status.Live || status.FilterReason == "" {
		t.Fatalf("status = %+v", status)
	}
	if cfg := app.ConfigSnapshot(); cfg.Channels[0].Destinations[0].LastStreamID != "" {
		t.Fatalf("filtered stream recorded a last stream ID: %+v", cfg.Channels[0].Destinations[0])
	}

	if err := app.SetChannelFilters(channel.ID, Filters{}); err != nil {
		t.Fatalf("clearing filters returned error: %v", err)
	}
	summary, err = app.CheckNow(context.Background())
	if err != nil {
		t.Fatalf("second CheckNow returned error: %v", err)
	}
	if summary.Posted != 1 || capture.count() != 1 {
		t.Fatalf("second summary = %+v, calls = %d", summary, capture.count())
	}
}

func TestInvalidDestinationIsMarkedAndSkipped(t *testing.T) {
	twitch := newTwitchTestServer(t, liveTwitchJSON)
	setTwitchTestURL(t, twitch)

	var failingCalls atomic.Int32
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		failingCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"message":"Invalid Webhook Token","code":50027}`)
	}))
	defer failing.Close()
	working, capture := newDiscordCapture(t)

	app := newTestApp(t)
	broken := injectWebhook(t, app, "Broken", failing.URL+"/api/webhooks/1/token")
	healthy := injectWebhook(t, app, "Healthy", working.URL+"/api/webhooks/2/token")
	channel, err := app.AddChannel("teststreamer")
	if err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if err := app.SetDestinations(channel.ID, []string{broken.ID, healthy.ID}); err != nil {
		t.Fatalf("SetDestinations returned error: %v", err)
	}

	summary, err := app.CheckNow(context.Background())
	if err != nil {
		t.Fatalf("CheckNow returned error: %v", err)
	}
	if summary.Posted != 1 || summary.Errors != 1 || capture.count() != 1 {
		t.Fatalf("summary = %+v, calls = %d", summary, capture.count())
	}
	cfg := app.ConfigSnapshot()
	if !cfg.Webhooks[0].Invalid || cfg.Webhooks[0].LastError == "" {
		t.Fatalf("broken webhook = %+v", cfg.Webhooks[0])
	}
	if cfg.Webhooks[1].LastSuccessAt == "" || cfg.Webhooks[1].Invalid {
		t.Fatalf("healthy webhook = %+v", cfg.Webhooks[1])
	}

	summary, err = app.CheckNow(context.Background())
	if err != nil {
		t.Fatalf("second CheckNow returned error: %v", err)
	}
	if summary.Posted != 0 || summary.Errors != 0 || failingCalls.Load() != 1 || capture.count() != 1 {
		t.Fatalf("second summary = %+v, failing calls = %d, healthy calls = %d", summary, failingCalls.Load(), capture.count())
	}

	replacement := "https://discord.com/api/webhooks/987654321012345678/replacementtoken12345"
	if _, err := app.UpdateWebhook(broken.ID, "Repaired", replacement, nil, true, defaultTemplate()); err != nil {
		t.Fatalf("UpdateWebhook returned error: %v", err)
	}
	if app.ConfigSnapshot().Webhooks[0].Invalid {
		t.Fatal("invalid flag survived a webhook URL change")
	}
}
