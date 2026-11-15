# Roadmap

**Current baseline:** 2.1.0  
**Research input:** [competitive-research.md](competitive-research.md)

## Product principles

1. **No bot required for the core experience.** A Twitch username and a Discord
   webhook should still be enough.
2. **Local-first and private.** Core alerts and settings run on the user’s
   machine; no cloud relay is required.
3. **Webhook-native.** Use Discord webhooks for delivery and message management
   where possible.
4. **Multi-streamer, multi-destination.** One streamer can post to several
   labelled destinations; one destination can receive many streamers.
5. **Safe mentions.** Never allow accidental `@everyone`/role pings. Only
   explicitly configured IDs are mentionable.
6. **Once per stream, per destination.** Dedupe state must survive restarts and
   manual checks.
7. **Simple enough for a community manager.** Clear labels, live previews,
   sample posts, and no config-file editing.
8. **Cross-platform.** Windows, Linux, and macOS are first-class dashboard
   targets.

## Non-goals

- A general moderation, music, economy, or leveling bot.
- A full Streamcord/MEE6/Carl feature clone.
- Requiring a Discord bot or Twitch OAuth for basic alerts.
- A hosted SaaS in the near term.
- Live-role automation or schedule sync until an optional bot mode exists.

## Priority definitions

- **P0:** Blocks reliability, safety, or meaningful adoption.
- **P1:** High-value parity or differentiation after P0.
- **P2:** Valuable expansion with larger scope/risk.
- **P3:** Nice-to-have or speculative.

## v2.2 — Reliability and community essentials

Goal: make the app dependable enough to recommend to a real community.

### P0: Docker / self-host package

Why: a desktop app cannot match cloud uptime when the machine sleeps. A small
Docker image lets a streamer run it on a VPS, NAS, or Raspberry Pi.

Deliverables:

- Multi-stage `Dockerfile`.
- `docker-compose.yml` with a persistent config volume and port `8787`.
- Healthcheck endpoint.
- `docs/docker.md` with VPS/NAS/Raspberry Pi instructions.
- Publish multi-arch images on release tags (at least `linux/amd64`,
  `linux/arm64`).

Acceptance criteria:

- `docker compose up -d` starts the dashboard and monitor with an empty config.
- Config survives container recreation.
- `GET /healthz` returns healthy/unhealthy with last check information.
- No webhook URLs appear in container logs.
- Documentation covers updating and backing up the volume.

### P0: Retry and rate-limit handling

Why: a transient Twitch or Discord failure should not lose an alert or hammer
the APIs.

Deliverables:

- Bounded retries with exponential backoff and jitter for Twitch 5xx/timeouts.
- Discord `429` handling using `Retry-After` / `retry_after`.
- Do not retry `401`, `403`, or `404` indefinitely; mark the destination as
  invalid and surface it in the UI.
- Per-destination status/error history.
- Tests using local mock servers for 429, 500, timeout, and 404.

Acceptance criteria:

- A `429` retries after the server-provided delay once, not in a tight loop.
- A failed post does not record `last_stream_id`, so the next check retries.
- A repeatedly failing destination does not block other destinations.
- Failures are visible in the activity feed and destination card.

### P0: Cross-platform single instance

Why: two instances monitoring the same config can race and double-post.

Deliverables:

- File-based single-instance lock for Linux/macOS `--serve`.
- Existing Windows mutex retained.
- If a second instance starts, either open the existing dashboard or exit with a
  clear message.
- Tests for lock acquisition/release and stale-lock recovery.

Acceptance criteria:

- Starting two `--serve` processes against one config results in one server.
- Killing the process releases the lock.
- A stale lock after a crash does not block a restart forever.

### P0: Role pings and mention controls

Why: communities often want to ping an `@Live` role, not only the streamer.

Deliverables:

- Optional role IDs per destination.
- Optional streamer ping and role ping toggles.
- `allowed_mentions.roles` scoped to configured IDs.
- Optional `@everyone`/`@here` behind a confirmation and an explicit setting.
- Template placeholders for role mentions if useful (`{role_mentions}` or
  destination-configured roles).
- UI warnings and preview support.

Acceptance criteria:

- Only configured role IDs can be mentioned.
- `@everyone` remains impossible unless explicitly enabled.
- Tests assert the exact `allowed_mentions` payload.
- Preview clearly shows what will ping.

### P0/P1: Notification filters and quiet hours

Why: competitors offer game/title filters and communities need control over
when alerts fire.

Deliverables:

- Include/exclude game or category rules.
- Title keyword allow/deny rules.
- Minimum viewers.
- Quiet hours and allowed days.
- Per-streamer cooldown and optional “only first stream of the day.”
- Optional alert delay (e.g. only after 5 minutes live).

Acceptance criteria:

- A filtered stream does not post but still updates live status.
- Quiet hours are evaluated in the user’s local timezone.
- Filter settings are visible and editable per streamer.
- Unit tests cover include/exclude precedence.

### P1: Installer, signing, and auto-update

Deliverables:

- Windows installer (Inno Setup, MSIX, or similar) and portable ZIP.
- Code-signing certificate for the Windows executable.
- winget manifest.
- macOS `.app`/menu-bar wrapper and notarization.
- Linux AppImage/deb package.
- GitHub-release update check and optional auto-update.

Acceptance criteria:

- A non-technical Windows user can install, launch, and find the tray icon.
- SmartScreen warnings disappear when signed.
- An update can be installed without losing settings.
- The app reports its version and update status.

## v2.3 — Stream lifecycle and message power

Goal: match the most visible notification features competitors offer.

### P1: Stream-end cleanup

Deliverables:

- Store Discord message ID per channel/destination.
- On stream end, optionally edit or delete the live message.
- Optional “stream ended” message with last title/game and VOD link where
  available.
- Per-destination cleanup policy.

Acceptance criteria:

- Deleting/editing uses the stored webhook message ID and token.
- A cleanup failure does not affect the next stream.
- Dedupe state still prevents duplicate live posts.
- UI previews/toggles the cleanup behavior.

### P1: Live message updates

Deliverables:

- Optional periodic edit of the live message while the stream is live.
- Update viewer count/title at a configurable interval.
- Avoid rate-limit abuse with a minimum interval.
- Stop updates when offline.

Acceptance criteria:

- The same Discord message is edited, not reposted.
- Failed edits are retried according to the retry policy.
- Turning updates off leaves the original message untouched.

### P1: Custom identity and thread targets

- Per-destination webhook username and avatar override.
- Post to a Discord thread/forum via `thread_id`.
- Optional link buttons/components where Discord supports them.

### P1: Per-route templates and template power

- Template override per streamer→destination route.
- Conditional/default syntax, e.g. `{game|Just Chatting}`.
- Custom fields and reorderable fields.
- Template import/export and “duplicate destination.”

### P1: Secret storage and config UX

- Encrypt webhook URLs using DPAPI (Windows), Keychain (macOS), and libsecret
  (Linux) where available.
- Export/import config with the webhook secret redacted by default.
- Optional local dashboard PIN/token.
- Backup/restore settings from the UI.

## v2.4 / v3.0 — Reach and advanced modes

### P2: Optional official Twitch API mode

Why: public GQL is convenient but fragile. Official Helix polling is more stable
and does not require a bot.

Deliverables:

- User-supplied Twitch Client ID/secret/app token.
- Helix `Get Streams` polling mode.
- Clear setup instructions and diagnostics.
- Fallback to public GQL if official mode is not configured.

Open question: official Helix polling vs EventSub. EventSub WebSocket requires
user OAuth, has connection/subscription limits, and can duplicate events; Helix
polling may be the better fit for unlimited streamers.

### P2: EventSub mode (optional)

- Twitch OAuth connect flow with loopback redirect.
- `stream.online` / `stream.offline` subscriptions.
- Dedupe by Twitch `message_id`.
- Reconnect/keepalive handling.
- Clear UI explanation of OAuth and subscription limits.

### P2: YouTube and Kick support

- YouTube Live via YouTube Data API.
- Kick live alerts via its public API where available.
- Platform-specific channel/destination fields.
- Shared template placeholders across platforms where possible.

### P2: Twitch Teams

- Subscribe to a Twitch Team and alert when any member goes live.
- Team-level destination routing.
- Dedupe by member + stream ID.

### P2: Optional Discord bot mode

Only if there is demand. A bot enables features webhooks cannot provide:

- Live role assignment/removal.
- Closing/creating Discord scheduled events from Twitch schedules.
- Discovering channel/role names.
- Editing messages without retaining webhook tokens.
- Approvals/moderation workflows.

This must be a separate opt-in mode with OAuth and clear permission scopes.

### P2: Team/manager configuration

- Shared config with multiple managers.
- Role-based permissions and audit log.
- Approval before posting for sensitive communities.

### P3: Analytics and hosted monitoring

- Delivery history persisted with message links.
- Failure-rate and latency dashboards.
- Optional hosted relay for users who cannot run a service.
- Privacy-preserving aggregate stats only if explicitly opted in.

## Cross-cutting engineering work

- Add health endpoint and structured local logs with webhook redaction.
- Add `context.Context` cancellation to network calls and shutdown.
- Add Discord webhook health state and “disable after repeated failure” UI.
- Add integration tests for multi-destination, retries, filters, cleanup, and
  role mentions.
- Expand Playwright UI tests for onboarding, settings, and failure states.
- Keep `go vet` for Windows/Linux/macOS and race tests in CI.
- Evaluate code signing and reproducible builds before a wider launch.

## Success metrics

- Setup completion: percentage of users who add a destination, a streamer, and
  route at least one alert.
- Delivery: successful alerts per check; failed alerts retried successfully.
- Duplicate rate: zero duplicate posts for the same stream/destination under
  normal single-instance operation.
- Uptime: percentage of expected checks completed; Docker/service health.
- Support load: most common errors and setup drop-off points.
- Optional: GitHub stars/downloads are vanity metrics; prioritize successful
  community deployments.

## Risks and open decisions

1. **Twitch API stability.** Public GQL can change. Decide whether official API
   mode is v2.3 or v3.0.
2. **EventSub complexity.** User OAuth, subscription limits, reconnect logic, and
   duplicate events may outweigh the benefits for unlimited channels.
3. **Bot-mode scope creep.** Live roles and schedules require a bot. Keep them
   optional and separate from the webhook core.
4. **Webhook secret storage.** Encrypt at rest before encouraging shared/team
   use.
5. **Installer and signing costs.** Budget for a Windows code-signing
   certificate and macOS notarization.
6. **Hosted service temptation.** A cloud option would improve uptime but
   changes the privacy/ops model. Prefer Docker/self-host first.

## Definition of done for a release

- Unit tests and UI smoke tests pass.
- Windows, Linux, and macOS builds pass in CI.
- Release artifacts and checksums are uploaded.
- README and changelog are updated.
- Upgrade path from the previous version is tested.
- No real webhook URLs or secrets are committed.
- Known limitations are documented.
