package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// deliveryHistoryFilename is stored next to config.json in the data folder.
	deliveryHistoryFilename = "delivery-history.json"
	deliveryHistoryVersion  = 1

	// deliveryHistoryCap keeps the persisted file bounded no matter how long
	// the app runs. Older entries are dropped first.
	deliveryHistoryCap = 1000

	defaultHistoryListLimit = 200
	maxHistoryListLimit     = deliveryHistoryCap
)

// Delivery history kinds and levels mirror the values used by the dashboard.
const (
	deliveryKindPosted  = "posted"
	deliveryKindEdited  = "edited"
	deliveryKindCleaned = "cleaned"
	deliveryKindSkipped = "skipped"
	deliveryKindError   = "error"

	deliveryLevelInfo    = "info"
	deliveryLevelSuccess = "success"
	deliveryLevelError   = "error"
)

// DeliveryRecord is one auditable delivery outcome: a post, edit, cleanup,
// filter/rate-limit skip, or failure. It intentionally carries no webhook URL
// or token, only the non-secret destination label and internal IDs.
type DeliveryRecord struct {
	ID           string `json:"id"`
	Time         string `json:"time"`
	Kind         string `json:"kind"`
	Level        string `json:"level"`
	ChannelID    string `json:"channel_id,omitempty"`
	ChannelLogin string `json:"channel_login,omitempty"`
	ChannelName  string `json:"channel_name,omitempty"`
	WebhookID    string `json:"webhook_id,omitempty"`
	WebhookLabel string `json:"webhook_label,omitempty"`
	StreamID     string `json:"stream_id,omitempty"`
	Title        string `json:"title,omitempty"`
	Game         string `json:"game,omitempty"`
	MessageID    string `json:"message_id,omitempty"`
	MessageURL   string `json:"message_url,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

// DeliveryFilters narrows a history listing. Empty fields match everything.
type DeliveryFilters struct {
	Kind      string
	Level     string
	ChannelID string
	// Query is a case-insensitive substring match over the streamer login,
	// streamer display name, stream title, and skip/error reason.
	Query string
}

type deliveryHistoryFile struct {
	Version int              `json:"version"`
	Entries []DeliveryRecord `json:"entries"`
}

// deliveryLog is a small append-only ring buffer persisted as JSON. All
// methods are safe for concurrent use. Entries are kept oldest-first in
// memory; List returns newest-first for the dashboard.
type deliveryLog struct {
	mu      sync.RWMutex
	path    string
	entries []DeliveryRecord
	max     int
	now     func() time.Time

	// loadErr records a non-fatal problem while reading the file so callers
	// can surface a diagnostic without refusing to start.
	loadErr error
}

// newDeliveryLog creates the history log for a data/config directory. The
// history lane constructs it with filepath.Dir(configPath); a missing file is
// normal and an unreadable or malformed file never prevents startup.
func newDeliveryLog(configDir string) *deliveryLog {
	dir := strings.TrimSpace(configDir)
	if dir == "" {
		return &deliveryLog{max: deliveryHistoryCap, now: time.Now}
	}
	return newDeliveryLogAtPath(filepath.Join(dir, deliveryHistoryFilename))
}

// newDeliveryLogAtPath is the explicit-path constructor used by tests.
func newDeliveryLogAtPath(path string) *deliveryLog {
	log := &deliveryLog{
		path: strings.TrimSpace(path),
		max:  deliveryHistoryCap,
		now:  time.Now,
	}
	log.load()
	return log
}

func (l *deliveryLog) load() {
	if l == nil || strings.TrimSpace(l.path) == "" {
		return
	}
	data, err := os.ReadFile(l.path)
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		l.loadErr = err
		return
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return
	}

	var file deliveryHistoryFile
	entries := []DeliveryRecord(nil)
	if err := json.Unmarshal(data, &file); err == nil {
		entries = file.Entries
	} else {
		// Tolerate a bare JSON array written by an early build or a hand-edited
		// file. Anything still unparseable is ignored rather than fatal.
		if err := json.Unmarshal(data, &entries); err != nil {
			l.loadErr = errors.New("the delivery history file is not valid JSON")
			return
		}
	}

	max := l.maximum()
	if len(entries) > max {
		entries = entries[len(entries)-max:]
	}
	l.entries = entries
}

func (l *deliveryLog) maximum() int {
	if l == nil || l.max <= 0 {
		return deliveryHistoryCap
	}
	return l.max
}

// Add normalizes and appends one record, trims the oldest entries beyond the
// cap, and atomically persists the file with 0600 permissions.
func (l *deliveryLog) Add(record DeliveryRecord) error {
	if l == nil {
		return nil
	}
	record = l.normalize(record)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, record)
	if max := l.maximum(); len(l.entries) > max {
		l.entries = l.entries[len(l.entries)-max:]
	}
	return l.persistLocked()
}

func (l *deliveryLog) normalize(record DeliveryRecord) DeliveryRecord {
	record.ID = strings.TrimSpace(record.ID)
	if record.ID == "" {
		record.ID = newID()
	}
	record.Time = strings.TrimSpace(record.Time)
	if record.Time == "" {
		now := time.Now
		if l != nil && l.now != nil {
			now = l.now
		}
		record.Time = now().UTC().Format(time.RFC3339)
	} else if parsed, err := time.Parse(time.RFC3339, record.Time); err == nil {
		record.Time = parsed.UTC().Format(time.RFC3339)
	}
	record.Kind = strings.ToLower(strings.TrimSpace(record.Kind))
	record.Level = strings.ToLower(strings.TrimSpace(record.Level))
	record.ChannelID = strings.TrimSpace(record.ChannelID)
	record.ChannelLogin = strings.TrimSpace(record.ChannelLogin)
	record.ChannelName = strings.TrimSpace(record.ChannelName)
	record.WebhookID = strings.TrimSpace(record.WebhookID)
	record.WebhookLabel = strings.TrimSpace(record.WebhookLabel)
	record.StreamID = strings.TrimSpace(record.StreamID)
	record.Title = strings.TrimSpace(record.Title)
	record.Game = strings.TrimSpace(record.Game)
	record.MessageID = strings.TrimSpace(record.MessageID)
	record.MessageURL = strings.TrimSpace(record.MessageURL)
	record.Reason = strings.TrimSpace(record.Reason)
	return record
}

// List returns up to limit newest-first matching records plus the total number
// of matching records (ignoring the limit). A limit of zero returns every
// stored record, up to the cap.
func (l *deliveryLog) List(limit int, filters DeliveryFilters) ([]DeliveryRecord, int) {
	if l == nil {
		return []DeliveryRecord{}, 0
	}
	if limit <= 0 || limit > l.maximum() {
		limit = l.maximum()
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	matches := make([]DeliveryRecord, 0, len(l.entries))
	total := 0
	for i := len(l.entries) - 1; i >= 0; i-- {
		record := l.entries[i]
		if !record.matches(filters) {
			continue
		}
		total++
		if len(matches) < limit {
			matches = append(matches, record)
		}
	}
	if matches == nil {
		matches = []DeliveryRecord{}
	}
	return matches, total
}

func (r DeliveryRecord) matches(filters DeliveryFilters) bool {
	if kind := strings.ToLower(strings.TrimSpace(filters.Kind)); kind != "" && strings.ToLower(r.Kind) != kind {
		return false
	}
	if level := strings.ToLower(strings.TrimSpace(filters.Level)); level != "" && strings.ToLower(r.Level) != level {
		return false
	}
	if channelID := strings.TrimSpace(filters.ChannelID); channelID != "" && r.ChannelID != channelID {
		return false
	}
	if query := strings.ToLower(strings.TrimSpace(filters.Query)); query != "" {
		haystack := strings.ToLower(strings.Join([]string{
			r.ChannelLogin,
			r.ChannelName,
			r.WebhookLabel,
			r.Title,
			r.Game,
			r.Reason,
			r.Kind,
			r.Level,
		}, "\n"))
		if !strings.Contains(haystack, query) {
			return false
		}
	}
	return true
}

// Clear removes every record and persists the empty history.
func (l *deliveryLog) Clear() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = nil
	return l.persistLocked()
}

// persistLocked writes the current entries atomically. Callers must hold l.mu
// (read callers never persist).
func (l *deliveryLog) persistLocked() error {
	if strings.TrimSpace(l.path) == "" {
		return nil
	}
	file := deliveryHistoryFile{
		Version: deliveryHistoryVersion,
		Entries: append([]DeliveryRecord(nil), l.entries...),
	}
	if file.Entries == nil {
		file.Entries = []DeliveryRecord{}
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return errors.New("could not encode the delivery history")
	}
	data = append(data, '\n')

	dir := filepath.Dir(l.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return errors.New("could not create the delivery history folder")
	}
	tmp, err := os.CreateTemp(dir, ".delivery-history-*.tmp")
	if err != nil {
		return errors.New("could not write the delivery history file")
	}
	tmpName := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}
	if err := tmp.Chmod(0o600); err != nil {
		cleanup()
		return errors.New("could not secure the delivery history file")
	}
	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return errors.New("could not write the delivery history file")
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return errors.New("could not write the delivery history file")
	}
	if err := os.Rename(tmpName, l.path); err != nil {
		_ = os.Remove(tmpName)
		return errors.New("could not save the delivery history file")
	}
	return nil
}

// appHistoryResolver returns the delivery log attached to an App. Tests can
// replace the resolver to exercise the HTTP routes without a real App.
var appHistoryResolver = func(app *App) *deliveryLog {
	if app == nil {
		return nil
	}
	return app.history
}

// deliveryHistoryLog returns the app's history log, or nil when history is not
// configured.
func (a *App) deliveryHistoryLog() *deliveryLog {
	if a == nil {
		return nil
	}
	return a.history
}

// recordDelivery appends a record to the app's persisted delivery history. It
// is safe to call when history is not configured (for example in tests).
func (a *App) recordDelivery(record DeliveryRecord) {
	if log := appHistoryResolver(a); log != nil {
		_ = log.Add(record)
	}
}
