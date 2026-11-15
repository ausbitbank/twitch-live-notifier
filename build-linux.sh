#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "$0")"

arch="${GOARCH:-amd64}"
mkdir -p dist

CGO_ENABLED=0 GOOS=linux GOARCH="${arch}" go build \
	-trimpath \
	-buildvcs=false \
	-ldflags="-s -w" \
	-o "dist/TwitchLiveNotifier-linux-${arch}" .

ls -lh "dist/TwitchLiveNotifier-linux-${arch}"
