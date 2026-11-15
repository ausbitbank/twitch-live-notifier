#!/usr/bin/env bash
# Build a .app bundle around a macOS TwitchLiveNotifier binary.
# Usage: packaging/macos/build-app.sh <version> <amd64|arm64> [binary]
set -euo pipefail

version="${1:?usage: build-app.sh <version> <amd64|arm64> [binary]}"
arch="${2:?usage: build-app.sh <version> <amd64|arm64> [binary]}"
binary="${3:-dist/TwitchLiveNotifier-macos-$arch}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
if [[ -f "$root/$binary" ]]; then
  binary="$root/$binary"
fi
if [[ ! -f "$binary" ]]; then
  echo "binary not found: $binary" >&2
  exit 1
fi

outdir="$root/dist/macos-$arch"
app="$outdir/Twitch Live to Discord.app"
rm -rf "$outdir"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources"

install -m 0755 "$binary" "$app/Contents/MacOS/TwitchLiveNotifier"
sed "s/__VERSION__/$version/g" "$root/packaging/macos/Info.plist" > "$app/Contents/Info.plist"

if command -v iconutil >/dev/null 2>&1; then
  iconset="$outdir/AppIcon.iconset"
  mkdir -p "$iconset"
  cp "$root/winres/icon16.png" "$iconset/icon_16x16.png"
  cp "$root/winres/icon32.png" "$iconset/icon_16x16@2x.png"
  cp "$root/winres/icon32.png" "$iconset/icon_32x32.png"
  cp "$root/winres/icon48.png" "$iconset/icon_32x32@2x.png"
  cp "$root/winres/icon128.png" "$iconset/icon_128x128.png"
  cp "$root/winres/icon256.png" "$iconset/icon_128x128@2x.png"
  cp "$root/winres/icon256.png" "$iconset/icon_256x256.png"
  if iconutil -c icns "$iconset" -o "$app/Contents/Resources/AppIcon.icns"; then
    rm -rf "$iconset"
  else
    echo "warning: could not create AppIcon.icns" >&2
  fi
fi

echo "$app"
