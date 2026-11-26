# Delivery history and diagnostics

## History

The dashboard's **History** view records delivery outcomes locally:

- live posts, live-message edits, and stream-end cleanup,
- filtered or rate-limited skips,
- Discord and Twitch delivery errors,
- test/sample posts.

Each record includes the streamer, destination label, stream ID, title, game,
message ID, a Discord message link when available, and a reason. History is
stored in `delivery-history.json` next to `config.json`, is capped at 1,000
entries, and is written with `0600` permissions.

History never stores webhook URLs, webhook tokens, secret-store IDs, or the
dashboard token. **Clear history** removes the records without touching
streamers, destinations, templates, or settings.

## Diagnostics

Use **Settings → Diagnostics → Copy diagnostics** or **Download diagnostics** to
produce a redacted support bundle. The same JSON is available from:

- `GET /api/diagnostics`
- `TwitchLiveNotifier --doctor`

The bundle contains:

- app/version, Go version, platform, data/config paths, uptime, last check,
- streamer counts (total/enabled/live/errors),
- destination labels and health (enabled, invalid, last error),
- recent errors and recent delivery history,
- update status and install-job state.

It never contains webhook URLs, tokens, secret IDs, or raw config secrets.
Strings are redacted before output, and the update status is round-tripped
through the redactor so future nested fields are covered too.

## Support workflow

1. Reproduce the problem and note the streamer, destination label, and time.
2. Open **History** and filter by that streamer or action.
3. If the post never appeared, check the record's reason and the destination
   health on the Destinations page.
4. If you need to share details, use **Copy diagnostics** rather than a
   screenshot of the destination editor.
5. Never post a with-secrets backup, `secrets.json`, or `secret.key` publicly.

## Why a stream was skipped

Common reasons include:

- a game/title include or exclude rule,
- minimum viewers,
- alert delay or cooldown,
- quiet hours or an allowed-day rule,
- the destination is paused, invalid, or rate-limited,
- the stream was already posted to that destination,
- the Twitch check failed.

Filtered streams still update live status and do not consume dedupe. A failed
Discord post does not record the stream ID, so the next check retries it.
