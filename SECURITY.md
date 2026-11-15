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
