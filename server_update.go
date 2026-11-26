package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// registerUpdateRoutes adds the install endpoints. The main dashboard handler
// calls this from newDashboardHandler so update.go/server.go stay focused on
// checking and status.
func registerUpdateRoutes(mux *http.ServeMux, app *App) {
	mux.HandleFunc("GET /api/updates/install", func(w http.ResponseWriter, r *http.Request) {
		manager := app.updates
		if manager == nil {
			writeJSON(w, http.StatusOK, map[string]string{"status": updateJobIdle})
			return
		}
		job := manager.Current()
		if job.ID == "" && job.Status == updateJobIdle {
			writeJSON(w, http.StatusOK, map[string]string{"status": updateJobIdle})
			return
		}
		writeJSON(w, http.StatusOK, job)
	})

	mux.HandleFunc("POST /api/updates/install", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		request := struct {
			Restart bool `json:"restart"`
		}{}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&request); err != nil && !errors.Is(err, io.EOF) {
			writeAPIError(w, http.StatusBadRequest, errors.New("invalid JSON request"))
			return
		}
		manager := app.updates
		if manager == nil {
			writeAPIError(w, http.StatusServiceUnavailable, errors.New("the update manager is not ready"))
			return
		}
		job, err := manager.Start(request.Restart)
		if errors.Is(err, errUpdateJobActive) {
			writeAPIError(w, http.StatusConflict, err)
			return
		}
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, job)
	})
}
