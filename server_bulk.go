package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type bulkActionResponse struct {
	Updated int `json:"updated"`
	Removed int `json:"removed"`
}

type bulkActionRequest struct {
	ChannelIDs []string `json:"channel_ids"`
	Action     string   `json:"action"`
	WebhookIDs []string `json:"webhook_ids"`
}

// registerBulkRoutes adds the multi-select channel action endpoint. The main
// dashboard handler calls this during route registration.
func registerBulkRoutes(mux *http.ServeMux, app *App) {
	mux.HandleFunc("POST /api/channels/bulk-action", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		var request bulkActionRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		result, err := app.BulkChannelAction(request.Action, request.ChannelIDs, request.WebhookIDs)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"updated": result.Updated,
			"removed": result.Removed,
			"state":   app.State(),
		})
	})
}

// BulkChannelAction applies one action to many channels and persists once.
// Deletes also drop the in-memory status. A check is triggered only for
// enable and assign, matching the single-channel endpoints.
func (a *App) BulkChannelAction(action string, channelIDs []string, webhookIDs []string) (bulkActionResponse, error) {
	action = strings.ToLower(strings.TrimSpace(action))
	switch action {
	case "enable", "disable", "delete", "assign", "unassign":
	default:
		return bulkActionResponse{}, errors.New("action must be enable, disable, delete, assign, or unassign")
	}

	ids := make([]string, 0, len(channelIDs))
	seenIDs := make(map[string]bool, len(channelIDs))
	for _, channelID := range channelIDs {
		channelID = strings.TrimSpace(channelID)
		if channelID == "" || seenIDs[channelID] {
			continue
		}
		seenIDs[channelID] = true
		ids = append(ids, channelID)
	}
	if len(ids) == 0 {
		return bulkActionResponse{}, errors.New("select at least one streamer")
	}

	var assignedWebhooks []string
	if action == "assign" || action == "unassign" {
		if len(webhookIDs) == 0 {
			return bulkActionResponse{}, errors.New("choose at least one destination")
		}
		selected, _, err := a.importDestinations(webhookIDs)
		if err != nil {
			return bulkActionResponse{}, err
		}
		if len(selected) == 0 {
			return bulkActionResponse{}, errors.New("choose at least one destination")
		}
		assignedWebhooks = selected
	}

	a.mu.Lock()
	backupChannels := append([]Channel(nil), a.cfg.Channels...)
	backupStatuses := make(map[string]ChannelStatus, len(a.statuses))
	for channelID, status := range a.statuses {
		backupStatuses[channelID] = status
	}

	response := bulkActionResponse{}
	changed := false
	switch action {
	case "enable", "disable":
		want := action == "enable"
		for _, channelID := range ids {
			index := a.channelIndexLocked(channelID)
			if index < 0 {
				continue
			}
			response.Updated++
			if a.cfg.Channels[index].Enabled != want {
				a.cfg.Channels[index].Enabled = want
				changed = true
			}
		}
	case "delete":
		selected := make(map[string]bool, len(ids))
		for _, channelID := range ids {
			selected[channelID] = true
		}
		remaining := make([]Channel, 0, len(a.cfg.Channels))
		for _, channel := range a.cfg.Channels {
			if selected[channel.ID] {
				delete(a.statuses, channel.ID)
				response.Removed++
				changed = true
				continue
			}
			remaining = append(remaining, channel)
		}
		a.cfg.Channels = remaining
	case "assign", "unassign":
		selected := make(map[string]bool, len(assignedWebhooks))
		for _, webhookID := range assignedWebhooks {
			selected[webhookID] = true
		}
		for _, channelID := range ids {
			index := a.channelIndexLocked(channelID)
			if index < 0 {
				continue
			}
			response.Updated++
			destinations := append([]Destination(nil), a.cfg.Channels[index].Destinations...)
			if action == "assign" {
				existing := make(map[string]bool, len(destinations))
				for _, destination := range destinations {
					existing[destination.WebhookID] = true
				}
				for _, webhookID := range assignedWebhooks {
					if existing[webhookID] {
						continue
					}
					existing[webhookID] = true
					destinations = append(destinations, Destination{WebhookID: webhookID})
					changed = true
				}
			} else {
				filtered := make([]Destination, 0, len(destinations))
				for _, destination := range destinations {
					if selected[destination.WebhookID] {
						changed = true
						continue
					}
					filtered = append(filtered, destination)
				}
				destinations = filtered
			}
			a.cfg.Channels[index].Destinations = destinations
		}
	}

	if !changed {
		a.mu.Unlock()
		return response, nil
	}
	if err := a.persistLocked(); err != nil {
		a.cfg.Channels = backupChannels
		a.statuses = backupStatuses
		a.mu.Unlock()
		return bulkActionResponse{}, err
	}
	a.mu.Unlock()

	a.addActivity("info", bulkActionMessage(action, response.Updated, response.Removed, len(assignedWebhooks)))
	if a.enableAutoCheck && (action == "enable" || action == "assign") {
		a.TriggerCheck()
	}
	return response, nil
}

func bulkActionMessage(action string, updated, removed, destinations int) string {
	switch action {
	case "enable":
		return fmt.Sprintf("Enabled %d streamer%s.", updated, pluralSuffix(updated))
	case "disable":
		return fmt.Sprintf("Disabled %d streamer%s.", updated, pluralSuffix(updated))
	case "delete":
		return fmt.Sprintf("Removed %d streamer%s.", removed, pluralSuffix(removed))
	case "assign":
		return fmt.Sprintf("Assigned %d streamer%s to %d destination%s.", updated, pluralSuffix(updated), destinations, pluralSuffix(destinations))
	case "unassign":
		return fmt.Sprintf("Unassigned %d streamer%s from %d destination%s.", updated, pluralSuffix(updated), destinations, pluralSuffix(destinations))
	}
	return "Updated streamers."
}
