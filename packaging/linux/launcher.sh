#!/bin/sh
# Desktop launcher: the app's no-argument mode starts the dashboard and opens
# the browser. Settings and secrets stay in the normal per-user folder.
set -eu

bin="${TWITCH_LIVE_BIN:-$(dirname "$0")/twitch-live-notifier}"
exec "$bin" "$@"
