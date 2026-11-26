package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fakeDiscordToken() string {
	return strings.Repeat("A", 24) + "." + strings.Repeat("B", 6) + "." + strings.Repeat("C", 27)
}

func newDiagnosticsTestApp(t *testing.T, dir string) *App {
	t.Helper()
	return &App{
		configPath: filepath.Join(dir, "config.json"),
		cfg: Config{
			Channels: []Channel{
				{ID: "channel-1", Login: "alpha", DisplayName: "Alpha", Enabled: true},
				{ID: "channel-2", Login: "beta", DisplayName: "Beta", Enabled: false},
			},
			Webhooks: []Webhook{
				{ID: "webhook-1", Label: "#live", Enabled: true},
			},
		},
		statuses: map[string]ChannelStatus{
			"channel-1": {Live: true, Title: "Speedrun night"},
			"channel-2": {Error: "network down"},
		},
		activity: []Activity{
			{ID: "activity-1", Time: time.Now(), Level: "error", Message: "Could not post to #live: boom"},
		},
		startedAt: time.Now().Add(-2 * time.Minute),
		lastCheck: time.Now().Add(-30 * time.Second),
		updateStatus: UpdateStatus{
			CurrentVersion:  Version,
			LatestVersion:   "2.6.0",
			UpdateAvailable: true,
			CheckedAt:       time.Now().UTC().Format(time.RFC3339),
		},
	}
}

func TestBuildDiagnosticsShape(t *testing.T) {
	dir := t.TempDir()
	app := newDiagnosticsTestApp(t, dir)
	history := newDeliveryLog(dir)
	if err := history.Add(DeliveryRecord{
		Kind:         deliveryKindPosted,
		Level:        deliveryLevelSuccess,
		ChannelID:    "channel-1",
		ChannelLogin: "alpha",
		ChannelName:  "Alpha",
		WebhookID:    "webhook-1",
		WebhookLabel: "#live",
		Title:        "Speedrun night",
	}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	diagnostics := buildDiagnostics(app, history)

	if diagnostics.App != appName {
		t.Errorf("app = %q", diagnostics.App)
	}
	if diagnostics.Version != Version {
		t.Errorf("version = %q", diagnostics.Version)
	}
	if !strings.HasPrefix(diagnostics.GoVersion, "go") {
		t.Errorf("go_version = %q", diagnostics.GoVersion)
	}
	if diagnostics.Platform == "" {
		t.Error("platform is empty")
	}
	if diagnostics.DataDir != dir || diagnostics.ConfigPath != app.configPath {
		t.Errorf("paths = %q / %q", diagnostics.DataDir, diagnostics.ConfigPath)
	}
	if diagnostics.UptimeSeconds < 100 {
		t.Errorf("uptime = %d, want at least 100", diagnostics.UptimeSeconds)
	}
	if diagnostics.LastCheck == "" {
		t.Error("last_check is empty")
	}
	if diagnostics.Channels != (DiagnosticsChannels{Total: 2, Enabled: 1, Live: 1, Errors: 1}) {
		t.Errorf("channels = %+v", diagnostics.Channels)
	}
	if len(diagnostics.Webhooks) != 1 {
		t.Fatalf("webhooks = %+v", diagnostics.Webhooks)
	}
	webhook := diagnostics.Webhooks[0]
	if webhook.Label != "#live" || !webhook.Enabled || webhook.Invalid || webhook.LastError != "" {
		t.Errorf("webhook = %+v", webhook)
	}
	if len(diagnostics.RecentErrors) == 0 {
		t.Fatal("recent_errors is empty")
	}
	joined := strings.Join(diagnostics.RecentErrors, "\n")
	if !strings.Contains(joined, "network down") || !strings.Contains(joined, "boom") {
		t.Errorf("recent_errors = %q", joined)
	}
	if len(diagnostics.RecentHistory) != 1 || diagnostics.RecentHistory[0].ChannelLogin != "alpha" {
		t.Errorf("recent_history = %+v", diagnostics.RecentHistory)
	}
	if diagnostics.Update.LatestVersion != "2.6.0" || !diagnostics.Update.UpdateAvailable {
		t.Errorf("update = %+v", diagnostics.Update)
	}
}

func TestBuildDiagnosticsRedactsSecrets(t *testing.T) {
	dir := t.TempDir()
	app := newDiagnosticsTestApp(t, dir)
	secretURL := "https://discord.com/api/webhooks/123456789012345678/superSecretTokenValue"
	token := fakeDiscordToken()
	app.cfg.Webhooks[0].URL = secretURL
	app.cfg.Webhooks[0].LastError = "posting to " + secretURL + " failed"
	app.activity = append(app.activity, Activity{
		ID:      "activity-2",
		Time:    time.Now(),
		Level:   "error",
		Message: "Discord token " + token + " rejected",
	})

	history := newDeliveryLog(dir)
	if err := history.Add(DeliveryRecord{
		Kind:         deliveryKindError,
		Level:        deliveryLevelError,
		ChannelLogin: "alpha",
		Title:        "bad " + secretURL,
		Reason:       "callback https://example.test/hook?token=abc123 rejected " + token,
	}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	data, err := json.Marshal(buildDiagnostics(app, history))
	if err != nil {
		t.Fatalf("marshal diagnostics: %v", err)
	}
	output := string(data)
	for _, secret := range []string{
		"superSecretTokenValue",
		"123456789012345678",
		"MTIzNDU2Nzg5MDEyMzQ1Njc4",
		"abc123",
		"api/webhooks",
	} {
		if strings.Contains(output, secret) {
			t.Errorf("diagnostics leaked %q:\n%s", secret, output)
		}
	}
	if !strings.Contains(output, "[redacted webhook]") {
		t.Errorf("diagnostics did not mark the redacted webhook:\n%s", output)
	}
	if !strings.Contains(output, "[redacted token]") {
		t.Errorf("diagnostics did not mark the redacted token:\n%s", output)
	}
}

func TestSanitizeDiagnosticText(t *testing.T) {
	input := "url https://discord.com/api/webhooks/123/secret and token=topsecret plus " + fakeDiscordToken()
	output := sanitizeDiagnosticText(input, 500)
	for _, secret := range []string{"secret and", "topsecret", strings.Repeat("A", 24)} {
		if strings.Contains(output, secret) {
			t.Errorf("sanitizeDiagnosticText leaked %q: %q", secret, output)
		}
	}
}

func TestDiagnosticsJSON(t *testing.T) {
	dir := t.TempDir()
	app := newDiagnosticsTestApp(t, dir)
	data, err := app.DiagnosticsJSON()
	if err != nil {
		t.Fatalf("DiagnosticsJSON: %v", err)
	}
	var diagnostics Diagnostics
	if err := json.Unmarshal(data, &diagnostics); err != nil {
		t.Fatalf("diagnostics JSON is invalid: %v", err)
	}
	if diagnostics.App != appName || diagnostics.Channels.Total != 2 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}

	empty := buildDiagnostics(nil, nil)
	if empty.App != appName || empty.RecentErrors == nil || empty.RecentHistory == nil {
		t.Fatalf("nil-app diagnostics = %+v", empty)
	}
}

func TestDeliveryHistoryRoutes(t *testing.T) {
	dir := t.TempDir()
	history := newDeliveryLog(dir)
	for _, record := range []DeliveryRecord{
		{ID: "1", Kind: deliveryKindPosted, Level: deliveryLevelSuccess, ChannelID: "channel-1", ChannelLogin: "alpha", ChannelName: "Alpha"},
		{ID: "2", Kind: deliveryKindPosted, Level: deliveryLevelSuccess, ChannelID: "channel-1", ChannelLogin: "alpha", ChannelName: "Alpha"},
		{ID: "3", Kind: deliveryKindSkipped, Level: deliveryLevelInfo, ChannelID: "channel-2", ChannelLogin: "beta", ChannelName: "Beta", Reason: "Quiet hours"},
	} {
		if err := history.Add(record); err != nil {
			t.Fatalf("Add %s: %v", record.ID, err)
		}
	}

	app := newDiagnosticsTestApp(t, dir)
	previous := appHistoryResolver
	appHistoryResolver = func(*App) *deliveryLog { return history }
	defer func() { appHistoryResolver = previous }()

	mux := http.NewServeMux()
	registerHistoryRoutes(mux, app)
	server := httptest.NewServer(mux)
	defer server.Close()

	type historyResponse struct {
		Entries []DeliveryRecord `json:"entries"`
		Total   int              `json:"total"`
		Limit   int              `json:"limit"`
	}

	response, err := http.Get(server.URL + "/api/history?kind=posted&limit=1")
	if err != nil {
		t.Fatalf("GET history: %v", err)
	}
	var listed historyResponse
	if err := json.NewDecoder(response.Body).Decode(&listed); err != nil {
		t.Fatalf("decode history: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET history status = %d", response.StatusCode)
	}
	if listed.Total != 2 || len(listed.Entries) != 1 || listed.Limit != 1 {
		t.Fatalf("history response = %+v", listed)
	}
	if listed.Entries[0].ChannelLogin != "alpha" {
		t.Fatalf("history entry = %+v", listed.Entries[0])
	}

	response, err = http.Get(server.URL + "/api/history?q=quiet")
	if err != nil {
		t.Fatalf("GET history query: %v", err)
	}
	if err := json.NewDecoder(response.Body).Decode(&listed); err != nil {
		t.Fatalf("decode queried history: %v", err)
	}
	response.Body.Close()
	if listed.Total != 1 || listed.Entries[0].ID != "3" {
		t.Fatalf("queried history = %+v", listed)
	}

	response, err = http.Get(server.URL + "/api/diagnostics")
	if err != nil {
		t.Fatalf("GET diagnostics: %v", err)
	}
	var diagnostics Diagnostics
	if err := json.NewDecoder(response.Body).Decode(&diagnostics); err != nil {
		t.Fatalf("decode diagnostics: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK || diagnostics.App != appName {
		t.Fatalf("diagnostics status = %d, body = %+v", response.StatusCode, diagnostics)
	}

	// A cross-origin DELETE is rejected and leaves the history intact.
	request, err := http.NewRequest(http.MethodDelete, server.URL+"/api/history", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	request.Header.Set("Origin", "http://evil.example")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("cross-origin DELETE: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin DELETE status = %d", response.StatusCode)
	}
	if _, total := history.List(0, DeliveryFilters{}); total != 3 {
		t.Fatalf("cross-origin DELETE cleared history: total = %d", total)
	}

	// A same-origin (or origin-less) DELETE clears the history.
	request, err = http.NewRequest(http.MethodDelete, server.URL+"/api/history", nil)
	if err != nil {
		t.Fatalf("build DELETE: %v", err)
	}
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("DELETE history: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("DELETE history status = %d", response.StatusCode)
	}
	if _, total := history.List(0, DeliveryFilters{}); total != 0 {
		t.Fatalf("history total after DELETE = %d", total)
	}
}

func TestHistoryLimit(t *testing.T) {
	if got := historyLimit(""); got != defaultHistoryListLimit {
		t.Errorf("empty limit = %d", got)
	}
	if got := historyLimit("bad"); got != defaultHistoryListLimit {
		t.Errorf("bad limit = %d", got)
	}
	if got := historyLimit("-4"); got != defaultHistoryListLimit {
		t.Errorf("negative limit = %d", got)
	}
	if got := historyLimit("12"); got != 12 {
		t.Errorf("limit 12 = %d", got)
	}
	if got := historyLimit("100000"); got != maxHistoryListLimit {
		t.Errorf("large limit = %d", got)
	}
}

func TestDiscordMessageURL(t *testing.T) {
	guild := "123456789012345678"
	channel := "234567890123456789"
	message := "345678901234567890"
	want := "https://discord.com/channels/" + guild + "/" + channel + "/" + message
	if got := discordMessageURL(guild, channel, message); got != want {
		t.Errorf("discordMessageURL = %q, want %q", got, want)
	}
	result := discordPostResult{MessageID: message, ChannelID: channel, GuildID: guild}
	if got := result.MessageURL(); got != want {
		t.Errorf("MessageURL = %q, want %q", got, want)
	}
	for _, test := range []struct{ guild, channel, message string }{
		{"", channel, message},
		{guild, "", message},
		{guild, channel, ""},
		{"abc", channel, message},
		{guild, "12", message},
		{guild, channel, "message123"},
	} {
		if got := discordMessageURL(test.guild, test.channel, test.message); got != "" {
			t.Errorf("discordMessageURL(%q, %q, %q) = %q, want empty", test.guild, test.channel, test.message, got)
		}
	}
}

func TestParseDiscordPostResult(t *testing.T) {
	body := []byte(`{"id":"345678901234567890","channel_id":"234567890123456789","guild_id":"123456789012345678"}`)
	result := parseDiscordPostResult(body, "")
	if result.MessageID != "345678901234567890" || result.ChannelID != "234567890123456789" || result.GuildID != "123456789012345678" {
		t.Fatalf("result = %+v", result)
	}
	if result.MessageURL() != "https://discord.com/channels/123456789012345678/234567890123456789/345678901234567890" {
		t.Fatalf("message URL = %q", result.MessageURL())
	}

	fallback := parseDiscordPostResult([]byte(`{"channel_id":"234567890123456789"}`), "345678901234567890")
	if fallback.MessageID != "345678901234567890" {
		t.Fatalf("fallback message ID = %q", fallback.MessageID)
	}
	empty := parseDiscordPostResult([]byte(`not json`), "")
	if empty.MessageID != "" || empty.ChannelID != "" || empty.GuildID != "" {
		t.Fatalf("empty result = %+v", empty)
	}
}

func TestPostAndEditDiscordExposeChannelAndGuild(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"345678901234567890","channel_id":"234567890123456789","guild_id":"123456789012345678"}`)
	}))
	defer server.Close()

	result, err := postDiscord(server.URL+"/api/webhooks/1/token", discordWebhookPayload{Content: "hello"})
	if err != nil {
		t.Fatalf("postDiscord: %v", err)
	}
	if result.ChannelID != "234567890123456789" || result.GuildID != "123456789012345678" {
		t.Fatalf("post result = %+v", result)
	}
	if result.MessageURL() == "" {
		t.Fatal("post result has no message URL")
	}

	edited, err := editDiscordContext(context.Background(), server.URL+"/api/webhooks/1/token", result.MessageID, discordWebhookPayload{Content: "updated"}, "")
	if err != nil {
		t.Fatalf("editDiscordContext: %v", err)
	}
	if edited.ChannelID != result.ChannelID || edited.GuildID != result.GuildID || edited.MessageID != result.MessageID {
		t.Fatalf("edit result = %+v", edited)
	}
}
