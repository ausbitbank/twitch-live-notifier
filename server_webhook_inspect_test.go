package main

import "testing"

func TestParseDiscordWebhookInfo(t *testing.T) {
	info, err := parseDiscordWebhookInfo([]byte(`{
		"id": "123456789012345678",
		"name": "Live alerts",
		"channel_id": "223456789012345678",
		"guild_id": "323456789012345678",
		"token": "should-not-leak"
	}`))
	if err != nil {
		t.Fatalf("parseDiscordWebhookInfo: %v", err)
	}
	if info.Name != "Live alerts" || info.ChannelID != "223456789012345678" || info.GuildID != "323456789012345678" {
		t.Fatalf("unexpected info: %+v", info)
	}
}

func TestParseDiscordWebhookInfoRejectsEmpty(t *testing.T) {
	if _, err := parseDiscordWebhookInfo([]byte(`{"id":"1"}`)); err == nil {
		t.Fatal("expected an error for a response without metadata")
	}
	if _, err := parseDiscordWebhookInfo([]byte(`not json`)); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}
