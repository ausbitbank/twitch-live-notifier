package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type StreamInfo struct {
	StreamID  string
	Login     string
	Streamer  string
	Title     string
	Game      string
	URL       string
	VODURL    string
	Thumbnail string
	AvatarURL string
	DiscordID string
	Viewers   int
	StartedAt time.Time
	Live      bool
}

type Placeholder struct {
	Token       string `json:"token"`
	Description string `json:"description"`
}

// placeholders may include a fallback, e.g. {game|Just Chatting}. The fallback
// is used when the placeholder resolves to an empty string.
var placeholderPattern = regexp.MustCompile(`\{([A-Za-z0-9_]+)(?:\|([^{}]*))?\}`)

var templatePlaceholders = []Placeholder{
	{"{streamer}", "Twitch display name"},
	{"{title}", "Current stream title"},
	{"{game}", "Game or category"},
	{"{url}", "Twitch channel URL"},
	{"{vod}", "VOD link for the last stream, when available"},
	{"{thumbnail}", "Twitch preview image URL"},
	{"{viewers}", "Viewer count"},
	{"{started}", "Stream start date and time"},
	{"{live_for}", "How long the stream has been live"},
	{"{webhook}", "Destination label"},
	{"{channel}", "Twitch login name"},
	{"{discord_mention}", "Optional ping for the streamer's Discord user ID"},
	{"{role_mentions}", "Configured Discord role pings for this destination"},
}

const (
	maxCustomFields     = 10
	maxCustomFieldName  = 256
	maxCustomFieldValue = 1024
)

func normalizeTemplate(t Template) Template {
	if t.Color == "" {
		t.Color = "#9146FF"
	}
	if t.Content == "" && t.EmbedTitle == "" && t.EmbedDescription == "" &&
		t.Footer == "" && !t.ShowAuthor && !t.ShowImage && !t.ShowGame &&
		!t.ShowViewers && !t.ShowLiveSince && len(t.CustomFields) == 0 {
		return defaultTemplate()
	}
	if t.PingStreamer == nil {
		value := true
		t.PingStreamer = &value
	}
	if t.PingRoles == nil {
		value := true
		t.PingRoles = &value
	}
	t.CustomFields = normalizeCustomFields(t.CustomFields)
	t.FieldOrder = normalizeFieldOrder(t.FieldOrder, len(t.CustomFields))
	return t
}

func normalizeCustomFields(input []TemplateField) []TemplateField {
	fields := make([]TemplateField, 0, len(input))
	for _, field := range input {
		name := strings.TrimSpace(field.Name)
		value := strings.TrimSpace(field.Value)
		if name == "" && value == "" {
			continue
		}
		if name == "" {
			name = "Details"
		}
		if utf8.RuneCountInString(name) > maxCustomFieldName {
			name = string([]rune(name)[:maxCustomFieldName])
		}
		if utf8.RuneCountInString(value) > maxCustomFieldValue {
			value = string([]rune(value)[:maxCustomFieldValue])
		}
		fields = append(fields, TemplateField{Name: name, Value: value, Inline: field.Inline})
		if len(fields) == maxCustomFields {
			break
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func normalizeFieldOrder(input []string, customCount int) []string {
	allowed := map[string]bool{"game": true, "viewers": true, "live": true}
	for i := 0; i < customCount; i++ {
		allowed[fmt.Sprintf("custom:%d", i)] = true
	}
	seen := map[string]bool{}
	order := make([]string, 0, len(input))
	for _, raw := range input {
		key := strings.TrimSpace(raw)
		if !allowed[key] || seen[key] {
			continue
		}
		seen[key] = true
		order = append(order, key)
	}
	if len(order) == 0 {
		return nil
	}
	return order
}

func cloneTemplateFields(fields []TemplateField) []TemplateField {
	if len(fields) == 0 {
		return nil
	}
	return append([]TemplateField(nil), fields...)
}

func cloneStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}

func streamerPingEnabled(template Template) bool {
	return boolPointerValue(template.PingStreamer, true)
}

func rolePingEnabled(template Template) bool {
	return boolPointerValue(template.PingRoles, true)
}

func validateTemplate(t Template) (Template, error) {
	if len(t.CustomFields) > maxCustomFields {
		return Template{}, fmt.Errorf("add at most %d custom fields", maxCustomFields)
	}
	for _, field := range t.CustomFields {
		if utf8.RuneCountInString(strings.TrimSpace(field.Name)) > maxCustomFieldName {
			return Template{}, fmt.Errorf("custom field names must be %d characters or fewer", maxCustomFieldName)
		}
		if utf8.RuneCountInString(strings.TrimSpace(field.Value)) > maxCustomFieldValue {
			return Template{}, fmt.Errorf("custom field values must be %d characters or fewer", maxCustomFieldValue)
		}
	}
	t = normalizeTemplate(t)
	if utf8.RuneCountInString(t.Content) > 2000 {
		return Template{}, fmt.Errorf("message content is over Discord's 2,000 character limit")
	}
	if utf8.RuneCountInString(t.EmbedTitle) > 256 {
		return Template{}, fmt.Errorf("embed title is over Discord's 256 character limit")
	}
	if utf8.RuneCountInString(t.EmbedDescription) > 4096 {
		return Template{}, fmt.Errorf("embed description is over Discord's 4,096 character limit")
	}
	if utf8.RuneCountInString(t.Footer) > 2048 {
		return Template{}, fmt.Errorf("footer is over Discord's 2,048 character limit")
	}
	embedLength := utf8.RuneCountInString(t.EmbedTitle) + utf8.RuneCountInString(t.EmbedDescription) + utf8.RuneCountInString(t.Footer)
	for _, field := range t.CustomFields {
		embedLength += utf8.RuneCountInString(field.Name) + utf8.RuneCountInString(field.Value)
	}
	if embedLength > 5600 {
		return Template{}, fmt.Errorf("the embed text is too long for Discord")
	}
	if _, err := parseHexColor(t.Color); err != nil {
		return Template{}, err
	}
	if strings.TrimSpace(t.Content) == "" && strings.TrimSpace(t.EmbedTitle) == "" && strings.TrimSpace(t.EmbedDescription) == "" && len(t.CustomFields) == 0 {
		return Template{}, fmt.Errorf("add some message content or embed text")
	}
	return t, nil
}

func parseHexColor(value string) (int, error) {
	value = strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(value) != 6 {
		return 0, fmt.Errorf("choose a six-digit hex color")
	}
	number, err := strconv.ParseInt(value, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("choose a valid hex color")
	}
	return int(number), nil
}

func renderTemplate(text string, info StreamInfo, webhookLabel string, now time.Time) string {
	return renderMessage(text, info, Webhook{Label: webhookLabel, Template: defaultTemplate()}, now)
}

func renderMessage(text string, info StreamInfo, webhook Webhook, now time.Time) string {
	return renderMessageWithTemplate(text, info, webhook, webhook.Template, now)
}

func renderMessageWithTemplate(text string, info StreamInfo, webhook Webhook, rawTemplate Template, now time.Time) string {
	template := normalizeTemplate(rawTemplate)
	liveFor := "Ended"
	if info.Live && !info.StartedAt.IsZero() {
		liveFor = formatDuration(now.Sub(info.StartedAt))
	}
	started := "—"
	if !info.StartedAt.IsZero() {
		started = info.StartedAt.Local().Format("Jan 2, 2006 at 3:04 PM")
	}
	viewers := "—"
	if info.Live {
		viewers = formatCount(info.Viewers)
	}
	discordMention := "@" + info.Streamer
	if streamerPingEnabled(template) {
		if id, err := normalizeDiscordID(info.DiscordID); err == nil && id != "" {
			discordMention = "<@" + id + ">"
		}
	}
	roleMentions := ""
	if rolePingEnabled(template) {
		roleMentions = formatRoleMentions(webhook.RoleIDs)
	}

	values := map[string]string{
		"streamer":        info.Streamer,
		"title":           info.Title,
		"game":            info.Game,
		"url":             info.URL,
		"vod":             info.VODURL,
		"thumbnail":       info.Thumbnail,
		"viewers":         viewers,
		"started":         started,
		"live_for":        liveFor,
		"webhook":         webhook.Label,
		"channel":         info.Login,
		"discord_mention": discordMention,
		"role_mentions":   roleMentions,
	}
	return placeholderPattern.ReplaceAllStringFunc(text, func(token string) string {
		match := placeholderPattern.FindStringSubmatch(token)
		if len(match) == 0 {
			return token
		}
		value, ok := values[match[1]]
		if !ok {
			return token
		}
		if strings.TrimSpace(value) == "" && len(match) > 2 {
			return match[2]
		}
		return value
	})
}

func formatRoleMentions(roleIDs []string) string {
	roles := sanitizeRoleIDs(roleIDs)
	if len(roles) == 0 {
		return ""
	}
	parts := make([]string, 0, len(roles))
	for _, id := range roles {
		parts = append(parts, "<@&"+id+">")
	}
	return strings.Join(parts, " ")
}

func unknownPlaceholders(text string) []string {
	known := make(map[string]bool, len(templatePlaceholders))
	for _, placeholder := range templatePlaceholders {
		known[placeholder.Token] = true
	}
	seen := map[string]bool{}
	var unknown []string
	for _, match := range placeholderPattern.FindAllStringSubmatch(text, -1) {
		token := match[0]
		name := "{" + match[1] + "}"
		if known[name] {
			continue
		}
		if !seen[token] {
			seen[token] = true
			unknown = append(unknown, token)
		}
	}
	return unknown
}

func buildPayload(webhook Webhook, info StreamInfo, now time.Time, example bool) discordWebhookPayload {
	return buildPayloadWithTemplate(webhook, webhook.Template, info, now, example)
}

func routeTemplateFor(webhook Webhook, destination *Destination) Template {
	if destination != nil && destination.Template != nil {
		template := *destination.Template
		// Mention controls are destination-wide safety settings, so a route
		// template cannot silently widen or narrow them.
		template.PingStreamer = webhook.Template.PingStreamer
		template.PingRoles = webhook.Template.PingRoles
		template.AllowEveryone = webhook.Template.AllowEveryone
		return template
	}
	return webhook.Template
}

func buildPayloadForDestination(webhook Webhook, destination *Destination, info StreamInfo, now time.Time, example bool) discordWebhookPayload {
	return buildPayloadWithTemplate(webhook, routeTemplateFor(webhook, destination), info, now, example)
}

func buildPayloadWithTemplate(webhook Webhook, rawTemplate Template, info StreamInfo, now time.Time, example bool) discordWebhookPayload {
	template := normalizeTemplate(rawTemplate)
	content := renderMessageWithTemplate(template.Content, info, webhook, template, now)
	if example {
		if content != "" {
			content = "🧪 **Example:** " + content
		} else {
			content = "🧪 **Example**"
		}
	}

	embed := discordEmbed{
		Title:       renderMessageWithTemplate(template.EmbedTitle, info, webhook, template, now),
		URL:         info.URL,
		Description: renderMessageWithTemplate(template.EmbedDescription, info, webhook, template, now),
		Timestamp:   now.UTC().Format(time.RFC3339),
	}
	if color, err := parseHexColor(template.Color); err == nil {
		embed.Color = color
	}
	if template.ShowAuthor {
		embed.Author = &discordAuthor{
			Name:    info.Streamer,
			URL:     info.URL,
			IconURL: info.AvatarURL,
		}
	}

	fields := map[string]discordField{}
	if template.ShowGame && strings.TrimSpace(info.Game) != "" {
		fields["game"] = discordField{Name: "Game", Value: info.Game, Inline: true}
	}
	if template.ShowViewers {
		value := "—"
		if info.Live {
			value = formatCount(info.Viewers)
		}
		fields["viewers"] = discordField{Name: "Viewers", Value: value, Inline: true}
	}
	if template.ShowLiveSince && !info.StartedAt.IsZero() {
		if info.Live {
			fields["live"] = discordField{
				Name:   "Live since",
				Value:  fmt.Sprintf("<t:%d:R>", info.StartedAt.Unix()),
				Inline: true,
			}
		} else {
			fields["live"] = discordField{
				Name:   "Last live",
				Value:  fmt.Sprintf("<t:%d:F>", info.StartedAt.Unix()),
				Inline: true,
			}
		}
	}
	for index, custom := range template.CustomFields {
		name := strings.TrimSpace(renderMessageWithTemplate(custom.Name, info, webhook, template, now))
		value := strings.TrimSpace(renderMessageWithTemplate(custom.Value, info, webhook, template, now))
		if name == "" || value == "" {
			continue
		}
		if utf8.RuneCountInString(name) > maxCustomFieldName {
			name = string([]rune(name)[:maxCustomFieldName])
		}
		if utf8.RuneCountInString(value) > maxCustomFieldValue {
			value = string([]rune(value)[:maxCustomFieldValue])
		}
		fields[fmt.Sprintf("custom:%d", index)] = discordField{Name: name, Value: value, Inline: custom.Inline}
	}
	embed.Fields = orderedFields(fields, template.FieldOrder)

	if template.ShowImage && strings.TrimSpace(info.Thumbnail) != "" {
		embed.Image = &discordImage{URL: withCacheBuster(info.Thumbnail)}
	}
	footer := renderMessageWithTemplate(template.Footer, info, webhook, template, now)
	if example {
		if footer != "" {
			footer += " • Example"
		} else {
			footer = "Example"
		}
	}
	if footer != "" {
		embed.Footer = &discordFooter{Text: footer}
	}

	return discordWebhookPayload{
		Content:         content,
		Username:        webhook.Username,
		AvatarURL:       webhook.AvatarURL,
		Embeds:          []discordEmbed{embed},
		AllowedMentions: allowedMentionsForTemplate(webhook, template, info),
		Components:      buildLinkComponents(webhook, template, info, now),
	}
}

func orderedFields(fields map[string]discordField, order []string) []discordField {
	if len(fields) == 0 {
		return nil
	}
	used := map[string]bool{}
	result := make([]discordField, 0, len(fields))
	appendField := func(key string) {
		if used[key] {
			return
		}
		field, ok := fields[key]
		if !ok {
			return
		}
		used[key] = true
		result = append(result, field)
	}
	for _, key := range order {
		appendField(key)
	}
	for _, key := range []string{"game", "viewers", "live"} {
		appendField(key)
	}
	for index := 0; index < maxCustomFields; index++ {
		appendField(fmt.Sprintf("custom:%d", index))
	}
	return result
}

func buildLinkComponents(webhook Webhook, template Template, info StreamInfo, now time.Time) []discordComponent {
	if len(webhook.Buttons) == 0 {
		return nil
	}
	buttons := make([]discordButton, 0, len(webhook.Buttons))
	for _, button := range webhook.Buttons {
		label := strings.TrimSpace(renderMessageWithTemplate(button.Label, info, webhook, template, now))
		target := strings.TrimSpace(renderMessageWithTemplate(button.URL, info, webhook, template, now))
		if label == "" || target == "" {
			continue
		}
		if utf8.RuneCountInString(label) > 80 {
			label = string([]rune(label)[:80])
		}
		parsed, err := url.Parse(target)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			continue
		}
		buttons = append(buttons, discordButton{Type: 2, Style: 5, Label: label, URL: target})
	}
	if len(buttons) == 0 {
		return nil
	}
	return []discordComponent{{Type: 1, Components: buttons}}
}

func allowedMentionsFor(webhook Webhook, info StreamInfo) *discordAllowedMentions {
	return allowedMentionsForTemplate(webhook, webhook.Template, info)
}

func allowedMentionsForTemplate(webhook Webhook, rawTemplate Template, info StreamInfo) *discordAllowedMentions {
	template := normalizeTemplate(rawTemplate)
	allowed := &discordAllowedMentions{Parse: []string{}}
	if streamerPingEnabled(template) {
		if id, err := normalizeDiscordID(info.DiscordID); err == nil && id != "" {
			allowed.Users = []string{id}
		}
	}
	if rolePingEnabled(template) {
		allowed.Roles = sanitizeRoleIDs(webhook.RoleIDs)
	}
	if template.AllowEveryone {
		allowed.Parse = []string{"everyone"}
	}
	return allowed
}

func buildTestWebhookPayload(now time.Time) discordWebhookPayload {
	return discordWebhookPayload{
		Content: "✅ **" + appName + " is connected.**",
		Embeds: []discordEmbed{{
			Title:       "Webhook test",
			Description: "If you can see this, your Discord destination is working.",
			Color:       0x57F287,
			Footer:      &discordFooter{Text: appName + " v" + Version},
			Timestamp:   now.UTC().Format(time.RFC3339),
		}},
		AllowedMentions: &discordAllowedMentions{Parse: []string{}},
	}
}

func streamInfoFromTwitch(user *twitchUser, live bool) StreamInfo {
	login := strings.ToLower(strings.TrimSpace(user.Login))
	if login == "" {
		login = strings.ToLower(strings.TrimSpace(user.DisplayName))
	}
	info := StreamInfo{
		Login:     login,
		Streamer:  displayNameFor(user, ""),
		URL:       channelURLFor(user, ""),
		VODURL:    latestVODURL(user, login),
		AvatarURL: user.ProfileImageURL,
		Live:      live,
	}
	if live && user.Stream != nil {
		stream := user.Stream
		info.StreamID = stream.ID
		info.Title = streamTitle(stream)
		info.Game = strings.TrimSpace(stream.Game.Name)
		info.Thumbnail = firstNonEmpty(stream.PreviewImageURL, user.BannerImageURL, user.ProfileImageURL)
		info.Viewers = stream.ViewersCount
		if started, err := time.Parse(time.RFC3339, stream.CreatedAt); err == nil {
			info.StartedAt = started
		}
		return info
	}
	if user.LastBroadcast != nil {
		broadcast := user.LastBroadcast
		info.Title = strings.TrimSpace(broadcast.Title)
		if info.Title == "" {
			info.Title = "Untitled stream"
		}
		info.Game = strings.TrimSpace(broadcast.Game.Name)
		info.Thumbnail = firstNonEmpty(user.BannerImageURL, broadcast.Game.BoxArtURL, user.ProfileImageURL)
		if started, err := time.Parse(time.RFC3339, broadcast.StartedAt); err == nil {
			info.StartedAt = started
		}
		return info
	}
	info.Title = "No streams yet"
	return info
}

func latestVODURL(user *twitchUser, login string) string {
	if user == nil {
		return ""
	}
	for _, edge := range user.Videos.Edges {
		if id := strings.TrimSpace(edge.Node.ID); id != "" {
			return "https://www.twitch.tv/videos/" + id
		}
	}
	if login != "" && user.LastBroadcast == nil && user.Stream == nil {
		return "https://www.twitch.tv/" + strings.ToLower(login) + "/videos"
	}
	return ""
}

func sampleStreamInfo() StreamInfo {
	now := time.Now().Add(-2*time.Hour - 14*time.Minute)
	return StreamInfo{
		StreamID:  "sample-stream",
		Login:     "your_channel",
		Streamer:  "Your Streamer",
		Title:     "Cozy Saturday hangout — come say hi!",
		Game:      "Just Chatting",
		URL:       "https://www.twitch.tv/",
		Thumbnail: "",
		Viewers:   128,
		StartedAt: now,
		Live:      true,
	}
}

func displayNameFor(user *twitchUser, fallback string) string {
	if user != nil {
		if name := strings.TrimSpace(user.DisplayName); name != "" {
			return name
		}
	}
	if fallback = strings.TrimSpace(fallback); fallback != "" {
		return fallback
	}
	return "This channel"
}

func streamTitle(stream *twitchStream) string {
	if stream == nil {
		return "Untitled stream"
	}
	if title := strings.TrimSpace(stream.Title); title != "" {
		return title
	}
	return "Untitled stream"
}

func channelURLFor(user *twitchUser, fallbackLogin string) string {
	login := strings.TrimSpace(fallbackLogin)
	if user != nil {
		login = firstNonEmpty(strings.ToLower(strings.TrimSpace(user.Login)), strings.ToLower(strings.TrimSpace(user.DisplayName)), login)
	}
	if login == "" {
		return "https://www.twitch.tv/"
	}
	return "https://www.twitch.tv/" + strings.ToLower(strings.ReplaceAll(login, " ", ""))
}

func withCacheBuster(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return raw
	}
	query := u.Query()
	query.Set("cb", strconv.FormatInt(time.Now().Unix(), 10))
	u.RawQuery = query.Encode()
	return u.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func formatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}
	totalMinutes := int(duration.Minutes())
	days := totalMinutes / (24 * 60)
	hours := (totalMinutes / 60) % 24
	minutes := totalMinutes % 60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}

func formatCount(value int) string {
	raw := strconv.Itoa(value)
	if len(raw) <= 3 {
		return raw
	}
	var parts []string
	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}
	parts = append([]string{raw}, parts...)
	return strings.Join(parts, ",")
}
