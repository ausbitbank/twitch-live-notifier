package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newImportTestServer(t *testing.T, app *App) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	registerImportRoutes(mux, app)
	registerBulkRoutes(mux, app)
	server := httptest.NewServer(withSecurity(mux))
	t.Cleanup(server.Close)
	return server
}

func postJSONBody(t *testing.T, url, body string) (int, []byte) {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	return response.StatusCode, data
}

func TestImportChannelsEndpoint(t *testing.T) {
	app := newTestApp(t)
	webhook := injectWebhook(t, app, "Discord", "https://discord.com/api/webhooks/1234567890/abcdefghij")
	twitch := newBulkTwitchServer(t)
	setTwitchTestURL(t, twitch)
	dashboard := newImportTestServer(t, app)

	body := `{"text":"alpha\nhttps://twitch.tv/beta","webhook_ids":["` + webhook.ID + `"],"source":"paste"}`
	status, data := postJSONBody(t, dashboard.URL+"/api/channels/import", body)
	if status != http.StatusOK {
		t.Fatalf("status = %d, body = %s", status, data)
	}
	var payload struct {
		Result bulkAddResponse `json:"result"`
		State  State           `json:"state"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Result.Added != 2 || payload.Result.Assigned != 2 || payload.Result.Total != 2 {
		t.Fatalf("result = %+v", payload.Result)
	}
	if len(payload.State.Config.Channels) != 2 {
		t.Fatalf("state channels = %d, want 2", len(payload.State.Config.Channels))
	}
	for _, channel := range payload.State.Config.Channels {
		if len(channel.Destinations) != 1 || channel.Destinations[0].WebhookID != webhook.ID {
			t.Fatalf("%s destinations = %+v", channel.Login, channel.Destinations)
		}
	}
}

func TestImportChannelsEndpointRejectsBadJSON(t *testing.T) {
	app := newTestApp(t)
	dashboard := newImportTestServer(t, app)

	status, _ := postJSONBody(t, dashboard.URL+"/api/channels/import", `{"text":`)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
}

func TestBulkActionEndpoint(t *testing.T) {
	app := newTestApp(t)
	injectWebhook(t, app, "First", "https://discord.com/api/webhooks/1234567890/abcdefghij")
	second := injectWebhook(t, app, "Second", "https://discord.com/api/webhooks/9876543210/abcdefghij")

	app.mu.Lock()
	app.cfg.Channels = []Channel{
		{ID: "c1", Login: "alpha", DisplayName: "Alpha", Enabled: true},
		{ID: "c2", Login: "beta", DisplayName: "Beta", Enabled: true},
		{ID: "c3", Login: "gamma", DisplayName: "Gamma", Enabled: true},
	}
	app.mu.Unlock()

	dashboard := newImportTestServer(t, app)
	assign := `{"channel_ids":["c1","c2"],"action":"assign","webhook_ids":["` + second.ID + `"]}`
	status, data := postJSONBody(t, dashboard.URL+"/api/channels/bulk-action", assign)
	if status != http.StatusOK {
		t.Fatalf("assign status = %d, body = %s", status, data)
	}
	var assigned struct {
		Updated int   `json:"updated"`
		Removed int   `json:"removed"`
		State   State `json:"state"`
	}
	if err := json.Unmarshal(data, &assigned); err != nil {
		t.Fatalf("decode assign response: %v", err)
	}
	if assigned.Updated != 2 || assigned.Removed != 0 {
		t.Fatalf("assign response = %+v", assigned)
	}
	for _, channel := range assigned.State.Config.Channels {
		if channel.ID != "c1" && channel.ID != "c2" {
			continue
		}
		if len(channel.Destinations) != 1 || channel.Destinations[0].WebhookID != second.ID {
			t.Fatalf("%s destinations = %+v", channel.ID, channel.Destinations)
		}
	}

	disable := `{"channel_ids":["c1","c2"],"action":"disable"}`
	status, data = postJSONBody(t, dashboard.URL+"/api/channels/bulk-action", disable)
	if status != http.StatusOK {
		t.Fatalf("disable status = %d, body = %s", status, data)
	}
	var disabled struct {
		Updated int   `json:"updated"`
		State   State `json:"state"`
	}
	if err := json.Unmarshal(data, &disabled); err != nil {
		t.Fatalf("decode disable response: %v", err)
	}
	if disabled.Updated != 2 {
		t.Fatalf("disable updated = %d", disabled.Updated)
	}
	for _, channel := range disabled.State.Config.Channels {
		if channel.ID == "c1" || channel.ID == "c2" {
			if channel.Enabled {
				t.Fatalf("%s is still enabled", channel.ID)
			}
		}
	}

	remove := `{"channel_ids":["c1"],"action":"delete"}`
	status, data = postJSONBody(t, dashboard.URL+"/api/channels/bulk-action", remove)
	if status != http.StatusOK {
		t.Fatalf("delete status = %d, body = %s", status, data)
	}
	var removed struct {
		Removed int   `json:"removed"`
		State   State `json:"state"`
	}
	if err := json.Unmarshal(data, &removed); err != nil {
		t.Fatalf("decode delete response: %v", err)
	}
	if removed.Removed != 1 || len(removed.State.Config.Channels) != 2 {
		t.Fatalf("delete response = %+v", removed)
	}

	loaded, err := loadConfig(app.configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(loaded.Channels) != 2 {
		t.Fatalf("persisted channels = %d, want 2", len(loaded.Channels))
	}
	if loaded.Channels[0].Destinations[0].WebhookID != second.ID {
		t.Fatalf("assignment was not persisted: %+v", loaded.Channels[0].Destinations)
	}

	invalidAction := `{"channel_ids":["c2"],"action":"explode"}`
	status, _ = postJSONBody(t, dashboard.URL+"/api/channels/bulk-action", invalidAction)
	if status != http.StatusBadRequest {
		t.Fatalf("invalid action status = %d, want 400", status)
	}

	invalidWebhook := `{"channel_ids":["c2"],"action":"assign","webhook_ids":["missing"]}`
	status, _ = postJSONBody(t, dashboard.URL+"/api/channels/bulk-action", invalidWebhook)
	if status != http.StatusBadRequest {
		t.Fatalf("invalid webhook status = %d, want 400", status)
	}
}

func TestBulkChannelActionUnassignRemovesOnlySelectedDestinations(t *testing.T) {
	app := newTestApp(t)
	first := injectWebhook(t, app, "First", "https://discord.com/api/webhooks/1234567890/abcdefghij")
	second := injectWebhook(t, app, "Second", "https://discord.com/api/webhooks/9876543210/abcdefghij")
	app.mu.Lock()
	app.cfg.Channels = []Channel{{
		ID: "c1", Login: "alpha", DisplayName: "Alpha", Enabled: true,
		Destinations: []Destination{{WebhookID: first.ID}, {WebhookID: second.ID}},
	}}
	app.mu.Unlock()

	result, err := app.BulkChannelAction("unassign", []string{"c1"}, []string{first.ID})
	if err != nil {
		t.Fatalf("BulkChannelAction returned error: %v", err)
	}
	if result.Updated != 1 || result.Removed != 0 {
		t.Fatalf("result = %+v", result)
	}
	destinations := app.ConfigSnapshot().Channels[0].Destinations
	if len(destinations) != 1 || destinations[0].WebhookID != second.ID {
		t.Fatalf("destinations = %+v", destinations)
	}
}

func TestBulkChannelActionValidatesInput(t *testing.T) {
	app := newTestApp(t)
	app.mu.Lock()
	app.cfg.Channels = []Channel{{ID: "c1", Login: "alpha", DisplayName: "Alpha", Enabled: true}}
	app.mu.Unlock()

	if _, err := app.BulkChannelAction("enable", []string{"missing"}, nil); err != nil {
		t.Fatalf("BulkChannelAction returned error: %v", err)
	}
	if _, err := app.BulkChannelAction("assign", []string{"c1"}, nil); err == nil {
		t.Fatal("assign without destinations was accepted")
	}
}
