# Project context for a clean session

This file is the hand-off brief for a new agent/session with no prior chat
context.

## Repository

- **Local path:** `/mnt/arty/twitch-live-notifier`
- **GitHub:** <https://github.com/ausbitbank/twitch-live-notifier>
- **Current version:** 2.1.0
- **Current release:** <https://github.com/ausbitbank/twitch-live-notifier/releases/tag/v2.1.0>
- **Branch:** `main`
- **License:** MIT

## What the product is

A cross-platform local app that monitors Twitch channels and posts rich alerts
to labelled Discord webhooks.

Core promises:

- No Discord bot, Twitch OAuth, or Twitch API registration required.
- Multiple streamers and multiple labelled Discord destinations.
- Per-destination message templates with placeholders and a live Discord preview.
- Once-per-stream, per-destination dedupe that survives restarts.
- Optional per-streamer Discord ping using `{discord_mention}`.
- Local-first privacy: dashboard and settings run on `127.0.0.1`.
- Windows tray app; Linux/macOS dashboard mode (`--serve`).

## Stack

- **Go 1.22+**, standard library only.
- Embedded HTML/CSS/JS dashboard via `go:embed web`.
- Native Win32 tray/message box code for Windows.
- Browser dashboard and HTTP API for all platforms.
- JSON config in the per-user config directory.
- Windows resources (icon/version/manifest) from `rsrc_windows_amd64.syso`.

## Architecture map

| File | Responsibility |
| --- | --- |
| `main.go` | CLI parsing, mode dispatch, usage, `--check`, `--serve`, `--background` |
| `app.go` | App state, monitor loop, Twitch checks, routing, dedupe, mutations, activity |
| `config.go` | Config types, defaults, migration, validation, persistence, Discord ID normalization |
| `server.go` | Local HTTP server, JSON API, embedded dashboard, security headers, quit/health |
| `templates.go` | Placeholders, rendering, Discord payload builder, test payload |
| `twitch.go` | Twitch public web GraphQL client |
| `discord.go` | Discord webhook POST, payload types, error formatting |
| `platform_windows.go` | Windows shell/file/startup/console helpers |
| `tray_windows.go` | Win32 tray icon, menu, hidden window, single-instance mutex |
| `platform_other.go` | Linux/macOS browser/folder/autostart stubs, desktop-mode error |
| `web/` | Embedded dashboard (`index.html`, `app.js`, `styles.css`, icon, mascot) |
| `winres/` | Icon PNGs and Windows version/manifest source |
| `docs/` | Screenshots, competitive research, roadmap, this hand-off |
| `build.sh` | Cross-compile Windows from Linux/macOS |
| `build.ps1` | Windows PowerShell build |
| `build-linux.sh` | Native Linux dashboard binary |
| `build-macos.sh` | Intel + Apple Silicon macOS dashboard binaries |
| `tools/generate-resources.sh` | Regenerate icons, web artwork, `.syso` |

## Current runtime behavior

- On start, the app begins monitoring after ~1.5 seconds (unless paused).
- Default check interval: **120 seconds**; minimum 30 seconds.
- Checks all enabled streamers, then posts to each assigned destination.
- A destination stores `last_stream_id` and `last_posted_at`; matching stream IDs
  are skipped.
- Manual **Check all** also respects dedupe.
- A new Twitch stream ID posts again.
- A failed Discord post does not record the stream ID, so it retries on the next
  check.
- Config is saved after profile changes and successful posts.
- Windows runs in the tray; Linux/macOS run the dashboard in the foreground.
- Optional streamer Discord ID renders `{discord_mention}` as `<@id>`.
- `allowed_mentions` uses `users: [id]` only, so `@everyone` and roles are
  blocked.

## Current endpoints/modes

- `GET /api/state` — full dashboard state.
- `POST /api/channels`, `PUT/DELETE /api/channels/{id}`.
- `POST /api/channels/{id}/check`, `/sample`.
- `PUT /api/channels/{id}/destinations`.
- `POST /api/webhooks`, `PUT/DELETE /api/webhooks/{id}`.
- `POST /api/webhooks/{id}/test`.
- `POST /api/preview`.
- `POST /api/check`, `PUT /api/settings`, `/api/open-data-folder`, `/api/quit`.
- Browser dashboard at `http://127.0.0.1:8787/` (falls back to the next port).

## Settings locations

- **Windows:** `%APPDATA%\TwitchLiveNotifier\config.json`
- **Linux:** `~/.config/TwitchLiveNotifier/config.json`
- **macOS:** `~/Library/Application Support/TwitchLiveNotifier/config.json`

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
./build-macos.sh    # macOS Intel + Apple Silicon binaries
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
```

## Release flow

- Main branch is pushed to GitHub.
- Tags matching `v*` trigger `.github/workflows/release.yml`.
- Release workflow:
  - runs tests,
  - installs `go-winres`,
  - builds Windows, Linux, and macOS binaries,
  - creates Windows ZIP, Linux tar.gz, macOS tar.gz,
  - uploads raw binaries plus archives and `SHA256SUMS.txt`.
- Current tag: `v2.1.0`.

## Known constraints

- Polling is every 2 minutes by default; very short streams can be missed.
- Twitch checks use the public web GraphQL endpoint, not the official API.
- No role pings, filters, quiet hours, stream-end cleanup, live message updates,
  live roles, schedule sync, or Twitch Teams yet.
- Windows has the native tray and single-instance mutex. Linux/macOS have no
  native tray and no cross-platform single-instance lock yet.
- Webhook URLs are stored in plain text in the per-user config.
- Windows binary is unsigned.
- No Docker/self-host package or healthcheck endpoint yet.
- No retry/backoff for Discord 429 or Twitch 5xx yet.
- No official Twitch API/EventSub mode.

## Do not regress

- Core operation must not require a Discord bot or Twitch login.
- Multi-destination routing and per-destination templates must remain.
- Dedupe must remain per stream/per destination and survive restarts.
- `@everyone` and role mentions must remain blocked unless explicitly enabled.
- The dashboard must remain local-only by default.
- Windows/Linux/macOS builds must keep passing CI.
- Do not commit `dist/`, config files, real webhook URLs, or secrets.

## Read next

1. [Competitive research](competitive-research.md)
2. [Roadmap](roadmap.md)
3. [README](../README.md)

## Fresh session start prompt

```text
We are continuing work on Twitch Live to Discord at
/mnt/arty/twitch-live-notifier. Read README.md, docs/project-context.md,
docs/competitive-research.md, and docs/roadmap.md first. The current release is
v2.1.0 on main.

Start with the P0 v2.2 roadmap item: Docker/self-host package plus a healthcheck
endpoint. Then proceed through the remaining v2.2 P0 items in roadmap order:
retry/backoff + Discord 429 handling, cross-platform single-instance lock, role
pings, and notification filters/quiet hours as time allows.

Before coding, restate the acceptance criteria and make a short implementation
plan. Preserve the core behavior: no bot/OAuth required, multi-streamer and
multi-destination routing, per-destination templates and live preview, persisted
per-stream/per-destination dedupe, and blocked @everyone/roles unless explicitly
configured.

Implement one slice at a time. Add/update tests, run go test ./..., go vet for
Linux and Windows, and race tests. Run UI/Playwright smoke tests where relevant.
Update README, CHANGELOG, and docs. Commit and push to main. If all checks pass,
tag and release v2.2.0 and verify the GitHub release assets and checksums.
```
