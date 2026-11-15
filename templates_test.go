package main

import (
	"strings"
	"testing"
	"time"
)

func testStreamInfo() StreamInfo {
	return StreamInfo{
		StreamID:  "stream123",
		Login:     "teststreamer",
		Streamer:  "TestStreamer",
		Title:     "A cozy test stream",
		Game:      "Just Chatting",
		URL:       "https://www.twitch.tv/teststreamer",
		Thumbnail: "https://example.com/preview.jpg",
		AvatarURL: "https://example.com/avatar.png",
		Viewers:   1234,
		StartedAt: time.Now().Add(-2*time.Hour - 11*time.Minute),
		Live:      true,
	}
}

func TestRenderTemplate(t *testing.T) {
	info := testStreamInfo()
	now := info.StartedAt.Add(2*time.Hour + 11*time.Minute)
	text := "{streamer} | {title} | {game} | {url} | {viewers} | {live_for} | {webhook} | {channel} | {started}"
	got := renderTemplate(text, info, "Community", now)
	for _, want := range []string{"TestStreamer", "A cozy test stream", "Just Chatting", "https://www.twitch.tv/teststreamer", "1,234", "2h 11m", "Community", "teststreamer"} {
		if !strings.Contains(got, want) {
			t.Errorf("renderTemplate() = %q, missing %q", got, want)
		}
	}
}

func TestUnknownPlaceholders(t *testing.T) {
	got := unknownPlaceholders("{streamer} {unknown} {title} {another_one} {unknown}")
	if len(got) != 2 || got[0] != "{unknown}" || got[1] != "{another_one}" {
		t.Fatalf("unknownPlaceholders() = %#v", got)
	}
}

func TestBuildPayload(t *testing.T) {
	template := defaultTemplate()
	webhook := Webhook{ID: "w1", Label: "Community", URL: "https://example.com", Enabled: true, Template: template}
	info := testStreamInfo()
	now := time.Now()
	payload := buildPayload(webhook, info, now, false)

	if !strings.Contains(payload.Content, "TestStreamer") || !strings.Contains(payload.Content, "Going") && !strings.Contains(payload.Content, "live") {
		t.Errorf("content = %q", payload.Content)
	}
	if len(payload.Embeds) != 1 {
		t.Fatalf("embeds = %d", len(payload.Embeds))
	}
	embed := payload.Embeds[0]
	if embed.Title != info.Title || embed.URL != info.URL {
		t.Errorf("title/url = %q / %q", embed.Title, embed.URL)
	}
	if embed.Color != 0x9146FF {
		t.Errorf("color = %#x", embed.Color)
	}
	if embed.Author == nil || embed.Author.Name != info.Streamer || embed.Author.IconURL != info.AvatarURL {
		t.Errorf("author = %+v", embed.Author)
	}
	if embed.Image == nil || !strings.Contains(embed.Image.URL, "preview.jpg") || !strings.Contains(embed.Image.URL, "cb=") {
		t.Errorf("image = %+v", embed.Image)
	}
	if len(embed.Fields) != 3 {
		t.Fatalf("fields = %+v", embed.Fields)
	}
	if embed.Fields[1].Name != "Viewers" || embed.Fields[1].Value != "1,234" {
		t.Errorf("viewers field = %+v", embed.Fields[1])
	}
	if payload.AllowedMentions == nil || len(payload.AllowedMentions.Parse) != 0 {
		t.Errorf("allowed mentions = %+v", payload.AllowedMentions)
	}
}

func TestBuildExamplePayload(t *testing.T) {
	template := defaultTemplate()
	webhook := Webhook{Label: "Community", Template: template}
	info := testStreamInfo()
	info.Live = false
	payload := buildPayload(webhook, info, time.Now(), true)
	if !strings.HasPrefix(payload.Content, "🧪 **Example:**") {
		t.Errorf("content = %q", payload.Content)
	}
	if payload.Embeds[0].Footer == nil || !strings.HasSuffix(payload.Embeds[0].Footer.Text, "Example") {
		t.Errorf("footer = %+v", payload.Embeds[0].Footer)
	}
	if payload.Embeds[0].Fields[1].Value != "—" {
		t.Errorf("offline viewers = %q", payload.Embeds[0].Fields[1].Value)
	}
}

func TestValidateTemplate(t *testing.T) {
	if _, err := validateTemplate(Template{Content: "hello", Color: "#9146FF"}); err != nil {
		t.Fatalf("valid template rejected: %v", err)
	}
	empty, err := validateTemplate(Template{Color: "#9146FF"})
	if err != nil || empty.Content == "" {
		t.Fatalf("empty template was not replaced by defaults: %+v, %v", empty, err)
	}
	if _, err := validateTemplate(Template{Content: "hello", Color: "purple"}); err == nil {
		t.Fatal("invalid color accepted")
	}
	long := strings.Repeat("x", 2001)
	if _, err := validateTemplate(Template{Content: long, Color: "#9146FF"}); err == nil {
		t.Fatal("oversized content accepted")
	}
}

func TestFormatHelpers(t *testing.T) {
	if got := formatCount(1234567); got != "1,234,567" {
		t.Errorf("formatCount = %q", got)
	}
	if got := formatDuration(2*time.Hour + 11*time.Minute); got != "2h 11m" {
		t.Errorf("formatDuration = %q", got)
	}
}
