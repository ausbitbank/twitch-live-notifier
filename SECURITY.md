# Security

## Discord webhook URLs

A Discord webhook URL is a credential. Anyone with the URL can post messages to
that channel. The app stores webhook URLs in the current user's profile:

- **Windows:** `%APPDATA%\TwitchLiveNotifier\`
- **Linux:** `~/.config/TwitchLiveNotifier/`
- **macOS:** `~/Library/Application Support/TwitchLiveNotifier/`

Webhook URLs are encrypted at rest in `secrets.json` using DPAPI on Windows,
Keychain on macOS, or libsecret on Linux when available. When no keyring is
available, the app uses a private 0600 `secret.key` file as a fallback, so the
settings folder must still be kept private.

The dashboard can reveal a webhook URL while editing a destination. The URLs are
not embedded in the executable and are not sent anywhere except to Discord. The
local dashboard only listens on `127.0.0.1` by default and requires same-origin
JSON requests for changes.

## Backups

**Download backup** redacts webhook URLs by default. **With secrets** is an
explicit action and includes live webhook URLs; treat that file like a password.
Restoring a redacted backup marks each destination invalid until its URL is
re-entered.

## Update checks and auto-update

When update checks are enabled (the default), the app requests the latest
release metadata from `api.github.com`. It sends no settings, webhook URLs, or
streamer data. Disable the check in Settings if you do not want that outbound
request.

Auto-update only runs after you click **Install update**. The app downloads the
release asset, verifies its SHA-256 checksum from the release manifest or
`SHA256SUMS.txt`, preflights the binary, and keeps the previous binary or bundle
for rollback. Releases may optionally publish an Ed25519-signed manifest; when a
public key is pinned, an unsigned or mismatched manifest is rejected. Windows
Authenticode and macOS notarization remain separate release-signing features.

Docker and system-package installs are not self-replaced; the dashboard shows the
appropriate `docker compose pull` or package-manager guidance.

## Follows import

The no-sign-in follows helper is a bookmarklet that you run on
`https://www.twitch.tv/directory/following/channels` in your existing browser
session. It reads links already rendered on that page, shows them in an editable
overlay, and copies them to your clipboard. It does not contact this app,
Twitch's API, Discord, or any other server. The app only receives the list you
paste. No Twitch token is created or stored.

The optional device-code sign-in fallback keeps its access token in memory only
for the import and discards it after the follows list is fetched.

## Delivery history and diagnostics

`delivery-history.json` stores delivery outcomes (streamer/destination labels,
stream ID, title, game, message ID/link, and error reason) and is capped at 1,000
entries. It never contains webhook URLs, tokens, or secret IDs. The same rule
applies to **Copy diagnostics**, **Download diagnostics**, `GET /api/diagnostics`,
and `--doctor`.

## Dashboard token

An optional dashboard token can be set under Settings. When present, `/api/*`
requires `X-Dashboard-Token` or a `?token=` query parameter. `/healthz` remains
unauthenticated for Docker and uptime monitors. A token is not a substitute for
keeping the port on loopback or behind an authenticating HTTPS reverse proxy.

## Exposing the dashboard

The dashboard has no login by default. If you use `--listen` or Docker to make it
reachable from other machines, set a dashboard token, put it behind an
authenticating HTTPS reverse proxy, and never publish port `8787` directly to the
internet. The Docker compose file publishes the port on `127.0.0.1` by default.

The `GET /healthz` endpoint is intentionally unauthenticated so Docker and
uptime monitors can reach it. It contains monitor status and counts, not webhook
URLs or message content. If your reverse proxy protects the dashboard, leave
`/healthz` reachable only from the local container/host where possible.

## Docker

Settings live in the `twitch-live-data` volume at `/data/config.json`, with
encrypted URLs in `/data/secrets.json` and the fallback key in
`/data/secret.key`. Protect access to the Docker host and volume backups. The
image runs as a non-root user.

Do not post your settings folder, a with-secrets backup, or an unmasked
screenshot of the destination editor in a public issue. If a webhook is exposed,
delete it in Discord and create a new one.

## Reporting a vulnerability

Open a private security advisory on the repository, or contact the maintainer
directly. Please do not disclose a working exploit publicly before it has been
addressed.
