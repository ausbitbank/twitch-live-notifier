package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Keep the test suite deterministic and avoid D-Bus/Keychain prompts on
	// developer machines. The fallback key-file path is exercised instead.
	_ = os.Setenv("TWITCH_LIVE_DISABLE_KEYRING", "1")
	os.Exit(m.Run())
}

func TestSecretStoreRoundTripAndCiphertext(t *testing.T) {
	dir := t.TempDir()
	store, err := openSecretStore(dir)
	if err != nil {
		t.Fatalf("openSecretStore returned error: %v", err)
	}
	const plaintext = "https://discord.com/api/webhooks/1234567890/supersecrettoken"
	if err := store.Set("url-1", plaintext); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if err := store.Save(); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, secretStoreFile))
	if err != nil {
		t.Fatalf("read secrets file: %v", err)
	}
	if strings.Contains(string(raw), "supersecrettoken") {
		t.Fatal("secrets file contains the plaintext webhook token")
	}

	reopened, err := openSecretStore(dir)
	if err != nil {
		t.Fatalf("reopen secret store: %v", err)
	}
	got, err := reopened.Get("url-1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != plaintext {
		t.Fatalf("Get = %q", got)
	}
	if reopened.Description() != "aes-gcm" {
		t.Fatalf("protector = %q", reopened.Description())
	}
}

func TestConfigEncryptsWebhookURLAtRest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	const secret = "abcdefghijklmnopqrstuvwxyz"
	cfg := defaultConfig()
	cfg.Webhooks = []Webhook{{
		ID:       "webhook1",
		Label:    "Community",
		URL:      "https://discord.com/api/webhooks/1234567890/" + secret,
		Enabled:  true,
		Template: defaultTemplate(),
	}}
	cfg.Channels = []Channel{{
		ID: "channel1", Login: "teststreamer", DisplayName: "TestStreamer", Enabled: true,
		Destinations: []Destination{{
			WebhookID: "webhook1",
			LiveMessage: &LiveMessage{
				MessageID:  "message1",
				WebhookURL: "https://discord.com/api/webhooks/1234567890/" + secret,
			},
		}},
	}}
	if err := saveConfig(path, cfg); err != nil {
		t.Fatalf("saveConfig returned error: %v", err)
	}

	rawConfig, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(rawConfig), secret) {
		t.Fatal("config file contains the plaintext webhook token")
	}
	rawSecrets, err := os.ReadFile(filepath.Join(dir, secretStoreFile))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(rawSecrets), secret) {
		t.Fatal("secrets file contains the plaintext webhook token")
	}

	loaded, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if loaded.Webhooks[0].URL != "https://discord.com/api/webhooks/1234567890/"+secret {
		t.Fatalf("webhook URL = %q", loaded.Webhooks[0].URL)
	}
	record := loaded.Channels[0].Destinations[0].LiveMessage
	if record == nil || record.WebhookURL != "https://discord.com/api/webhooks/1234567890/"+secret {
		t.Fatalf("live message = %+v", record)
	}
}

func TestExportConfigRedactsSecretsByDefault(t *testing.T) {
	app := newTestApp(t)
	const secret = "abcdefghijklmnopqrstuvwxyz"
	webhook := injectWebhook(t, app, "Community", "https://discord.com/api/webhooks/1234567890/"+secret)

	redacted, err := app.ExportConfigJSON(false)
	if err != nil {
		t.Fatalf("ExportConfigJSON returned error: %v", err)
	}
	if strings.Contains(string(redacted), secret) {
		t.Fatal("redacted export contains the webhook token")
	}
	if !strings.Contains(string(redacted), `"redacted": true`) {
		t.Fatalf("redacted export = %s", redacted)
	}
	if !strings.Contains(string(redacted), `"url_redacted": true`) {
		t.Fatalf("redacted export missing webhook marker: %s", redacted)
	}

	withSecrets, err := app.ExportConfigJSON(true)
	if err != nil {
		t.Fatalf("ExportConfigJSON(includeSecrets) returned error: %v", err)
	}
	if !strings.Contains(string(withSecrets), secret) {
		t.Fatal("full export is missing the webhook URL")
	}
	if webhook.ID == "" {
		t.Fatal("test webhook was not created")
	}
}

func TestConfigExportAPI(t *testing.T) {
	app := newTestApp(t)
	const secret = "abcdefghijklmnopqrstuvwxyz"
	injectWebhook(t, app, "Community", "https://discord.com/api/webhooks/1234567890/"+secret)
	server := httptest.NewServer(withSecurity(withDashboardAuth(app, newDashboardHandler(app, func() {}))))
	defer server.Close()

	response, err := http.Get(server.URL + "/api/config/export")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("export status = %d", response.StatusCode)
	}
	if strings.Contains(string(body), secret) {
		t.Fatal("export endpoint leaked the webhook token")
	}
	if !strings.Contains(response.Header.Get("Content-Disposition"), "attachment") {
		t.Fatalf("content disposition = %q", response.Header.Get("Content-Disposition"))
	}

	if err := app.SetDashboardToken("export-token"); err != nil {
		t.Fatalf("SetDashboardToken returned error: %v", err)
	}
	response, err = http.Get(server.URL + "/api/config/export")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("export without token = %d", response.StatusCode)
	}
	response, err = http.Get(server.URL + "/api/config/export?token=export-token")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("export with query token = %d", response.StatusCode)
	}
}

func TestImportConfigRestoresFullBackup(t *testing.T) {
	source := newTestApp(t)
	const secret = "abcdefghijklmnopqrstuvwxyz"
	injectWebhook(t, source, "Community", "https://discord.com/api/webhooks/1234567890/"+secret)
	backup, err := source.ExportConfigJSON(true)
	if err != nil {
		t.Fatalf("ExportConfigJSON returned error: %v", err)
	}
	target := newTestApp(t)
	if err := target.ImportConfig(backup); err != nil {
		t.Fatalf("ImportConfig returned error: %v", err)
	}
	cfg := target.ConfigSnapshot()
	if len(cfg.Webhooks) != 1 || cfg.Webhooks[0].URL != "https://discord.com/api/webhooks/1234567890/"+secret {
		t.Fatalf("imported webhook = %+v", cfg.Webhooks)
	}
	if cfg.Webhooks[0].Invalid {
		t.Fatal("full backup import marked the webhook invalid")
	}
}

func TestImportConfigRestoresRedactedBackup(t *testing.T) {
	source := newTestApp(t)
	injectWebhook(t, source, "Community", "https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz")
	source.mu.Lock()
	source.cfg.Channels = []Channel{{
		ID: "channel1", Login: "teststreamer", DisplayName: "TestStreamer", Enabled: true,
		Destinations: []Destination{{WebhookID: source.cfg.Webhooks[0].ID}},
	}}
	_ = source.persistLocked()
	source.mu.Unlock()

	backup, err := source.ExportConfigJSON(false)
	if err != nil {
		t.Fatalf("ExportConfigJSON returned error: %v", err)
	}

	target := newTestApp(t)
	if err := target.ImportConfig(backup); err != nil {
		t.Fatalf("ImportConfig returned error: %v", err)
	}
	cfg := target.ConfigSnapshot()
	if len(cfg.Channels) != 1 || len(cfg.Webhooks) != 1 {
		t.Fatalf("imported config = %+v", cfg)
	}
	if !cfg.Webhooks[0].Invalid || cfg.Webhooks[0].URL != "" {
		t.Fatalf("redacted webhook = %+v", cfg.Webhooks[0])
	}
	if cfg.Channels[0].Destinations[0].LastStreamID != "" {
		t.Fatalf("unexpected dedupe state: %+v", cfg.Channels[0].Destinations[0])
	}
}
