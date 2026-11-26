package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type webhookInspectRequest struct {
	URL string `json:"url"`
}

type webhookInspectResponse struct {
	Name      string `json:"name,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
	GuildID   string `json:"guild_id,omitempty"`
}

// registerWebhookInspectRoutes adds the onboarding helper that reads a
// webhook's non-secret display metadata. It never posts and never returns the
// webhook URL or token.
func registerWebhookInspectRoutes(mux *http.ServeMux, app *App) {
	mux.HandleFunc("POST /api/webhooks/inspect", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		var request webhookInspectRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		webhookURL, err := normalizeWebhookURL(request.URL)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
		defer cancel()
		info, err := inspectDiscordWebhook(ctx, webhookURL)
		if err != nil {
			writeAPIError(w, http.StatusBadGateway, err)
			return
		}
		writeJSON(w, http.StatusOK, info)
	})
}

func inspectDiscordWebhook(ctx context.Context, webhookURL string) (webhookInspectResponse, error) {
	body, err := discordRequest(ctx, http.MethodGet, webhookURL, nil)
	if err != nil {
		return webhookInspectResponse{}, err
	}
	return parseDiscordWebhookInfo(body)
}

func parseDiscordWebhookInfo(body []byte) (webhookInspectResponse, error) {
	var payload struct {
		Name      string `json:"name"`
		ChannelID string `json:"channel_id"`
		GuildID   string `json:"guild_id"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return webhookInspectResponse{}, errors.New("Discord returned an invalid webhook response")
	}
	info := webhookInspectResponse{
		Name:      strings.TrimSpace(payload.Name),
		ChannelID: strings.TrimSpace(payload.ChannelID),
		GuildID:   strings.TrimSpace(payload.GuildID),
	}
	if info.Name == "" && info.ChannelID == "" && info.GuildID == "" {
		return webhookInspectResponse{}, errors.New("Discord did not return webhook details")
	}
	return info, nil
}
