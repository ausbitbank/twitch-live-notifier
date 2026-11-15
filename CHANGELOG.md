# Changelog

## 2.1.0 - 2026-11-15

- Added an optional Discord user ID to each streamer.
- Added the `{discord_mention}` template placeholder.
- Discord payloads now allow only the configured streamer mention while still
  blocking `@everyone` and role mentions.
- Added a Discord ping editor and preview guidance on the Streamers page.

## 2.0.0 - 2026-11-15

- New dark dashboard hosted by the app and opened in the browser.
- Unlimited Twitch streamers and labelled Discord webhook destinations.
- Per-streamer destination routing.
- Per-destination message templates with placeholders and a live Discord preview.
- Background monitoring from the system tray with once-per-stream deduplication.
- Test webhook, send-last-stream sample, pause/resume, and check-now actions.
- Start with Windows, configurable check interval, dark/light theme.
- New mascot artwork, app icon, dashboard hero, README art, and release screenshots.
- Settings migration from the single-channel 1.x configuration.

## 1.0.0 - 2026-11-15

- First public release.
- Native Windows setup window for Twitch username and Discord webhook.
- Live stream alerts with title, game, viewers, live duration, and preview image.
- Test webhook and last-stream example actions.
- Local settings stored per Windows user.
- Optional CLI and Task Scheduler support with once-per-stream deduplication.
- Custom app icon, version metadata, and Windows DPI-aware manifest.
