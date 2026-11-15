# Security

## Discord webhook URLs

A Discord webhook URL is a credential. Anyone with the URL can post messages to
that channel. The app stores webhook URLs in the current user's profile:

- **Windows:** `%APPDATA%\TwitchLiveNotifier\config.json`
- **Linux:** `~/.config/TwitchLiveNotifier/config.json`
- **macOS:** `~/Library/Application Support/TwitchLiveNotifier/config.json`

The webhook URLs are not embedded in the executable, logged, or sent anywhere
except to Discord. The local dashboard only listens on `127.0.0.1` and requires
same-origin JSON requests for changes.

Do not post your `config.json` or an unmasked screenshot of the destination
editor in a public issue. If a webhook is exposed, delete it in Discord and
create a new one.

## Reporting a vulnerability

Open a private security advisory on the repository, or contact the maintainer
directly. Please do not disclose a working exploit publicly before it has been
addressed.
