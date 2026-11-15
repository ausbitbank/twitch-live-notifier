package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNormalizeTwitchLogin(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{"TestStreamer", "teststreamer", true},
		{"@Foo_Bar", "foo_bar", true},
		{"  foo  ", "foo", true},
		{"https://www.twitch.tv/Foo?x=1", "foo", true},
		{"twitch.tv/foo/", "foo", true},
		{"www.twitch.tv/foo", "foo", true},
		{"https://m.twitch.tv/foo/videos", "foo", true},
		{"", "", false},
		{"foo bar", "", false},
		{"foo-bar", "", false},
		{"https://example.com/foo", "", false},
		{"https://twitch.tv/", "", false},
		{"this_name_is_way_too_long_for_twitch", "", false},
	}

	for _, test := range tests {
		got, err := normalizeTwitchLogin(test.input)
		if test.ok && err != nil {
			t.Errorf("normalizeTwitchLogin(%q) returned error: %v", test.input, err)
			continue
		}
		if !test.ok && err == nil {
			t.Errorf("normalizeTwitchLogin(%q) = %q, want an error", test.input, got)
			continue
		}
		if got != test.want {
			t.Errorf("normalizeTwitchLogin(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}

func TestNormalizeWebhookURL(t *testing.T) {
	valid := []string{
		"https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
		"https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz?wait=true",
		"https://discordapp.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
		"https://ptb.discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
		"https://canary.discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
	}
	for _, input := range valid {
		got, err := normalizeWebhookURL(input)
		if err != nil {
			t.Errorf("normalizeWebhookURL(%q) returned error: %v", input, err)
			continue
		}
		if input == valid[0] || input == valid[1] {
			want := "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz"
			if got != want {
				t.Errorf("normalizeWebhookURL(%q) = %q, want %q", input, got, want)
			}
		}
	}

	invalid := []string{
		"",
		"http://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
		"https://example.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
		"https://discord.com/api/webhooks/notanid/abcdefghijklmnopqrstuvwxyz",
		"https://discord.com/api/webhooks/1234567890/short",
		"https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz/extra",
		"https://discord.com:8443/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
		"https://user:pass@discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
	}
	for _, input := range invalid {
		if got, err := normalizeWebhookURL(input); err == nil {
			t.Errorf("normalizeWebhookURL(%q) = %q, want an error", input, got)
		}
	}
}

func TestConfigRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	login := "teststreamer"
	webhookURL := "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz"
	want := Config{
		Version: configVersion,
		Channels: []Channel{{
			ID:          "channel1",
			Login:       login,
			DisplayName: "TestStreamer",
			Enabled:     true,
			Destinations: []Destination{{
				WebhookID:    "webhook1",
				LastStreamID: "319913347671",
				LastPostedAt: "2026-11-15T06:00:00Z",
			}},
			Filters: Filters{
				IncludeGames:    []string{"Just Chatting"},
				ExcludeTitles:   []string{"spoilers"},
				MinViewers:      10,
				QuietStart:      "23:00",
				QuietEnd:        "07:00",
				AllowedDays:     []int{1, 3},
				CooldownMinutes: 60,
				DelayMinutes:    5,
			},
			LastAlertAt:       "2026-11-15T06:00:00Z",
			LastAlertDate:     "2026-11-15",
			LastAlertStreamID: "319913347671",
		}},
		Webhooks: []Webhook{{
			ID:       "webhook1",
			Label:    "Community",
			URL:      webhookURL,
			Enabled:  true,
			RoleIDs:  []string{"123456789012345678"},
			Template: defaultTemplate(),
		}},
		Settings: Settings{PollSeconds: 120, Theme: "dark"},
	}
	if err := saveConfig(path, want); err != nil {
		t.Fatalf("saveConfig returned error: %v", err)
	}
	got, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if got.Webhooks[0].URLSecretID == "" {
		t.Fatal("webhook URL was not moved into the secret store")
	}
	want.Webhooks[0].URLSecretID = got.Webhooks[0].URLSecretID
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded config = %+v, want %+v", got, want)
	}
}

func TestLegacyConfigMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	legacy := `{
		"twitch_login": "TestStreamer",
		"discord_webhook": "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
		"last_notified_stream_id": "319913347671",
		"last_notified_at": "2026-11-15T06:00:00Z"
	}`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if cfg.Version != configVersion {
		t.Errorf("version = %d, want %d", cfg.Version, configVersion)
	}
	if len(cfg.Channels) != 1 || len(cfg.Webhooks) != 1 {
		t.Fatalf("migrated config = %+v", cfg)
	}
	if cfg.Channels[0].Login != "teststreamer" || !cfg.Channels[0].Enabled {
		t.Errorf("channel = %+v", cfg.Channels[0])
	}
	if len(cfg.Channels[0].Destinations) != 1 || cfg.Channels[0].Destinations[0].LastStreamID != "319913347671" {
		t.Errorf("destinations = %+v", cfg.Channels[0].Destinations)
	}
	if cfg.Webhooks[0].Label != "Discord" || !cfg.Webhooks[0].Enabled {
		t.Errorf("webhook = %+v", cfg.Webhooks[0])
	}
}

func TestVersion3ConfigMigratesWithoutLosingDedupe(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	v3 := `{
		"version": 3,
		"channels": [{
			"id": "channel1",
			"login": "teststreamer",
			"display_name": "TestStreamer",
			"enabled": true,
			"destinations": [{
				"webhook_id": "webhook1",
				"last_stream_id": "319913347671",
				"last_posted_at": "2026-11-15T06:00:00Z"
			}]
		}],
		"webhooks": [{
			"id": "webhook1",
			"label": "Community",
			"url": "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz",
			"enabled": true,
			"template": {
				"content": "{streamer} is live",
				"embed_title": "{title}",
				"embed_description": "",
				"footer": "",
				"color": "#9146FF",
				"show_author": true,
				"show_image": true,
				"show_game": true,
				"show_viewers": true,
				"show_live_since": true
			}
		}],
		"settings": {"poll_seconds": 120, "theme": "dark"}
	}`
	if err := os.WriteFile(path, []byte(v3), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if cfg.Version != configVersion {
		t.Fatalf("version = %d", cfg.Version)
	}
	if len(cfg.Channels) != 1 || cfg.Channels[0].Destinations[0].LastStreamID != "319913347671" {
		t.Fatalf("dedupe state = %+v", cfg.Channels)
	}
	if cfg.Webhooks[0].URL == "" || cfg.Webhooks[0].Cleanup.Mode != cleanupOff {
		t.Fatalf("v3 webhook migration = %+v", cfg.Webhooks[0])
	}
}

func TestLoadConfigMissingAndMalformed(t *testing.T) {
	missing, err := loadConfig(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("missing config error: %v", err)
	}
	if missing.Version != configVersion || missing.Settings.Theme != "dark" || missing.Settings.PollSeconds != 120 {
		t.Fatalf("missing config = %+v", missing)
	}

	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil {
		t.Fatal("loadConfig accepted malformed JSON")
	}
}

func TestCloneConfigDoesNotShareSlices(t *testing.T) {
	cfg := defaultConfig()
	cfg.Channels = []Channel{{ID: "a", Destinations: []Destination{{WebhookID: "b"}}}}
	clone := cloneConfig(cfg)
	clone.Channels[0].Destinations[0].WebhookID = "c"
	if cfg.Channels[0].Destinations[0].WebhookID != "b" {
		t.Fatal("cloneConfig shared destination storage")
	}
}

func TestNormalizeRoleIDs(t *testing.T) {
	const first = "123456789012345678"
	const second = "987654321098765432"
	roles, err := normalizeRoleIDs([]string{first, "<@&" + second + ">", first, ""})
	if err != nil {
		t.Fatalf("normalizeRoleIDs returned error: %v", err)
	}
	if len(roles) != 2 || roles[0] != first || roles[1] != second {
		t.Fatalf("roles = %#v", roles)
	}
	if _, err := normalizeRoleIDs([]string{"@everyone"}); err == nil {
		t.Fatal("invalid role ID accepted")
	}
	empty, err := normalizeRoleIDs(nil)
	if err != nil || empty != nil {
		t.Fatalf("empty roles = %#v, %v", empty, err)
	}
}

func TestNormalizeDiscordID(t *testing.T) {
	const id = "123456789012345678"
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{id, id, true},
		{"<@" + id + ">", id, true},
		{"<@!" + id + ">", id, true},
		{"https://discord.com/users/" + id, id, true},
		{"https://discordapp.com/users/" + id + "/", id, true},
		{"", "", true},
		{"@somebody", "", false},
		{"123", "", false},
		{"<@123>", "", false},
	}
	for _, test := range tests {
		got, err := normalizeDiscordID(test.input)
		if test.ok && err != nil {
			t.Errorf("normalizeDiscordID(%q) returned error: %v", test.input, err)
			continue
		}
		if !test.ok && err == nil {
			t.Errorf("normalizeDiscordID(%q) = %q, want an error", test.input, got)
			continue
		}
		if got != test.want {
			t.Errorf("normalizeDiscordID(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}
