# Project context for a clean session

This file is the hand-off brief for a new agent/session with no prior chat
context.

## Repository

- **Local path:** `/mnt/arty/twitch-live-notifier`
- **GitHub:** <https://github.com/ausbitbank/twitch-live-notifier>
- **Current version:** 2.5.0
- **Current release:** <https://github.com/ausbitbank/twitch-live-notifier/releases/tag/v2.5.0>
- **Next focus:** free OSS signing application (SignPath Foundation/OSSign),
  signed installer rollout, and v2.6 planning.
- **Branch:** `main`
- **License:** MIT

## What the product is

A cross-platform local app that monitors Twitch channels and posts rich alerts
to labelled Discord webhooks.

Core promises:

- No Discord bot, Twitch OAuth, or Twitch API registration required for alerts.
  Follows can be imported with a browser-session bookmarklet or a pasted list;
  the optional device-code sign-in is a fallback.
- Multiple streamers and multiple labelled Discord destinations.
- Per-destination message templates with placeholders and a live Discord preview.
- Per-route template overrides, conditional fallbacks, custom fields, and field
  ordering.
- Once-per-stream, per-destination dedupe that survives restarts.
- Stored live message IDs so live messages can be edited and cleaned up on
  stream end.
- Optional per-streamer Discord ping using `{discord_mention}`.
- Optional per-destination role pings using `{role_mentions}`.
- Per-destination webhook name/avatar overrides, thread targets, and link buttons.
- Game/title/viewer filters, quiet hours, allowed days, cooldowns, and delays.
- Search, bulk management, and a streamer × destination routing matrix.
- Encrypted webhook URLs at rest, redacted backups, and an optional dashboard
  token.
- Local-first privacy: dashboard and settings run on `127.0.0.1`.
- Windows tray app; Linux/macOS dashboard mode (`--serve`).
- Docker/self-host mode with `/healthz` and multi-arch images.
- Release packaging for Windows, macOS, Linux, and Docker, with checksum-verified
  in-app auto-update (and optional Ed25519 manifest signing).
- Persisted delivery history, redacted diagnostics, and `--doctor`.

## Stack

- **Go 1.22+**, standard library only.
- Embedded HTML/CSS/JS dashboard via `go:embed web`.
- Native Win32 tray/message box code for Windows.
- Browser dashboard and HTTP API for all platforms.
- JSON config in the per-user config directory (or `/data` in Docker).
- Windows resources (icon/version/manifest) from `rsrc_windows_amd64.syso`.

## Architecture map

| File | Responsibility |
| --- | --- |
| `main.go` | CLI parsing, mode dispatch, usage, `--check`, `--serve`, `--background`, `--healthcheck` |
| `app.go` | App state, monitor loop, Twitch checks, routing, dedupe, stream lifecycle, filters, mutations, activity |
| `config.go` | Config types, defaults, migration, validation, persistence, Discord ID/role/button normalization |
| `secret.go` | Secret store, AES-GCM protection, fallback key file |
| `secret_unix.go` | macOS Keychain / Linux libsecret key lookup and storage |
| `secret_windows.go` | Windows DPAPI protect/unprotect |
| `secret_config.go` | Resolve/encrypt config secrets, redacted export/import |
| `twitch_oauth.go` | Device-code sign-in, Helix user/follows lookups |
| `twitch_batch.go` | Aliased Twitch GraphQL batching with bounded concurrency |
| `import.go` | Bulk add, follows import sessions, HTML/paste parsing, one-write imports |
| `server_import.go` | `POST /api/channels/import` and the no-login import endpoint |
| `server_bulk.go` | `POST /api/channels/bulk-action` and bulk channel mutations |
| `auth.go` | Optional dashboard token hashing and API middleware |
| `update.go` | Cached GitHub-release update check, manifest parsing, asset selection |
| `update_manager.go` | Download/verify/apply job, progress, restart orchestration |
| `update_apply*.go` | Platform-specific binary, AppImage, Windows helper, macOS bundle swap |
| `update_restart.go` | Restart-child wait and relaunch environment |
| `update_key.go` | Optional pinned Ed25519 manifest verification |
| `server_update.go` | `GET/POST /api/updates/install` |
| `history.go` | Persisted delivery-history log and filters |
| `diagnostics.go` | Redacted support bundle and `DiagnosticsJSON` |
| `server_history.go` | `/api/history` and `/api/diagnostics` |
| `server_webhook_inspect.go` | Non-secret webhook metadata for onboarding labels |
| `filters.go` | Per-streamer filter normalization, validation, and evaluation |
| `health.go` | `/healthz` payload, monitor liveness rules, healthcheck CLI probe |
| `retry.go` | Exponential backoff with jitter and context-aware retry sleeps |
| `server.go` | Local HTTP server, JSON API, embedded dashboard, security headers, listen/health/quit |
| `templates.go` | Placeholders, fallback syntax, custom fields, route templates, mention scoping, payload builder |
| `twitch.go` | Twitch public web GraphQL client with bounded retries |
| `discord.go` | Discord webhook POST, 429 handling, payload types, error classification |
| `discord_limiter.go` | Per-webhook proactive rate-limit buckets and deferral |
| `instance_lock.go` | Shared lock path resolution |
| `instance_lock_unix.go` | Advisory `flock` single-instance lock for Linux/macOS |
| `instance_lock_windows.go` | Named-mutex single-instance lock for Windows |
| `platform_windows.go` | Windows shell/file/startup/console helpers |
| `tray_windows.go` | Win32 tray icon, menu, hidden window, single-instance mutex |
| `platform_other.go` | Linux/macOS browser/folder/autostart stubs, desktop-mode error |
| `web/` | Embedded dashboard (`index.html`, `app.js`, `styles.css`, icon, mascot) plus the no-login follows helper/bookmarklet |
| `Dockerfile` | Multi-stage amd64/arm64 image with healthcheck |
| `docker-compose.yml` | Persistent volume, loopback port, timezone, restart policy |
| `winres/` | Icon PNGs and Windows version/manifest source |
| `docs/` | Screenshots, Docker/signing guides, competitive research, roadmap, this hand-off |
| `packaging/windows/` | Inno Setup script, installer/signing PowerShell |
| `packaging/macos/` | `.app` bundle, launcher, Info.plist, sign/notarize script |
| `packaging/linux/` | `.deb`, AppImage, desktop entry, systemd user service |
| `packaging/winget/` | winget manifest templates and renderer |
| `build.sh` | Cross-compile Windows from Linux/macOS |
| `build.ps1` | Windows PowerShell build |
| `build-linux.sh` | Native Linux dashboard binary |
| `build-macos.sh` | macOS dashboard binaries plus `.app` bundles |
| `tools/generate-resources.sh` | Regenerate icons, web artwork, `.syso` |

## Current runtime behavior

- On start, the app begins monitoring after ~1.5 seconds (unless paused).
- Default check interval: **120 seconds**; minimum 30 seconds.
- Checks all enabled streamers in aliased GraphQL batches (up to 20 logins per
  request, at most 3 requests in flight), then posts to each assigned
  destination sequentially to respect Discord rate limits.
- Twitch 5xx, 429, timeout, and network failures retry up to three times with
  jittered exponential backoff. `404`/API errors do not retry. Batched requests
  fall back to the single-channel path if Twitch rejects the batch shape.
- A destination stores `last_stream_id` and `last_posted_at`; matching stream IDs
  are skipped.
- A posted live alert also stores a `LiveMessage` record with the Discord message
  ID, original webhook URL, thread ID, stream ID, title/game, VOD URL, and last
  edit time.
- While a stream is live, `update_seconds` can edit the same message at a minimum
  60-second interval. Off leaves it untouched.
- On stream end, the per-destination cleanup mode can leave, delete, edit into an
  ended message, or post a separate ended message. Ended messages default to no
  pings.
- Cleanup/edit failures are logged and retried, but never block a future stream.
- Manual **Check all** also respects dedupe.
- A new Twitch stream ID posts again.
- A failed Discord post does not record the stream ID, so it retries on the next
  check.
- Discord `429` honors `Retry-After`/`retry_after`. Short delays retry inline
  once; long delays persist a per-destination `retry_until` window and the route
  is skipped until it expires. Proactive `X-RateLimit-*` headers are honored.
- Rate-limited destinations do not record dedupe, so the alert is retried after
  the window. Destination cards and streamer routing rows show the rate-limit
  state.
- Scheduled and manual checks have a 15-minute timeout.
- Discord `401`/`403`/`404` marks the destination invalid and skips it until the
  URL changes or a test succeeds. Other destinations keep working.
- Config is saved after profile changes, successful posts, live edits, cleanup,
  and settings changes.
- Windows runs in the tray. Linux/macOS with no arguments (including the macOS
  `.app` bundle and Linux desktop launcher) start the dashboard and open the
  browser; `--serve` runs the dashboard in the foreground service mode.
- Optional streamer Discord ID renders `{discord_mention}` as `<@id>`.
- Optional destination role IDs render `{role_mentions}` as `<@&id>`.
- `allowed_mentions` always includes an explicit `parse` list; `users` and
  `roles` contain only configured IDs. `@everyone`/`@here` require an explicit
  per-destination opt-in.
- Per-destination webhook username/avatar, thread ID, and up to five link
  buttons are applied to posts and previews.
- A route-level template override changes only that streamer→destination message;
  mention controls stay destination-wide.
- Webhook URLs are encrypted into `secrets.json` with DPAPI/Keychain/libsecret,
  or a 0600 `secret.key` fallback. Redacted backups omit URLs; with-secrets
  backups include them.
- An optional dashboard token protects `/api/*`; `/healthz` remains public.
- The app checks the latest GitHub release after startup and every 24 hours when
  enabled. Results are cached; Settings shows current/latest versions, release
  notes, and a download. When the install target is writable and the asset has a
  SHA-256 checksum, **Install update** downloads, verifies, swaps, and restarts.
  Windows stages a detached helper; macOS swaps the `.app` bundle; Docker and
  system packages show a guided/manual path.
- Update checks can be disabled in Settings; the app still works offline.
- Delivery outcomes (posts, edits, cleanup, skips, errors) are persisted in
  `delivery-history.json` (1,000 entries) and exposed in the History view.
  Diagnostics are redacted and available from Settings or `--doctor`.
- First-run onboarding tracks destination, streamer, routing, test/sample, and
  first-check steps in the dashboard.
- Import accepts a pasted list, saved HTML, JSON/CSV, or the no-login browser
  helper, up to 2,000 unique logins per request.
- Per-streamer filters evaluate game/title rules, minimum viewers, delay,
  quiet hours, allowed days, cooldown, and first-stream-of-day. Excludes win.
- Filtered streams still update live status and do not record dedupe.
- One instance per settings file. Unix uses `flock`; Windows uses the named
  mutex. Stale lock files do not block because the kernel releases the lock.
- `/healthz` is healthy while checks are current, unhealthy when stale, and
  healthy when monitoring is intentionally paused.

## Current endpoints/modes

- `GET /api/state` — full dashboard state.
- `GET /healthz` — monitor health, last check, counts; HTTP 200/503.
- `POST /api/channels`, `PUT/DELETE /api/channels/{id}` (enabled, Discord ID,
  filters).
- `POST /api/channels/{id}/check`, `/sample`.
- `POST /api/channels/import` (pasted/file/helper list, up to 2,000, optional
  `webhook_ids`).
- `POST /api/channels/bulk-action` (enable/disable/delete/assign/unassign many
  channels).
- `POST /api/channels/bulk` (legacy paste-only import, up to 2,000).
- `GET /follows-helper` (local bookmarklet helper page; `/static/follows-*.js`
  assets).
- `POST /api/twitch/imports`, `GET /api/twitch/imports/{id}`,
  `POST /api/twitch/imports/{id}/import`, `DELETE /api/twitch/imports/{id}`
  (optional device-code follows import).
- `GET /api/history`, `DELETE /api/history` (delivery history).
- `GET /api/diagnostics` (redacted support bundle).
- `PUT /api/channels/{id}/destinations` (routing).
- `PUT /api/channels/{id}/destinations/{webhookID}` (route template override).
- `POST /api/webhooks`, `PUT/DELETE /api/webhooks/{id}` (label, URL, roles,
  template, identity, thread, buttons, update interval, cleanup, enabled).
- `POST /api/webhooks/{id}/test`.
- `POST /api/webhooks/inspect` (non-secret webhook metadata for label
  suggestions).
- `POST /api/preview` (template, channel, label, role IDs, identity, buttons,
  ended preview).
- `GET /api/config/export` (redacted by default, `?secrets=1` for full).
- `POST /api/config/import` (restore backup).
- `PUT /api/dashboard-token` (set or clear the optional API token).
- `GET /api/updates` (cached status + install job), `POST /api/updates/check`
  (manual check), `POST/GET /api/updates/install` (start/poll install).
- `POST /api/check`, `PUT /api/settings`, `/api/open-data-folder`, `/api/quit`.
- Browser dashboard at `http://127.0.0.1:8787/` (falls back to the next port).
  `--listen` binds an explicit address; Docker uses `0.0.0.0:8787` internally.

## Settings locations

- **Windows:** `%APPDATA%\TwitchLiveNotifier\config.json`
- **Linux:** `~/.config/TwitchLiveNotifier/config.json`
- **macOS:** `~/Library/Application Support/TwitchLiveNotifier/config.json`
- **Docker:** `/data/config.json` in the `twitch-live-data` volume

The same folder also contains `secrets.json` (encrypted webhook URLs),
`delivery-history.json` (bounded delivery outcomes, no secrets), and, when no OS
keyring is available, `secret.key` (the fallback 0600 AES key). Back up the whole
folder or use the with-secrets export.

## Commands

```bash
# Tests and static checks
go test ./...
go vet ./...
GOOS=windows GOARCH=amd64 go vet ./...
go test -race ./...

# Builds
./build.sh          # Windows exe into dist/
./build-linux.sh    # Linux dashboard binary
./build-macos.sh    # macOS Intel + Apple Silicon binaries and .app bundles
docker build -t twitch-live-notifier .
docker compose up -d --build

# Packages
./packaging/linux/build-deb.sh <version>
APPIMAGETOOL=/path/to/appimagetool ./packaging/linux/build-appimage.sh <version>
./packaging/macos/build-app.sh <version> <amd64|arm64>
./packaging/windows/build-installer.ps1 -Version <version>   # on Windows
./packaging/winget/render.sh <version> <installer.exe>
```

Local artifacts are written under `dist/` and are ignored by Git.

Run the dashboard without the tray:

```bash
./dist/TwitchLiveNotifier-linux-amd64 --serve
```

Windows CLI:

```text
TwitchLiveNotifier.exe
TwitchLiveNotifier.exe --background
TwitchLiveNotifier.exe --check
TwitchLiveNotifier.exe --serve
TwitchLiveNotifier.exe --healthcheck
TwitchLiveNotifier.exe --doctor
```

## Release flow

- Main branch is pushed to GitHub.
- Tags matching `v*` trigger `.github/workflows/release.yml`.
- Release workflow:
  - runs tests,
  - installs `go-winres`,
  - builds Windows, Linux amd64/arm64, and macOS binaries,
  - builds Linux `.deb` and AppImage packages,
  - builds macOS `.app` bundles,
  - creates Windows ZIP, Linux amd64/arm64 tar.gz, macOS tar.gz, and
    `SHA256SUMS.txt`,
  - uploads raw binaries, packages, archives, and checksums,
  - builds the Windows installer, winget manifests, and signed Windows artifacts
    on a Windows runner,
  - optionally signs/notarizes macOS artifacts when secrets are configured,
  - builds and pushes `linux/amd64` + `linux/arm64` images to GHCR,
  - generates `update-manifest.json` for self-update and optionally signs it
    when `UPDATE_SIGNING_KEY` is configured.
- Current tag: `v2.5.0`.

## Known constraints

- Polling is every 2 minutes by default; very short streams can be missed.
  Batched checks keep hundreds of streamers inside the poll window, but detection
  is still poll-based.
- Twitch checks use the public web GraphQL endpoint, not the official API.
- No live-role automation, schedule sync, or Twitch Teams yet.
- Windows has the native tray. Linux/macOS have no native tray.
- All platforms prevent multiple instances per settings file.
- Webhook URLs are encrypted at rest where a keyring is available. Without a
  keyring, the fallback key file is only as private as the settings folder.
- Windows and macOS artifacts are unsigned unless signing is configured. Free
  Windows signing is possible via SignPath Foundation or OSSign for eligible OSS;
  macOS Developer ID signing requires a paid Apple Developer Program membership.
  A free Homebrew/source-build path for macOS is included.
- Auto-update verifies the release SHA-256 checksum. Optional Ed25519 manifest
  signing is supported when `updateSigningPublicKeyHex` is pinned and
  `UPDATE_SIGNING_KEY` is configured in release. Windows Authenticode and macOS
  notarization remain separate.
- The dashboard has no authentication by default. It can be protected with an
  optional dashboard token, but it should still stay on loopback or behind an
  authenticating reverse proxy when exposed.
- No official Twitch API/EventSub mode.
- Stream lifecycle, identity/threads, route templates, and secret storage
  shipped in v2.3. Installers/packages, update checks, and rate-limit hardening
  shipped in v2.4. No-login import, batched polling, verified auto-update,
  bulk routing, onboarding, and delivery diagnostics shipped in v2.5.

## Do not regress

- Core operation must not require a Discord bot or Twitch login.
- The no-login follows import must stay local: no Twitch token to the app, no
  backend, no relay. Paste/file import must keep working without the helper.
- Multi-destination routing and per-destination templates must remain.
- Dedupe must remain per stream/per destination and survive restarts.
- Stored message IDs must keep edits/cleanup targeting the original post, and a
  cleanup failure must never block the next stream.
- Batched polling must preserve the per-stream/per-destination dedupe and routing
  semantics of the single-channel path.
- Auto-update must verify the checksum before swapping and must not touch a
  non-writable/system-managed install without a clear manual path.
- Diagnostics and delivery history must never include webhook URLs, tokens,
  secret IDs, or raw config secrets.
- `@everyone` and role mentions must remain blocked unless explicitly enabled.
- The dashboard must remain local-only by default.
- Windows/Linux/macOS builds must keep passing CI.
- Do not commit `dist/`, config files, real webhook URLs, or secrets.

## Read next

1. [Competitive research](competitive-research.md)
2. [Roadmap](roadmap.md)
3. [Twitch follows import](twitch-import.md)
4. [Verified auto-update](auto-update.md)
5. [Delivery history and diagnostics](diagnostics.md)
6. [Docker and self-hosting](docker.md)
7. [README](../README.md)

## Fresh session start prompt

```text
We are continuing work on Twitch Live to Discord at
/mnt/arty/twitch-live-notifier. Read README.md, docs/project-context.md,
docs/competitive-research.md, docs/roadmap.md, docs/twitch-import.md,
docs/auto-update.md, and docs/diagnostics.md first. The current release is
v2.5.0 on main.

Next focus: verify the published v2.5.0 assets, then the free OSS signing
application (SignPath Foundation or OSSign), signed installer rollout, and v2.6
planning. Keep unsigned builds and checks passing while signing credentials are
absent.

Before coding, restate the acceptance criteria and make a short implementation
plan. Preserve the core behavior: no bot/OAuth required, no-login follows import
without a token/backend, multi-streamer and multi-destination routing,
per-destination templates and live preview, persisted per-stream/per-destination
dedupe, verified auto-update, scoped mentions, filters/quiet hours, and blocked
@everyone/roles unless explicitly configured.

Implement one slice at a time. Add/update tests, run go test ./..., go vet for
Linux/Windows/macOS, and race tests. Run UI smoke tests where relevant. Update
README, CHANGELOG, and docs. Commit and push to main. Tag and release only after
all checks and published assets are verified.
```
