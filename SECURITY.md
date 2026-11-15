# Security

## Discord webhook URLs

A Discord webhook URL is a credential. Anyone with the URL can post messages to
that channel. The app stores webhook URLs in the current user's profile:

- **Windows:** `%APPDATA%\TwitchLiveNotifier\config.json`
- **Linux:** `~/.config/TwitchLiveNotifier/config.json`
- **macOS:** `~/Library/Application Support/TwitchLiveNotifier/config.json`

The webhook URLs are not embedded in the executable, logged, or sent anywhere
except to Discord. The local dashboard only listens on `127.0.0.1` by default
and requires same-origin JSON requests for changes.

## Exposing the dashboard

The dashboard has no login. If you use `--listen` or Docker to make it reachable
from other machines, put it behind an authenticating HTTPS reverse proxy and
never publish port `8787` directly to the internet. The Docker compose file
publishes the port on `127.0.0.1` by default. An exposed dashboard can add
streamers, change routing, and reveal masked webhook URLs.

The `GET /healthz` endpoint is intentionally unauthenticated so Docker and
uptime monitors can reach it. It contains monitor status and counts, not webhook
URLs or message content. If your reverse proxy protects the dashboard, leave
`/healthz` reachable only from the local container/host where possible.

## Docker

Settings live in the `twitch-live-data` volume at `/data/config.json`. Protect
access to the Docker host and volume backups; the file contains webhook URLs in
plain text. The image runs as a non-root user.

Do not post your `config.json` or an unmasked screenshot of the destination
editor in a public issue. If a webhook is exposed, delete it in Discord and
create a new one.

## Reporting a vulnerability

Open a private security advisory on the repository, or contact the maintainer
directly. Please do not disclose a working exploit publicly before it has been
addressed.
