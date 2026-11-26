package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func setTwitchImportTestURLs(t *testing.T, server *httptest.Server) {
	t.Helper()
	oldOAuth := twitchOAuthBaseURL
	oldHelix := twitchHelixBaseURL
	twitchOAuthBaseURL = server.URL
	twitchHelixBaseURL = server.URL
	t.Cleanup(func() {
		twitchOAuthBaseURL = oldOAuth
		twitchHelixBaseURL = oldHelix
	})
}

func newTwitchImportServer(t *testing.T, mode string) *httptest.Server {
	t.Helper()
	var mu sync.Mutex
	tokenCalls := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/device" && r.Method == http.MethodPost:
			if err := r.ParseForm(); err != nil {
				t.Errorf("parse device form: %v", err)
			}
			if r.Form.Get("client_id") != "test-client" {
				t.Errorf("client_id = %q", r.Form.Get("client_id"))
			}
			if r.Form.Get("scopes") != twitchFollowScope {
				t.Errorf("scopes = %q", r.Form.Get("scopes"))
			}
			io.WriteString(w, `{"device_code":"device-code","expires_in":1800,"interval":1,"user_code":"ABCD1234","verification_uri":"https://www.twitch.tv/activate?public=true&device-code=ABCD1234"}`)
		case r.URL.Path == "/oauth2/token" && r.Method == http.MethodPost:
			if err := r.ParseForm(); err != nil {
				t.Errorf("parse token form: %v", err)
			}
			if r.Form.Get("grant_type") != "urn:ietf:params:oauth:grant-type:device_code" {
				t.Errorf("grant_type = %q", r.Form.Get("grant_type"))
			}
			mu.Lock()
			tokenCalls++
			call := tokenCalls
			mu.Unlock()
			switch mode {
			case "expired":
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, `{"status":400,"message":"expired token"}`)
				return
			case "denied":
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, `{"status":400,"message":"access_denied"}`)
				return
			case "fatal":
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, `{"status":400,"message":"invalid device code"}`)
				return
			}
			if call == 1 {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, `{"status":400,"message":"authorization_pending"}`)
				return
			}
			io.WriteString(w, `{"access_token":"secret-token","expires_in":14400,"refresh_token":"refresh-secret","scope":["user:read:follows"],"token_type":"bearer"}`)
		case r.URL.Path == "/helix/users" && r.Method == http.MethodGet:
			if r.Header.Get("Authorization") != "Bearer secret-token" {
				t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
			}
			if r.Header.Get("Client-Id") != "test-client" {
				t.Errorf("Client-Id = %q", r.Header.Get("Client-Id"))
			}
			logins := r.URL.Query()["login"]
			if len(logins) == 0 {
				io.WriteString(w, `{"data":[{"id":"42","login":"me","display_name":"Me","profile_image_url":"https://example.com/me.png"}]}`)
				return
			}
			users := make([]map[string]string, 0, len(logins))
			for _, login := range logins {
				users = append(users, map[string]string{
					"id":                "id-" + login,
					"login":             login,
					"display_name":      strings.ToUpper(login[:1]) + login[1:],
					"profile_image_url": "https://example.com/" + login + ".png",
				})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": users})
		case r.URL.Path == "/helix/channels/followed" && r.Method == http.MethodGet:
			if r.URL.Query().Get("user_id") != "42" {
				t.Errorf("user_id = %q", r.URL.Query().Get("user_id"))
			}
			if r.URL.Query().Get("first") != "100" {
				t.Errorf("first = %q", r.URL.Query().Get("first"))
			}
			switch r.URL.Query().Get("after") {
			case "":
				io.WriteString(w, `{"data":[
					{"broadcaster_id":"1","broadcaster_login":"alpha","broadcaster_name":"Alpha","followed_at":"2026-01-01T00:00:00Z"},
					{"broadcaster_id":"2","broadcaster_login":"beta","broadcaster_name":"Beta","followed_at":"2026-01-02T00:00:00Z"}
				],"total":3,"pagination":{"cursor":"next"}}`)
			case "next":
				io.WriteString(w, `{"data":[
					{"broadcaster_id":"3","broadcaster_login":"gamma","broadcaster_name":"Gamma","followed_at":"2026-01-03T00:00:00Z"}
				],"total":3,"pagination":{}}`)
			default:
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, `{"status":400,"message":"invalid cursor"}`)
			}
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func forceTwitchPoll(t *testing.T, app *App, id string) twitchImportView {
	t.Helper()
	app.twitchMu.Lock()
	if session := app.twitchImports[id]; session != nil {
		session.NextPoll = time.Time{}
	}
	app.twitchMu.Unlock()
	view, err := app.PollTwitchImport(context.Background(), id)
	if err != nil {
		t.Fatalf("PollTwitchImport returned error: %v", err)
	}
	return view
}

func TestTwitchImportFlow(t *testing.T) {
	app := newTestApp(t)
	injectWebhook(t, app, "Discord", "https://discord.com/api/webhooks/1234567890/abcdefghij")
	app.mu.Lock()
	app.cfg.Channels = append(app.cfg.Channels, Channel{ID: "existing", Login: "alpha", DisplayName: "Alpha", Enabled: true})
	app.mu.Unlock()

	server := newTwitchImportServer(t, "ok")
	setTwitchImportTestURLs(t, server)
	t.Setenv(twitchOAuthClientIDEnv, "test-client")

	start, err := app.StartTwitchImport(context.Background())
	if err != nil {
		t.Fatalf("StartTwitchImport returned error: %v", err)
	}
	if start.Status != twitchImportPending {
		t.Fatalf("status = %q, want pending", start.Status)
	}
	if start.UserCode != "ABCD1234" {
		t.Errorf("user code = %q", start.UserCode)
	}
	if start.Interval != 1 {
		t.Errorf("interval = %d", start.Interval)
	}
	id := start.ID

	if view := forceTwitchPoll(t, app, id); view.Status != twitchImportPending {
		t.Fatalf("first poll status = %q, want pending", view.Status)
	}

	view := forceTwitchPoll(t, app, id)
	if view.Status != twitchImportReady {
		t.Fatalf("second poll status = %q (%s), want ready", view.Status, view.Message)
	}
	if view.Account == nil || view.Account.Login != "me" || view.Account.DisplayName != "Me" {
		t.Fatalf("account = %+v", view.Account)
	}
	if view.Total != 3 || len(view.Follows) != 3 {
		t.Fatalf("total = %d, follows = %d", view.Total, len(view.Follows))
	}
	added := map[string]bool{}
	for _, follow := range view.Follows {
		added[follow.Login] = follow.Added
		if follow.Login != "alpha" && follow.AvatarURL == "" {
			t.Errorf("%s has no avatar", follow.Login)
		}
	}
	if !added["alpha"] {
		t.Error("existing channel should be marked as added")
	}
	if added["beta"] || added["gamma"] {
		t.Error("new follows should not be marked as added")
	}

	result, err := app.ImportTwitchFollows(id, []string{"alpha", "beta", "gamma"})
	if err != nil {
		t.Fatalf("ImportTwitchFollows returned error: %v", err)
	}
	if result.Added != 2 {
		t.Errorf("added = %d, want 2", result.Added)
	}
	if len(result.Skipped) != 1 || result.Skipped[0].Login != "alpha" {
		t.Errorf("skipped = %+v", result.Skipped)
	}

	app.mu.RLock()
	channels := append([]Channel(nil), app.cfg.Channels...)
	app.mu.RUnlock()
	if len(channels) != 3 {
		t.Fatalf("channels = %d, want 3", len(channels))
	}
	var beta Channel
	for _, channel := range channels {
		if channel.Login == "beta" {
			beta = channel
		}
	}
	if beta.DisplayName != "Beta" || beta.AvatarURL != "https://example.com/beta.png" {
		t.Errorf("beta = %+v", beta)
	}
	if len(beta.Destinations) != 1 {
		t.Errorf("beta destinations = %+v", beta.Destinations)
	}

	data, err := os.ReadFile(app.configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	for _, secret := range []string{"secret-token", "refresh-secret", "device-code"} {
		if strings.Contains(string(data), secret) {
			t.Errorf("config contains %q", secret)
		}
	}

	app.twitchMu.Lock()
	_, exists := app.twitchImports[id]
	app.twitchMu.Unlock()
	if exists {
		t.Error("import session should be removed after import")
	}
}

func TestTwitchImportSignInStates(t *testing.T) {
	cases := []struct {
		mode string
		want string
	}{
		{"expired", twitchImportExpired},
		{"denied", twitchImportDenied},
		{"fatal", twitchImportError},
	}
	for _, tc := range cases {
		t.Run(tc.mode, func(t *testing.T) {
			app := newTestApp(t)
			server := newTwitchImportServer(t, tc.mode)
			setTwitchImportTestURLs(t, server)
			t.Setenv(twitchOAuthClientIDEnv, "test-client")

			start, err := app.StartTwitchImport(context.Background())
			if err != nil {
				t.Fatalf("StartTwitchImport returned error: %v", err)
			}
			view := forceTwitchPoll(t, app, start.ID)
			if view.Status != tc.want {
				t.Fatalf("status = %q, want %q (%s)", view.Status, tc.want, view.Message)
			}
			if view.Follows != nil {
				t.Errorf("failed sign-in should not return follows")
			}
		})
	}
}

func TestTwitchImportRequiresClientID(t *testing.T) {
	t.Setenv(twitchOAuthClientIDEnv, "")
	app := newTestApp(t)
	if app.State().TwitchConnectEnabled {
		t.Error("connect should be disabled without a client ID")
	}
	if _, err := app.StartTwitchImport(context.Background()); err == nil {
		t.Error("StartTwitchImport should fail without a client ID")
	}
}

func newBulkTwitchServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		var request struct {
			Query     string            `json:"query"`
			Variables map[string]string `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
		}
		buildUser := func(login string) any {
			if login == "missing" {
				return nil
			}
			return map[string]any{
				"login":           login,
				"displayName":     strings.ToUpper(login),
				"profileImageURL": "https://example.com/" + login + ".png",
				"bannerImageURL":  "",
				"stream":          nil,
				"lastBroadcast":   nil,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if logins := twitchRequestLogins(request.Query); len(logins) > 0 {
			data := make(map[string]any, len(logins))
			for index, login := range logins {
				data["u"+strconv.Itoa(index)] = buildUser(login)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
			return
		}
		payload := map[string]any{
			"data": map[string]any{
				"user": buildUser(request.Variables["login"]),
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	t.Cleanup(server.Close)
	return server
}

func TestAddChannelsBulk(t *testing.T) {
	app := newTestApp(t)
	injectWebhook(t, app, "Discord", "https://discord.com/api/webhooks/1234567890/abcdefghij")
	app.mu.Lock()
	app.cfg.Channels = append(app.cfg.Channels, Channel{ID: "existing", Login: "existing", DisplayName: "Existing", Enabled: true})
	app.mu.Unlock()

	server := newBulkTwitchServer(t)
	setTwitchTestURL(t, server)

	result, err := app.AddChannelsBulk(context.Background(), []string{
		"newone\nhttps://twitch.tv/another, NEWONE; existing; missing",
	})
	if err != nil {
		t.Fatalf("AddChannelsBulk returned error: %v", err)
	}
	if result.Added != 2 {
		t.Errorf("added = %d, want 2", result.Added)
	}
	if len(result.Skipped) != 1 || result.Skipped[0].Login != "existing" {
		t.Errorf("skipped = %+v", result.Skipped)
	}
	if len(result.Failed) != 1 || result.Failed[0].Login != "missing" {
		t.Errorf("failed = %+v", result.Failed)
	}

	loaded, err := loadConfig(app.configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(loaded.Channels) != 3 {
		t.Fatalf("persisted channels = %d, want 3", len(loaded.Channels))
	}
	if loaded.Channels[1].Login != "newone" || loaded.Channels[2].Login != "another" {
		t.Errorf("order = %q, %q", loaded.Channels[1].Login, loaded.Channels[2].Login)
	}
	if loaded.Channels[1].DisplayName != "NEWONE" {
		t.Errorf("display name = %q", loaded.Channels[1].DisplayName)
	}
	if len(loaded.Channels[1].Destinations) != 1 {
		t.Errorf("destinations = %+v", loaded.Channels[1].Destinations)
	}
}

func TestAddChannelsBulkLimit(t *testing.T) {
	app := newTestApp(t)
	logins := make([]string, 0, bulkImportMaxLogins+1)
	for index := 0; index <= bulkImportMaxLogins; index++ {
		logins = append(logins, fmt.Sprintf("user%d", index))
	}
	if _, err := app.AddChannelsBulk(context.Background(), logins); err == nil {
		t.Error("AddChannelsBulk should reject lists over the limit")
	}
}

func TestAddChannelsBulkInvalidInput(t *testing.T) {
	app := newTestApp(t)
	if _, err := app.AddChannelsBulk(context.Background(), []string{"not a channel!"}); err == nil {
		t.Error("AddChannelsBulk should reject invalid usernames")
	}
}

func TestImportChannelsParsesHTMLAndURLs(t *testing.T) {
	app := newTestApp(t)
	injectWebhook(t, app, "Discord", "https://discord.com/api/webhooks/1234567890/abcdefghij")
	server := newBulkTwitchServer(t)
	setTwitchTestURL(t, server)

	savedPage := `<html><body>
		<a href="/alpha">Alpha</a>
		<a href='/directory'>Directory</a>
		<a href="https://www.twitch.tv/Beta?filter=all">Beta</a>
		<a href="https://example.com/gamma">Not Twitch</a>
		<span>twitch.tv/delta</span>
		<a href="/alpha/">Duplicate</a>
	</body></html>`

	result, err := app.ImportChannels(context.Background(), []string{savedPage}, nil)
	if err != nil {
		t.Fatalf("ImportChannels returned error: %v", err)
	}
	if result.Added != 3 || result.Assigned != 3 || result.Total != 3 {
		t.Fatalf("result = %+v", result)
	}
	cfg := app.ConfigSnapshot()
	if len(cfg.Channels) != 3 {
		t.Fatalf("channels = %d, want 3", len(cfg.Channels))
	}
	if cfg.Channels[0].Login != "alpha" || cfg.Channels[1].Login != "beta" || cfg.Channels[2].Login != "delta" {
		t.Fatalf("order = %q, %q, %q", cfg.Channels[0].Login, cfg.Channels[1].Login, cfg.Channels[2].Login)
	}
	imported := 0
	for _, activity := range app.State().Activity {
		if strings.Contains(activity.Message, "Imported") {
			imported++
		}
	}
	if imported != 1 {
		t.Fatalf("import activities = %d, want 1", imported)
	}
}

func TestImportChannelsParsesJSON(t *testing.T) {
	app := newTestApp(t)
	server := newBulkTwitchServer(t)
	setTwitchTestURL(t, server)

	payloads := []string{
		`[{"login":"alpha"},{"channel_login":"beta"},"gamma"]`,
		`{"follows":[{"login":"delta"},{"display_name":"Epsilon","login":"epsilon"},{"login":"zeta","display_name":"ZetaTV"}]}`,
	}
	for _, payload := range payloads {
		if _, err := app.ImportChannels(context.Background(), []string{payload}, nil); err != nil {
			t.Fatalf("ImportChannels(%s) returned error: %v", payload, err)
		}
	}
	cfg := app.ConfigSnapshot()
	if len(cfg.Channels) != 6 {
		t.Fatalf("channels = %d, want 6: %+v", len(cfg.Channels), cfg.Channels)
	}
	want := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta"}
	for index, login := range want {
		if cfg.Channels[index].Login != login {
			t.Fatalf("channel %d = %q, want %q", index, cfg.Channels[index].Login, login)
		}
	}
}

func TestImportChannelsAssignsSelectedDestinations(t *testing.T) {
	app := newTestApp(t)
	injectWebhook(t, app, "First", "https://discord.com/api/webhooks/1234567890/abcdefghij")
	second := injectWebhook(t, app, "Second", "https://discord.com/api/webhooks/9876543210/abcdefghij")
	server := newBulkTwitchServer(t)
	setTwitchTestURL(t, server)

	result, err := app.ImportChannels(context.Background(), []string{"alpha, beta"}, []string{second.ID, second.ID})
	if err != nil {
		t.Fatalf("ImportChannels returned error: %v", err)
	}
	if result.Added != 2 || result.Assigned != 2 || result.Total != 2 {
		t.Fatalf("result = %+v", result)
	}
	for _, channel := range app.ConfigSnapshot().Channels {
		if len(channel.Destinations) != 1 || channel.Destinations[0].WebhookID != second.ID {
			t.Fatalf("%s destinations = %+v", channel.Login, channel.Destinations)
		}
	}
}

func TestImportChannelsRejectsUnknownDestination(t *testing.T) {
	app := newTestApp(t)
	injectWebhook(t, app, "First", "https://discord.com/api/webhooks/1234567890/abcdefghij")
	server := newBulkTwitchServer(t)
	setTwitchTestURL(t, server)

	if _, err := app.ImportChannels(context.Background(), []string{"alpha"}, []string{"missing"}); err == nil {
		t.Fatal("ImportChannels accepted an unknown destination")
	}
	if channels := app.ConfigSnapshot().Channels; len(channels) != 0 {
		t.Fatalf("channels = %d, want 0", len(channels))
	}
}

func TestImportChannelsRejectsHugeText(t *testing.T) {
	app := newTestApp(t)
	huge := strings.Repeat("a", bulkImportMaxTextBytes+1)
	if _, err := app.ImportChannels(context.Background(), []string{huge}, nil); err == nil {
		t.Fatal("ImportChannels accepted text over 2 MB")
	}
}

func TestBulkAddEndpoint(t *testing.T) {
	app := newTestApp(t)
	injectWebhook(t, app, "Discord", "https://discord.com/api/webhooks/1234567890/abcdefghij")
	server := newBulkTwitchServer(t)
	setTwitchTestURL(t, server)

	dashboard := httptest.NewServer(withSecurity(newDashboardHandler(app, func() {})))
	defer dashboard.Close()

	request, _ := http.NewRequest(http.MethodPost, dashboard.URL+"/api/channels/bulk", strings.NewReader(`{"text":"newone, missing"}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	var payload struct {
		Result bulkAddResponse `json:"result"`
		State  State           `json:"state"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Result.Added != 1 {
		t.Errorf("added = %d, want 1", payload.Result.Added)
	}
	if len(payload.Result.Failed) != 1 || payload.Result.Failed[0].Login != "missing" {
		t.Errorf("failed = %+v", payload.Result.Failed)
	}
	if len(payload.State.Config.Channels) != 1 {
		t.Errorf("state channels = %d, want 1", len(payload.State.Config.Channels))
	}
}

func TestTwitchImportEndpointErrors(t *testing.T) {
	app := newTestApp(t)
	dashboard := httptest.NewServer(withSecurity(newDashboardHandler(app, func() {})))
	defer dashboard.Close()

	t.Setenv(twitchOAuthClientIDEnv, "")
	request, _ := http.NewRequest(http.MethodPost, dashboard.URL+"/api/twitch/imports", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("disabled import status = %d, want 400", response.StatusCode)
	}

	t.Setenv(twitchOAuthClientIDEnv, "test-client")
	response, err = http.Get(dashboard.URL + "/api/twitch/imports/unknown")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("unknown session status = %d, want 400", response.StatusCode)
	}
}

func TestImportChannelsSkipsReservedLogins(t *testing.T) {
	app := newTestApp(t)
	server := newBulkTwitchServer(t)
	setTwitchTestURL(t, server)

	result, err := app.ImportChannels(context.Background(), []string{"directory\nlogin\nalpha"}, nil)
	if err != nil {
		t.Fatalf("ImportChannels returned error: %v", err)
	}
	if result.Added != 1 || len(result.Failed) != 2 {
		t.Fatalf("result = %+v", result)
	}
	cfg := app.ConfigSnapshot()
	if len(cfg.Channels) != 1 || cfg.Channels[0].Login != "alpha" {
		t.Fatalf("channels = %+v", cfg.Channels)
	}
}
