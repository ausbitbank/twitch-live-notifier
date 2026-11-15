#!/usr/bin/env bash
# Build a Debian package for the Linux dashboard binary.
# Usage: packaging/linux/build-deb.sh <version> [binary] [arch]
set -euo pipefail

version="${1:?usage: build-deb.sh <version> [binary] [arch]}"
binary="${2:-dist/TwitchLiveNotifier-linux-amd64}"
arch="${3:-amd64}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
stage="$(mktemp -d)"
trap 'rm -rf "$stage"' EXIT

if [[ ! -f "$root/$binary" && ! -f "$binary" ]]; then
  echo "binary not found: $binary" >&2
  exit 1
fi
binary_path="$binary"
if [[ -f "$root/$binary" ]]; then
  binary_path="$root/$binary"
fi

mkdir -p "$stage/DEBIAN" \
  "$stage/usr/bin" \
  "$stage/usr/share/applications" \
  "$stage/usr/share/icons/hicolor/256x256/apps" \
  "$stage/usr/share/metainfo" \
  "$stage/usr/lib/systemd/user" \
  "$stage/usr/share/doc/twitch-live-notifier"

install -m 0755 "$binary_path" "$stage/usr/bin/twitch-live-notifier"
install -m 0755 "$root/packaging/linux/launcher.sh" "$stage/usr/bin/twitch-live-notifier-launch"
install -m 0644 "$root/packaging/linux/twitch-live-notifier.desktop" "$stage/usr/share/applications/twitch-live-notifier.desktop"
install -m 0644 "$root/winres/icon256.png" "$stage/usr/share/icons/hicolor/256x256/apps/twitch-live-notifier.png"
sed "s/__VERSION__/${version#v}/g" "$root/packaging/linux/twitch-live-notifier.appdata.xml" > "$stage/usr/share/metainfo/twitch-live-notifier.appdata.xml"
install -m 0644 "$root/packaging/linux/twitch-live-notifier.service" "$stage/usr/lib/systemd/user/twitch-live-notifier.service"
install -m 0644 "$root/README.md" "$stage/usr/share/doc/twitch-live-notifier/README.md"
install -m 0644 "$root/LICENSE" "$stage/usr/share/doc/twitch-live-notifier/LICENSE"

installed_size="$(du -sk "$stage/usr" | awk '{print $1}')"
cat > "$stage/DEBIAN/control" <<EOF
Package: twitch-live-notifier
Version: ${version#v}
Section: net
Priority: optional
Architecture: $arch
Installed-Size: $installed_size
Maintainer: ausbitbank <ausbitbank@users.noreply.github.com>
Homepage: https://github.com/ausbitbank/twitch-live-notifier
Description: Twitch live alerts for labelled Discord webhooks
 A local dashboard and monitor that posts Twitch live alerts to one or more
 Discord webhooks. No bot, OAuth, or Twitch API registration is required.
EOF

mkdir -p "$root/dist"
output="$root/dist/TwitchLiveNotifier-v${version#v}-Linux-x64.deb"
if command -v dpkg-deb >/dev/null 2>&1; then
  dpkg-deb --build --root-owner-group "$stage" "$output" >/dev/null
else
  # Portable fallback for environments without dpkg-deb.
  pkgdir="$(mktemp -d)"
  trap 'rm -rf "$stage" "$pkgdir"' EXIT
  printf '2.0\n' > "$pkgdir/debian-binary"
  tar --owner=0 --group=0 -czf "$pkgdir/control.tar.gz" -C "$stage/DEBIAN" .
  tar --owner=0 --group=0 -czf "$pkgdir/data.tar.gz" -C "$stage" --exclude=DEBIAN .
  rm -f "$output"
  (cd "$pkgdir" && ar r "$output" debian-binary control.tar.gz data.tar.gz >/dev/null)
fi
echo "$output"
