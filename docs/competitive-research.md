# Competitive research: Twitch → Discord alerts

**Research date:** 2026-11-15  
**Product version reviewed:** 2.1.0  
**Scope:** Twitch live alerts for Discord communities, not general-purpose Discord bots.

## Method

I reviewed public product pages and documentation for the main cloud bots,
automation platforms, self-hosted bots, and local stream automation tools. I also
reviewed Twitch EventSub documentation and Discord rate-limit documentation
because those constrain what a local webhook app can realistically do.

MEE6 and Carl-bot product pages were JavaScript-heavy or blocked direct fetching,
so they are treated here as the general-purpose bot category rather than used
for precise feature claims.

## Market categories

1. **Cloud notification bots** — Streamcord, MEE6, Carl-bot. One-click install,
   persistent uptime, role pings, live roles, dashboards, paid tiers.
2. **General automation** — IFTTT, Zapier. OAuth to Twitch and Discord, simple
   message actions, limited formatting, polling intervals/task quotas.
3. **General-purpose self-hosted bots** — YAGPDB. Twitch feeds plus broad
   server automation, advanced templates, self-hosting.
4. **Local stream automation platforms** — Streamer.bot. Very powerful local
   actions, Twitch/YouTube/Kick events, Discord webhooks, C# scripting.
5. **Small open-source scripts** — usually one streamer, config-file setup,
   Node/Python, often no multi-destination routing or UI.

## Competitor snapshot

| Product | Model | Platforms | Relevant strengths | Gaps / tradeoffs |
| --- | --- | --- | --- | --- |
| **Streamcord** | Cloud Discord bot | Twitch, YouTube, Kick | Fast notifications, custom message/embed color/thumbnail, game/title filters, Teams, auto-crosspost, stream-end cleanup, live roles, schedule sync, custom webhooks. Free 5 Twitch channels; Pro $2.99/mo for 250. | Requires bot install/OAuth and cloud dependency. Notification model is channel-oriented; multi-destination routing is not the core mental model. |
| **YAGPDB** | Cloud bot, self-hostable | Twitch, YouTube, RSS, Reddit | Twitch feeds within a few minutes, advanced templates, role/user pings, broad Discord automation, open-source self-host option. | General-purpose bot complexity. Not focused on a polished Twitch alert setup for non-technical communities. |
| **Zapier** | Cloud automation | Twitch + Discord | OAuth triggers for new live streams, filters by game/language/streamer/min viewers, Discord message action with username/avatar, huge integration catalog. | Task-based pricing, generic message shape, more setup for rich embeds, not a dedicated alert dashboard. |
| **IFTTT** | Cloud automation | Twitch + Discord | Very simple “when live, post to Discord” applet. | Polling intervals: 1 hour free, 15 min Starter, 5 min Pro. Limited formatting and controls. |
| **Streamer.bot** | Local automation platform | Twitch, YouTube, Kick | Local-first, extremely flexible actions, Discord webhooks, OBS integration, C# scripting, donations and chat events. | Steep learning curve. Not a simple “add streamer + webhook” experience. |
| **Siddhart/Twitch-Discord-Bot** | Self-hosted OSS bot | Twitch | Official Twitch API, unlimited streamers, role tagging, message updates every 10 minutes, cron-configurable. | Requires Discord bot token plus Twitch client ID/secret. Config-file setup, one Discord channel per deployment. |
| **alepouna/twitch-discord-alerts** | Self-hosted OSS helper | Anything Streamer.bot supports | Small Node helper that triggers from Streamer.bot. | Manual Streamer.bot/Node setup, one streamer, not a standalone product. |
| **Twitch Live to Discord** | Local webhook app | Twitch | No bot/OAuth/API registration, multi-streamer + multi-destination routing, per-destination templates and live preview, optional streamer ping, persisted dedupe, cross-platform dashboard, MIT. | Polling-based, local app must stay running, no role pings/live roles/schedule sync, no official Twitch API mode yet, no macOS/Linux tray. |

## Feature comparison

Legend: **Yes** = first-class; **Partial** = possible with extra setup or a
generic bot; **No** = not a core feature.

| Feature | Streamcord | YAGPDB | Zapier/IFTTT | Streamer.bot | Our app |
| --- | --- | --- | --- | --- | --- |
| Twitch live alerts | Yes | Yes | Yes | Yes | Yes |
| YouTube/Kick alerts | Yes / Yes | YouTube | Limited | Yes / Yes | No |
| No bot or OAuth required | No | No | No | Yes (local app) | **Yes** |
| Multiple streamers | Yes | Yes | Per-task | Yes | **Yes** |
| Multiple labelled webhooks/destinations | Partial | Partial | Partial | Yes with actions | **Yes, core** |
| Per-destination message templates | Partial | Yes | Limited | Yes via actions | **Yes** |
| Live Discord preview | Dashboard | No | No | No | **Yes** |
| Optional streamer ping | Role/user mentions | Role/user mentions | Partial | Yes | **Yes** |
| Role pings | Yes | Yes | Partial | Yes | No |
| Game/title filters | Yes | Templates | Zapier filters | Yes via actions | No |
| Quiet hours/cooldowns | Cooldowns | Templates | Limited | Yes via actions | No |
| Stream-end cleanup | Yes | No | No | Yes via actions | No |
| Message updates while live | Partial | No | No | Yes via actions | No |
| Live roles | Yes | Partial | No | Yes via actions | No |
| Schedule sync | Yes (Pro) | No | No | Yes via actions | No |
| Twitch Teams | Yes | No | No | No | No |
| Self-hosted/local-first | No | Yes | No | Yes | **Yes** |
| Cloud 24/7 uptime | Yes | Yes | Yes | No | No |
| Persisted per-stream dedupe | Assumed | Assumed | Not guaranteed | Manual/action-dependent | **Yes, persisted** |
| Free without streamer limits | No (5 free) | Yes-ish | Task limits | Yes | **Yes** |
| Installer/polish for non-technical users | Yes | Dashboard | Yes | Desktop app | Partial |

## Where our app already wins

- **No bot, OAuth flow, Twitch API registration, or cloud account.** This is a
  meaningfully different onboarding story.
- **Multi-destination routing.** Users can send one streamer to several labelled
  Discord channels/servers, and one destination can receive many streamers.
- **Message control with live preview.** Per-destination content, embed styling,
  field toggles, placeholders, and a Discord-like preview are stronger than most
  simple webhook tools.
- **Local-first privacy.** Settings and webhook URLs stay on the user’s machine;
  no third-party alert service sees them.
- **Cross-platform dashboard.** Windows, Linux, and macOS use the same UI.
- **Dedupe that survives restarts.** The last stream ID is stored per
  destination, so manual checks and restarts do not repost the same stream.
- **Optional streamer ping with narrow allowed mentions.** `{discord_mention}`
  allows only the configured user; `@everyone` and roles remain blocked.
- **MIT and free.** No streamer-count paywall.

## Biggest gaps and opportunities

### 1. Availability and 24/7 uptime

Cloud bots are always online. A local desktop app is not when the machine sleeps,
reboots, or is closed. This is the largest product gap.

Opportunity:

- Docker image + compose for VPS/NAS/Raspberry Pi.
- Healthcheck endpoint and optional external heartbeat.
- Service installers for Windows/Linux/macOS.
- Clear dashboard health status: last check, next check, last error.
- Cross-platform single-instance lock to prevent duplicate processes.

### 2. Real-time or more reliable Twitch detection

Our current default is a 2-minute poll. That can miss very short streams.

Twitch EventSub is event-driven, but the local implications are significant:
EventSub WebSocket requires a user access token, has connection/subscription
limits, and can deliver duplicate events that must be deduped by `message_id`.
Webhook transport requires a public HTTPS callback. For an app promising
unlimited streamers without OAuth, official Helix polling with a user-supplied
app credential may be a more practical “reliable mode” than EventSub.

Opportunities:

- Lower minimum poll interval to 30 seconds and add adaptive polling.
- Optional official Twitch API/Helix mode.
- Optional EventSub WebSocket mode for users who connect Twitch.
- Deduplicate EventSub events by `message_id` if adopted.

### 3. Role pings and mention controls

Our streamer ping is a start. Communities also need a Live role or an
`@everyone`/`@here` option. Webhooks can mention roles with a scoped
`allowed_mentions.roles` list.

Opportunities:

- Per-destination role ID pings.
- Separate streamer ping, role ping, and optional `@everyone`/`@here`.
- Explicit confirmation for broad mentions.
- Tests that only intended IDs are allowed.

### 4. Notification rules and filters

Competitors offer game/title filtering and cooldowns. These are cheap to add and
high value for busy communities.

Opportunities:

- Include/exclude games/categories.
- Title keyword allow/deny rules.
- Minimum viewers.
- Quiet hours, allowed days, per-streamer cooldown.
- Alert delay (e.g. only after 5 minutes live).
- “Only first stream of the day.”

### 5. Stream lifecycle actions

Discord webhooks can edit and delete their own messages. That makes several
competitor features possible without a bot:

- Delete or edit the live message when the stream ends.
- Add a VOD/last-stream link on stream end.
- Edit the live message every N minutes to update viewers/title.
- Thread/forum target support.
- Custom webhook username/avatar per destination.

### 6. Packaging and trust

A public utility needs lower-friction installation and updates:

- Windows installer, code signing, winget.
- macOS `.app`/menu-bar wrapper and notarization.
- Linux AppImage/deb and a native tray.
- Auto-update from GitHub releases.
- Encrypt webhook URLs with DPAPI/Keychain/libsecret.
- Optional local dashboard PIN/token for local-network access.

### 7. Platform reach

- YouTube Live via YouTube Data API.
- Kick live alerts.
- Twitch Teams.
- Schedule sync to Discord events.
- Live role automation.

The last two require a Discord bot/OAuth. They should be optional advanced
features, not requirements for the core webhook app.

## Platform constraints worth remembering

### Twitch

- EventSub is event-driven, but WebSocket uses user access tokens and has
  subscription/connection limits.
- EventSub notifications are at-least-once; duplicate `message_id` values are
  possible.
- If a WebSocket disconnects, subscriptions must be recreated; lost events are
  not replayed.
- Our current public GQL endpoint can change without notice. An official API mode
  is the long-term reliability answer.

### Discord

- Respect `429` `Retry-After` / `retry_after`; do not hardcode rate limits.
- Do not retry `401`, `403`, or `404` indefinitely. Repeated invalid requests
  can lead to temporary IP restrictions.
- Webhooks can edit/delete their own messages and override username/avatar.
- Webhooks cannot assign roles, create scheduled events, or act as a general
  bot. Live roles and schedule sync require bot mode.

## Strategic read

Our best position is:

> **The private, local-first, webhook-native alert router for communities that
> follow many streamers and post to several Discord destinations.**

Do not try to become a general-purpose moderation/music/economy bot. Competing
with Streamcord/MEE6/Carl on bot breadth is a losing game. Compete on setup
simplicity, message control, multi-destination routing, privacy, and reliability.

The most valuable next steps are:

1. Docker/self-host and health checks.
2. Retry/backoff and hard reliability guarantees.
3. Role pings.
4. Filters/quiet hours.
5. Stream-end cleanup and message updates.
6. Installer/signing/auto-update.

See [roadmap.md](roadmap.md) for the proposed releases and acceptance criteria.

## Sources

- Streamcord: <https://streamcord.io/> · <https://docs.streamcord.io/notifications/add-twitch.html> · <https://streamcord.io/pro>
- IFTTT Twitch: <https://ifttt.com/twitch>
- Zapier Twitch + Discord: <https://zapier.com/apps/twitch/integrations/discord>
- YAGPDB: <https://yagpdb.xyz/>
- Streamer.bot: <https://streamer.bot/>
- Siddhart/Twitch-Discord-Bot: <https://github.com/Siddhart/Twitch-Discord-Bot>
- alepouna/twitch-discord-alerts: <https://github.com/alepouna/twitch-discord-alerts>
- Twitch EventSub: <https://dev.twitch.tv/docs/eventsub/> · <https://dev.twitch.tv/docs/eventsub/handling-websocket-events/>
- Discord rate limits: <https://discord.com/developers/docs/topics/rate-limits>
