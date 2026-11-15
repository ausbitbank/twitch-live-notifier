#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "$0")"

version="${VERSION:-2.0.0}"
rm -rf dist/docs
mkdir -p dist/docs

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 GOAMD64=v1 go build \
	-trimpath \
	-buildvcs=false \
	-ldflags "-s -w -H=windowsgui -buildid= -X main.Version=${version}" \
	-o dist/TwitchLiveNotifier.exe .

cp README.md LICENSE dist/
cp docs/*.png dist/docs/

if command -v sha256sum >/dev/null 2>&1; then
	sha256sum dist/TwitchLiveNotifier.exe > dist/SHA256SUMS.txt
else
	shasum -a 256 dist/TwitchLiveNotifier.exe > dist/SHA256SUMS.txt
fi

ls -lh dist/TwitchLiveNotifier.exe
cat dist/SHA256SUMS.txt
