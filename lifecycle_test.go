package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// lifecycleDiscord records posts, edits, deletes, and the query string used for
// each call so tests can prove that the original message is targeted.
type lifecycleDiscord struct {
	mu      sync.Mutex
	posts   []discordWebhookPayload
	edits   []discordWebhookPayload
	deletes []string
	paths   []string
	queries []string
}

func (c *lifecycleDiscord) addPost(payload discordWebhookPayload) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.posts = append(c.posts, payload)
}

func (c *lifecycleDiscord) addEdit(payload discordWebhookPayload) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.edits = append(c.edits, payload)
}

func (c *lifecycleDiscord) addDelete(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deletes = append(c.deletes, path)
}

func (c *lifecycleDiscord) record(path, query string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paths = append(c.paths, path)
	c.queries = append(c.queries, query)
}

func (c *lifecycleDiscord) counts() (posts, edits, deletes int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.posts), len(c.edits), len(c.deletes)
}

func newLifecycleDiscord(t *testing.T, deleteStatus int) (*httptest.Server, *lifecycleDiscord) {
	t.Helper()
	capture := &lifecycleDiscord{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capture.record(r.URL.Path, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var payload discordWebhookPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode Discord post: %v", err)
			}
			capture.addPost(payload)
			io.WriteString(w, `{"id":"discordmsg1"}`)
		case http.MethodPatch:
			var payload discordWebhookPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode Discord edit: %v", err)
			}
			capture.addEdit(payload)
			io.WriteString(w, `{"id":"discordmsg1"}`)
		case http.MethodDelete:
			capture.addDelete(r.URL.Path)
			if deleteStatus != 0 {
				w.WriteHeader(deleteStatus)
				io.WriteString(w, `{"message":"cleanup failed","code":500}`)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected Discord method %s", r.Method)
		}
	}))
	t.Cleanup(server.Close)
	return server, capture
}

type mutableTwitchServer struct {
	mu   sync.Mutex
	body string
}

func (m *mutableTwitchServer) set(body string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.body = body
}

func (m *mutableTwitchServer) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		body := m.body
		m.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, body)
	}
}

func TestLiveMessageIsEditedThenDeletedOnStreamEnd(t *testing.T) {
	twitchState := &mutableTwitchServer{body: liveTwitchJSON}
	twitch := httptest.NewServer(twitchState.handler(t))
	defer twitch.Close()
	setTwitchTestURL(t, twitch)

	discord, capture := newLifecycleDiscord(t, 0)
	app := newTestApp(t)
	app.mu.Lock()
	app.cfg.Webhooks = []Webhook{{
		ID:            "webhook1",
		Label:         "Community",
		URL:           discord.URL + "/api/webhooks/1/token",
		Enabled:       true,
		Template:      defaultTemplate(),
		UpdateSeconds: 60,
		Cleanup:       CleanupSettings{Mode: cleanupDelete},
	}}
	app.mu.Unlock()

	channel, err := app.AddChannel("teststreamer")
	if err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}

	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("first CheckNow returned error: %v", err)
	}
	posts, edits, deletes := capture.counts()
	if posts != 1 || edits != 0 || deletes != 0 {
		t.Fatalf("counts after post = %d/%d/%d", posts, edits, deletes)
	}
	cfg := app.ConfigSnapshot()
	record := cfg.Channels[0].Destinations[0].LiveMessage
	if record == nil || record.MessageID != "discordmsg1" || record.StreamID != "319913347671" {
		t.Fatalf("live message = %+v", record)
	}
	if record.WebhookURL != discord.URL+"/api/webhooks/1/token" {
		t.Fatalf("recorded webhook URL = %q", record.WebhookURL)
	}

	// Pretend the configured update interval has elapsed.
	app.mu.Lock()
	app.cfg.Channels[0].Destinations[0].LiveMessage.LastEditedAt = time.Now().Add(-2 * time.Minute).UTC().Format(time.RFC3339)
	app.mu.Unlock()

	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("edit CheckNow returned error: %v", err)
	}
	posts, edits, deletes = capture.counts()
	if posts != 1 || edits != 1 || deletes != 0 {
		t.Fatalf("counts after live update = %d/%d/%d", posts, edits, deletes)
	}
	if !strings.Contains(capture.paths[len(capture.paths)-1], "/messages/discordmsg1") {
		t.Fatalf("edit path = %q", capture.paths[len(capture.paths)-1])
	}

	// A second check immediately afterwards must respect the minimum interval.
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("interval CheckNow returned error: %v", err)
	}
	if _, edits, _ = capture.counts(); edits != 1 {
		t.Fatalf("edits after interval check = %d", edits)
	}

	twitchState.set(offlineTwitchJSON)
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("offline CheckNow returned error: %v", err)
	}
	if _, _, deletes = capture.counts(); deletes != 1 {
		t.Fatalf("deletes after stream end = %d", deletes)
	}
	if record := app.ConfigSnapshot().Channels[0].Destinations[0].LiveMessage; record != nil {
		t.Fatalf("record survived cleanup: %+v", record)
	}

	// No further cleanup calls once the record is gone.
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("second offline CheckNow returned error: %v", err)
	}
	if _, _, deletes = capture.counts(); deletes != 1 {
		t.Fatalf("deletes after second offline check = %d", deletes)
	}

	// The streamer still deduplicates the original stream.
	if cfg := app.ConfigSnapshot(); cfg.Channels[0].ID != channel.ID || cfg.Channels[0].Destinations[0].LastStreamID != "319913347671" {
		t.Fatalf("dedupe state = %+v", cfg.Channels[0].Destinations[0])
	}
}

func TestStreamEndEditUsesEndedTemplateAndVOD(t *testing.T) {
	twitchState := &mutableTwitchServer{body: liveTwitchJSON}
	twitch := httptest.NewServer(twitchState.handler(t))
	defer twitch.Close()
	setTwitchTestURL(t, twitch)

	discord, capture := newLifecycleDiscord(t, 0)
	app := newTestApp(t)
	ended := defaultEndedTemplate()
	ended.EmbedDescription = "Last playing **{game}**. VOD: {vod}"
	app.mu.Lock()
	app.cfg.Webhooks = []Webhook{{
		ID:       "webhook1",
		Label:    "Community",
		URL:      discord.URL + "/api/webhooks/1/token",
		Enabled:  true,
		Template: defaultTemplate(),
		Cleanup:  CleanupSettings{Mode: cleanupEdit, Template: &ended},
	}}
	app.mu.Unlock()

	if _, err := app.AddChannel("teststreamer"); err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("first CheckNow returned error: %v", err)
	}

	// The offline response includes an archive video id.
	twitchState.set(`{"data":{"user":{"login":"teststreamer","displayName":"TestStreamer","profileImageURL":"https://example.com/avatar.png","stream":null,"lastBroadcast":{"id":"316308419689","title":"Karaoke time","startedAt":"2026-11-14T17:56:38Z","game":{"name":"Music"}},"videos":{"edges":[{"node":{"id":"1234567890"}}]}}}}`)
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("offline CheckNow returned error: %v", err)
	}

	posts, edits, _ := capture.counts()
	if posts != 1 || edits != 1 {
		t.Fatalf("counts = %d/%d", posts, edits)
	}
	edited := capture.edits[0]
	if !strings.Contains(edited.Content, "stream has ended") {
		t.Fatalf("ended content = %q", edited.Content)
	}
	if len(edited.Embeds) != 1 || !strings.Contains(edited.Embeds[0].Description, "https://www.twitch.tv/videos/1234567890") {
		t.Fatalf("ended embed = %+v", edited.Embeds)
	}
	if record := app.ConfigSnapshot().Channels[0].Destinations[0].LiveMessage; record != nil {
		t.Fatalf("record survived edit cleanup: %+v", record)
	}
}

func TestCleanupFailureDoesNotBlockNextStream(t *testing.T) {
	twitchState := &mutableTwitchServer{body: liveTwitchJSON}
	twitch := httptest.NewServer(twitchState.handler(t))
	defer twitch.Close()
	setTwitchTestURL(t, twitch)

	// Every DELETE fails, which used to be the kind of error that could stop
	// later lifecycle processing.
	discord, capture := newLifecycleDiscord(t, http.StatusInternalServerError)
	app := newTestApp(t)
	app.mu.Lock()
	app.cfg.Webhooks = []Webhook{{
		ID:       "webhook1",
		Label:    "Community",
		URL:      discord.URL + "/api/webhooks/1/token",
		Enabled:  true,
		Template: defaultTemplate(),
		Cleanup:  CleanupSettings{Mode: cleanupDelete},
	}}
	app.mu.Unlock()
	if _, err := app.AddChannel("teststreamer"); err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}

	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("first CheckNow returned error: %v", err)
	}
	twitchState.set(offlineTwitchJSON)
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("offline CheckNow returned error: %v", err)
	}
	if record := app.ConfigSnapshot().Channels[0].Destinations[0].LiveMessage; record == nil {
		t.Fatal("transient cleanup failure lost the message record")
	}

	// A new stream must post even though the old message could not be removed.
	twitchState.set(strings.Replace(liveTwitchJSON, `"319913347671"`, `"419913347672"`, 1))
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("new stream CheckNow returned error: %v", err)
	}
	posts, _, deletes := capture.counts()
	if posts != 2 {
		t.Fatalf("posts after new stream = %d", posts)
	}
	if deletes < 2 {
		t.Fatalf("cleanup was not retried before the new stream: %d", deletes)
	}
	record := app.ConfigSnapshot().Channels[0].Destinations[0].LiveMessage
	if record == nil || record.StreamID != "419913347672" {
		t.Fatalf("new live message = %+v", record)
	}
}

func TestLiveEditsUseOriginalThreadAfterDestinationChange(t *testing.T) {
	twitchState := &mutableTwitchServer{body: liveTwitchJSON}
	twitch := httptest.NewServer(twitchState.handler(t))
	defer twitch.Close()
	setTwitchTestURL(t, twitch)

	const originalThread = "111111111111111111"
	const newThread = "222222222222222222"
	discord, capture := newLifecycleDiscord(t, 0)
	app := newTestApp(t)
	app.mu.Lock()
	app.cfg.Webhooks = []Webhook{{
		ID:            "webhook1",
		Label:         "Community",
		URL:           discord.URL + "/api/webhooks/1/token",
		Enabled:       true,
		Template:      defaultTemplate(),
		ThreadID:      originalThread,
		UpdateSeconds: 60,
	}}
	app.mu.Unlock()
	if _, err := app.AddChannel("teststreamer"); err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("first CheckNow returned error: %v", err)
	}
	app.mu.Lock()
	app.cfg.Webhooks[0].ThreadID = newThread
	app.cfg.Channels[0].Destinations[0].LiveMessage.LastEditedAt = time.Now().Add(-2 * time.Minute).UTC().Format(time.RFC3339)
	app.mu.Unlock()
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("edit CheckNow returned error: %v", err)
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if len(capture.queries) < 2 {
		t.Fatalf("queries = %#v", capture.queries)
	}
	lastQuery := capture.queries[len(capture.queries)-1]
	if !strings.Contains(lastQuery, "thread_id="+originalThread) || strings.Contains(lastQuery, newThread) {
		t.Fatalf("edit query = %q", lastQuery)
	}
}

func TestLiveUpdatesOffLeavesMessageUntouched(t *testing.T) {
	twitchState := &mutableTwitchServer{body: liveTwitchJSON}
	twitch := httptest.NewServer(twitchState.handler(t))
	defer twitch.Close()
	setTwitchTestURL(t, twitch)

	discord, capture := newLifecycleDiscord(t, 0)
	app := newTestApp(t)
	app.mu.Lock()
	app.cfg.Webhooks = []Webhook{{
		ID:       "webhook1",
		Label:    "Community",
		URL:      discord.URL + "/api/webhooks/1/token",
		Enabled:  true,
		Template: defaultTemplate(),
	}}
	app.mu.Unlock()
	if _, err := app.AddChannel("teststreamer"); err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("first CheckNow returned error: %v", err)
	}
	app.mu.Lock()
	app.cfg.Channels[0].Destinations[0].LiveMessage.LastEditedAt = time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	app.mu.Unlock()
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("second CheckNow returned error: %v", err)
	}
	posts, edits, _ := capture.counts()
	if posts != 1 || edits != 0 {
		t.Fatalf("updates off made edits: posts=%d edits=%d", posts, edits)
	}
}

func TestCleanupOffLeavesMessageUntouched(t *testing.T) {
	twitchState := &mutableTwitchServer{body: liveTwitchJSON}
	twitch := httptest.NewServer(twitchState.handler(t))
	defer twitch.Close()
	setTwitchTestURL(t, twitch)

	discord, capture := newLifecycleDiscord(t, 0)
	app := newTestApp(t)
	app.mu.Lock()
	app.cfg.Webhooks = []Webhook{{
		ID:       "webhook1",
		Label:    "Community",
		URL:      discord.URL + "/api/webhooks/1/token",
		Enabled:  true,
		Template: defaultTemplate(),
	}}
	app.mu.Unlock()
	if _, err := app.AddChannel("teststreamer"); err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("first CheckNow returned error: %v", err)
	}
	twitchState.set(offlineTwitchJSON)
	if _, err := app.CheckNow(context.Background()); err != nil {
		t.Fatalf("offline CheckNow returned error: %v", err)
	}
	_, edits, deletes := capture.counts()
	if edits != 0 || deletes != 0 {
		t.Fatalf("cleanup off made calls: edits=%d deletes=%d", edits, deletes)
	}
	if record := app.ConfigSnapshot().Channels[0].Destinations[0].LiveMessage; record != nil {
		t.Fatalf("cleanup off kept a record: %+v", record)
	}
}

func TestLiveMessageRecordPersistsAcrossRestart(t *testing.T) {
	twitch := newTwitchTestServer(t, liveTwitchJSON)
	setTwitchTestURL(t, twitch)
	discord, _ := newLifecycleDiscord(t, 0)
	path := t.TempDir() + "/config.json"

	first, err := NewApp(path)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	first.enableAutoCheck = false
	first.mu.Lock()
	first.cfg.Webhooks = []Webhook{{
		ID:       "webhook1",
		Label:    "Community",
		URL:      discord.URL + "/api/webhooks/1/token",
		Enabled:  true,
		Template: defaultTemplate(),
	}}
	first.mu.Unlock()
	if _, err := first.AddChannel("teststreamer"); err != nil {
		t.Fatalf("AddChannel returned error: %v", err)
	}
	if _, err := first.CheckNow(context.Background()); err != nil {
		t.Fatalf("CheckNow returned error: %v", err)
	}

	second, err := NewApp(path)
	if err != nil {
		t.Fatalf("second NewApp returned error: %v", err)
	}
	record := second.ConfigSnapshot().Channels[0].Destinations[0].LiveMessage
	if record == nil || record.MessageID != "discordmsg1" {
		t.Fatalf("persisted record = %+v", record)
	}
	if second.ConfigSnapshot().Channels[0].Destinations[0].LastStreamID != "319913347671" {
		t.Fatal("dedupe state did not survive restart")
	}
}
