package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
)

func hashDashboardToken(token string) string {
	sum := sha256.Sum256([]byte(appName + ":dashboard:" + token))
	return hex.EncodeToString(sum[:])
}

func (a *App) SetDashboardToken(token string) error {
	token = strings.TrimSpace(token)
	if token != "" && len([]rune(token)) < 4 {
		return errors.New("use at least four characters for the dashboard token")
	}
	a.mu.Lock()
	if token == "" {
		a.cfg.Settings.DashboardTokenHash = ""
	} else {
		a.cfg.Settings.DashboardTokenHash = hashDashboardToken(token)
	}
	err := a.persistLocked()
	a.mu.Unlock()
	if err != nil {
		return err
	}
	if token == "" {
		a.addActivity("info", "Dashboard token removed.")
	} else {
		a.addActivity("info", "Dashboard token enabled.")
	}
	return nil
}

func (a *App) DashboardTokenRequired() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cfg.Settings.DashboardTokenHash != ""
}

func (a *App) dashboardTokenValid(r *http.Request) bool {
	a.mu.RLock()
	expected := a.cfg.Settings.DashboardTokenHash
	a.mu.RUnlock()
	if expected == "" {
		return true
	}
	token := r.Header.Get("X-Dashboard-Token")
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	if token == "" {
		return false
	}
	actual := hashDashboardToken(token)
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

// withDashboardAuth protects the local JSON API when a dashboard token is
// configured. Static assets, the dashboard shell, and /healthz stay reachable
// so the page can prompt for the token and Docker can keep health-checking.
func withDashboardAuth(app *App, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") && !app.dashboardTokenValid(r) {
			writeAPIError(w, http.StatusUnauthorized, errors.New("dashboard token required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
