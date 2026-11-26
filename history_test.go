package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestDeliveryLogPersistsAndReloads(t *testing.T) {
	dir := t.TempDir()
	log := newDeliveryLog(dir)
	if log == nil {
		t.Fatal("newDeliveryLog returned nil")
	}
	record := DeliveryRecord{
		ID:           "rec-1",
		Time:         "2026-01-02T03:04:05Z",
		Kind:         deliveryKindPosted,
		Level:        deliveryLevelSuccess,
		ChannelID:    "channel-1",
		ChannelLogin: "alpha",
		ChannelName:  "Alpha",
		WebhookID:    "webhook-1",
		WebhookLabel: "#live",
		StreamID:     "stream-1",
		Title:        "Speedrun night",
		Game:         "Celeste",
		MessageID:    "345678901234567890",
		MessageURL:   "https://discord.com/channels/123456789012345678/234567890123456789/345678901234567890",
	}
	if err := log.Add(record); err != nil {
		t.Fatalf("Add: %v", err)
	}

	path := filepath.Join(dir, deliveryHistoryFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read history file: %v", err)
	}
	var file deliveryHistoryFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("history file is not valid JSON: %v", err)
	}
	if len(file.Entries) != 1 || file.Entries[0].ID != "rec-1" {
		t.Fatalf("persisted entries = %+v", file.Entries)
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat history file: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Fatalf("history file permissions = %o, want 600", perm)
		}
	}
	leftovers, _ := filepath.Glob(filepath.Join(dir, ".delivery-history-*.tmp"))
	if len(leftovers) != 0 {
		t.Fatalf("temporary files left behind: %v", leftovers)
	}

	reloaded := newDeliveryLog(dir)
	entries, total := reloaded.List(0, DeliveryFilters{})
	if total != 1 || len(entries) != 1 {
		t.Fatalf("reloaded total = %d, entries = %d", total, len(entries))
	}
	if entries[0].MessageURL != record.MessageURL || entries[0].Game != "Celeste" {
		t.Fatalf("reloaded record = %+v", entries[0])
	}
}

func TestDeliveryLogDefaultsAndNormalization(t *testing.T) {
	log := newDeliveryLogAtPath(filepath.Join(t.TempDir(), deliveryHistoryFilename))
	fixed := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)
	log.now = func() time.Time { return fixed }

	if err := log.Add(DeliveryRecord{
		Kind:  " POSTED ",
		Level: " SUCCESS ",
		Time:  "2026-02-03T05:05:06+01:00",
	}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	entries, _ := log.List(1, DeliveryFilters{})
	if len(entries) != 1 {
		t.Fatalf("entries = %d", len(entries))
	}
	record := entries[0]
	if record.ID == "" {
		t.Error("generated ID is empty")
	}
	if record.Kind != deliveryKindPosted || record.Level != deliveryLevelSuccess {
		t.Errorf("kind/level = %q/%q", record.Kind, record.Level)
	}
	if record.Time != "2026-02-03T04:05:06Z" {
		t.Errorf("time = %q, want normalized UTC", record.Time)
	}

	if err := log.Add(DeliveryRecord{Kind: deliveryKindError}); err != nil {
		t.Fatalf("Add without time: %v", err)
	}
	entries, _ = log.List(1, DeliveryFilters{})
	if entries[0].Time != "2026-02-03T04:05:06Z" {
		t.Errorf("default time = %q", entries[0].Time)
	}
}

func TestDeliveryLogCapKeepsNewest(t *testing.T) {
	dir := t.TempDir()
	log := newDeliveryLog(dir)
	log.max = 3
	for i := 0; i < 5; i++ {
		if err := log.Add(DeliveryRecord{ID: string(rune('a' + i)), Kind: deliveryKindPosted}); err != nil {
			t.Fatalf("Add %d: %v", i, err)
		}
	}
	entries, total := log.List(0, DeliveryFilters{})
	if total != 3 || len(entries) != 3 {
		t.Fatalf("total = %d, len = %d", total, len(entries))
	}
	want := []string{"e", "d", "c"}
	for i, id := range want {
		if entries[i].ID != id {
			t.Fatalf("entry %d = %q, want %q", i, entries[i].ID, id)
		}
	}

	reloaded := newDeliveryLog(dir)
	reloadedEntries, _ := reloaded.List(0, DeliveryFilters{})
	if len(reloadedEntries) != 3 || reloadedEntries[0].ID != "e" {
		t.Fatalf("reloaded entries = %+v", reloadedEntries)
	}
}

func TestDeliveryLogListsNewestFirstWithLimit(t *testing.T) {
	log := newDeliveryLogAtPath(filepath.Join(t.TempDir(), deliveryHistoryFilename))
	for _, id := range []string{"one", "two", "three"} {
		if err := log.Add(DeliveryRecord{ID: id, Kind: deliveryKindPosted}); err != nil {
			t.Fatalf("Add %s: %v", id, err)
		}
	}
	entries, total := log.List(2, DeliveryFilters{})
	if total != 3 || len(entries) != 2 {
		t.Fatalf("total = %d, len = %d", total, len(entries))
	}
	if entries[0].ID != "three" || entries[1].ID != "two" {
		t.Fatalf("entries = %+v", entries)
	}
}

func TestDeliveryLogFilters(t *testing.T) {
	log := newDeliveryLogAtPath(filepath.Join(t.TempDir(), deliveryHistoryFilename))
	records := []DeliveryRecord{
		{ID: "1", Kind: deliveryKindPosted, Level: deliveryLevelSuccess, ChannelID: "a", ChannelLogin: "alpha", ChannelName: "Alpha", WebhookLabel: "Community", Title: "Speedrun night", Game: "Celeste"},
		{ID: "2", Kind: deliveryKindSkipped, Level: deliveryLevelInfo, ChannelID: "b", ChannelLogin: "beta", ChannelName: "Beta", Title: "Chill stream", Reason: "Quiet hours"},
		{ID: "3", Kind: deliveryKindError, Level: deliveryLevelError, ChannelID: "a", ChannelLogin: "alpha", ChannelName: "Alpha", Reason: "Discord returned HTTP 500"},
		{ID: "4", Kind: deliveryKindEdited, Level: deliveryLevelSuccess, ChannelID: "b", ChannelLogin: "beta", ChannelName: "Beta", Title: "Chill stream"},
	}
	for _, record := range records {
		if err := log.Add(record); err != nil {
			t.Fatalf("Add %s: %v", record.ID, err)
		}
	}

	tests := []struct {
		name    string
		filters DeliveryFilters
		want    int
	}{
		{"kind", DeliveryFilters{Kind: deliveryKindPosted}, 1},
		{"kind case-insensitive", DeliveryFilters{Kind: "POSTED"}, 1},
		{"level", DeliveryFilters{Level: deliveryLevelError}, 1},
		{"channel", DeliveryFilters{ChannelID: "a"}, 2},
		{"query login", DeliveryFilters{Query: "alpha"}, 2},
		{"query title", DeliveryFilters{Query: "chill"}, 2},
		{"query reason", DeliveryFilters{Query: "quiet"}, 1},
		{"query destination", DeliveryFilters{Query: "community"}, 1},
		{"query game", DeliveryFilters{Query: "celeste"}, 1},
		{"query is case-insensitive", DeliveryFilters{Query: "SPEEDRUN"}, 1},
		{"combined", DeliveryFilters{Kind: deliveryKindSkipped, ChannelID: "b"}, 1},
		{"no match", DeliveryFilters{Query: "missing"}, 0},
	}
	for _, test := range tests {
		entries, total := log.List(0, test.filters)
		if total != test.want || len(entries) != test.want {
			t.Errorf("%s: total = %d, len = %d, want %d", test.name, total, len(entries), test.want)
		}
	}

	entries, total := log.List(2, DeliveryFilters{ChannelID: "a"})
	if total != 2 || len(entries) != 2 {
		t.Fatalf("limited filtered total = %d, len = %d", total, len(entries))
	}
	if entries[0].ID != "3" || entries[1].ID != "1" {
		t.Fatalf("filtered order = %+v", entries)
	}
}

func TestDeliveryLogClear(t *testing.T) {
	dir := t.TempDir()
	log := newDeliveryLog(dir)
	if err := log.Add(DeliveryRecord{Kind: deliveryKindPosted}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := log.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	entries, total := log.List(0, DeliveryFilters{})
	if total != 0 || len(entries) != 0 {
		t.Fatalf("after clear total = %d, len = %d", total, len(entries))
	}
	reloaded := newDeliveryLog(dir)
	reloadedEntries, _ := reloaded.List(0, DeliveryFilters{})
	if len(reloadedEntries) != 0 {
		t.Fatalf("reloaded entries after clear = %+v", reloadedEntries)
	}
}

func TestDeliveryLogConcurrentAdds(t *testing.T) {
	dir := t.TempDir()
	log := newDeliveryLog(dir)
	var wg sync.WaitGroup
	for worker := 0; worker < 20; worker++ {
		wg.Add(1)
		go func(target *deliveryLog) {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				if err := target.Add(DeliveryRecord{Kind: deliveryKindPosted, Level: deliveryLevelSuccess}); err != nil {
					t.Errorf("concurrent Add: %v", err)
					return
				}
			}
		}(log)
	}
	wg.Wait()
	entries, total := log.List(0, DeliveryFilters{})
	if total != 100 || len(entries) != 100 {
		t.Fatalf("total = %d, len = %d", total, len(entries))
	}
	reloaded := newDeliveryLog(dir)
	reloadedEntries, _ := reloaded.List(0, DeliveryFilters{})
	if len(reloadedEntries) != 100 {
		t.Fatalf("reloaded entries = %d", len(reloadedEntries))
	}
}

func TestDeliveryLogToleratesCorruptFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, deliveryHistoryFilename)
	if err := os.WriteFile(path, []byte("{ this is not json"), 0o600); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}
	log := newDeliveryLog(dir)
	if log.loadErr == nil {
		t.Error("loadErr is nil for a corrupt file")
	}
	if entries, total := log.List(0, DeliveryFilters{}); total != 0 || len(entries) != 0 {
		t.Fatalf("corrupt file produced entries: %+v", entries)
	}
	if err := log.Add(DeliveryRecord{Kind: deliveryKindError}); err != nil {
		t.Fatalf("Add after corrupt load: %v", err)
	}
	reloaded := newDeliveryLog(dir)
	entries, total := reloaded.List(0, DeliveryFilters{})
	if total != 1 || len(entries) != 1 {
		t.Fatalf("reloaded total = %d, len = %d", total, len(entries))
	}
}

func TestDeliveryLogLoadsBareArray(t *testing.T) {
	dir := t.TempDir()
	data, err := json.Marshal([]DeliveryRecord{{ID: "legacy", Kind: deliveryKindPosted}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, deliveryHistoryFilename), data, 0o600); err != nil {
		t.Fatalf("write legacy file: %v", err)
	}
	log := newDeliveryLog(dir)
	entries, total := log.List(0, DeliveryFilters{})
	if total != 1 || len(entries) != 1 || entries[0].ID != "legacy" {
		t.Fatalf("legacy entries = %+v", entries)
	}
}

func TestDeliveryLogLoadsEmptyEnvelope(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, deliveryHistoryFilename), []byte(`{"version":1}`), 0o600); err != nil {
		t.Fatalf("write envelope: %v", err)
	}
	log := newDeliveryLog(dir)
	if log.loadErr != nil {
		t.Fatalf("loadErr = %v", log.loadErr)
	}
	if entries, total := log.List(0, DeliveryFilters{}); total != 0 || len(entries) != 0 {
		t.Fatalf("entries = %+v", entries)
	}
}

func TestRecordDeliveryWithoutHistoryIsSafe(t *testing.T) {
	app := &App{}
	app.recordDelivery(DeliveryRecord{Kind: deliveryKindPosted, Level: deliveryLevelSuccess})
	// The integration lane adds App.history; until then this must be a no-op
	// rather than a nil dereference.
	if log := app.deliveryHistoryLog(); log != nil {
		t.Fatalf("deliveryHistoryLog = %v, want nil without App.history", log)
	}
}

func TestDeliveryLogNilReceiver(t *testing.T) {
	var log *deliveryLog
	if err := log.Add(DeliveryRecord{Kind: deliveryKindPosted}); err != nil {
		t.Fatalf("nil Add: %v", err)
	}
	entries, total := log.List(10, DeliveryFilters{})
	if total != 0 || len(entries) != 0 {
		t.Fatalf("nil List = %d/%d", total, len(entries))
	}
	if err := log.Clear(); err != nil {
		t.Fatalf("nil Clear: %v", err)
	}
}
