package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

type retryRecorder struct {
	mu     sync.Mutex
	delays []time.Duration
}

func (r *retryRecorder) sleep(_ context.Context, delay time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.delays = append(r.delays, delay)
	return nil
}

func (r *retryRecorder) snapshot() []time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]time.Duration(nil), r.delays...)
}

func stubRetrySleep(t *testing.T) *retryRecorder {
	t.Helper()
	recorder := &retryRecorder{}
	old := retrySleep
	retrySleep = recorder.sleep
	t.Cleanup(func() { retrySleep = old })
	return recorder
}
