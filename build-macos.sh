#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "$0")"

version="${VERSION:-$(sed -n 's/.*Version = "\(.*\)"/\1/p' version.go)}"
mkdir -p dist

for arch in amd64 arm64; do
	CGO_ENABLED=0 GOOS=darwin GOARCH="${arch}" go build \
		-trimpath \
		-buildvcs=false \
		-ldflags="-s -w" \
		-o "dist/TwitchLiveNotifier-macos-${arch}" .
	./packaging/macos/build-app.sh "$version" "$arch" "dist/TwitchLiveNotifier-macos-${arch}" >/dev/null
done

ls -lh dist/TwitchLiveNotifier-macos-* dist/macos-*/*.app/Contents/MacOS/twitch-live-notifier
