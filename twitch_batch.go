package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// twitchBatchSize keeps GraphQL alias queries small enough that Twitch
	// answers every alias in one response. twitchBatchConcurrency bounds the
	// number of HTTP requests in flight so a large streamer list does not
	// trigger a rate-limit wave.
	twitchBatchSize        = 20
	twitchBatchConcurrency = 3
)

// twitchBatchResult is the outcome for one login in a batched fetch. A login
// either has a User or an Err, never both.
type twitchBatchResult struct {
	Login string
	User  *twitchUser
	Err   error
}

type twitchBatchGQLResponse struct {
	Data   map[string]*twitchUser `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// fetchTwitchChannelsBatchContext resolves many logins with aliased GraphQL
// queries. It returns a result per unique, normalized login. Invalid logins
// get an error result without making a request. Missing users map to
// errTwitchChannelNotFound. A terminal request failure is reported for every
// login in the affected batch; a rejected batch shape falls back to the
// single-channel path.
func fetchTwitchChannelsBatchContext(ctx context.Context, logins []string) map[string]twitchBatchResult {
	results := make(map[string]twitchBatchResult, len(logins))
	normalized := make([]string, 0, len(logins))
	seen := make(map[string]bool, len(logins))

	for _, raw := range logins {
		login, err := normalizeTwitchLogin(raw)
		if err != nil {
			key := strings.ToLower(strings.TrimSpace(raw))
			if key == "" {
				key = raw
			}
			if _, exists := results[key]; !exists {
				results[key] = twitchBatchResult{Login: strings.TrimSpace(raw), Err: err}
			}
			continue
		}
		if seen[login] {
			continue
		}
		seen[login] = true
		normalized = append(normalized, login)
	}
	if len(normalized) == 0 {
		return results
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, twitchBatchConcurrency)
	for start := 0; start < len(normalized); start += twitchBatchSize {
		end := start + twitchBatchSize
		if end > len(normalized) {
			end = len(normalized)
		}
		batch := normalized[start:end]
		wg.Add(1)
		go func(batch []string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				mu.Lock()
				for _, login := range batch {
					results[login] = twitchBatchResult{Login: login, Err: ctx.Err()}
				}
				mu.Unlock()
				return
			}
			defer func() { <-sem }()

			batchResults := fetchTwitchBatchWithRetry(ctx, batch)
			mu.Lock()
			for login, result := range batchResults {
				results[login] = result
			}
			mu.Unlock()
		}(batch)
	}
	wg.Wait()

	// A batch cancelled before it started still needs a result per login.
	for _, login := range normalized {
		if _, ok := results[login]; !ok {
			results[login] = twitchBatchResult{Login: login, Err: errors.New("Twitch did not return this channel")}
		}
	}
	return results
}

// fetchTwitchBatchWithRetry mirrors the single-channel retry policy: retry
// transient transport/5xx/429 failures with backoff, honor Retry-After, and
// never retry a per-request protocol error. Protocol errors fall back to
// single-channel requests for the same batch.
func fetchTwitchBatchWithRetry(ctx context.Context, logins []string) map[string]twitchBatchResult {
	var lastErr error
	lastRetryable := false

	for attempt := 0; attempt < twitchMaxAttempts; attempt++ {
		users, retryAfter, retryable, err := fetchTwitchBatchOnce(ctx, logins)
		if err == nil {
			return twitchBatchResults(logins, users)
		}
		lastErr = err
		lastRetryable = retryable
		if !retryable || attempt == twitchMaxAttempts-1 {
			break
		}
		delay := backoffDelay(attempt, twitchRetryBase, twitchRetryMax)
		if retryAfter > delay {
			delay = retryAfter
		}
		if delay > twitchRetryMax {
			delay = twitchRetryMax
		}
		if sleepErr := retrySleep(ctx, delay); sleepErr != nil {
			lastErr = sleepErr
			lastRetryable = false
			break
		}
	}
	if lastErr == nil {
		lastErr = errors.New("Twitch did not return this channel")
	}
	if !lastRetryable && !errors.Is(lastErr, context.Canceled) && !errors.Is(lastErr, context.DeadlineExceeded) {
		return fetchTwitchBatchSingles(ctx, logins)
	}

	results := make(map[string]twitchBatchResult, len(logins))
	for _, login := range logins {
		results[login] = twitchBatchResult{Login: login, Err: lastErr}
	}
	return results
}

func fetchTwitchBatchOnce(ctx context.Context, logins []string) (map[string]*twitchUser, time.Duration, bool, error) {
	requestBody, err := json.Marshal(map[string]string{
		"query": twitchBatchQuery(logins),
	})
	if err != nil {
		return nil, 0, false, errors.New("could not prepare the Twitch request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, twitchGQLURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, 0, false, errors.New("could not prepare the Twitch request")
	}
	req.Header.Set("Client-ID", twitchWebClientID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", "https://www.twitch.tv")
	req.Header.Set("Referer", "https://www.twitch.tv/directory/following/channels")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, true, fmt.Errorf("could not reach Twitch: %s", networkErrorText(err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, 0, true, errors.New("could not read the Twitch response")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retryAfter := retryAfterFromHeaders(resp.Header.Get("Retry-After"))
		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, retryAfter, true, fmt.Errorf("Twitch rate limited the request (HTTP 429)")
		}
		return nil, retryAfter, retryable, fmt.Errorf("Twitch returned HTTP %d", resp.StatusCode)
	}

	var parsed twitchBatchGQLResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, 0, false, errors.New("Twitch returned an invalid response")
	}
	if len(parsed.Data) == 0 {
		if len(parsed.Errors) > 0 {
			message := strings.TrimSpace(parsed.Errors[0].Message)
			if message == "" {
				message = "unknown API error"
			}
			return nil, 0, false, fmt.Errorf("Twitch API error: %s", message)
		}
		return nil, 0, false, errors.New("Twitch returned an empty batch response")
	}

	// If the response has none of the aliases we asked for, the endpoint did
	// not understand the batched query. Report a protocol error so the caller
	// can fall back to single-channel requests.
	hasAlias := false
	for index := range logins {
		if _, ok := parsed.Data[batchAlias(index)]; ok {
			hasAlias = true
			break
		}
	}
	if !hasAlias {
		return nil, 0, false, errors.New("Twitch did not accept the batched request")
	}
	return parsed.Data, 0, false, nil
}

func twitchBatchResults(logins []string, data map[string]*twitchUser) map[string]twitchBatchResult {
	results := make(map[string]twitchBatchResult, len(logins))
	for index, login := range logins {
		user := data[batchAlias(index)]
		if user == nil {
			results[login] = twitchBatchResult{Login: login, Err: errTwitchChannelNotFound}
			continue
		}
		if strings.TrimSpace(user.DisplayName) == "" {
			user.DisplayName = login
		}
		results[login] = twitchBatchResult{Login: login, User: user}
	}
	return results
}

// fetchTwitchBatchSingles is the compatibility fallback for endpoints (or
// test doubles) that do not understand alias queries. One request per batch at
// a time keeps the total in-flight count bounded by the batch concurrency.
func fetchTwitchBatchSingles(ctx context.Context, logins []string) map[string]twitchBatchResult {
	results := make(map[string]twitchBatchResult, len(logins))
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 1)
	for _, login := range logins {
		wg.Add(1)
		go func(login string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				mu.Lock()
				results[login] = twitchBatchResult{Login: login, Err: ctx.Err()}
				mu.Unlock()
				return
			}
			defer func() { <-sem }()

			user, err := fetchTwitchChannelContext(ctx, login)
			mu.Lock()
			results[login] = twitchBatchResult{Login: login, User: user, Err: err}
			mu.Unlock()
		}(login)
	}
	wg.Wait()
	return results
}

// twitchBatchQuery builds one query with an alias per login. Logins are
// normalized before reaching this function, but strconv.Quote keeps the query
// safe if that ever changes.
func twitchBatchQuery(logins []string) string {
	var builder strings.Builder
	builder.WriteString("query ChannelBatch {\n")
	for index, login := range logins {
		builder.WriteString("  ")
		builder.WriteString(batchAlias(index))
		builder.WriteString(": user(login: ")
		builder.WriteString(strconv.Quote(login))
		builder.WriteString(") {\n")
		builder.WriteString(twitchUserFields)
		builder.WriteString("\n  }\n")
	}
	builder.WriteString("}")
	return builder.String()
}

func batchAlias(index int) string {
	return "u" + strconv.Itoa(index)
}
