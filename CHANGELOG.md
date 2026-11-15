# Changelog

## 2.2.0 - 2026-11-16

### Self-hosting and health

- Added a multi-stage Dockerfile and `docker-compose.yml` with a persistent
  config volume, non-root user, and built-in healthcheck.
- Added `GET /healthz` with monitor status, uptime, last-check age, pause state,
  streamer/live/destination counts, and the latest error.
- Added `--listen`, `--healthcheck`, and `--healthcheck-url` CLI options.
- Release tags now publish `linux/amd64` and `linux/arm64` images to GHCR.
- Added `docs/docker.md` for VPS, NAS, Raspberry Pi, backups, updates, timezone,
  and public-access safety.

### Reliability

- Twitch checks retry 5xx, 429, and timeout failures with bounded, jittered
  exponential backoff.
- Discord 429 responses honor `Retry-After` / `retry_after` and retry once.
- Discord `401`, `403`, and `404` responses mark a destination invalid instead
  of retrying forever.
- Destinations now store last success, last error, failure count, and invalid
  state. The dashboard shows failures and offers a retry/test action.
- Failed posts still do not record `last_stream_id`, so the next check retries.
- Added a cross-platform single-instance lock. A second `--serve` opens the
  existing dashboard; `--check` skips while another instance is running.

### Mention controls

- Added up to ten role IDs per destination and the `{role_mentions}` placeholder.
- Added streamer and role ping toggles plus a per-destination
  `@everyone`/`@here` opt-in with confirmation.
- `allowed_mentions` now always sends an explicit `parse` list and scopes
  `roles`/`users` to configured IDs only.
- The live preview lists exactly what will ping and warns about blocked or
  missing mentions.

### Notification filters

- Added per-streamer game/category include and exclude rules.
- Added title keyword include and exclude rules. Excludes always win.
- Added minimum viewers, alert delay, cooldown, and first-stream-of-the-day.
- Added quiet hours and allowed weekdays, evaluated in local time.
- Filtered streams still update live status and keep dedupe intact.
- Streamer cards show filter chips, a **Not posting** reason, and a filter
  editor.

### Upgrade notes

- The config schema is now version 3. Existing version 2 configs load
  automatically; new filters default to off.
- Docker users should set `TZ` if quiet hours should use a non-UTC timezone.

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
