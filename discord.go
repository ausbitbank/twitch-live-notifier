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
	discordInlineRetryMax = 10 * time.Second
	discordMaxDefer       = time.Hour
	discordDefaultRetryIn = time.Second
)

type discordWebhookPayload struct {
	Content         string                  `json:"content,omitempty"`
	Username        string                  `json:"username,omitempty"`
	AvatarURL       string                  `json:"avatar_url,omitempty"`
	Embeds          []discordEmbed          `json:"embeds,omitempty"`
	AllowedMentions *discordAllowedMentions `json:"allowed_mentions,omitempty"`
	Components      []discordComponent      `json:"components,omitempty"`
}

type discordComponent struct {
	Type       int             `json:"type"`
	Components []discordButton `json:"components"`
}

type discordButton struct {
	Type  int    `json:"type"`
	Style int    `json:"style"`
	Label string `json:"label"`
	URL   string `json:"url"`
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
	ChannelID string
	GuildID   string
}

// MessageURL builds a jump link for the posted message when Discord returned
// the guild and channel IDs. It returns an empty string for DMs, group DMs,
// or a partial response.
func (r discordPostResult) MessageURL() string {
	return discordMessageURL(r.GuildID, r.ChannelID, r.MessageID)
}

// discordMessageURL builds https://discord.com/channels/<guild>/<channel>/<message>.
// IDs are validated as Discord snowflakes so a malformed response cannot inject
// a different URL into the dashboard.
func discordMessageURL(guildID, channelID, messageID string) string {
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	messageID = strings.TrimSpace(messageID)
	if !discordIDPattern.MatchString(guildID) || !discordIDPattern.MatchString(channelID) || !discordIDPattern.MatchString(messageID) {
		return ""
	}
	return "https://discord.com/channels/" + guildID + "/" + channelID + "/" + messageID
}

// parseDiscordPostResult reads the message, channel, and guild IDs from a
// webhook POST/PATCH response. Discord includes channel_id/guild_id on the
// message object; older responses or DMs may omit the guild.
func parseDiscordPostResult(body []byte, fallbackMessageID string) discordPostResult {
	var response struct {
		ID        string `json:"id"`
		ChannelID string `json:"channel_id"`
		GuildID   string `json:"guild_id"`
	}
	_ = json.Unmarshal(body, &response)
	result := discordPostResult{
		MessageID: strings.TrimSpace(response.ID),
		ChannelID: strings.TrimSpace(response.ChannelID),
		GuildID:   strings.TrimSpace(response.GuildID),
	}
	if result.MessageID == "" {
		result.MessageID = strings.TrimSpace(fallbackMessageID)
	}
	return result
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
	return postDiscordContextWithOptions(ctx, webhookURL, payload, discordWebhookOptions{})
}

type discordWebhookOptions struct {
	ThreadID string
}

func postDiscordContextWithOptions(ctx context.Context, webhookURL string, payload discordWebhookPayload, options discordWebhookOptions) (discordPostResult, error) {
	u, err := discordURL(webhookURL, "")
	if err != nil {
		return discordPostResult{}, err
	}
	query := u.Query()
	query.Set("wait", "true")
	if threadID := strings.TrimSpace(options.ThreadID); threadID != "" {
		query.Set("thread_id", threadID)
	}
	u.RawQuery = query.Encode()

	respBody, err := discordRequest(ctx, http.MethodPost, u.String(), payload)
	if err != nil {
		return discordPostResult{}, err
	}
	return parseDiscordPostResult(respBody, ""), nil
}

// editDiscordContext edits an existing webhook message. The caller passes the
// webhook URL that originally posted the message so a later destination URL
// change cannot redirect the edit to a different channel.
func editDiscordContext(ctx context.Context, webhookURL, messageID string, payload discordWebhookPayload, threadID string) (discordPostResult, error) {
	u, err := discordURL(webhookURL, messageID)
	if err != nil {
		return discordPostResult{}, err
	}
	query := u.Query()
	if value := strings.TrimSpace(threadID); value != "" {
		query.Set("thread_id", value)
	}
	u.RawQuery = query.Encode()

	respBody, err := discordRequest(ctx, http.MethodPatch, u.String(), payload)
	if err != nil {
		return discordPostResult{}, err
	}
	return parseDiscordPostResult(respBody, messageID), nil
}

func deleteDiscordContext(ctx context.Context, webhookURL, messageID, threadID string) error {
	u, err := discordURL(webhookURL, messageID)
	if err != nil {
		return err
	}
	query := u.Query()
	if value := strings.TrimSpace(threadID); value != "" {
		query.Set("thread_id", value)
	}
	u.RawQuery = query.Encode()

	_, err = discordRequest(ctx, http.MethodDelete, u.String(), nil)
	return err
}

func discordURL(webhookURL, messageID string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(webhookURL))
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return nil, errors.New("invalid Discord webhook URL")
	}
	if messageID != "" {
		u.Path = strings.TrimSuffix(u.Path, "/") + "/messages/" + url.PathEscape(messageID)
	}
	return u, nil
}

// discordRequest performs one logical Discord request with the retry policy:
// 429 waits for the server-provided delay and retries once. Other errors are
// surfaced immediately for the caller to classify.
func discordRequest(ctx context.Context, method, rawURL string, payload any) ([]byte, error) {
	var body []byte
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, errors.New("could not build the Discord payload")
		}
		body = encoded
	}

	if err := globalDiscordRateLimiter.wait(ctx, rawURL); err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt < discordMaxAttempts; attempt++ {
		var reader io.Reader
		if body != nil {
			reader = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, rawURL, reader)
		if err != nil {
			return nil, errors.New("could not build the Discord request")
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", userAgent)

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("could not reach Discord: %s", networkErrorText(err))
		}

		globalDiscordRateLimiter.observe(rawURL, resp.Header)
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		httpErr := &discordHTTPError{
			StatusCode: resp.StatusCode,
			Detail:     discordErrorDetails(respBody),
			RetryAfter: discordRetryAfter(resp, respBody),
		}
		lastErr = httpErr
		if resp.StatusCode == http.StatusTooManyRequests {
			globalDiscordRateLimiter.deferFor(rawURL, httpErr.RetryAfter)
		}
		if resp.StatusCode != http.StatusTooManyRequests || attempt == discordMaxAttempts-1 {
			return nil, httpErr
		}

		delay := httpErr.RetryAfter
		if delay <= 0 {
			// Spread fallback retries when Discord does not tell us how long to
			// wait. Server-provided delays are respected exactly.
			delay = backoffDelay(0, discordDefaultRetryIn, 2*discordDefaultRetryIn)
		}
		if delay > discordInlineRetryMax {
			// Do not hold the streamer check open for a long window. The caller
			// persists RetryUntil and retries on a later check.
			return nil, httpErr
		}
		if sleepErr := retrySleep(ctx, delay); sleepErr != nil {
			return nil, sleepErr
		}
	}
	return nil, lastErr
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
