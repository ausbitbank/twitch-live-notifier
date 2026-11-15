package main

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// discordRateLimiter keeps a small in-memory view of Discord's per-webhook
// rate-limit buckets. It complements the persisted per-destination retry
// window: the in-memory view handles proactive X-RateLimit headers during a
// process lifetime, while RetryUntil survives restarts.
type discordRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]time.Time
}

func newDiscordRateLimiter() *discordRateLimiter {
	return &discordRateLimiter{buckets: map[string]time.Time{}}
}

var globalDiscordRateLimiter = newDiscordRateLimiter()

func discordRateKey(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Host == "" {
		return strings.TrimSpace(rawURL)
	}
	path := strings.TrimSuffix(parsed.Path, "/")
	if index := strings.Index(path, "/messages/"); index >= 0 {
		path = path[:index]
	}
	parsed.Path = path
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func (l *discordRateLimiter) wait(ctx context.Context, rawURL string) error {
	l.mu.Lock()
	notBefore := l.buckets[discordRateKey(rawURL)]
	l.mu.Unlock()
	delay := time.Until(notBefore)
	if delay <= 0 {
		return nil
	}
	if delay > discordInlineRetryMax {
		return &discordHTTPError{
			StatusCode: http.StatusTooManyRequests,
			Detail:     ": deferred by local rate limiter",
			RetryAfter: delay,
		}
	}
	return retrySleep(ctx, delay)
}

func (l *discordRateLimiter) deferFor(rawURL string, delay time.Duration) {
	if delay <= 0 {
		return
	}
	key := discordRateKey(rawURL)
	until := time.Now().Add(delay)
	l.mu.Lock()
	if current := l.buckets[key]; until.After(current) {
		l.buckets[key] = until
	}
	l.mu.Unlock()
}

func (l *discordRateLimiter) clear(rawURL string) {
	l.mu.Lock()
	delete(l.buckets, discordRateKey(rawURL))
	l.mu.Unlock()
}

// observe reads Discord's proactive rate-limit headers. When the bucket is
// empty, it defers the next request until the reset window instead of waiting
// for a 429.
func (l *discordRateLimiter) observe(rawURL string, headers http.Header) {
	if headers == nil {
		return
	}
	remaining, ok := parseHeaderInt(headers.Get("X-RateLimit-Remaining"))
	if !ok {
		return
	}
	if remaining > 0 {
		l.clear(rawURL)
		return
	}
	resetAfter := parseHeaderDuration(headers.Get("X-RateLimit-Reset-After"))
	if resetAfter <= 0 {
		resetAfter = parseHeaderDuration(headers.Get("Retry-After"))
	}
	l.deferFor(rawURL, resetAfter)
}

func parseHeaderInt(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func parseHeaderDuration(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}

func resetDiscordRateLimiter() {
	globalDiscordRateLimiter = newDiscordRateLimiter()
}
