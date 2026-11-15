#!/usr/bin/env bash
# Build a Linux AppImage. Requires appimagetool (or set APPIMAGETOOL).
# Usage: packaging/linux/build-appimage.sh <version> [binary] [arch]
set -euo pipefail

version="${1:?usage: build-appimage.sh <version> [binary] [arch]}"
binary="${2:-dist/TwitchLiveNotifier-linux-amd64}"
arch="${3:-x86_64}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
stage="$(mktemp -d)"
trap 'rm -rf "$stage"' EXIT

binary_path="$binary"
if [[ -f "$root/$binary" ]]; then
  binary_path="$root/$binary"
fi
if [[ ! -f "$binary_path" ]]; then
  echo "binary not found: $binary" >&2
  exit 1
fi

appdir="$stage/TwitchLiveNotifier.AppDir"
mkdir -p "$appdir/usr/bin" "$appdir/usr/share/applications" \
  "$appdir/usr/share/icons/hicolor/256x256/apps" "$appdir/usr/share/metainfo"
install -m 0755 "$binary_path" "$appdir/usr/bin/twitch-live-notifier"
install -m 0755 "$root/packaging/linux/launcher.sh" "$appdir/usr/bin/twitch-live-notifier-launch"
install -m 0644 "$root/packaging/linux/twitch-live-notifier.desktop" "$appdir/twitch-live-notifier.desktop"
install -m 0644 "$root/packaging/linux/twitch-live-notifier.desktop" "$appdir/usr/share/applications/twitch-live-notifier.desktop"
install -m 0644 "$root/winres/icon256.png" "$appdir/twitch-live-notifier.png"
install -m 0644 "$root/winres/icon256.png" "$appdir/usr/share/icons/hicolor/256x256/apps/twitch-live-notifier.png"
sed "s/__VERSION__/${version#v}/g" "$root/packaging/linux/twitch-live-notifier.appdata.xml" > "$appdir/usr/share/metainfo/twitch-live-notifier.appdata.xml"

cat > "$appdir/AppRun" <<'EOF'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
export TWITCH_LIVE_BIN="$HERE/usr/bin/twitch-live-notifier"
exec "$HERE/usr/bin/twitch-live-notifier-launch" "$@"
EOF
chmod 0755 "$appdir/AppRun"

tool="${APPIMAGETOOL:-}"
if [[ -z "$tool" ]]; then
  tool="$(command -v appimagetool || true)"
fi
if [[ -z "$tool" && -x "$root/dist/appimagetool-x86_64.AppImage" ]]; then
  tool="$root/dist/appimagetool-x86_64.AppImage"
fi
if [[ -z "$tool" ]]; then
  echo "appimagetool not found. Set APPIMAGETOOL or put appimagetool-x86_64.AppImage in dist/." >&2
  exit 1
fi

mkdir -p "$root/dist"
output="$root/dist/TwitchLiveNotifier-v${version#v}-Linux-x64.AppImage"
args=(--no-appstream "$appdir" "$output")
if [[ "$tool" == *.AppImage ]]; then
  args=(--appimage-extract-and-run "${args[@]}")
fi
ARCH="$arch" "$tool" "${args[@]}" >/dev/null
chmod 0755 "$output"
echo "$output"
