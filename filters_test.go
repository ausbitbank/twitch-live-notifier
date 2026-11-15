package main

import (
	"strings"
	"testing"
	"time"
)

func filterTestInfo() StreamInfo {
	return StreamInfo{
		StreamID:  "stream123",
		Login:     "teststreamer",
		Streamer:  "TestStreamer",
		Title:     "Cozy Saturday hangout",
		Game:      "Just Chatting",
		Viewers:   42,
		StartedAt: time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC),
		Live:      true,
	}
}

func filterTestChannel(filters Filters) Channel {
	return Channel{ID: "channel1", DisplayName: "TestStreamer", Filters: normalizeFilters(filters)}
}

func TestFilterExcludeWinsOverInclude(t *testing.T) {
	channel := filterTestChannel(Filters{
		IncludeGames:  []string{"Just Chatting"},
		ExcludeGames:  []string{"Just Chatting"},
		IncludeTitles: []string{"hangout"},
		ExcludeTitles: []string{"hangout"},
	})
	now := filterTestInfo().StartedAt.Add(time.Hour)
	allowed, reason := evaluateFilters(channel, filterTestInfo(), now, time.UTC)
	if allowed {
		t.Fatal("excluded game/title was allowed")
	}
	if !strings.Contains(reason, "excluded") {
		t.Fatalf("reason = %q", reason)
	}

	channel.Filters = normalizeFilters(Filters{IncludeGames: []string{"Just Chatting"}, ExcludeGames: []string{"Nothing"}})
	if allowed, reason := evaluateFilters(channel, filterTestInfo(), now, time.UTC); !allowed {
		t.Fatalf("included game was rejected: %s", reason)
	}
}

func TestFilterIncludeLists(t *testing.T) {
	channel := filterTestChannel(Filters{IncludeGames: []string{"Music"}})
	allowed, reason := evaluateFilters(channel, filterTestInfo(), time.Now(), time.UTC)
	if allowed || !strings.Contains(reason, "include") {
		t.Fatalf("allowed = %v, reason = %q", allowed, reason)
	}

	channel.Filters = normalizeFilters(Filters{IncludeTitles: []string{"karaoke"}})
	allowed, reason = evaluateFilters(channel, filterTestInfo(), time.Now(), time.UTC)
	if allowed || !strings.Contains(reason, "included keyword") {
		t.Fatalf("allowed = %v, reason = %q", allowed, reason)
	}

	channel.Filters = normalizeFilters(Filters{IncludeTitles: []string{"cozy"}})
	if allowed, reason := evaluateFilters(channel, filterTestInfo(), time.Now(), time.UTC); !allowed {
		t.Fatalf("matching title keyword was rejected: %s", reason)
	}
}

func TestFilterMinimumViewers(t *testing.T) {
	channel := filterTestChannel(Filters{MinViewers: 100})
	allowed, reason := evaluateFilters(channel, filterTestInfo(), time.Now(), time.UTC)
	if allowed || !strings.Contains(reason, "viewers") {
		t.Fatalf("allowed = %v, reason = %q", allowed, reason)
	}
	info := filterTestInfo()
	info.Viewers = 100
	if allowed, reason := evaluateFilters(channel, info, time.Now(), time.UTC); !allowed {
		t.Fatalf("exact minimum was rejected: %s", reason)
	}
}

func TestFilterAlertDelay(t *testing.T) {
	info := filterTestInfo()
	now := info.StartedAt.Add(2 * time.Minute)
	channel := filterTestChannel(Filters{DelayMinutes: 5})
	allowed, reason := evaluateFilters(channel, info, now, time.UTC)
	if allowed || !strings.Contains(reason, "delay") {
		t.Fatalf("allowed = %v, reason = %q", allowed, reason)
	}
	if allowed, reason := evaluateFilters(channel, info, info.StartedAt.Add(5*time.Minute), time.UTC); !allowed {
		t.Fatalf("delay boundary was rejected: %s", reason)
	}
}

func TestFilterQuietHoursOvernight(t *testing.T) {
	channel := filterTestChannel(Filters{QuietStart: "23:00", QuietEnd: "07:00"})
	location := time.FixedZone("Test", 2*60*60)
	tests := []struct {
		hour, minute int
		blocked      bool
	}{
		{22, 59, false},
		{23, 0, true},
		{23, 30, true},
		{0, 0, true},
		{6, 59, true},
		{7, 0, false},
		{12, 0, false},
	}
	for _, test := range tests {
		now := time.Date(2026, 11, 18, test.hour, test.minute, 0, 0, location)
		allowed, reason := evaluateFilters(channel, filterTestInfo(), now, location)
		if allowed == test.blocked {
			t.Errorf("%02d:%02d allowed = %v, want blocked = %v (reason %q)", test.hour, test.minute, allowed, test.blocked, reason)
		}
	}
}

func TestFilterQuietHoursSameDay(t *testing.T) {
	channel := filterTestChannel(Filters{QuietStart: "09:00", QuietEnd: "17:00"})
	now := time.Date(2026, 11, 18, 10, 30, 0, 0, time.UTC)
	if allowed, reason := evaluateFilters(channel, filterTestInfo(), now, time.UTC); allowed || !strings.Contains(reason, "quiet hours") {
		t.Fatalf("allowed = %v, reason = %q", allowed, reason)
	}
	now = time.Date(2026, 11, 18, 18, 0, 0, 0, time.UTC)
	if allowed, reason := evaluateFilters(channel, filterTestInfo(), now, time.UTC); !allowed {
		t.Fatalf("outside quiet hours was rejected: %s", reason)
	}
}

func TestFilterQuietHoursUseLocalTimezone(t *testing.T) {
	channel := filterTestChannel(Filters{QuietStart: "09:00", QuietEnd: "17:00"})
	// 08:30 UTC is 10:30 in a +2 zone: quiet locally even though it is
	// outside the window in UTC.
	now := time.Date(2026, 11, 18, 8, 30, 0, 0, time.UTC)
	location := time.FixedZone("Test", 2*60*60)
	allowed, reason := evaluateFilters(channel, filterTestInfo(), now, location)
	if allowed || !strings.Contains(reason, "quiet hours") {
		t.Fatalf("allowed = %v, reason = %q", allowed, reason)
	}
	if allowed, reason := evaluateFilters(channel, filterTestInfo(), now, time.UTC); !allowed {
		t.Fatalf("UTC evaluation was rejected: %s", reason)
	}
}

func TestFilterAllowedDays(t *testing.T) {
	channel := filterTestChannel(Filters{AllowedDays: []int{1, 3}})
	wednesday := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC) // Wednesday
	if allowed, reason := evaluateFilters(channel, filterTestInfo(), wednesday, time.UTC); !allowed {
		t.Fatalf("allowed day was rejected: %s", reason)
	}
	thursday := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	if allowed, reason := evaluateFilters(channel, filterTestInfo(), thursday, time.UTC); allowed || !strings.Contains(reason, "allowed day") {
		t.Fatalf("allowed = %v, reason = %q", allowed, reason)
	}
}

func TestFilterCooldownAndFirstStreamOfDay(t *testing.T) {
	info := filterTestInfo()
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	channel := filterTestChannel(Filters{CooldownMinutes: 60})
	channel.LastAlertAt = now.Add(-10 * time.Minute).Format(time.RFC3339)
	channel.LastAlertStreamID = "old-stream"
	if allowed, reason := evaluateFilters(channel, info, now, time.UTC); allowed || !strings.Contains(reason, "cooldown") {
		t.Fatalf("allowed = %v, reason = %q", allowed, reason)
	}
	channel.LastAlertStreamID = info.StreamID
	if allowed, reason := evaluateFilters(channel, info, now, time.UTC); !allowed {
		t.Fatalf("same stream was blocked by cooldown: %s", reason)
	}

	channel.Filters = normalizeFilters(Filters{FirstStreamOfDay: true})
	channel.LastAlertStreamID = "old-stream"
	channel.LastAlertDate = now.Format("2006-01-02")
	if allowed, reason := evaluateFilters(channel, info, now, time.UTC); allowed || !strings.Contains(reason, "first stream") {
		t.Fatalf("allowed = %v, reason = %q", allowed, reason)
	}
	channel.LastAlertStreamID = info.StreamID
	if allowed, reason := evaluateFilters(channel, info, now, time.UTC); !allowed {
		t.Fatalf("same stream was blocked by first-of-day: %s", reason)
	}
}

func TestFilterDefaultsAllow(t *testing.T) {
	channel := filterTestChannel(Filters{})
	if allowed, reason := evaluateFilters(channel, filterTestInfo(), time.Now(), time.UTC); !allowed {
		t.Fatalf("empty filters rejected: %s", reason)
	}
}

func TestValidateFilters(t *testing.T) {
	valid, err := validateFilters(Filters{
		IncludeGames:    []string{" Just Chatting ", "Just Chatting"},
		QuietStart:      "23:00",
		QuietEnd:        "07:00",
		AllowedDays:     []int{3, 1, 3},
		CooldownMinutes: 60,
		DelayMinutes:    5,
	})
	if err != nil {
		t.Fatalf("validateFilters returned error: %v", err)
	}
	if len(valid.IncludeGames) != 1 || len(valid.AllowedDays) != 2 || valid.AllowedDays[0] != 1 || valid.AllowedDays[1] != 3 {
		t.Fatalf("normalized filters = %+v", valid)
	}

	for _, filters := range []Filters{
		{QuietStart: "25:00", QuietEnd: "07:00"},
		{QuietStart: "23:00", QuietEnd: "7:00"},
		{MinViewers: -1},
		{CooldownMinutes: -1},
		{DelayMinutes: -1},
		{AllowedDays: []int{7}},
	} {
		if _, err := validateFilters(filters); err == nil {
			t.Errorf("validateFilters accepted %+v", filters)
		}
	}
}
