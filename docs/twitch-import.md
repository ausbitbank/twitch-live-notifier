# Importing your Twitch follows

The Streamers page can import the channels a Twitch account already follows.
This is an optional convenience; monitoring, alerts, and everything else work
without connecting Twitch.

## Why some Twitch context is needed

Twitch no longer exposes a user's followed channels to anonymous requests.
Reading a private follows list requires the account's own browser session or a
user access token. The app gives you three ways to provide that context, and
none of them sends a token to a server:

| Option | What it needs | What the app receives | Best for |
| --- | --- | --- | --- |
| **Browser helper (recommended)** | You are already logged into Twitch in your browser | A pasted list of channel names | Most users; no Twitch app registration |
| **Paste or file** | A list from anywhere | The list | Users who already have an export, or want no Twitch interaction |
| **Sign in to Twitch** | A public Twitch client ID configured in the build | A temporary memory-only token; follows are fetched directly | Users who prefer a guided picker |

The browser helper is deliberately not an OAuth connection. It runs only on
`https://www.twitch.tv/directory/following/channels` in your existing browser
session, reads the channel links already rendered on that page, and copies them
to your clipboard. The app never receives a Twitch token, and there is no
backend or relay.

## Browser helper (no sign-in)

1. Open **Streamers → Import follows → No sign-in**.
2. Click **Open helper**. The helper page is served locally by the app.
3. Drag the purple **Export follows** button to your bookmarks bar. If the bar is
   hidden, show it first (usually `Ctrl/Cmd + Shift + B`). You can also copy the
   bookmarklet code and create a bookmark manually.
4. Open <https://www.twitch.tv/directory/following/channels> and wait for the
   channel cards to load.
5. Click the bookmark. It scrolls the page, collects unique channel logins, and
   shows an editable list. Review the list and click **Copy list**.
6. Back in the app, choose **Paste a list**, paste, optionally choose
   destinations, and click **Import follows**.

The bookmarklet only reads links already in your browser page. It does not call
Twitch's API, Discord, this app, or any other server. Because the overlay is
editable, you can remove anything that is not a streamer before importing.

### Troubleshooting

- **No channels found:** make sure you opened the full following page
  (`/directory/following/channels`), not the live-only sidebar, then scroll once
  and run the bookmark again.
- **Browser blocks the bookmarklet:** create a bookmark manually and paste the
  code from the helper page as its URL. Some browsers strip the `javascript:`
  prefix when pasting; add it back if needed.
- **The bookmarklet does nothing:** some pages block `javascript:` URLs through
  their security policy. Open DevTools (`F12`) → **Console** on the Twitch
  following page, click **Copy console script** in the helper page, paste it
  into the console, and press Enter.
- **The list includes something unexpected:** edit the overlay before copying.
  The import validates every entry with Twitch and marks channels that are
  already on your list.

## Paste or file import

- Paste usernames, channel URLs, comma-separated lists, or save a follows page
  and import the `.html` file.
- Up to **2,000 unique channels** per import.
- Channels are validated with Twitch in batched requests before anything is
  written.
- You can assign one or more destinations during the import. With exactly one
  destination configured, new streamers are routed to it automatically.
- Already-added streamers are skipped; unresolved entries stay in the box for a
  retry.

## Optional Twitch sign-in

The app also supports Twitch's **Device Code Grant Flow** (DCF) with the
`user:read:follows` scope. It is a public-client flow, so the project ships a
client ID and **no client secret**.

From **Streamers → Import follows → Sign in to Twitch**:

1. The app shows a short code and opens `twitch.tv/activate`.
2. Approve access for the requested scope (`user:read:follows` only).
3. Pick the channels to add. Channels already on your list are marked.
4. Click **Import selected**.

The access token is kept in memory only for the import and then discarded. No
Twitch token is written to the config file, the secret store, or backups. Start
the flow again for a later import. The button is hidden unless a public client
ID is configured.

## For maintainers and self-hosters

Without a client ID the optional sign-in button stays hidden; the browser helper
and paste/file import still work. To enable the DCF fallback, register a Twitch
application once:

1. Open <https://dev.twitch.tv/console/apps> and sign in.
2. Choose **Register Your Application**.
3. Set **OAuth Redirect URLs** to any localhost URL (the device flow does not use
   it, but the field is required), for example `http://localhost`.
4. Set **Client Type** to **Public**. Do not generate a secret.
5. Copy the **Client ID** (about 30 characters).

Then provide it in one of these ways:

- **Source or ldflags:** set `defaultTwitchOAuthClientID` in `twitch_oauth.go`,
  or build with
  `-ldflags "-X main.defaultTwitchOAuthClientID=<client id>"`.
- **Runtime environment:** set `TWITCH_LIVE_CLIENT_ID` for the app process,
  including the Docker container.

The environment variable takes precedence. Because the client ID is public, none
of these options require storing a secret.

## API surface

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/api/channels/import` | Import a pasted/file list, with optional `webhook_ids` |
| `POST` | `/api/channels/bulk-action` | Enable, disable, delete, assign, or unassign many streamers |
| `GET` | `/follows-helper` | Local helper page with the draggable bookmarklet |
| `POST` | `/api/channels/bulk` | Legacy paste-only import |
| `POST` | `/api/twitch/imports` | Start the optional device-code sign-in |
| `GET` | `/api/twitch/imports/{id}` | Poll sign-in state and load follows |
| `POST` | `/api/twitch/imports/{id}/import` | Add the selected follows |
| `DELETE` | `/api/twitch/imports/{id}` | Cancel and discard the session |

Bulk and helper imports write the config once and trigger a single monitor
check. The DCF import uses Helix data and is capped at 1,000 selected channels
per session.
