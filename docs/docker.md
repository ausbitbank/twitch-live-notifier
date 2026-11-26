# Running with Docker

The Docker image runs the same dashboard and monitor as the desktop app. It is
the recommended way to keep alerts running 24/7 on a VPS, NAS, or Raspberry Pi.

The image supports `linux/amd64` and `linux/arm64`. Use a 64-bit host OS; 32-bit
Raspberry Pi OS is not supported.

## Quick start

```bash
git clone https://github.com/ausbitbank/twitch-live-notifier.git
cd twitch-live-notifier
docker compose up -d
```

Open <http://127.0.0.1:8787/>. Add your Discord destinations and Twitch
streamers exactly as you would in the desktop app.

Configuration lives in the `twitch-live-data` volume at `/data/config.json`.
Encrypted webhook URLs live in `/data/secrets.json`, with `/data/secret.key` used
when the container has no OS keyring. Delivery history is stored in
`/data/delivery-history.json` and never contains webhook URLs or tokens. The
volume survives `docker compose down`, image upgrades, and container recreation.

Check status and logs:

```bash
docker compose ps
docker compose logs -f app
curl -fsS http://127.0.0.1:8787/healthz
```

The container has a built-in healthcheck that probes `/healthz` every 30
seconds. `healthy` means the monitor is running and its last check is recent;
`unhealthy` includes a reason such as `no check has completed yet`.

## Accessing the dashboard safely

**The dashboard has no login by default.** It can add streamers, inspect the
encrypted secret store while editing, and change routing. Set an optional
**Dashboard token** under Settings to require a token for API access, but still
keep the port on loopback or put it behind an authenticating reverse proxy.

The default compose mapping is:

```yaml
ports:
  - "127.0.0.1:8787:8787"
```

- **On the same machine:** open <http://127.0.0.1:8787/>.
- **On a VPS/headless host:** use an SSH tunnel:

  ```bash
  ssh -L 8787:127.0.0.1:8787 user@your-server
  ```

  Then open <http://127.0.0.1:8787/> locally.

- **On a LAN:** set `TWITCH_LIVE_BIND=0.0.0.0` in `.env`, or change the port
  mapping to `"8787:8787"`. Only do this on a trusted network.

- **Public access:** put Caddy, nginx, or Traefik in front with HTTPS and
  authentication. Do not publish port 8787 directly to the internet.

The app only makes outbound requests to Twitch and Discord. It has no
analytics, telemetry, or cloud relay.

## Optional Twitch follows import

The **no-sign-in** follows import works in Docker without any extra
configuration: use the browser helper or paste a list into the dashboard. The
helper runs in your browser, not in the container, and the app only receives the
pasted list.

To also show the optional **Sign in to Twitch** device-code fallback, set a
public Twitch application client ID:

```yaml
environment:
  TWITCH_LIVE_CLIENT_ID: "your-public-client-id"
```

No client secret is used. The import token stays in memory for one request and
is never written to `/data`. See
[Importing your Twitch follows](twitch-import.md) for the app registration
steps.

## Configuration volume

The named volume is created automatically. You can also use **Download backup**
in Settings for a portable, redacted-by-default export, or **With secrets** when
you need to move the configuration to another machine. For a full filesystem
backup, archive the named volume:

```bash
docker run --rm \
  -v twitch-live-notifier_twitch-live-data:/data:ro \
  -v "$PWD":/backup \
  alpine tar -czf /backup/twitch-live-backup.tar.gz -C /data .
```

Restore into a fresh volume:

```bash
docker volume create twitch-live-notifier_twitch-live-data
docker run --rm \
  -v twitch-live-notifier_twitch-live-data:/data \
  -v "$PWD":/backup \
  alpine sh -c "tar -xzf /backup/twitch-live-backup.tar.gz -C /data && chown -R 100:101 /data"
```

The volume name is `<compose-project>_twitch-live-data`. With a custom compose
project name, use `docker volume ls` to find the exact name.

You can also use a bind mount:

```yaml
volumes:
  - ./data:/data
```

Ensure the host directory is writable by the container user. On Linux, run
with `user: "${UID}:${GID}"` or `chown` the directory to the container user.

## Updating

```bash
docker compose pull
docker compose up -d
```

or, when building from a checkout:

```bash
git pull
docker compose build --pull
docker compose up -d
```

Settings are not touched by an update. The previous image remains available for
rollback by pinning a tag:

```yaml
image: ghcr.io/ausbitbank/twitch-live-notifier:v2.5.0
```

The in-app updater is disabled inside Docker; the dashboard shows the
`docker compose pull` / `up -d` guidance instead. That keeps the container
immutable and avoids replacing a running image.

## Timezone and quiet hours

Quiet hours, allowed days, cooldowns, and "first stream of the day" use the
container's local timezone. Set it in `.env` or compose:

```yaml
environment:
  TZ: Europe/London
```

The default is `UTC`. The image includes `tzdata`.

## VPS, NAS, and Raspberry Pi notes

- **VPS:** keep the port on loopback and use an SSH tunnel, or terminate HTTPS
  in a reverse proxy with authentication.
- **Synology/TrueNAS/QNAP:** create the project from `docker-compose.yml` in the
  NAS UI, or use `docker compose` over SSH. Keep the config on a persistent
  volume, not inside the container.
- **Raspberry Pi 4/5:** use a 64-bit OS and the `linux/arm64` image. The compose
  file pulls the correct architecture automatically. A Pi can comfortably run
  dozens of streamers at the default two-minute interval.
- **Sleep/hibernation:** containers do not run while the host sleeps. Use a
  always-on host for reliable 24/7 alerts.

## Health endpoint

`GET /healthz` returns JSON and either HTTP 200 or 503:

```json
{
  "status": "healthy",
  "version": "2.5.0",
  "uptime_seconds": 3600,
  "monitoring": true,
  "paused": false,
  "checking": false,
  "last_check": "2026-11-15T12:00:00Z",
  "last_check_age_seconds": 42,
  "streamers": 4,
  "live": 1,
  "destinations": 2
}
```

- `healthy` — the monitor is running and the last check is recent.
- `unhealthy` — no check has completed after startup, or the last check is
  stale (more than three poll intervals, minimum six minutes). HTTP 503.
- Pausing monitoring intentionally keeps the app `healthy`.

The image healthcheck uses the binary itself, so no `curl` or shell is needed:

```bash
docker inspect --format '{{.State.Health.Status}}' twitch-live-notifier
docker inspect --format '{{json .State.Health}}' twitch-live-notifier
```

You can run the same probe manually:

```bash
docker compose exec app /app/TwitchLiveNotifier --healthcheck
```

## Privacy and logs

- Webhook URLs are encrypted in `/data/secrets.json` inside the volume. The
  fallback key is `/data/secret.key` when no keyring is available.
- The app does not print webhook URLs to logs. Startup logs include the version,
  dashboard URL, and config path only.
- Container logs contain check results and destination labels, not webhook
  tokens.
- Do not share `docker compose config`, unmasked dashboard screenshots, backups
  with secrets, or the config volume in bug reports.

## Troubleshooting

**Port 8787 is already in use.** Stop the other process, or change the host
side of the mapping (for example `127.0.0.1:8788:8787`).

**The dashboard is unreachable from another machine.** The default mapping is
loopback-only by design. Use an SSH tunnel, set `TWITCH_LIVE_BIND=0.0.0.0`, or
configure a reverse proxy.

**Health is unhealthy after startup.** Check `docker compose logs app`. The
first check runs about 1.5 seconds after start. If Twitch is unreachable, the
status still records the failed check, so `healthz` normally recovers quickly.
A permanently unhealthy state usually means the monitor loop is stuck.

**Quiet hours fire at the wrong time.** Set `TZ` to your IANA timezone and
recreate the container.

**A webhook returns 401/403/404.** The destination is marked invalid in the
dashboard and is skipped to avoid retry storms. Recreate the webhook in Discord,
update the URL, and use **Test webhook** to clear the invalid state.
