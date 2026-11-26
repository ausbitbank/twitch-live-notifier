package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestV25RoutesAreRegistered(t *testing.T) {
	app := newTestApp(t)
	server := httptest.NewServer(withSecurity(newDashboardHandler(app, func() {})))
	defer server.Close()

	for _, path := range []string{
		"/follows-helper",
		"/static/follows-bookmarklet.js",
		"/api/history",
		"/api/diagnostics",
		"/api/updates/install",
	} {
		response, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("GET %s status = %d: %s", path, response.StatusCode, body)
		}
	}

	helper, err := http.Get(server.URL + "/follows-helper")
	if err != nil {
		t.Fatal(err)
	}
	helperBody, _ := io.ReadAll(helper.Body)
	helper.Body.Close()
	if !strings.Contains(string(helperBody), "Export follows") {
		t.Fatal("helper page does not contain the bookmarklet button")
	}

	script, err := http.Get(server.URL + "/static/follows-bookmarklet.js")
	if err != nil {
		t.Fatal(err)
	}
	scriptBody, _ := io.ReadAll(script.Body)
	script.Body.Close()
	if !strings.Contains(string(scriptBody), "tln-follows-export") {
		t.Fatal("bookmarklet source is missing its overlay id")
	}
}

func TestV25ImportAndBulkRoutesValidateInput(t *testing.T) {
	app := newTestApp(t)
	server := httptest.NewServer(withSecurity(newDashboardHandler(app, func() {})))
	defer server.Close()

	post := func(path, body string) *http.Response {
		request, _ := http.NewRequest(http.MethodPost, server.URL+path, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatalf("POST %s: %v", path, err)
		}
		return response
	}

	response := post("/api/channels/import", `{"text":""}`)
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty import status = %d", response.StatusCode)
	}

	response = post("/api/channels/bulk-action", `{"action":"assign","channel_ids":["missing"]}`)
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("bulk action without destinations status = %d", response.StatusCode)
	}

	response = post("/api/webhooks/inspect", `{"url":"https://example.com/not-a-webhook"}`)
	response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("inspect invalid URL status = %d", response.StatusCode)
	}
}
