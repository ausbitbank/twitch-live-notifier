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
	"time"
)

const (
	twitchWebClientID = "kimne78kx3ncx6brgo4mv6wki5h1ko"
	userAgent         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

const twitchGQLQuery = `query ChannelData($login: String!) {
  user(login: $login) {
    login
    displayName
    profileImageURL(width: 600)
    bannerImageURL
    stream {
      id
      title
      previewImageURL(width: 1280, height: 720)
      viewersCount
      createdAt
      game {
        name
      }
    }
    lastBroadcast {
      id
      title
      startedAt
      game {
        id
        name
        boxArtURL(width: 600, height: 800)
      }
    }
  }
}`

var (
	twitchGQLURL = "https://gql.twitch.tv/gql"
	httpClient   = &http.Client{Timeout: 20 * time.Second}
)

type twitchGame struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"boxArtURL"`
}

type twitchStream struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	PreviewImageURL string     `json:"previewImageURL"`
	ViewersCount    int        `json:"viewersCount"`
	CreatedAt       string     `json:"createdAt"`
	Game            twitchGame `json:"game"`
}

type twitchBroadcast struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	StartedAt string     `json:"startedAt"`
	Game      twitchGame `json:"game"`
}

type twitchUser struct {
	Login           string           `json:"login"`
	DisplayName     string           `json:"displayName"`
	ProfileImageURL string           `json:"profileImageURL"`
	BannerImageURL  string           `json:"bannerImageURL"`
	Stream          *twitchStream    `json:"stream"`
	LastBroadcast   *twitchBroadcast `json:"lastBroadcast"`
}

type twitchGQLResponse struct {
	Data struct {
		User *twitchUser `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func fetchTwitchChannel(login string) (*twitchUser, error) {
	login, err := normalizeTwitchLogin(login)
	if err != nil {
		return nil, err
	}

	requestBody, err := json.Marshal(map[string]any{
		"query": twitchGQLQuery,
		"variables": map[string]string{
			"login": login,
		},
	})
	if err != nil {
		return nil, errors.New("could not prepare the Twitch request")
	}

	req, err := http.NewRequest(http.MethodPost, twitchGQLURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, errors.New("could not prepare the Twitch request")
	}
	req.Header.Set("Client-ID", twitchWebClientID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", "https://www.twitch.tv")
	req.Header.Set("Referer", "https://www.twitch.tv/"+login)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach Twitch: %s", networkErrorText(err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, errors.New("could not read the Twitch response")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Twitch returned HTTP %d", resp.StatusCode)
	}

	var parsed twitchGQLResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, errors.New("Twitch returned an invalid response")
	}
	if len(parsed.Errors) > 0 {
		message := strings.TrimSpace(parsed.Errors[0].Message)
		if message == "" {
			message = "unknown API error"
		}
		return nil, fmt.Errorf("Twitch API error: %s", message)
	}
	if parsed.Data.User == nil {
		return nil, fmt.Errorf("Twitch channel %q was not found", login)
	}
	if strings.TrimSpace(parsed.Data.User.DisplayName) == "" {
		parsed.Data.User.DisplayName = login
	}
	return parsed.Data.User, nil
}

func networkErrorText(err error) string {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return "request timed out"
		}
		if urlErr.Err != nil {
			return urlErr.Err.Error()
		}
		return "network error"
	}
	return "network error"
}
