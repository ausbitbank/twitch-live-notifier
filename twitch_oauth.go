package main

// Twitch no longer returns a user's followed channels to anonymous GraphQL
// callers, so the optional follows import uses the documented Helix endpoint
// with a user token from the device-code grant. Only a public client ID is
// needed; there is no client secret.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	twitchFollowScope      = "user:read:follows"
	twitchHelixPageSize    = 100
	twitchFollowsMaxPages  = 100
	twitchOAuthClientIDEnv = "TWITCH_LIVE_CLIENT_ID"
)

var (
	// defaultTwitchOAuthClientID is empty in source checkouts. Release builds
	// set it to the project's registered public Twitch application. Client IDs
	// are public identifiers; a public client has no secret.
	defaultTwitchOAuthClientID = ""

	twitchOAuthBaseURL = "https://id.twitch.tv"
	twitchHelixBaseURL = "https://api.twitch.tv"
)

// twitchOAuthClientID returns the public Twitch application client ID used for
// the optional follows import. An empty value disables the feature.
func twitchOAuthClientID() string {
	if value := strings.TrimSpace(os.Getenv(twitchOAuthClientIDEnv)); value != "" {
		return value
	}
	return strings.TrimSpace(defaultTwitchOAuthClientID)
}

type twitchDeviceStart struct {
	DeviceCode      string `json:"device_code"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
}

type twitchToken struct {
	AccessToken  string   `json:"access_token"`
	ExpiresIn    int      `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

const (
	twitchPollAuthorized = "authorized"
	twitchPollPending    = "pending"
	twitchPollSlowDown   = "slow_down"
	twitchPollExpired    = "expired"
	twitchPollDenied     = "denied"
)

type twitchHelixUser struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	ProfileImageURL string `json:"profile_image_url"`
}

type twitchHelixFollow struct {
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName  string `json:"broadcaster_name"`
	FollowedAt       string `json:"followed_at"`
}

type twitchHelixUsersResponse struct {
	Data []twitchHelixUser `json:"data"`
}

type twitchHelixFollowedResponse struct {
	Data       []twitchHelixFollow `json:"data"`
	Total      int                 `json:"total"`
	Pagination struct {
		Cursor string `json:"cursor"`
	} `json:"pagination"`
}

type twitchAPIErrorPayload struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func startTwitchDeviceFlow(ctx context.Context, clientID string) (twitchDeviceStart, error) {
	form := url.Values{
		"client_id": {clientID},
		"scopes":    {twitchFollowScope},
	}
	status, body, err := twitchFormPost(ctx, twitchOAuthBaseURL+"/oauth2/device", form)
	if err != nil {
		return twitchDeviceStart{}, err
	}
	if status != http.StatusOK {
		message := twitchAPIErrorMessage(body, fmt.Sprintf("Twitch returned HTTP %d", status))
		if strings.Contains(strings.ToLower(message), "invalid client") {
			return twitchDeviceStart{}, fmt.Errorf("Twitch rejected the configured client ID. Register a public Twitch app and set %s.", twitchOAuthClientIDEnv)
		}
		return twitchDeviceStart{}, fmt.Errorf("Twitch could not start the sign-in: %s", message)
	}

	var start twitchDeviceStart
	if err := json.Unmarshal(body, &start); err != nil {
		return twitchDeviceStart{}, errors.New("Twitch returned an invalid device-code response")
	}
	if strings.TrimSpace(start.DeviceCode) == "" || strings.TrimSpace(start.UserCode) == "" {
		return twitchDeviceStart{}, errors.New("Twitch returned an incomplete device-code response")
	}
	if start.Interval < 1 {
		start.Interval = 5
	}
	if start.ExpiresIn < 1 {
		start.ExpiresIn = 1800
	}
	return start, nil
}

func pollTwitchDeviceToken(ctx context.Context, clientID, deviceCode string) (twitchToken, string, bool, error) {
	form := url.Values{
		"client_id":   {clientID},
		"scopes":      {twitchFollowScope},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	}
	status, body, err := twitchFormPost(ctx, twitchOAuthBaseURL+"/oauth2/token", form)
	if err != nil {
		return twitchToken{}, "", true, err
	}
	if status == http.StatusOK {
		var token twitchToken
		if err := json.Unmarshal(body, &token); err != nil || strings.TrimSpace(token.AccessToken) == "" {
			return twitchToken{}, "", false, errors.New("Twitch returned an invalid token response")
		}
		return token, twitchPollAuthorized, false, nil
	}

	message := strings.ToLower(twitchAPIErrorMessage(body, ""))
	switch {
	case strings.Contains(message, "authorization_pending"):
		return twitchToken{}, twitchPollPending, false, nil
	case strings.Contains(message, "slow_down"):
		return twitchToken{}, twitchPollSlowDown, false, nil
	case strings.Contains(message, "expired"):
		return twitchToken{}, twitchPollExpired, false, nil
	case strings.Contains(message, "denied") || strings.Contains(message, "access_denied"):
		return twitchToken{}, twitchPollDenied, false, nil
	}
	if message == "" {
		message = fmt.Sprintf("Twitch returned HTTP %d", status)
	}
	retryable := status >= 500 || status == http.StatusTooManyRequests
	return twitchToken{}, "", retryable, fmt.Errorf("Twitch sign-in failed: %s", message)
}

func twitchFormPost(ctx context.Context, endpoint string, form url.Values) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return 0, nil, errors.New("could not prepare the Twitch sign-in request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("could not reach Twitch: %s", networkErrorText(err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resp.StatusCode, nil, errors.New("could not read the Twitch response")
	}
	return resp.StatusCode, body, nil
}

func twitchAPIErrorMessage(body []byte, fallback string) string {
	var payload twitchAPIErrorPayload
	if err := json.Unmarshal(body, &payload); err == nil {
		if message := strings.TrimSpace(payload.Message); message != "" {
			return message
		}
	}
	return fallback
}

func twitchHelixGet(ctx context.Context, clientID, accessToken, path string, query url.Values) ([]byte, error) {
	endpoint := twitchHelixBaseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, errors.New("could not prepare the Twitch request")
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach Twitch: %s", networkErrorText(err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, errors.New("could not read the Twitch response")
	}
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return nil, errors.New("the Twitch sign-in is no longer valid. Connect again.")
	case resp.StatusCode == http.StatusTooManyRequests:
		return nil, errors.New("Twitch rate limited the import. Wait a minute and try again.")
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		message := twitchAPIErrorMessage(body, fmt.Sprintf("HTTP %d", resp.StatusCode))
		return nil, fmt.Errorf("Twitch API error: %s", message)
	}
	return body, nil
}

func fetchTwitchHelixUser(ctx context.Context, clientID, accessToken string) (twitchHelixUser, error) {
	body, err := twitchHelixGet(ctx, clientID, accessToken, "/helix/users", nil)
	if err != nil {
		return twitchHelixUser{}, err
	}
	var payload twitchHelixUsersResponse
	if err := json.Unmarshal(body, &payload); err != nil || len(payload.Data) == 0 {
		return twitchHelixUser{}, errors.New("Twitch did not return the signed-in account")
	}
	return payload.Data[0], nil
}

func fetchTwitchHelixFollows(ctx context.Context, clientID, accessToken, userID string) ([]twitchHelixFollow, int, error) {
	follows := make([]twitchHelixFollow, 0, 100)
	total := 0
	cursor := ""
	for page := 0; page < twitchFollowsMaxPages; page++ {
		query := url.Values{
			"user_id": {userID},
			"first":   {strconv.Itoa(twitchHelixPageSize)},
		}
		if cursor != "" {
			query.Set("after", cursor)
		}
		body, err := twitchHelixGet(ctx, clientID, accessToken, "/helix/channels/followed", query)
		if err != nil {
			return nil, 0, err
		}
		var payload twitchHelixFollowedResponse
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, 0, errors.New("Twitch returned an invalid follows response")
		}
		if payload.Total > total {
			total = payload.Total
		}
		follows = append(follows, payload.Data...)
		cursor = strings.TrimSpace(payload.Pagination.Cursor)
		if cursor == "" || len(payload.Data) == 0 {
			break
		}
	}
	return follows, total, nil
}

// fetchTwitchHelixProfiles batches profile lookups so imported channels get
// display names and avatars without one GraphQL request per channel. Partial
// results are returned if a batch fails.
func fetchTwitchHelixProfiles(ctx context.Context, clientID, accessToken string, logins []string) map[string]twitchHelixUser {
	profiles := make(map[string]twitchHelixUser, len(logins))
	for start := 0; start < len(logins); start += 100 {
		end := start + 100
		if end > len(logins) {
			end = len(logins)
		}
		query := url.Values{}
		for _, login := range logins[start:end] {
			query.Add("login", login)
		}
		body, err := twitchHelixGet(ctx, clientID, accessToken, "/helix/users", query)
		if err != nil {
			continue
		}
		var payload twitchHelixUsersResponse
		if err := json.Unmarshal(body, &payload); err != nil {
			continue
		}
		for _, user := range payload.Data {
			profiles[strings.ToLower(user.Login)] = user
		}
	}
	return profiles
}
