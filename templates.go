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
	Thumbnail string
	AvatarURL string
	Viewers   int
	StartedAt time.Time
	Live      bool
}

type Placeholder struct {
	Token       string `json:"token"`
	Description string `json:"description"`
}

var placeholderPattern = regexp.MustCompile(`\{[A-Za-z0-9_]+\}`)

var templatePlaceholders = []Placeholder{
	{"{streamer}", "Twitch display name"},
	{"{title}", "Current stream title"},
	{"{game}", "Game or category"},
	{"{url}", "Twitch channel URL"},
	{"{thumbnail}", "Twitch preview image URL"},
	{"{viewers}", "Viewer count"},
	{"{started}", "Stream start date and time"},
	{"{live_for}", "How long the stream has been live"},
	{"{webhook}", "Destination label"},
	{"{channel}", "Twitch login name"},
}

func normalizeTemplate(t Template) Template {
	if t.Color == "" {
		t.Color = "#9146FF"
	}
	if t.Content == "" && t.EmbedTitle == "" && t.EmbedDescription == "" &&
		t.Footer == "" && !t.ShowAuthor && !t.ShowImage && !t.ShowGame &&
		!t.ShowViewers && !t.ShowLiveSince {
		return defaultTemplate()
	}
	return t
}

func validateTemplate(t Template) (Template, error) {
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
	if embedLength > 5600 {
		return Template{}, fmt.Errorf("the embed text is too long for Discord")
	}
	if _, err := parseHexColor(t.Color); err != nil {
		return Template{}, err
	}
	if strings.TrimSpace(t.Content) == "" && strings.TrimSpace(t.EmbedTitle) == "" && strings.TrimSpace(t.EmbedDescription) == "" {
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

	replacer := strings.NewReplacer(
		"{streamer}", info.Streamer,
		"{title}", info.Title,
		"{game}", info.Game,
		"{url}", info.URL,
		"{thumbnail}", info.Thumbnail,
		"{viewers}", viewers,
		"{started}", started,
		"{live_for}", liveFor,
		"{webhook}", webhookLabel,
		"{channel}", info.Login,
	)
	return replacer.Replace(text)
}

func unknownPlaceholders(text string) []string {
	known := make(map[string]bool, len(templatePlaceholders))
	for _, placeholder := range templatePlaceholders {
		known[placeholder.Token] = true
	}
	seen := map[string]bool{}
	var unknown []string
	for _, token := range placeholderPattern.FindAllString(text, -1) {
		if !known[token] && !seen[token] {
			seen[token] = true
			unknown = append(unknown, token)
		}
	}
	return unknown
}

func buildPayload(webhook Webhook, info StreamInfo, now time.Time, example bool) discordWebhookPayload {
	template := normalizeTemplate(webhook.Template)
	content := renderTemplate(template.Content, info, webhook.Label, now)
	if example {
		if content != "" {
			content = "🧪 **Example:** " + content
		} else {
			content = "🧪 **Example**"
		}
	}

	embed := discordEmbed{
		Title:       renderTemplate(template.EmbedTitle, info, webhook.Label, now),
		URL:         info.URL,
		Description: renderTemplate(template.EmbedDescription, info, webhook.Label, now),
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

	if template.ShowGame && strings.TrimSpace(info.Game) != "" {
		embed.Fields = append(embed.Fields, discordField{Name: "Game", Value: info.Game, Inline: true})
	}
	if template.ShowViewers {
		value := "—"
		if info.Live {
			value = formatCount(info.Viewers)
		}
		embed.Fields = append(embed.Fields, discordField{Name: "Viewers", Value: value, Inline: true})
	}
	if template.ShowLiveSince && !info.StartedAt.IsZero() {
		if info.Live {
			embed.Fields = append(embed.Fields, discordField{
				Name:   "Live since",
				Value:  fmt.Sprintf("<t:%d:R>", info.StartedAt.Unix()),
				Inline: true,
			})
		} else {
			embed.Fields = append(embed.Fields, discordField{
				Name:   "Last live",
				Value:  fmt.Sprintf("<t:%d:F>", info.StartedAt.Unix()),
				Inline: true,
			})
		}
	}

	if template.ShowImage && strings.TrimSpace(info.Thumbnail) != "" {
		embed.Image = &discordImage{URL: withCacheBuster(info.Thumbnail)}
	}
	footer := renderTemplate(template.Footer, info, webhook.Label, now)
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
		Embeds:          []discordEmbed{embed},
		AllowedMentions: &discordAllowedMentions{Parse: []string{}},
	}
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
