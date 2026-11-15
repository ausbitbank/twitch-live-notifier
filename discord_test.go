package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
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
