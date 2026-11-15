package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	discordMaxAttempts    = 2
	discordMaxRetryDelay  = 60 * time.Second
	discordDefaultRetryIn = time.Second
)

type discordWebhookPayload struct {
	Content         string                  `json:"content,omitempty"`
	Embeds          []discordEmbed          `json:"embeds,omitempty"`
	AllowedMentions *discordAllowedMentions `json:"allowed_mentions,omitempty"`
}

type discordAllowedMentions struct {
	Parse []string `json:"parse"`
	Roles []string `json:"roles,omitempty"`
	Users []string `json:"users,omitempty"`
}

type discordEmbed struct {
	Author      *discordAuthor `json:"author,omitempty"`
	Title       string         `json:"title,omitempty"`
	URL         string         `json:"url,omitempty"`
	Description string         `json:"description,omitempty"`
	Color       int            `json:"color,omitempty"`
	Fields      []discordField `json:"fields,omitempty"`
	Image       *discordImage  `json:"image,omitempty"`
	Footer      *discordFooter `json:"footer,omitempty"`
	Timestamp   string         `json:"timestamp,omitempty"`
}

type discordAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type discordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type discordImage struct {
	URL string `json:"url"`
}

type discordFooter struct {
	Text string `json:"text"`
}

type discordPostResult struct {
	MessageID string
}

// discordHTTPError carries the status code so callers can distinguish a
// permanent credential/target problem from a transient failure.
type discordHTTPError struct {
	StatusCode int
	Detail     string
	RetryAfter time.Duration
}

func (e *discordHTTPError) Error() string {
	return fmt.Sprintf("Discord returned HTTP %d%s", e.StatusCode, e.Detail)
}

func isPermanentDiscordError(err error) bool {
	var httpErr *discordHTTPError
	if !errors.As(err, &httpErr) {
		return false
	}
	switch httpErr.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return true
	default:
		return false
	}
}

func postDiscord(webhookURL string, payload discordWebhookPayload) (discordPostResult, error) {
	return postDiscordContext(context.Background(), webhookURL, payload)
}

func postDiscordContext(ctx context.Context, webhookURL string, payload discordWebhookPayload) (discordPostResult, error) {
	u, err := url.Parse(webhookURL)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return discordPostResult{}, errors.New("invalid Discord webhook URL")
	}

	query := u.Query()
	query.Set("wait", "true")
	u.RawQuery = query.Encode()

	body, err := json.Marshal(payload)
	if err != nil {
		return discordPostResult{}, errors.New("could not build the Discord payload")
	}

	var lastErr error
	for attempt := 0; attempt < discordMaxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(body))
		if err != nil {
			return discordPostResult{}, errors.New("could not build the Discord request")
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", userAgent)

		resp, err := httpClient.Do(req)
		if err != nil {
			return discordPostResult{}, fmt.Errorf("could not reach Discord: %s", networkErrorText(err))
		}

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var posted struct {
				ID string `json:"id"`
			}
			_ = json.Unmarshal(respBody, &posted)
			return discordPostResult{MessageID: posted.ID}, nil
		}

		httpErr := &discordHTTPError{
			StatusCode: resp.StatusCode,
			Detail:     discordErrorDetails(respBody),
			RetryAfter: discordRetryAfter(resp, respBody),
		}
		lastErr = httpErr
		if resp.StatusCode != http.StatusTooManyRequests || attempt == discordMaxAttempts-1 {
			return discordPostResult{}, httpErr
		}

		delay := httpErr.RetryAfter
		if delay <= 0 {
			delay = discordDefaultRetryIn
		}
		if delay > discordMaxRetryDelay {
			delay = discordMaxRetryDelay
		}
		if sleepErr := retrySleep(ctx, delay); sleepErr != nil {
			return discordPostResult{}, sleepErr
		}
	}
	return discordPostResult{}, lastErr
}

func discordRetryAfter(resp *http.Response, body []byte) time.Duration {
	var payload struct {
		RetryAfter float64 `json:"retry_after"`
	}
	if err := json.Unmarshal(body, &payload); err == nil && payload.RetryAfter > 0 {
		return time.Duration(payload.RetryAfter * float64(time.Second))
	}
	if resp != nil {
		return retryAfterFromHeaders(resp.Header.Get("Retry-After"))
	}
	return 0
}

func discordErrorDetails(body []byte) string {
	var apiError struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &apiError)

	message := strings.TrimSpace(apiError.Message)
	if message == "" {
		message = strings.TrimSpace(string(body))
	}
	if message == "" {
		return ""
	}
	message = strings.Join(strings.Fields(message), " ")
	if len(message) > 300 {
		message = message[:300] + "..."
	}
	return ": " + message
}
