package main

import (
	"bytes"
	"context"
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

const (
	twitchMaxAttempts = 3
)

// twitchUserFields is shared by the single-channel query and the batched
// alias query so both paths always request the same profile, stream, and
// archive fields.
const twitchUserFields = `    login
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
    videos(first: 1, type: ARCHIVE) {
      edges {
        node {
          id
        }
      }
    }`

const twitchGQLQuery = `query ChannelData($login: String!) {
  user(login: $login) {
` + twitchUserFields + `
  }
}`

// errTwitchChannelNotFound is used by the batch path so callers can report a
// consistent "channel not found" reason without leaking the alias index.
var errTwitchChannelNotFound = errors.New("Twitch channel not found")

var (
	twitchGQLURL    = "https://gql.twitch.tv/gql"
	httpClient      = &http.Client{Timeout: 15 * time.Second}
	twitchRetryBase = 400 * time.Millisecond
	twitchRetryMax  = 5 * time.Second
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

type twitchVideo struct {
	ID string `json:"id"`
}

type twitchUser struct {
	Login           string           `json:"login"`
	DisplayName     string           `json:"displayName"`
	ProfileImageURL string           `json:"profileImageURL"`
	BannerImageURL  string           `json:"bannerImageURL"`
	Stream          *twitchStream    `json:"stream"`
	LastBroadcast   *twitchBroadcast `json:"lastBroadcast"`
	Videos          struct {
		Edges []struct {
			Node twitchVideo `json:"node"`
		} `json:"edges"`
	} `json:"videos"`
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
	return fetchTwitchChannelContext(context.Background(), login)
}

func fetchTwitchChannelContext(ctx context.Context, login string) (*twitchUser, error) {
	login, err := normalizeTwitchLogin(login)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt < twitchMaxAttempts; attempt++ {
		user, retryAfter, retryable, err := fetchTwitchOnce(ctx, login)
		if err == nil {
			return user, nil
		}
		lastErr = err
		if !retryable || attempt == twitchMaxAttempts-1 {
			break
		}
		delay := backoffDelay(attempt, twitchRetryBase, twitchRetryMax)
		if retryAfter > delay {
			delay = retryAfter
		}
		if delay > twitchRetryMax {
			delay = twitchRetryMax
		}
		if sleepErr := retrySleep(ctx, delay); sleepErr != nil {
			return nil, sleepErr
		}
	}
	return nil, lastErr
}

func fetchTwitchOnce(ctx context.Context, login string) (*twitchUser, time.Duration, bool, error) {
	requestBody, err := json.Marshal(map[string]any{
		"query": twitchGQLQuery,
		"variables": map[string]string{
			"login": login,
		},
	})
	if err != nil {
		return nil, 0, false, errors.New("could not prepare the Twitch request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, twitchGQLURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, 0, false, errors.New("could not prepare the Twitch request")
	}
	req.Header.Set("Client-ID", twitchWebClientID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Origin", "https://www.twitch.tv")
	req.Header.Set("Referer", "https://www.twitch.tv/"+login)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, true, fmt.Errorf("could not reach Twitch: %s", networkErrorText(err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, 0, true, errors.New("could not read the Twitch response")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		retryAfter := retryAfterFromHeaders(resp.Header.Get("Retry-After"))
		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, retryAfter, true, fmt.Errorf("Twitch rate limited the request (HTTP 429)")
		}
		return nil, retryAfter, retryable, fmt.Errorf("Twitch returned HTTP %d", resp.StatusCode)
	}

	var parsed twitchGQLResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, 0, false, errors.New("Twitch returned an invalid response")
	}
	if len(parsed.Errors) > 0 {
		message := strings.TrimSpace(parsed.Errors[0].Message)
		if message == "" {
			message = "unknown API error"
		}
		return nil, 0, false, fmt.Errorf("Twitch API error: %s", message)
	}
	if parsed.Data.User == nil {
		return nil, 0, false, fmt.Errorf("Twitch channel %q was not found", login)
	}
	if strings.TrimSpace(parsed.Data.User.DisplayName) == "" {
		parsed.Data.User.DisplayName = login
	}
	return parsed.Data.User, 0, false, nil
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
