#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "$0")"

mkdir -p dist

for arch in amd64 arm64; do
	CGO_ENABLED=0 GOOS=darwin GOARCH="${arch}" go build \
		-trimpath \
		-buildvcs=false \
		-ldflags="-s -w" \
		-o "dist/TwitchLiveNotifier-macos-${arch}" .
done

ls -lh dist/TwitchLiveNotifier-macos-*
