package main

import (
	"context"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// retrySleep is a variable so tests can avoid real waiting. It must honor
// context cancellation so shutdown does not hang on a long rate-limit delay.
var retrySleep = func(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// backoffDelay returns base * 2^attempt with up to 50% positive jitter, capped
// at max. Jitter spreads retries when many streamers fail together.
func backoffDelay(attempt int, base, max time.Duration) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	delay := base
	for i := 0; i < attempt && delay < max; i++ {
		delay *= 2
	}
	if delay <= 0 || delay > max {
		delay = max
	}
	if delay > 0 {
		jitter := time.Duration(rand.Int63n(int64(delay/2) + 1))
		delay += jitter
		if delay > max {
			delay = max
		}
	}
	return delay
}

func retryAfterFromHeaders(value string) time.Duration {
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

func minDuration(value, minimum time.Duration) time.Duration {
	if value < minimum {
		return minimum
	}
	return value
}
