package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	twitchImportPending     = "pending"
	twitchImportAuthorizing = "authorizing"
	twitchImportReady       = "ready"
	twitchImportExpired     = "expired"
	twitchImportDenied      = "denied"
	twitchImportError       = "error"

	twitchImportSessionTTL   = 30 * time.Minute
	twitchImportMaxSessions  = 5
	twitchImportMaxSelection = 1000

	// bulkImportMaxLogins and bulkImportMaxTextBytes bound a single import
	// request. The legacy /api/channels/bulk endpoint shares the same cap.
	bulkImportMaxLogins    = 2000
	bulkImportMaxTextBytes = 2 << 20
	bulkAddMaxLogins       = bulkImportMaxLogins
)

type twitchAccountView struct {
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type twitchFollowView struct {
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	FollowedAt  string `json:"followed_at,omitempty"`
	Added       bool   `json:"added"`
}

type twitchFollowRecord struct {
	Login       string
	DisplayName string
	AvatarURL   string
	FollowedAt  string
}

type twitchImportView struct {
	ID              string             `json:"id"`
	Status          string             `json:"status"`
	UserCode        string             `json:"user_code,omitempty"`
	VerificationURI string             `json:"verification_uri,omitempty"`
	ExpiresIn       int                `json:"expires_in,omitempty"`
	Interval        int                `json:"interval,omitempty"`
	Account         *twitchAccountView `json:"account,omitempty"`
	Follows         []twitchFollowView `json:"follows,omitempty"`
	Total           int                `json:"total,omitempty"`
	Message         string             `json:"message,omitempty"`
}

type twitchImportSession struct {
	ID              string
	DeviceCode      string
	UserCode        string
	VerificationURI string
	Interval        time.Duration
	ExpiresAt       time.Time
	NextPoll        time.Time
	CreatedAt       time.Time
	Status          string
	Message         string
	Polling         bool
	Account         *twitchAccountView
	Follows         []twitchFollowRecord
	Total           int
}

type bulkAddIssue struct {
	Login  string `json:"login"`
	Reason string `json:"reason"`
}

type bulkAddResponse struct {
	Added    int            `json:"added"`
	Assigned int            `json:"assigned"`
	Total    int            `json:"total"`
	Skipped  []bulkAddIssue `json:"skipped,omitempty"`
	Failed   []bulkAddIssue `json:"failed,omitempty"`
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func (a *App) channelLogins() map[string]bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	logins := make(map[string]bool, len(a.cfg.Channels))
	for _, channel := range a.cfg.Channels {
		logins[strings.ToLower(channel.Login)] = true
	}
	return logins
}

// twitchImportViewLocked builds a client-safe view. Callers must hold
// a.twitchMu. It reads channels under a.mu; no code path locks these in the
// opposite order.
func (a *App) twitchImportViewLocked(session *twitchImportSession) twitchImportView {
	existing := a.channelLogins()
	view := twitchImportView{
		ID:              session.ID,
		Status:          session.Status,
		UserCode:        session.UserCode,
		VerificationURI: session.VerificationURI,
		Interval:        int(session.Interval / time.Second),
		Account:         session.Account,
		Total:           session.Total,
		Message:         session.Message,
	}
	if session.Status == twitchImportAuthorizing {
		view.Status = twitchImportPending
		view.Message = "Loading your follows…"
	}
	if view.Interval < 1 {
		view.Interval = 5
	}
	if !session.ExpiresAt.IsZero() {
		remaining := int(time.Until(session.ExpiresAt).Seconds())
		if remaining < 0 {
			remaining = 0
		}
		view.ExpiresIn = remaining
	}
	if session.Status == twitchImportReady {
		view.Follows = make([]twitchFollowView, 0, len(session.Follows))
		for _, follow := range session.Follows {
			view.Follows = append(view.Follows, twitchFollowView{
				Login:       follow.Login,
				DisplayName: follow.DisplayName,
				AvatarURL:   follow.AvatarURL,
				FollowedAt:  follow.FollowedAt,
				Added:       existing[strings.ToLower(follow.Login)],
			})
		}
	}
	return view
}

func (a *App) cleanupTwitchImportsLocked() {
	now := time.Now()
	for id, session := range a.twitchImports {
		if now.Sub(session.CreatedAt) > twitchImportSessionTTL {
			delete(a.twitchImports, id)
			continue
		}
		if !session.ExpiresAt.IsZero() && now.After(session.ExpiresAt.Add(5*time.Minute)) {
			delete(a.twitchImports, id)
		}
	}
	if len(a.twitchImports) < twitchImportMaxSessions {
		return
	}
	oldestID := ""
	var oldest time.Time
	for id, session := range a.twitchImports {
		if oldestID == "" || session.CreatedAt.Before(oldest) {
			oldestID = id
			oldest = session.CreatedAt
		}
	}
	if oldestID != "" {
		delete(a.twitchImports, oldestID)
	}
}

// StartTwitchImport begins the device-code sign-in. Tokens are held in memory
// only and are discarded once the follows list has been fetched.
func (a *App) StartTwitchImport(ctx context.Context) (twitchImportView, error) {
	clientID := twitchOAuthClientID()
	if clientID == "" {
		return twitchImportView{}, errors.New("Twitch import is not configured in this build")
	}
	start, err := startTwitchDeviceFlow(ctx, clientID)
	if err != nil {
		return twitchImportView{}, err
	}

	now := time.Now()
	session := &twitchImportSession{
		ID:              newID(),
		DeviceCode:      start.DeviceCode,
		UserCode:        start.UserCode,
		VerificationURI: start.VerificationURI,
		Interval:        time.Duration(start.Interval) * time.Second,
		ExpiresAt:       now.Add(time.Duration(start.ExpiresIn) * time.Second),
		NextPoll:        now.Add(time.Duration(start.Interval) * time.Second),
		CreatedAt:       now,
		Status:          twitchImportPending,
	}

	a.twitchMu.Lock()
	defer a.twitchMu.Unlock()
	a.cleanupTwitchImportsLocked()
	a.twitchImports[session.ID] = session
	return a.twitchImportViewLocked(session), nil
}

// PollTwitchImport checks the device-code grant and, once authorized, loads the
// account's followed channels. It is safe to call repeatedly.
func (a *App) PollTwitchImport(ctx context.Context, id string) (twitchImportView, error) {
	clientID := twitchOAuthClientID()
	if clientID == "" {
		return twitchImportView{}, errors.New("Twitch import is not configured in this build")
	}

	a.twitchMu.Lock()
	session := a.twitchImports[id]
	if session == nil {
		a.twitchMu.Unlock()
		return twitchImportView{}, errors.New("this Twitch import expired. Start again")
	}
	now := time.Now()
	if session.Status == twitchImportPending && now.After(session.ExpiresAt) {
		session.Status = twitchImportExpired
		session.Message = "The sign-in code expired. Start again."
	}
	if session.Polling || session.Status != twitchImportPending || now.Before(session.NextPoll) {
		view := a.twitchImportViewLocked(session)
		a.twitchMu.Unlock()
		return view, nil
	}
	session.Polling = true
	deviceCode := session.DeviceCode
	a.twitchMu.Unlock()

	token, state, retryable, err := pollTwitchDeviceToken(ctx, clientID, deviceCode)

	a.twitchMu.Lock()
	if a.twitchImports[id] != session {
		a.twitchMu.Unlock()
		return twitchImportView{ID: id, Status: twitchImportExpired, Message: "This import was cancelled."}, nil
	}
	session.Polling = false
	if err != nil {
		session.Message = err.Error()
		if retryable {
			session.NextPoll = time.Now().Add(session.Interval)
		} else {
			session.Status = twitchImportError
		}
		view := a.twitchImportViewLocked(session)
		a.twitchMu.Unlock()
		return view, nil
	}

	switch state {
	case twitchPollPending:
		session.Message = ""
		session.NextPoll = time.Now().Add(session.Interval)
	case twitchPollSlowDown:
		session.Interval += 5 * time.Second
		session.NextPoll = time.Now().Add(session.Interval)
	case twitchPollExpired:
		session.Status = twitchImportExpired
		session.Message = "The sign-in code expired. Start again."
	case twitchPollDenied:
		session.Status = twitchImportDenied
		session.Message = "Twitch sign-in was cancelled."
	default:
		session.Status = twitchImportAuthorizing
		session.Message = "Loading your follows…"
		session.DeviceCode = ""
		session.UserCode = ""
	}
	if session.Status != twitchImportAuthorizing {
		view := a.twitchImportViewLocked(session)
		a.twitchMu.Unlock()
		return view, nil
	}
	a.twitchMu.Unlock()

	user, fetchErr := fetchTwitchHelixUser(ctx, clientID, token.AccessToken)
	var follows []twitchHelixFollow
	var total int
	if fetchErr == nil {
		follows, total, fetchErr = fetchTwitchHelixFollows(ctx, clientID, token.AccessToken, user.ID)
	}
	var profiles map[string]twitchHelixUser
	if fetchErr == nil {
		logins := make([]string, 0, len(follows))
		seen := make(map[string]bool, len(follows))
		for _, follow := range follows {
			login := strings.ToLower(strings.TrimSpace(follow.BroadcasterLogin))
			if login == "" || seen[login] {
				continue
			}
			seen[login] = true
			logins = append(logins, login)
		}
		profiles = fetchTwitchHelixProfiles(ctx, clientID, token.AccessToken, logins)
	}
	token.AccessToken = ""
	token.RefreshToken = ""

	a.twitchMu.Lock()
	defer a.twitchMu.Unlock()
	if a.twitchImports[id] != session {
		return twitchImportView{ID: id, Status: twitchImportExpired, Message: "This import was cancelled."}, nil
	}
	if fetchErr != nil {
		session.Status = twitchImportError
		session.Message = fetchErr.Error()
		return a.twitchImportViewLocked(session), nil
	}

	records := make([]twitchFollowRecord, 0, len(follows))
	seen := make(map[string]bool, len(follows))
	for _, follow := range follows {
		login := strings.ToLower(strings.TrimSpace(follow.BroadcasterLogin))
		if login == "" || seen[login] {
			continue
		}
		seen[login] = true
		record := twitchFollowRecord{
			Login:       login,
			DisplayName: strings.TrimSpace(follow.BroadcasterName),
			FollowedAt:  follow.FollowedAt,
		}
		if record.DisplayName == "" {
			record.DisplayName = login
		}
		if profile, ok := profiles[login]; ok {
			if name := strings.TrimSpace(profile.DisplayName); name != "" {
				record.DisplayName = name
			}
			record.AvatarURL = profile.ProfileImageURL
		}
		records = append(records, record)
	}
	accountName := strings.TrimSpace(user.DisplayName)
	if accountName == "" {
		accountName = user.Login
	}
	session.Account = &twitchAccountView{
		Login:       user.Login,
		DisplayName: accountName,
		AvatarURL:   user.ProfileImageURL,
	}
	session.Follows = records
	session.Total = total
	if session.Total < len(records) {
		session.Total = len(records)
	}
	session.Status = twitchImportReady
	session.Message = ""
	return a.twitchImportViewLocked(session), nil
}

// ImportTwitchFollows adds the selected follows in one config write. The
// in-memory token is already gone; the cached follows drive display names and
// avatars.
func (a *App) ImportTwitchFollows(id string, logins []string) (bulkAddResponse, error) {
	a.twitchMu.Lock()
	session := a.twitchImports[id]
	if session == nil || session.Status != twitchImportReady {
		a.twitchMu.Unlock()
		return bulkAddResponse{}, errors.New("this Twitch import is not ready. Connect Twitch again")
	}
	records := make(map[string]twitchFollowRecord, len(session.Follows))
	for _, record := range session.Follows {
		records[strings.ToLower(record.Login)] = record
	}
	a.twitchMu.Unlock()

	if len(logins) == 0 {
		return bulkAddResponse{}, errors.New("choose at least one streamer")
	}
	if len(logins) > twitchImportMaxSelection {
		return bulkAddResponse{}, fmt.Errorf("import at most %d streamers at a time", twitchImportMaxSelection)
	}

	result := bulkAddResponse{}
	selected := make([]twitchFollowRecord, 0, len(logins))
	seen := make(map[string]bool, len(logins))
	for _, raw := range logins {
		login, err := normalizeTwitchLogin(raw)
		if err != nil {
			result.Failed = append(result.Failed, bulkAddIssue{Login: raw, Reason: err.Error()})
			continue
		}
		if seen[login] {
			continue
		}
		seen[login] = true
		record, ok := records[login]
		if !ok {
			result.Skipped = append(result.Skipped, bulkAddIssue{Login: login, Reason: "not in your Twitch follows"})
			continue
		}
		selected = append(selected, record)
	}
	if len(selected) == 0 {
		return result, errors.New("none of the selected streamers are in your Twitch follows")
	}

	a.mu.Lock()
	originalLen := len(a.cfg.Channels)
	addedIDs := make([]string, 0, len(selected))
	added := 0
	for _, record := range selected {
		if a.channelByLoginLocked(record.Login) != nil {
			result.Skipped = append(result.Skipped, bulkAddIssue{Login: record.Login, Reason: "already added"})
			continue
		}
		channel := Channel{
			ID:          newID(),
			Login:       record.Login,
			DisplayName: record.DisplayName,
			AvatarURL:   record.AvatarURL,
			Enabled:     true,
		}
		if len(a.cfg.Webhooks) == 1 {
			channel.Destinations = []Destination{{WebhookID: a.cfg.Webhooks[0].ID}}
		}
		a.cfg.Channels = append(a.cfg.Channels, channel)
		addedIDs = append(addedIDs, channel.ID)
		added++
	}
	var persistErr error
	if added > 0 {
		persistErr = a.persistLocked()
	}
	if persistErr != nil {
		a.cfg.Channels = a.cfg.Channels[:originalLen]
		for _, channelID := range addedIDs {
			delete(a.statuses, channelID)
		}
		a.mu.Unlock()
		return bulkAddResponse{}, persistErr
	}
	a.mu.Unlock()

	result.Added = added
	if added > 0 {
		a.addActivity("success", fmt.Sprintf("Imported %d streamer%s from Twitch.", added, pluralSuffix(added)))
		if a.enableAutoCheck {
			a.TriggerCheck()
		}
	}

	a.twitchMu.Lock()
	delete(a.twitchImports, id)
	a.twitchMu.Unlock()
	return result, nil
}

func (a *App) CancelTwitchImport(id string) {
	a.twitchMu.Lock()
	delete(a.twitchImports, id)
	a.twitchMu.Unlock()
}

func parseBulkLoginInput(values []string) []string {
	entries := make([]string, 0, len(values))
	for _, value := range values {
		for _, part := range strings.FieldsFunc(value, func(character rune) bool {
			return character == '\n' || character == '\r' || character == ',' || character == ';' ||
				character == '\t'
		}) {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				entries = append(entries, trimmed)
			}
		}
	}
	return entries
}

var (
	importHrefPattern      = regexp.MustCompile(`(?is)href\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))`)
	importTwitchURLPattern = regexp.MustCompile(`(?i)(?:https?://)?(?:www\.)?twitch\.tv/([A-Za-z0-9_]{1,25})`)
)

// reservedTwitchLogins are first-path segments that are Twitch pages rather
// than channels. The bookmarklet keeps the same list in JavaScript.
var reservedTwitchLogins = map[string]bool{
	"about": true, "activate": true, "app": true, "apps": true, "auth": true,
	"browse": true, "careers": true, "channel": true, "channels": true,
	"collections": true, "communities": true, "contact": true, "creatorcamp": true,
	"dashboard": true, "directory": true, "download": true, "downloads": true,
	"drops": true, "embed": true, "esports": true, "following": true,
	"friends": true, "gift": true, "help": true, "home": true, "index": true,
	"inventory": true, "jobs": true, "legal": true, "login": true, "logout": true,
	"messages": true, "mobile": true, "moderation": true, "notifications": true,
	"p": true, "partner": true, "player": true, "popout": true,
	"preferences": true, "prime": true, "privacy": true, "search": true,
	"settings": true, "signup": true, "store": true, "subscriptions": true,
	"support": true, "team": true, "teams": true, "tos": true, "turbo": true,
	"videos": true, "wallet": true, "watch": true,
}

func isReservedTwitchLogin(login string) bool {
	return reservedTwitchLogins[strings.ToLower(strings.TrimSpace(login))]
}

// parseImportJSON accepts common exports such as [{"login":"alpha"}, ...],
// {"follows":[...]}, or an array of plain strings. It is deliberately
// conservative about object keys so a title or unrelated field is not imported
// as a channel.
func parseImportJSON(text string) ([]string, bool) {
	var value any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		return nil, false
	}
	entries := make([]string, 0)
	var walk func(any)
	walk = func(node any) {
		switch typed := node.(type) {
		case string:
			entries = append(entries, typed)
		case []any:
			for _, item := range typed {
				walk(item)
			}
		case map[string]any:
			// Prefer explicit identity fields. Falling back to display names
			// only when no login-like key exists avoids importing "AlphaTV"
			// when the real login is "alpha".
			walkedIdentity := false
			for _, key := range []string{"login", "channel_login", "username", "user"} {
				if item, ok := typed[key]; ok {
					walk(item)
					walkedIdentity = true
				}
			}
			for _, key := range []string{"logins", "follows", "following", "followed_channels", "channels", "streamers", "data", "items"} {
				if item, ok := typed[key]; ok {
					walk(item)
				}
			}
			if !walkedIdentity {
				for _, key := range []string{"name", "display_name", "channel"} {
					if item, ok := typed[key]; ok {
						walk(item)
						walkedIdentity = true
					}
				}
			}
		}
	}
	walk(value)
	return entries, true
}

// parseImportLogins accepts plain paste (newline/comma/semicolon), Twitch
// URLs, JSON exports, and HTML saved from a following page. Only the raw text
// size is bounded here; ImportChannels enforces the unique-login cap after
// parsing.
func parseImportLogins(values []string) ([]string, error) {
	total := 0
	for _, value := range values {
		total += len(value)
		if total > bulkImportMaxTextBytes {
			return nil, errors.New("the pasted list is too large (max 2 MB)")
		}
	}

	entries := make([]string, 0, len(values))
	for _, value := range values {
		text := strings.TrimSpace(value)
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "[") || strings.HasPrefix(text, "{") {
			if parsed, ok := parseImportJSON(text); ok {
				entries = append(entries, parsed...)
				continue
			}
		}
		if strings.ContainsAny(text, "<>") {
			entries = append(entries, parseImportHTML(text)...)
			continue
		}
		entries = append(entries, parseBulkLoginInput([]string{text})...)
	}
	return entries, nil
}

// parseImportHTML is deliberately tolerant of saved pages: it reads quoted,
// single-quoted, and unquoted href attributes as well as absolute
// twitch.tv/login mentions anywhere in the markup.
func parseImportHTML(text string) []string {
	entries := make([]string, 0)
	seen := make(map[string]bool)
	add := func(candidate string) {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			return
		}
		key := strings.ToLower(candidate)
		if seen[key] {
			return
		}
		seen[key] = true
		entries = append(entries, candidate)
	}

	for _, match := range importHrefPattern.FindAllStringSubmatch(text, -1) {
		value := ""
		for index := 1; index < len(match); index++ {
			if strings.TrimSpace(match[index]) != "" {
				value = match[index]
				break
			}
		}
		if login, ok := loginFromImportHref(value); ok {
			add(login)
		}
	}
	for _, match := range importTwitchURLPattern.FindAllStringSubmatch(text, -1) {
		if login, ok := loginFromImportHref(match[0]); ok {
			add(login)
		}
	}
	return entries
}

// loginFromImportHref turns a saved-page link into a channel login. It accepts
// relative channel paths ("/streamer"), absolute twitch.tv URLs, and bare
// logins, and rejects reserved Twitch routes.
func loginFromImportHref(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "#") ||
		strings.HasPrefix(strings.ToLower(value), "javascript:") {
		return "", false
	}

	lower := strings.ToLower(value)
	path := ""
	switch {
	case strings.HasPrefix(lower, "http://"), strings.HasPrefix(lower, "https://"), strings.HasPrefix(value, "//"),
		strings.HasPrefix(lower, "twitch.tv/"), strings.HasPrefix(lower, "www.twitch.tv/"):
		parseValue := value
		if strings.HasPrefix(parseValue, "//") {
			parseValue = "https:" + parseValue
		} else if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
			parseValue = "https://" + parseValue
		}
		parsed, err := url.Parse(parseValue)
		if err != nil {
			return "", false
		}
		host := strings.ToLower(parsed.Hostname())
		if host != "twitch.tv" && !strings.HasSuffix(host, ".twitch.tv") {
			return "", false
		}
		path = parsed.Path
	case strings.HasPrefix(value, "/"):
		path = value
	case !strings.ContainsAny(value, "/?#"):
		path = value
	default:
		return "", false
	}

	path = strings.TrimPrefix(path, "/")
	if index := strings.IndexAny(path, "?#"); index >= 0 {
		path = path[:index]
	}
	path = strings.Trim(path, "/")
	if path == "" || strings.Contains(path, "/") {
		return "", false
	}
	if decoded, err := url.PathUnescape(path); err == nil {
		path = decoded
	}
	login, err := normalizeTwitchLogin(path)
	if err != nil || isReservedTwitchLogin(login) {
		return "", false
	}
	return login, true
}

// importDestinations resolves the requested destinations. An empty request
// keeps the historical single-webhook auto-assign behavior.
func (a *App) importDestinations(webhookIDs []string) ([]string, bool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(webhookIDs) == 0 {
		if len(a.cfg.Webhooks) == 1 {
			return nil, true, nil
		}
		return nil, false, nil
	}

	selected := make([]string, 0, len(webhookIDs))
	seen := make(map[string]bool, len(webhookIDs))
	for _, webhookID := range webhookIDs {
		webhookID = strings.TrimSpace(webhookID)
		if webhookID == "" || seen[webhookID] {
			continue
		}
		seen[webhookID] = true
		if a.webhookIndexLocked(webhookID) < 0 {
			return nil, false, fmt.Errorf("destination %q was not found", webhookID)
		}
		selected = append(selected, webhookID)
	}
	if len(selected) == 0 {
		return nil, false, errors.New("choose at least one destination")
	}
	return selected, false, nil
}

// ImportChannels parses a pasted list or saved HTML, validates the logins in
// batched GraphQL requests, and appends the new channels in one config write.
func (a *App) ImportChannels(ctx context.Context, raw []string, webhookIDs []string) (bulkAddResponse, error) {
	entries, err := parseImportLogins(raw)
	if err != nil {
		return bulkAddResponse{}, err
	}
	if len(entries) == 0 {
		return bulkAddResponse{}, errors.New("enter at least one Twitch username")
	}

	result := bulkAddResponse{}
	candidates := make([]string, 0, len(entries))
	seenLogins := make(map[string]bool, len(entries))
	seenInvalid := make(map[string]bool, len(entries))
	for _, entry := range entries {
		login, err := normalizeTwitchLogin(entry)
		if err != nil {
			key := strings.ToLower(strings.TrimSpace(entry))
			if key == "" || seenInvalid[key] {
				continue
			}
			seenInvalid[key] = true
			result.Failed = append(result.Failed, bulkAddIssue{Login: strings.TrimSpace(entry), Reason: err.Error()})
			continue
		}
		if seenLogins[login] {
			continue
		}
		if isReservedTwitchLogin(login) {
			result.Failed = append(result.Failed, bulkAddIssue{Login: login, Reason: "that is a Twitch page, not a channel"})
			continue
		}
		seenLogins[login] = true
		candidates = append(candidates, login)
	}
	if len(candidates) > bulkImportMaxLogins {
		return bulkAddResponse{}, fmt.Errorf("import at most %d streamers at a time", bulkImportMaxLogins)
	}
	if len(candidates) == 0 {
		if len(result.Failed) > 0 {
			return result, errors.New("none of those look like Twitch usernames")
		}
		return result, errors.New("enter at least one Twitch username")
	}

	assignedWebhooks, autoAssign, err := a.importDestinations(webhookIDs)
	if err != nil {
		return bulkAddResponse{}, err
	}

	existing := a.channelLogins()
	pending := make([]string, 0, len(candidates))
	for _, login := range candidates {
		if existing[login] {
			result.Skipped = append(result.Skipped, bulkAddIssue{Login: login, Reason: "already added"})
			continue
		}
		pending = append(pending, login)
	}

	if len(pending) > 0 {
		fetched := fetchTwitchChannelsBatchContext(ctx, pending)

		a.mu.Lock()
		originalLen := len(a.cfg.Channels)
		addedIDs := make([]string, 0, len(pending))
		added := 0
		assigned := 0
		for _, login := range pending {
			outcome, ok := fetched[login]
			if !ok {
				result.Failed = append(result.Failed, bulkAddIssue{Login: login, Reason: "Twitch did not return this channel"})
				continue
			}
			if outcome.Err != nil {
				reason := outcome.Err.Error()
				if errors.Is(outcome.Err, context.Canceled) || errors.Is(outcome.Err, context.DeadlineExceeded) {
					reason = "the import was cancelled"
				}
				result.Failed = append(result.Failed, bulkAddIssue{Login: login, Reason: reason})
				continue
			}
			if outcome.User == nil {
				result.Failed = append(result.Failed, bulkAddIssue{Login: login, Reason: "Twitch did not return this channel"})
				continue
			}
			if a.channelByLoginLocked(login) != nil {
				result.Skipped = append(result.Skipped, bulkAddIssue{Login: login, Reason: "already added"})
				continue
			}
			channel := Channel{
				ID:          newID(),
				Login:       login,
				DisplayName: displayNameFor(outcome.User, login),
				AvatarURL:   outcome.User.ProfileImageURL,
				BannerURL:   outcome.User.BannerImageURL,
				Enabled:     true,
			}
			if len(assignedWebhooks) > 0 {
				channel.Destinations = make([]Destination, 0, len(assignedWebhooks))
				for _, webhookID := range assignedWebhooks {
					channel.Destinations = append(channel.Destinations, Destination{WebhookID: webhookID})
				}
			} else if autoAssign {
				channel.Destinations = []Destination{{WebhookID: a.cfg.Webhooks[0].ID}}
			}
			if len(channel.Destinations) > 0 {
				assigned++
			}
			a.cfg.Channels = append(a.cfg.Channels, channel)
			a.statuses[channel.ID] = statusFromUser(outcome.User, time.Now())
			addedIDs = append(addedIDs, channel.ID)
			added++
		}
		var persistErr error
		if added > 0 {
			persistErr = a.persistLocked()
		}
		if persistErr != nil {
			a.cfg.Channels = a.cfg.Channels[:originalLen]
			for _, channelID := range addedIDs {
				delete(a.statuses, channelID)
			}
			a.mu.Unlock()
			return bulkAddResponse{}, persistErr
		}
		a.mu.Unlock()

		result.Added = added
		result.Assigned = assigned
	}

	result.Total = result.Added + len(result.Skipped) + len(result.Failed)
	if result.Added > 0 {
		a.addActivity("success", fmt.Sprintf("Imported %d streamer%s from a list.", result.Added, pluralSuffix(result.Added)))
		if a.enableAutoCheck {
			a.TriggerCheck()
		}
	}
	return result, nil
}

// AddChannelsBulk is the legacy entry point used by /api/channels/bulk. It
// shares parsing and batched validation with ImportChannels and keeps the
// single-webhook auto-assign behavior.
func (a *App) AddChannelsBulk(ctx context.Context, raw []string) (bulkAddResponse, error) {
	return a.ImportChannels(ctx, raw, nil)
}
