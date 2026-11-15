package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const liveTwitchJSON = `{
  "data": {
    "user": {
      "login": "teststreamer",
      "displayName": "TestStreamer",
      "profileImageURL": "https://example.com/avatar.png",
      "bannerImageURL": "https://example.com/banner.png",
      "stream": {
        "id": "319913347671",
        "title": "Testing the live notifier",
        "previewImageURL": "https://example.com/preview.jpg",
        "viewersCount": 42,
        "createdAt": "2026-11-15T06:00:00Z",
        "game": {"name": "Just Chatting"}
      },
      "lastBroadcast": {
        "id": "316308419689",
        "title": "Karaoke time",
        "startedAt": "2026-11-14T17:56:38Z",
        "game": {"id": "26936", "name": "Music", "boxArtURL": "https://example.com/music.jpg"}
      }
    }
  }
}`

const offlineTwitchJSON = `{
  "data": {
    "user": {
      "login": "teststreamer",
      "displayName": "TestStreamer",
      "profileImageURL": "https://example.com/avatar.png",
      "bannerImageURL": "https://example.com/banner.png",
      "stream": null,
      "lastBroadcast": {
        "id": "316308419689",
        "title": "Karaoke time",
        "startedAt": "2026-11-14T17:56:38Z",
        "game": {"id": "26936", "name": "Music", "boxArtURL": "https://example.com/music.jpg"}
      }
    }
  }
}`

func newTwitchTestServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Client-ID") != twitchWebClientID {
			t.Errorf("missing Twitch Client-ID header")
		}
		var request struct {
			Query     string            `json:"query"`
			Variables map[string]string `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode Twitch request: %v", err)
		}
		if request.Variables["login"] == "" {
			t.Errorf("missing login variable")
		}
		if !strings.Contains(request.Query, "previewImageURL") || !strings.Contains(request.Query, "lastBroadcast") {
			t.Errorf("query is missing expected fields")
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, body)
	}))
	t.Cleanup(server.Close)
	return server
}

func setTwitchTestURL(t *testing.T, server *httptest.Server) {
	t.Helper()
	old := twitchGQLURL
	twitchGQLURL = server.URL
	t.Cleanup(func() { twitchGQLURL = old })
}

func TestFetchTwitchChannel(t *testing.T) {
	server := newTwitchTestServer(t, liveTwitchJSON)
	setTwitchTestURL(t, server)

	user, err := fetchTwitchChannel("https://www.twitch.tv/TestStreamer")
	if err != nil {
		t.Fatalf("fetchTwitchChannel returned error: %v", err)
	}
	if user.DisplayName != "TestStreamer" {
		t.Errorf("display name = %q", user.DisplayName)
	}
	if user.Stream == nil || user.Stream.Title != "Testing the live notifier" {
		t.Fatalf("stream = %+v", user.Stream)
	}
	if user.LastBroadcast == nil || user.LastBroadcast.Game.Name != "Music" {
		t.Fatalf("last broadcast = %+v", user.LastBroadcast)
	}
}

func TestFetchTwitchChannelOffline(t *testing.T) {
	server := newTwitchTestServer(t, offlineTwitchJSON)
	setTwitchTestURL(t, server)

	user, err := fetchTwitchChannel("teststreamer")
	if err != nil {
		t.Fatalf("fetchTwitchChannel returned error: %v", err)
	}
	if user.Stream != nil {
		t.Fatalf("stream = %+v, want nil", user.Stream)
	}
	if user.LastBroadcast == nil {
		t.Fatal("last broadcast is nil")
	}
}

func TestFetchTwitchChannelNotFound(t *testing.T) {
	server := newTwitchTestServer(t, `{"data":{"user":null}}`)
	setTwitchTestURL(t, server)

	if _, err := fetchTwitchChannel("nobody"); err == nil {
		t.Fatal("fetchTwitchChannel returned nil error")
	}
}

func TestFetchTwitchChannelAPIError(t *testing.T) {
	server := newTwitchTestServer(t, `{"data":null,"errors":[{"message":"service error"}]}`)
	setTwitchTestURL(t, server)

	if _, err := fetchTwitchChannel("someone"); err == nil || !strings.Contains(err.Error(), "service error") {
		t.Fatalf("error = %v", err)
	}
}
