# Changelog

## 2.5.0 - 2026-11-26

### Streamer import without a Twitch sign-in

- Added a local **Export follows** bookmarklet/helper that reads your existing
  Twitch following page in your browser and copies the channel list. The app
  never receives a Twitch token, and there is no backend or relay. See
  [docs/twitch-import.md](docs/twitch-import.md).
- Added **Paste a list** and file import (`.txt`, `.csv`, `.json`, saved
  `.html`) with a 2,000-channel limit, batched Twitch validation, already-added
  detection, and optional destination assignment.
- Kept the optional device-code Twitch sign-in as a fallback when a public
  client ID is configured. Tokens remain memory-only.
- Raised the legacy bulk-add limit from 100 to 2,000 and made validation batched
  instead of one request per channel.

### Scale and reliability

- Twitch checks now use aliased GraphQL batches (up to 20 logins per request,
  bounded concurrency) instead of one request per streamer. Hundreds of
  streamers fit inside the normal poll window.
- Bulk import and bulk validation share the same batched path.
- Added retry/backoff and rate-limit handling to batched requests, with a
  single-channel fallback for endpoints that reject the batch shape.

### Verified auto-update

- Added an in-app installer that downloads the correct OS/architecture asset,
  verifies its SHA-256 checksum, preflights the new binary, swaps it in, and
  restarts. The previous binary or bundle is retained for rollback.
- Added `update-manifest.json` generation to the release workflow, plus optional
  Ed25519 manifest signing when `UPDATE_SIGNING_KEY` is configured.
- Added Windows, Linux raw/AppImage, macOS bundle, `.deb`/system-package, and
  Docker update paths. Non-writable or system-managed installs show a clear
  manual-download fallback.
- Added update progress, restart status, and automatic dashboard reconnection.
- Windows uses a detached helper instead of renaming the running `.exe`.
- See [docs/auto-update.md](docs/auto-update.md).

### Streamer management at scale

- Added search, status filters, sorting, multi-select, and bulk enable, disable,
  delete, assign, and unassign actions.
- Added a streamer × destination routing matrix with sticky headers and
  responsive horizontal scrolling.
- Imported streamers can be assigned to destinations during import.

### Onboarding and operations

- Added a first-run setup checklist covering destination, streamer, routing,
  test/sample, and the first check. It can be dismissed per version.
- Pasting a Discord webhook URL can now suggest a destination label from
  Discord's webhook metadata.
- Added a persisted **History** view for posts, edits, cleanup, skips, and
  errors, with filters, Discord message links, and a clear action.
- Added redacted diagnostics via **Copy diagnostics**, **Download diagnostics**,
  `GET /api/diagnostics`, and `--doctor`. Webhook URLs, tokens, and secret IDs
  are never included.
- Added clearer empty states and next-step actions throughout the dashboard.

### Documentation

- Added [docs/twitch-import.md](docs/twitch-import.md),
  [docs/auto-update.md](docs/auto-update.md), and
  [docs/diagnostics.md](docs/diagnostics.md).
- Updated README, roadmap, project context, Docker guide, and release notes.

## 2.4.0 - 2026-11-16

### Packaging

- Added a per-user Inno Setup Windows installer with Start Menu and optional
  desktop shortcuts. Upgrades preserve `%APPDATA%\TwitchLiveNotifier`.
- Added optional Windows code-signing hooks for the executable and installer.
- Added winget manifest templates, a render script, and a release-time winget zip.
- Added macOS `.app` bundles for Intel and Apple Silicon plus optional signing
  and notarization hooks.
- Added Linux `.deb` and AppImage packages, a desktop entry, icon, and systemd
  user service.

### Updates

- Added a cached GitHub-release update check with version comparison, download
  link, release-notes link, and a Settings card.
- Added a Settings toggle to disable outbound update checks.
- The app reports its version and update status but does not silently replace its
  own binary.

### Rate limits

- Persisted a per-destination `retry_until` window after Discord `429`s, so a
  rate-limited destination is skipped until Discord's window expires instead of
  being retried every check.
- Long `429` windows no longer hold a check open; they are deferred and retried
  on a later check.
- Added proactive `X-RateLimit-Remaining` / `X-RateLimit-Reset-After` handling
  and a per-webhook in-memory limiter.
- Added jittered fallback delays when Discord does not provide a retry delay.
- Added a 15-minute hard timeout for scheduled and manual checks.
- Destination cards and streamer routing rows show when a webhook is rate
  limited.

## 2.3.0 - 2026-11-16

### Stream lifecycle

- Each streamer→destination route now stores the Discord message ID, original
  webhook URL, and thread ID, so later edits and cleanup target the same post.
- Added per-destination stream-end behavior: leave the message, delete it, edit
  it into an ended message, or post a separate ended message.
- Ended messages support last title/game, the `{vod}` placeholder, and their own
  template, with pings off by default.
- Live messages can be edited while a stream is live at a configurable interval
  (minimum 60 seconds). Turning updates off leaves the original post untouched.
- Cleanup and edit failures are logged and retried safely; they never block the
  next stream.
- Existing `last_stream_id` / `last_posted_at` dedupe still survives restarts.

### Destination identity and targets

- Added per-destination webhook name and avatar overrides.
- Added an optional thread ID for posting into an existing Discord thread or
  forum post.
- Added up to five link buttons per destination, with placeholders in labels and
  URLs.

### Template power

- Added an optional template override for each streamer→destination route.
- Added conditional fallbacks such as `{game|Just Chatting}`.
- Added custom embed fields and a configurable field order.
- Added template import/export and destination duplication.

### Secret storage and config UX

- Webhook URLs are encrypted at rest using DPAPI on Windows, Keychain on macOS,
  and libsecret on Linux when available. A private 0600 key file is used as the
  fallback when no keyring is available.
- Added redacted-by-default config export, an explicit with-secrets export, and
  restore from backup in Settings.
- Added an optional dashboard token for API access when the dashboard is exposed
  beyond loopback. `/healthz` remains available for Docker.
- Added a friendly warning when a redacted backup is restored and the webhook URL
  needs to be re-entered.

### Upgrade notes

- The config schema is now version 4. Version 3 configs load automatically;
  new lifecycle, identity, and template fields default to off/empty.
- Webhook URLs move into `secrets.json` (plus `secret.key` when using the
  fallback) next to `config.json`. Back up the whole settings folder, or use the
  with-secrets export if you need to move a configuration to another machine.
- A redacted export intentionally contains no webhook URLs. Re-enter them on the
  Destinations page after restoring it.
- Backups never include the dashboard token. Set it again after restoring if you
  still need API protection.

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
