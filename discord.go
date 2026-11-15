package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type discordWebhookPayload struct {
	Content         string                  `json:"content,omitempty"`
	Embeds          []discordEmbed          `json:"embeds,omitempty"`
	AllowedMentions *discordAllowedMentions `json:"allowed_mentions,omitempty"`
}

type discordAllowedMentions struct {
	Parse []string `json:"parse,omitempty"`
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

func postDiscord(webhookURL string, payload discordWebhookPayload) (discordPostResult, error) {
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

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(body))
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
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return discordPostResult{}, fmt.Errorf("Discord returned HTTP %d%s", resp.StatusCode, discordErrorDetails(respBody))
	}

	var posted struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(respBody, &posted)
	return discordPostResult{MessageID: posted.ID}, nil
}

func discordErrorDetails(body []byte) string {
	var apiError struct {
		Message    string  `json:"message"`
		RetryAfter float64 `json:"retry_after"`
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
	if apiError.RetryAfter > 0 {
		message = fmt.Sprintf("%s (retry after %.1fs)", message, apiError.RetryAfter)
	}
	if len(message) > 300 {
		message = message[:300] + "..."
	}
	return ": " + message
}
