package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//go:embed web
var webFiles embed.FS

type dashboardServer struct {
	app    *App
	server *http.Server
	port   int
	host   string
}

// startDashboard starts the local HTTP server. An empty listen address keeps
// the historical loopback default and port fallback. An explicit address (for
// example 0.0.0.0:8787 in Docker) is bound exactly and fails loudly if busy.
func startDashboard(app *App, onQuit func(), listen string) (*dashboardServer, error) {
	handler := withSecurity(newDashboardHandler(app, onQuit))
	var (
		listener net.Listener
		err      error
	)
	listen = strings.TrimSpace(listen)
	if listen == "" {
		for port := 8787; port <= 8799; port++ {
			listener, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
			if err == nil {
				break
			}
		}
		if listener == nil {
			listener, err = net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				return nil, fmt.Errorf("could not start the local dashboard: %w", err)
			}
		}
	} else {
		address := listen
		if _, _, splitErr := net.SplitHostPort(address); splitErr != nil {
			if port, portErr := strconv.Atoi(strings.TrimPrefix(address, ":")); portErr == nil {
				address = fmt.Sprintf("127.0.0.1:%d", port)
			} else {
				return nil, fmt.Errorf("invalid --listen address %q: use host:port, for example 127.0.0.1:8787 or 0.0.0.0:8787", listen)
			}
		}
		listener, err = net.Listen("tcp", address)
		if err != nil {
			return nil, fmt.Errorf("could not listen on %s: %w", address, err)
		}
	}

	tcpAddress := listener.Addr().(*net.TCPAddr)
	port := tcpAddress.Port
	host := tcpAddress.IP.String()
	if host == "" || host == "::" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		_ = server.Serve(listener)
	}()
	return &dashboardServer{app: app, server: server, port: port, host: host}, nil
}

func (s *dashboardServer) URL() string {
	return fmt.Sprintf("http://%s/", net.JoinHostPort(s.host, strconv.Itoa(s.port)))
}

func (s *dashboardServer) Shutdown(ctx context.Context) {
	_ = s.server.Shutdown(ctx)
}

func runServerOnly(configPath, listen string) int {
	prepareConsole()
	app, err := NewApp(configPath)
	if err != nil {
		return cliError(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app.Start(ctx)

	quit := make(chan struct{}, 1)
	dashboard, err := startDashboard(app, func() {
		select {
		case quit <- struct{}{}:
		default:
		}
	}, listen)
	if err != nil {
		return cliError(err)
	}
	portFile := serverPortPath(configPath)
	_ = writePortFile(portFile, dashboard.port)
	fmt.Printf("%s v%s\n", appName, Version)
	fmt.Println("Config:", configPath)
	fmt.Println("Dashboard:", dashboard.URL())
	fmt.Println("Health:", dashboard.URL()+"healthz")
	fmt.Println("Press Ctrl+C to stop.")

	signals, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSignals()
	select {
	case <-signals.Done():
	case <-quit:
	}
	app.Shutdown()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()
	dashboard.Shutdown(shutdownCtx)
	_ = os.Remove(portFile)
	return 0
}

func newDashboardHandler(app *App, onQuit func()) http.Handler {
	mux := http.NewServeMux()
	staticFS, _ := fs.Sub(webFiles, "web")
	indexHTML, _ := fs.ReadFile(webFiles, "web/index.html")
	favicon, _ := fs.ReadFile(webFiles, "web/icon.png")

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(favicon)
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(indexHTML)
	})

	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, app.State())
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		health := app.Health()
		code := http.StatusOK
		if health.Status != healthStatusHealthy {
			code = http.StatusServiceUnavailable
		}
		writeJSON(w, code, health)
	})

	mux.HandleFunc("POST /api/channels", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		var request struct {
			Login string `json:"login"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		if _, err := app.AddChannel(request.Login); err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, app.State())
	})
	mux.HandleFunc("PUT /api/channels/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		var request struct {
			Enabled   *bool   `json:"enabled"`
			DiscordID *string `json:"discord_id"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		if request.Enabled == nil && request.DiscordID == nil {
			writeAPIError(w, http.StatusBadRequest, errors.New("no changes supplied"))
			return
		}
		if request.Enabled != nil {
			if err := app.SetChannelEnabled(r.PathValue("id"), *request.Enabled); err != nil {
				writeAPIError(w, http.StatusBadRequest, err)
				return
			}
		}
		if request.DiscordID != nil {
			if err := app.SetChannelDiscordID(r.PathValue("id"), *request.DiscordID); err != nil {
				writeAPIError(w, http.StatusBadRequest, err)
				return
			}
		}
		writeJSON(w, http.StatusOK, app.State())
	})
	mux.HandleFunc("DELETE /api/channels/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		if err := app.RemoveChannel(r.PathValue("id")); err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, app.State())
	})
	mux.HandleFunc("POST /api/channels/{id}/check", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		started, err := app.CheckChannelNow(r.PathValue("id"))
		if err != nil {
			writeAPIError(w, http.StatusConflict, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"started": started})
	})
	mux.HandleFunc("PUT /api/channels/{id}/destinations", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		var request struct {
			WebhookIDs []string `json:"webhook_ids"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		if err := app.SetDestinations(r.PathValue("id"), request.WebhookIDs); err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, app.State())
	})
	mux.HandleFunc("POST /api/channels/{id}/sample", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		var request struct {
			WebhookIDs []string `json:"webhook_ids"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		count, err := app.SendSample(r.PathValue("id"), request.WebhookIDs)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"message": fmt.Sprintf("Posted a sample to %d destination(s).", count),
		})
	})

	mux.HandleFunc("POST /api/webhooks", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		var request struct {
			Label    string   `json:"label"`
			URL      string   `json:"url"`
			Template Template `json:"template"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		if _, err := app.AddWebhook(request.Label, request.URL, &request.Template); err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, app.State())
	})
	mux.HandleFunc("PUT /api/webhooks/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		var request struct {
			Label    string   `json:"label"`
			URL      string   `json:"url"`
			Enabled  bool     `json:"enabled"`
			Template Template `json:"template"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		if _, err := app.UpdateWebhook(r.PathValue("id"), request.Label, request.URL, request.Enabled, request.Template); err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, app.State())
	})
	mux.HandleFunc("DELETE /api/webhooks/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		if err := app.RemoveWebhook(r.PathValue("id")); err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, app.State())
	})
	mux.HandleFunc("POST /api/webhooks/{id}/test", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		if _, err := app.TestWebhook(r.PathValue("id")); err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"message": "Test message posted."})
	})

	mux.HandleFunc("POST /api/preview", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		var request struct {
			Template     Template `json:"template"`
			ChannelID    string   `json:"channel_id"`
			WebhookLabel string   `json:"webhook_label"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		payload, unknown := app.Preview(request.Template, request.ChannelID, request.WebhookLabel)
		writeJSON(w, http.StatusOK, map[string]any{
			"payload":              payload,
			"unknown_placeholders": unknown,
		})
	})

	mux.HandleFunc("POST /api/check", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		started := app.TriggerCheck()
		writeJSON(w, http.StatusOK, map[string]any{"started": started})
	})

	mux.HandleFunc("PUT /api/settings", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		var request struct {
			PollSeconds  int    `json:"poll_seconds"`
			StartAtLogin bool   `json:"start_at_login"`
			Paused       bool   `json:"paused"`
			Theme        string `json:"theme"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		if err := app.UpdateSettings(Settings{
			PollSeconds:  request.PollSeconds,
			StartAtLogin: request.StartAtLogin,
			Paused:       request.Paused,
			Theme:        request.Theme,
		}); err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, app.State())
	})

	mux.HandleFunc("POST /api/open-data-folder", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		if err := app.OpenDataFolder(); err != nil {
			writeAPIError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	mux.HandleFunc("POST /api/quit", func(w http.ResponseWriter, r *http.Request) {
		if !requireLocalJSON(w, r) {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		go func() {
			time.Sleep(250 * time.Millisecond)
			onQuit()
		}()
	})

	return mux
}

func withSecurity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' https: data:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; font-src 'self';")
		next.ServeHTTP(w, r)
	})
}

func requireLocalJSON(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		writeAPIError(w, http.StatusUnsupportedMediaType, errors.New("expected application/json"))
		return false
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		parsed, err := url.Parse(origin)
		if err != nil || !strings.EqualFold(parsed.Host, r.Host) {
			writeAPIError(w, http.StatusForbidden, errors.New("cross-origin request blocked"))
			return false
		}
	}
	return true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		writeAPIError(w, http.StatusBadRequest, errors.New("invalid JSON request"))
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)
}

func writeAPIError(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]string{"error": err.Error()})
}

func serverPortPath(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "server.port")
}

func writePortFile(path string, port int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(port)), 0o600)
}

func readPortFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	port, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || port < 1 || port > 65535 {
		return 0, errors.New("invalid dashboard port")
	}
	return port, nil
}
