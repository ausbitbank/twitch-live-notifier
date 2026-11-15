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

	_, err := app.UpdateWebhook(webhook.ID, "Community", "https://discord.com/api/webhooks/1234567890/zyxwvutsrqponmlkjihgfedcba", true, defaultTemplate())
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
	payload, unknown := app.Preview(template, channel.ID, "Community")
	if len(payload.Embeds) != 1 || payload.Embeds[0].Title != "Testing the live notifier" {
		t.Fatalf("payload = %+v", payload)
	}
	if len(unknown) != 1 || unknown[0] != "{mystery}" {
		t.Fatalf("unknown = %#v", unknown)
	}
}

func TestAddWebhookValidation(t *testing.T) {
	app := newTestApp(t)
	_, err := app.AddWebhook("Community", "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz", nil)
	if err != nil {
		t.Fatalf("AddWebhook returned error: %v", err)
	}
	if _, err := app.AddWebhook("Community", "https://discord.com/api/webhooks/1234567891/abcdefghijklmnopqrstuvwxyz", nil); err == nil {
		t.Fatal("duplicate label accepted")
	}
	if _, err := app.AddWebhook("Other", "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz", nil); err == nil {
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

	webhook, err := app.AddWebhook("Community", "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz", nil)
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
