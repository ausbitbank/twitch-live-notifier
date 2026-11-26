package main

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// registerHistoryRoutes wires the delivery-history and diagnostics endpoints.
// The integration lane calls this from newDashboardHandler.
func registerHistoryRoutes(mux *http.ServeMux, app *App) {
	if mux == nil {
		return
	}

	mux.HandleFunc("GET /api/history", func(w http.ResponseWriter, r *http.Request) {
		limit := historyLimit(r.URL.Query().Get("limit"))
		filters := DeliveryFilters{
			Kind:      r.URL.Query().Get("kind"),
			Level:     r.URL.Query().Get("level"),
			ChannelID: r.URL.Query().Get("channel_id"),
			Query:     firstNonEmpty(r.URL.Query().Get("q"), r.URL.Query().Get("query"), r.URL.Query().Get("search")),
		}
		entries, total := appHistoryResolver(app).List(limit, filters)
		if entries == nil {
			entries = []DeliveryRecord{}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"entries": entries,
			"total":   total,
			"limit":   limit,
		})
	})

	mux.HandleFunc("DELETE /api/history", func(w http.ResponseWriter, r *http.Request) {
		// Clearing history has no request body, so do not insist on a JSON
		// Content-Type. Still block a cross-origin browser request.
		if !historyOriginAllowed(w, r) {
			return
		}
		if err := appHistoryResolver(app).Clear(); err != nil {
			writeAPIError(w, http.StatusInternalServerError, errors.New("could not clear the delivery history"))
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	mux.HandleFunc("GET /api/diagnostics", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, buildDiagnostics(app, appHistoryResolver(app)))
	})
}

// historyLimit parses the limit query parameter, falling back to the dashboard
// default and clamping to the persisted cap.
func historyLimit(raw string) int {
	limit := defaultHistoryListLimit
	if value, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil {
		limit = value
	}
	if limit <= 0 {
		limit = defaultHistoryListLimit
	}
	if limit > maxHistoryListLimit {
		limit = maxHistoryListLimit
	}
	return limit
}

// historyOriginAllowed mirrors the same-origin check used by requireLocalJSON
// without requiring a JSON body.
func historyOriginAllowed(w http.ResponseWriter, r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || !strings.EqualFold(parsed.Host, r.Host) {
		writeAPIError(w, http.StatusForbidden, errors.New("cross-origin request blocked"))
		return false
	}
	return true
}
