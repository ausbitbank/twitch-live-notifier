package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestPostDiscord(t *testing.T) {
	var calls atomic.Int32
	var query string
	var payload discordWebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		query = r.URL.RawQuery
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"message123"}`)
	}))
	defer server.Close()

	result, err := postDiscord(server.URL+"/api/webhooks/1/token", discordWebhookPayload{Content: "hello"})
	if err != nil {
		t.Fatalf("postDiscord returned error: %v", err)
	}
	if calls.Load() != 1 || result.MessageID != "message123" {
		t.Fatalf("calls = %d, result = %+v", calls.Load(), result)
	}
	if !strings.Contains(query, "wait=true") {
		t.Fatalf("query = %q", query)
	}
	if payload.Content != "hello" {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestDiscordErrorDoesNotLeakWebhookToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"message":"Invalid Webhook Token","code":50027}`)
	}))
	defer server.Close()

	_, err := postDiscord(server.URL+"/api/webhooks/1/supersecret", discordWebhookPayload{Content: "test"})
	if err == nil {
		t.Fatal("postDiscord returned nil error")
	}
	if !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "Invalid Webhook Token") {
		t.Fatalf("error = %q", err)
	}
	if strings.Contains(err.Error(), "supersecret") {
		t.Fatalf("error leaks the webhook token: %q", err)
	}
}

func TestPostDiscordRetries429WithBodyDelay(t *testing.T) {
	recorder := stubRetrySleep(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(w, `{"message":"You are being rate limited.","retry_after":0.25,"global":false}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"message456"}`)
	}))
	defer server.Close()

	result, err := postDiscord(server.URL+"/api/webhooks/1/token", discordWebhookPayload{Content: "hello"})
	if err != nil {
		t.Fatalf("postDiscord returned error: %v", err)
	}
	if calls.Load() != 2 || result.MessageID != "message456" {
		t.Fatalf("calls = %d, result = %+v", calls.Load(), result)
	}
	delays := recorder.snapshot()
	if len(delays) != 1 || delays[0] != 250*time.Millisecond {
		t.Fatalf("retry delays = %#v", delays)
	}
}

func TestPostDiscordRetries429WithHeaderDelay(t *testing.T) {
	recorder := stubRetrySleep(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "2")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(w, `{"message":"You are being rate limited."}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"message789"}`)
	}))
	defer server.Close()

	result, err := postDiscord(server.URL+"/api/webhooks/1/token", discordWebhookPayload{Content: "hello"})
	if err != nil {
		t.Fatalf("postDiscord returned error: %v", err)
	}
	if calls.Load() != 2 || result.MessageID != "message789" {
		t.Fatalf("calls = %d, result = %+v", calls.Load(), result)
	}
	delays := recorder.snapshot()
	if len(delays) != 1 || delays[0] != 2*time.Second {
		t.Fatalf("retry delays = %#v", delays)
	}
}

func TestPostDiscordStopsAfterOne429Retry(t *testing.T) {
	stubRetrySleep(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		io.WriteString(w, `{"message":"still rate limited","retry_after":0.01}`)
	}))
	defer server.Close()

	if _, err := postDiscord(server.URL+"/api/webhooks/1/token", discordWebhookPayload{Content: "hello"}); err == nil {
		t.Fatal("postDiscord returned nil error")
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
}

func TestPermanentDiscordErrorsAreClassified(t *testing.T) {
	for _, code := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound} {
		err := &discordHTTPError{StatusCode: code, Detail: ": nope"}
		if !isPermanentDiscordError(err) {
			t.Errorf("HTTP %d was not classified as permanent", code)
		}
	}
	if isPermanentDiscordError(&discordHTTPError{StatusCode: http.StatusTooManyRequests}) {
		t.Error("429 was classified as permanent")
	}
	if isPermanentDiscordError(errors.New("network")) {
		t.Error("plain network error was classified as permanent")
	}
}

func TestPostDiscordThreadTargetAndCustomIdentity(t *testing.T) {
	var query string
	var payload discordWebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"message123"}`)
	}))
	defer server.Close()

	info := sampleStreamInfo()
	info.VODURL = "https://www.twitch.tv/videos/123"
	webhook := Webhook{
		Label:     "Community",
		Username:  "Community Alerts",
		AvatarURL: "https://example.com/avatar.png",
		Template:  defaultTemplate(),
		Buttons: []TemplateButton{
			{Label: "Watch {streamer}", URL: "{url}"},
			{Label: "VOD", URL: "{vod}"},
		},
	}
	built := buildPayload(webhook, info, time.Now(), false)
	if built.Username != "Community Alerts" || built.AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("identity = %+v", built)
	}
	if len(built.Components) != 1 || len(built.Components[0].Components) != 2 {
		t.Fatalf("components = %+v", built.Components)
	}
	if built.Components[0].Components[0].URL != info.URL {
		t.Fatalf("button URL = %q", built.Components[0].Components[0].URL)
	}

	if _, err := postDiscordContextWithOptions(context.Background(), server.URL+"/api/webhooks/1/token", built, discordWebhookOptions{ThreadID: "987654321012345678"}); err != nil {
		t.Fatalf("postDiscordContextWithOptions returned error: %v", err)
	}
	if !strings.Contains(query, "thread_id=987654321012345678") || !strings.Contains(query, "wait=true") {
		t.Fatalf("query = %q", query)
	}
	if payload.Username != "Community Alerts" || len(payload.Components) != 1 {
		t.Fatalf("posted payload = %+v", payload)
	}
}

func TestEditAndDeleteDiscordMessageTargetStoredID(t *testing.T) {
	var method, path, query string
	var payload discordWebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path, query = r.Method, r.URL.Path, r.URL.RawQuery
		if r.Method == http.MethodPatch {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode payload: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"id":"message123"}`)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	result, err := editDiscordContext(context.Background(), server.URL+"/api/webhooks/1/token", "message123", discordWebhookPayload{Content: "updated"}, "987654321012345678")
	if err != nil {
		t.Fatalf("editDiscordContext returned error: %v", err)
	}
	if method != http.MethodPatch || path != "/api/webhooks/1/token/messages/message123" {
		t.Fatalf("edit request = %s %s", method, path)
	}
	if !strings.Contains(query, "thread_id=987654321012345678") {
		t.Fatalf("edit query = %q", query)
	}
	if result.MessageID != "message123" || payload.Content != "updated" {
		t.Fatalf("result = %+v, payload = %+v", result, payload)
	}

	if err := deleteDiscordContext(context.Background(), server.URL+"/api/webhooks/1/token", "message123", ""); err != nil {
		t.Fatalf("deleteDiscordContext returned error: %v", err)
	}
	if method != http.MethodDelete || path != "/api/webhooks/1/token/messages/message123" {
		t.Fatalf("delete request = %s %s", method, path)
	}
}

func TestEditDiscordRetries429(t *testing.T) {
	recorder := stubRetrySleep(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(w, `{"message":"slow down","retry_after":0.5}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"message123"}`)
	}))
	defer server.Close()

	if _, err := editDiscordContext(context.Background(), server.URL+"/api/webhooks/1/token", "message123", discordWebhookPayload{Content: "updated"}, ""); err != nil {
		t.Fatalf("editDiscordContext returned error: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d", calls.Load())
	}
	delays := recorder.snapshot()
	if len(delays) != 1 || delays[0] != 500*time.Millisecond {
		t.Fatalf("retry delays = %#v", delays)
	}
}

func TestPostDiscordPermanentErrorDoesNotRetry(t *testing.T) {
	recorder := stubRetrySleep(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"message":"Unknown Webhook","code":10015}`)
	}))
	defer server.Close()

	_, err := postDiscord(server.URL+"/api/webhooks/1/token", discordWebhookPayload{Content: "hello"})
	if err == nil || !isPermanentDiscordError(err) {
		t.Fatalf("error = %v", err)
	}
	if calls.Load() != 1 || len(recorder.snapshot()) != 0 {
		t.Fatalf("calls = %d, delays = %#v", calls.Load(), recorder.snapshot())
	}
}
