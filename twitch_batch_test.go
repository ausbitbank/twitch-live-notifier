package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func batchUserJSON(login string) map[string]any {
	return map[string]any{
		"login":           login,
		"displayName":     login,
		"profileImageURL": "https://example.com/" + login + ".png",
		"bannerImageURL":  "",
		"stream":          nil,
		"lastBroadcast":   nil,
	}
}

func TestTwitchBatchFetchBatchesAliases(t *testing.T) {
	var requests atomic.Int32
	var maxAliases atomic.Int32
	var inFlight atomic.Int32
	var maxInFlight atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		current := inFlight.Add(1)
		defer inFlight.Add(-1)
		for {
			old := maxInFlight.Load()
			if current <= old || maxInFlight.CompareAndSwap(old, current) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)

		var request struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
		}
		logins := twitchRequestLogins(request.Query)
		if int32(len(logins)) > maxAliases.Load() {
			maxAliases.Store(int32(len(logins)))
		}
		data := make(map[string]any, len(logins))
		for index, login := range logins {
			data["u"+strconv.Itoa(index)] = batchUserJSON(login)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
	}))
	setTwitchTestURL(t, server)

	logins := make([]string, 45)
	for index := range logins {
		logins[index] = fmt.Sprintf("user%d", index)
	}
	results := fetchTwitchChannelsBatchContext(context.Background(), logins)
	if len(results) != 45 {
		t.Fatalf("results = %d, want 45", len(results))
	}
	for _, login := range logins {
		result, ok := results[login]
		if !ok || result.Err != nil || result.User == nil {
			t.Fatalf("result for %s = %+v", login, result)
		}
	}
	if requests.Load() != 3 {
		t.Fatalf("HTTP requests = %d, want 3 batches", requests.Load())
	}
	if maxAliases.Load() > twitchBatchSize {
		t.Fatalf("largest batch = %d, want <= %d", maxAliases.Load(), twitchBatchSize)
	}
	if maxInFlight.Load() > twitchBatchConcurrency {
		t.Fatalf("concurrent requests = %d, want <= %d", maxInFlight.Load(), twitchBatchConcurrency)
	}
}

func TestTwitchBatchFetchPartialResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		logins := twitchRequestLogins(request.Query)
		data := make(map[string]any, len(logins))
		for index, login := range logins {
			switch login {
			case "missing":
				data["u"+strconv.Itoa(index)] = nil
			case "noname":
				data["u"+strconv.Itoa(index)] = map[string]any{
					"login":           login,
					"displayName":     "",
					"profileImageURL": "",
					"stream":          nil,
					"lastBroadcast":   nil,
				}
			default:
				data["u"+strconv.Itoa(index)] = batchUserJSON(login)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
	}))
	setTwitchTestURL(t, server)

	results := fetchTwitchChannelsBatchContext(context.Background(), []string{"alpha", "missing", "noname"})
	if results["alpha"].User == nil || results["alpha"].Err != nil {
		t.Fatalf("alpha result = %+v", results["alpha"])
	}
	if results["missing"].User != nil || results["missing"].Err == nil {
		t.Fatalf("missing result = %+v", results["missing"])
	}
	if results["missing"].Err.Error() != errTwitchChannelNotFound.Error() {
		t.Fatalf("missing error = %v, want channel not found", results["missing"].Err)
	}
	if results["noname"].User == nil || results["noname"].User.DisplayName != "noname" {
		t.Fatalf("noname result = %+v", results["noname"])
	}
}

func TestTwitchBatchFetchRetriesRateLimit(t *testing.T) {
	recorder := stubRetrySleep(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		var request struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		logins := twitchRequestLogins(request.Query)
		data := make(map[string]any, len(logins))
		for index, login := range logins {
			data["u"+strconv.Itoa(index)] = batchUserJSON(login)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
	}))
	setTwitchTestURL(t, server)

	results := fetchTwitchChannelsBatchContext(context.Background(), []string{"alpha", "beta"})
	for _, login := range []string{"alpha", "beta"} {
		if results[login].Err != nil || results[login].User == nil {
			t.Fatalf("result for %s = %+v", login, results[login])
		}
	}
	if calls.Load() != 2 {
		t.Fatalf("calls = %d, want 2", calls.Load())
	}
	delays := recorder.snapshot()
	if len(delays) != 1 || delays[0] != 2*time.Second {
		t.Fatalf("retry delays = %#v, want one 2s delay", delays)
	}
}

func TestTwitchBatchFetchFallsBackToSingles(t *testing.T) {
	var batchCalls atomic.Int32
	var singleCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Query     string            `json:"query"`
			Variables map[string]string `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&request)
		w.Header().Set("Content-Type", "application/json")
		if logins := twitchRequestLogins(request.Query); len(logins) > 0 {
			batchCalls.Add(1)
			// An endpoint that does not understand aliases answers with the
			// plain single-user shape.
			_, _ = io.WriteString(w, `{"data":{"user":{"login":"ignored","displayName":"Ignored"}}}`)
			return
		}
		singleCalls.Add(1)
		login := request.Variables["login"]
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"user": batchUserJSON(login)}})
	}))
	setTwitchTestURL(t, server)

	results := fetchTwitchChannelsBatchContext(context.Background(), []string{"alpha", "beta"})
	for _, login := range []string{"alpha", "beta"} {
		if results[login].Err != nil || results[login].User == nil || results[login].User.Login != login {
			t.Fatalf("result for %s = %+v", login, results[login])
		}
	}
	if batchCalls.Load() != 1 || singleCalls.Load() != 2 {
		t.Fatalf("batch calls = %d, single calls = %d", batchCalls.Load(), singleCalls.Load())
	}
}

func TestTwitchBatchFetchGlobalFailure(t *testing.T) {
	stubRetrySleep(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	setTwitchTestURL(t, server)

	results := fetchTwitchChannelsBatchContext(context.Background(), []string{"alpha", "beta"})
	for _, login := range []string{"alpha", "beta"} {
		if results[login].Err == nil {
			t.Fatalf("result for %s = %+v, want error", login, results[login])
		}
	}
	if calls.Load() != twitchMaxAttempts {
		t.Fatalf("calls = %d, want %d", calls.Load(), twitchMaxAttempts)
	}
}

func TestTwitchBatchFetchInvalidLoginMakesNoRequest(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	setTwitchTestURL(t, server)

	results := fetchTwitchChannelsBatchContext(context.Background(), []string{"not a channel!"})
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if calls.Load() != 0 {
		t.Fatalf("calls = %d, want 0", calls.Load())
	}
}
