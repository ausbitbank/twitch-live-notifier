package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type importChannelsRequest struct {
	Text       string   `json:"text"`
	Logins     []string `json:"logins"`
	WebhookIDs []string `json:"webhook_ids"`
	Source     string   `json:"source"`
}

// registerImportRoutes adds the no-login import endpoint. The main dashboard
// handler calls this during route registration.
func registerImportRoutes(mux *http.ServeMux, app *App) {
	mux.HandleFunc("POST /api/channels/import", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}

		// The text field may carry up to 2 MB; allow JSON overhead on top.
		r.Body = http.MaxBytesReader(w, r.Body, int64(bulkImportMaxTextBytes)+(1<<20))
		var request importChannelsRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				writeAPIError(w, http.StatusBadRequest, errors.New("the pasted list is too large (max 2 MB)"))
				return
			}
			writeAPIError(w, http.StatusBadRequest, errors.New("invalid JSON request"))
			return
		}

		entries := append([]string(nil), request.Logins...)
		if strings.TrimSpace(request.Text) != "" {
			entries = append(entries, request.Text)
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()
		result, err := app.ImportChannels(ctx, entries, request.WebhookIDs)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"result": result, "state": app.State()})
	})
}
