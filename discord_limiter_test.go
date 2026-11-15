package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDiscordRateKeyGroupsMessageEndpoints(t *testing.T) {
	post := "https://discord.com/api/webhooks/123/token?wait=true"
	edit := "https://discord.com/api/webhooks/123/token/messages/456"
	deleteURL := "https://discord.com/api/webhooks/123/token/messages/456?thread_id=789"
	if discordRateKey(post) != discordRateKey(edit) || discordRateKey(edit) != discordRateKey(deleteURL) {
		t.Fatalf("rate keys differ: %q %q %q", discordRateKey(post), discordRateKey(edit), discordRateKey(deleteURL))
	}
}

func TestDiscordRateLimiterHonorsProactiveHeaders(t *testing.T) {
	recorder := stubRetrySleep(t)
	limiter := newDiscordRateLimiter()
	url := "https://discord.com/api/webhooks/1/token"

	headers := http.Header{}
	headers.Set("X-RateLimit-Remaining", "0")
	headers.Set("X-RateLimit-Reset-After", "2")
	limiter.observe(url, headers)

	if err := limiter.wait(context.Background(), url); err != nil {
		t.Fatalf("wait returned error: %v", err)
	}
	delays := recorder.snapshot()
	if len(delays) != 1 || delays[0] < 1900*time.Millisecond || delays[0] > 2100*time.Millisecond {
		t.Fatalf("delays = %#v", delays)
	}

	headers.Set("X-RateLimit-Remaining", "5")
	limiter.observe(url, headers)
	if err := limiter.wait(context.Background(), url); err != nil {
		t.Fatalf("wait after reset returned error: %v", err)
	}
	if len(recorder.snapshot()) != 1 {
		t.Fatalf("limiter kept a stale delay: %#v", recorder.snapshot())
	}
}

func TestDiscordRateLimiterDefersLongWindows(t *testing.T) {
	recorder := stubRetrySleep(t)
	limiter := newDiscordRateLimiter()
	url := "https://discord.com/api/webhooks/1/token"
	limiter.deferFor(url, 5*time.Minute)

	err := limiter.wait(context.Background(), url)
	var httpErr *discordHTTPError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("wait error = %v", err)
	}
	if httpErr.RetryAfter < 4*time.Minute {
		t.Fatalf("retry after = %v", httpErr.RetryAfter)
	}
	if len(recorder.snapshot()) != 0 {
		t.Fatalf("long window slept inline: %#v", recorder.snapshot())
	}
}

func TestLong429IsDeferredWithoutInlineRetry(t *testing.T) {
	recorder := stubRetrySleep(t)
	resetDiscordRateLimiter()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		io.WriteString(w, `{"message":"rate limited","retry_after":120}`)
	}))
	defer server.Close()

	_, err := postDiscord(server.URL+"/api/webhooks/1/token", discordWebhookPayload{Content: "hello"})
	var httpErr *discordHTTPError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("error = %v", err)
	}
	if httpErr.RetryAfter != 120*time.Second {
		t.Fatalf("retry after = %v", httpErr.RetryAfter)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls = %d", calls.Load())
	}
	if len(recorder.snapshot()) != 0 {
		t.Fatalf("long 429 slept inline: %#v", recorder.snapshot())
	}
}

func TestDiscordProactiveHeadersDelayNextRequest(t *testing.T) {
	recorder := stubRetrySleep(t)
	resetDiscordRateLimiter()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset-After", "1.5")
		io.WriteString(w, `{"id":"message123"}`)
	}))
	defer server.Close()

	url := server.URL + "/api/webhooks/1/token"
	if _, err := postDiscord(url, discordWebhookPayload{Content: "first"}); err != nil {
		t.Fatalf("first post returned error: %v", err)
	}
	if _, err := postDiscord(url, discordWebhookPayload{Content: "second"}); err != nil {
		t.Fatalf("second post returned error: %v", err)
	}
	delays := recorder.snapshot()
	if len(delays) != 1 || delays[0] < 1400*time.Millisecond || delays[0] > 1600*time.Millisecond {
		t.Fatalf("delays = %#v", delays)
	}
}

func TestRateLimitHelpers(t *testing.T) {
	if destinationRateLimited(Destination{RetryUntil: time.Now().Add(time.Minute).UTC().Format(time.RFC3339)}, time.Now()) != true {
		t.Fatal("future RetryUntil was not rate limited")
	}
	if destinationRateLimited(Destination{RetryUntil: time.Now().Add(-time.Minute).UTC().Format(time.RFC3339)}, time.Now()) {
		t.Fatal("past RetryUntil was rate limited")
	}
	if destinationRateLimited(Destination{RetryUntil: "not a time"}, time.Now()) {
		t.Fatal("invalid RetryUntil was rate limited")
	}
	delay := rateLimitDelay(&discordHTTPError{StatusCode: http.StatusTooManyRequests, RetryAfter: 2 * time.Minute})
	if delay != 2*time.Minute {
		t.Fatalf("delay = %v", delay)
	}
	if rateLimitDelay(errors.New("nope")) != 0 {
		t.Fatal("non-429 error produced a retry delay")
	}
	if normalizeRetryUntil("not a time") != "" {
		t.Fatal("invalid RetryUntil survived normalization")
	}
	if !strings.Contains(discordRateKey("https://discord.com/api/webhooks/1/token"), "/api/webhooks/1/token") {
		t.Fatal("unexpected rate key")
	}
}
