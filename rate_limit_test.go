package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimitedDestinationDefersUntilWindowExpires(t *testing.T) {
	twitch := newTwitchTestServer(t, liveTwitchJSON)
	setTwitchTestURL(t, twitch)

	var calls atomic.Int32
	discord := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = io.WriteString(w, `{"message":"rate limited","retry_after":120}`)
			return
		}
		_, _ = io.WriteString(w, `{"id":"message123"}`)
	}))
	defer discord.Close()

	app := newTestApp(t)
	webhook := injectWebhook(t, app, "Community", discord.URL+"/api/webhooks/1/token")
	if _, err := app.AddChannel("teststreamer"); err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if err := app.SetDestinations(app.ConfigSnapshot().Channels[0].ID, []string{webhook.ID}); err != nil {
		t.Fatalf("SetDestinations returned error: %v", err)
	}

	summary, err := app.CheckNow(context.Background())
	if err != nil {
		t.Fatalf("first CheckNow returned error: %v", err)
	}
	if summary.Posted != 0 || summary.Errors != 1 || calls.Load() != 1 {
		t.Fatalf("first summary = %+v, calls = %d", summary, calls.Load())
	}
	destination := app.ConfigSnapshot().Channels[0].Destinations[0]
	if destination.RetryUntil == "" {
		t.Fatal("429 did not persist a retry window")
	}
	if destination.LastStreamID != "" {
		t.Fatalf("429 recorded dedupe state: %+v", destination)
	}

	// The retry window should make the next check skip Discord entirely.
	summary, err = app.CheckNow(context.Background())
	if err != nil {
		t.Fatalf("second CheckNow returned error: %v", err)
	}
	if summary.Posted != 0 || summary.Errors != 0 || calls.Load() != 1 {
		t.Fatalf("second summary = %+v, calls = %d", summary, calls.Load())
	}

	// Simulate the persisted window and in-memory bucket expiring.
	app.mu.Lock()
	app.cfg.Channels[0].Destinations[0].RetryUntil = time.Now().Add(-time.Minute).UTC().Format(time.RFC3339)
	app.mu.Unlock()
	resetDiscordRateLimiter()

	summary, err = app.CheckNow(context.Background())
	if err != nil {
		t.Fatalf("third CheckNow returned error: %v", err)
	}
	if summary.Posted != 1 || calls.Load() != 2 {
		t.Fatalf("third summary = %+v, calls = %d", summary, calls.Load())
	}
	destination = app.ConfigSnapshot().Channels[0].Destinations[0]
	if destination.LastStreamID != "319913347671" || destination.RetryUntil != "" {
		t.Fatalf("destination after success = %+v", destination)
	}
}

func TestLongRateLimitDoesNotBlockOtherDestinations(t *testing.T) {
	twitch := newTwitchTestServer(t, liveTwitchJSON)
	setTwitchTestURL(t, twitch)

	var limitedCalls atomic.Int32
	limited := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limitedCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"message":"rate limited","retry_after":120}`)
	}))
	defer limited.Close()
	healthy, capture := newDiscordCapture(t)

	app := newTestApp(t)
	limitedWebhook := injectWebhook(t, app, "Limited", limited.URL+"/api/webhooks/1/token")
	healthyWebhook := injectWebhook(t, app, "Healthy", healthy.URL+"/api/webhooks/2/token")
	channel, err := app.AddChannel("teststreamer")
	if err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if err := app.SetDestinations(channel.ID, []string{limitedWebhook.ID, healthyWebhook.ID}); err != nil {
		t.Fatalf("SetDestinations returned error: %v", err)
	}

	summary, err := app.CheckNow(context.Background())
	if err != nil {
		t.Fatalf("CheckNow returned error: %v", err)
	}
	if summary.Posted != 1 || summary.Errors != 1 || capture.count() != 1 || limitedCalls.Load() != 1 {
		t.Fatalf("summary = %+v, healthy calls = %d, limited calls = %d", summary, capture.count(), limitedCalls.Load())
	}
	for _, destination := range app.ConfigSnapshot().Channels[0].Destinations {
		if destination.WebhookID == limitedWebhook.ID && destination.RetryUntil == "" {
			t.Fatal("limited destination did not persist a retry window")
		}
		if destination.WebhookID == healthyWebhook.ID && destination.RetryUntil != "" {
			t.Fatalf("healthy destination was rate limited: %+v", destination)
		}
	}
}
