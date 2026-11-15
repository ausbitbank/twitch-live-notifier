package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	maxFilterRules      = 50
	maxFilterRuleLength = 100
	maxFilterMinutes    = 7 * 24 * 60
	maxFilterViewers    = 10_000_000
)

func normalizeFilters(filters Filters) Filters {
	filters.IncludeGames = normalizeRuleList(filters.IncludeGames)
	filters.ExcludeGames = normalizeRuleList(filters.ExcludeGames)
	filters.IncludeTitles = normalizeRuleList(filters.IncludeTitles)
	filters.ExcludeTitles = normalizeRuleList(filters.ExcludeTitles)

	if filters.MinViewers < 0 {
		filters.MinViewers = 0
	}
	if filters.MinViewers > maxFilterViewers {
		filters.MinViewers = maxFilterViewers
	}
	if filters.CooldownMinutes < 0 {
		filters.CooldownMinutes = 0
	}
	if filters.CooldownMinutes > maxFilterMinutes {
		filters.CooldownMinutes = maxFilterMinutes
	}
	if filters.DelayMinutes < 0 {
		filters.DelayMinutes = 0
	}
	if filters.DelayMinutes > maxFilterMinutes {
		filters.DelayMinutes = maxFilterMinutes
	}

	filters.QuietStart = normalizeClock(filters.QuietStart)
	filters.QuietEnd = normalizeClock(filters.QuietEnd)
	if filters.QuietStart == "" || filters.QuietEnd == "" {
		filters.QuietStart = ""
		filters.QuietEnd = ""
	}

	filters.AllowedDays = normalizeAllowedDays(filters.AllowedDays)
	return filters
}

func validateFilters(filters Filters) (Filters, error) {
	if filters.MinViewers < 0 || filters.MinViewers > maxFilterViewers {
		return Filters{}, fmt.Errorf("minimum viewers must be between 0 and %d", maxFilterViewers)
	}
	if filters.CooldownMinutes < 0 || filters.CooldownMinutes > maxFilterMinutes {
		return Filters{}, fmt.Errorf("cooldown must be between 0 and %d minutes", maxFilterMinutes)
	}
	if filters.DelayMinutes < 0 || filters.DelayMinutes > maxFilterMinutes {
		return Filters{}, fmt.Errorf("alert delay must be between 0 and %d minutes", maxFilterMinutes)
	}
	for _, list := range [][]string{filters.IncludeGames, filters.ExcludeGames, filters.IncludeTitles, filters.ExcludeTitles} {
		if len(list) > maxFilterRules {
			return Filters{}, fmt.Errorf("use at most %d rules per filter list", maxFilterRules)
		}
		for _, rule := range list {
			if len(rule) > maxFilterRuleLength {
				return Filters{}, fmt.Errorf("filter rules must be %d characters or fewer", maxFilterRuleLength)
			}
		}
	}
	if filters.QuietStart != "" || filters.QuietEnd != "" {
		if _, _, ok := parseClock(filters.QuietStart); !ok {
			return Filters{}, fmt.Errorf("quiet hours start must be a time like 23:00")
		}
		if _, _, ok := parseClock(filters.QuietEnd); !ok {
			return Filters{}, fmt.Errorf("quiet hours end must be a time like 07:00")
		}
	}
	for _, day := range filters.AllowedDays {
		if day < 0 || day > 6 {
			return Filters{}, fmt.Errorf("allowed days must use 0 (Sunday) through 6 (Saturday)")
		}
	}
	normalized := normalizeFilters(filters)
	if normalized.QuietStart == normalized.QuietEnd && normalized.QuietStart != "" {
		// A zero-length window is confusing; treat it as disabled.
		normalized.QuietStart = ""
		normalized.QuietEnd = ""
	}
	return normalized, nil
}

func normalizeRuleList(input []string) []string {
	seen := make(map[string]bool, len(input))
	rules := make([]string, 0, len(input))
	for _, raw := range input {
		rule := strings.TrimSpace(raw)
		if rule == "" {
			continue
		}
		if len(rule) > maxFilterRuleLength {
			rule = rule[:maxFilterRuleLength]
		}
		key := strings.ToLower(rule)
		if seen[key] {
			continue
		}
		seen[key] = true
		rules = append(rules, rule)
		if len(rules) == maxFilterRules {
			break
		}
	}
	if len(rules) == 0 {
		return nil
	}
	return rules
}

func normalizeAllowedDays(input []int) []int {
	seen := make(map[int]bool, len(input))
	days := make([]int, 0, len(input))
	for _, day := range input {
		if day < 0 || day > 6 || seen[day] {
			continue
		}
		seen[day] = true
		days = append(days, day)
	}
	sort.Ints(days)
	if len(days) == 0 {
		return nil
	}
	return days
}

func parseClock(value string) (int, int, bool) {
	value = strings.TrimSpace(value)
	if len(value) != 5 || value[2] != ':' {
		return 0, 0, false
	}
	hour, minute := 0, 0
	for i, r := range value {
		if i == 2 {
			continue
		}
		if r < '0' || r > '9' {
			return 0, 0, false
		}
		digit := int(r - '0')
		if i < 2 {
			hour = hour*10 + digit
		} else {
			minute = minute*10 + digit
		}
	}
	if hour > 23 || minute > 59 {
		return 0, 0, false
	}
	return hour, minute, true
}

func normalizeClock(value string) string {
	hour, minute, ok := parseClock(value)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%02d:%02d", hour, minute)
}

// evaluateFilters returns whether this live stream should alert and a short
// reason when it should not. Exclude rules always take precedence.
func evaluateFilters(channel Channel, info StreamInfo, now time.Time, loc *time.Location) (bool, string) {
	if loc == nil {
		loc = time.Local
	}
	localNow := now.In(loc)
	filters := channel.Filters

	if filters.MinViewers > 0 && info.Viewers < filters.MinViewers {
		return false, fmt.Sprintf("needs at least %s viewers (currently %s)", formatCount(filters.MinViewers), formatCount(info.Viewers))
	}
	if filters.DelayMinutes > 0 && !info.StartedAt.IsZero() {
		delay := time.Duration(filters.DelayMinutes) * time.Minute
		if now.Sub(info.StartedAt) < delay {
			return false, fmt.Sprintf("waiting for the %d-minute alert delay", filters.DelayMinutes)
		}
	}
	if len(filters.AllowedDays) > 0 && !containsInt(filters.AllowedDays, int(localNow.Weekday())) {
		return false, "today is not an allowed day"
	}
	if reason := quietHoursReason(localNow, filters); reason != "" {
		return false, reason
	}

	game := strings.TrimSpace(info.Game)
	title := strings.TrimSpace(info.Title)
	if rule, ok := matchRule(game, filters.ExcludeGames); ok {
		return false, fmt.Sprintf("excluded game/category %q", rule)
	}
	if len(filters.IncludeGames) > 0 {
		if _, ok := matchRule(game, filters.IncludeGames); !ok {
			return false, "game/category is not in the include list"
		}
	}
	if rule, ok := matchRule(title, filters.ExcludeTitles); ok {
		return false, fmt.Sprintf("title contains excluded keyword %q", rule)
	}
	if len(filters.IncludeTitles) > 0 {
		if _, ok := matchRule(title, filters.IncludeTitles); !ok {
			return false, "title does not contain an included keyword"
		}
	}

	// Cooldown and first-of-day apply to new streams. A stream that already
	// alerted to one destination must still be allowed to reach the others.
	if channel.LastAlertStreamID != info.StreamID {
		if filters.CooldownMinutes > 0 && channel.LastAlertAt != "" {
			if last, err := time.Parse(time.RFC3339, channel.LastAlertAt); err == nil {
				if localNow.Sub(last) < time.Duration(filters.CooldownMinutes)*time.Minute {
					return false, fmt.Sprintf("cooldown is %d minutes between alerts", filters.CooldownMinutes)
				}
			}
		}
		if filters.FirstStreamOfDay && channel.LastAlertDate != "" && channel.LastAlertDate == localNow.Format("2006-01-02") {
			return false, "already posted the first stream of the day"
		}
	}
	return true, ""
}

func quietHoursReason(localNow time.Time, filters Filters) string {
	startHour, startMinute, startOK := parseClock(filters.QuietStart)
	endHour, endMinute, endOK := parseClock(filters.QuietEnd)
	if !startOK || !endOK {
		return ""
	}
	start := startHour*60 + startMinute
	end := endHour*60 + endMinute
	if start == end {
		return ""
	}
	current := localNow.Hour()*60 + localNow.Minute()
	inQuiet := false
	if start < end {
		inQuiet = current >= start && current < end
	} else {
		inQuiet = current >= start || current < end
	}
	if !inQuiet {
		return ""
	}
	return fmt.Sprintf("quiet hours (%s–%s)", filters.QuietStart, filters.QuietEnd)
}

func matchRule(value string, rules []string) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "", false
	}
	for _, rule := range rules {
		needle := strings.ToLower(strings.TrimSpace(rule))
		if needle != "" && strings.Contains(value, needle) {
			return rule, true
		}
	}
	return "", false
}

func containsInt(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
