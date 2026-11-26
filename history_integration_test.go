package main

import (
	"context"
	"testing"
)

func TestCheckNowRecordsPostDelivery(t *testing.T) {
	twitch := newTwitchTestServer(t, liveTwitchJSON)
	setTwitchTestURL(t, twitch)
	discord, _ := newDiscordCapture(t)

	app := newTestApp(t)
	webhook := injectWebhook(t, app, "Community", discord.URL+"/api/webhooks/1/token")
	channel, err := app.AddChannel("teststreamer")
	if err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if len(channel.Destinations) != 1 {
		t.Fatalf("channel destinations = %+v", channel.Destinations)
	}

	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("CheckNow returned error: %v", err)
	}

	entries, total := app.history.List(10, DeliveryFilters{Kind: deliveryKindPosted})
	if total != 1 || len(entries) != 1 {
		t.Fatalf("posted history = %d/%d, want 1", len(entries), total)
	}
	record := entries[0]
	if record.ChannelLogin != "teststreamer" || record.ChannelName != "TestStreamer" {
		t.Fatalf("record identity = %+v", record)
	}
	if record.WebhookID != webhook.ID || record.WebhookLabel != "Community" {
		t.Fatalf("record destination = %+v", record)
	}
	if record.StreamID != "319913347671" || record.Title != "Testing the live notifier" || record.Game != "Just Chatting" {
		t.Fatalf("record stream = %+v", record)
	}
	if record.MessageID != "message123" {
		t.Fatalf("record message = %+v", record)
	}
}

func TestCheckNowRecordsFilterSkip(t *testing.T) {
	twitch := newTwitchTestServer(t, liveTwitchJSON)
	setTwitchTestURL(t, twitch)
	discord, _ := newDiscordCapture(t)

	app := newTestApp(t)
	injectWebhook(t, app, "Community", discord.URL+"/api/webhooks/1/token")
	channel, err := app.AddChannel("teststreamer")
	if err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if err := app.SetChannelFilters(channel.ID, Filters{ExcludeGames: []string{"Just Chatting"}}); err != nil {
		t.Fatalf("SetChannelFilters returned error: %v", err)
	}

	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("CheckNow returned error: %v", err)
	}

	entries, total := app.history.List(10, DeliveryFilters{Kind: deliveryKindSkipped})
	if total != 1 || len(entries) != 1 {
		t.Fatalf("skipped history = %d/%d, want 1", len(entries), total)
	}
	if entries[0].Reason == "" {
		t.Fatalf("skipped record has no reason: %+v", entries[0])
	}
	if entries[0].ChannelLogin != "teststreamer" {
		t.Fatalf("skipped record identity = %+v", entries[0])
	}
}
