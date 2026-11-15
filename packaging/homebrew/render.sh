#!/usr/bin/env bash
# Render a Homebrew formula for a release source tarball.
# Usage: packaging/homebrew/render.sh <version> <source.tar.gz> [output]
set -euo pipefail

version="${1:?usage: render.sh <version> <source.tar.gz> [output]}"
source_tarball="${2:?usage: render.sh <version> <source.tar.gz> [output]}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
output="${3:-$root/dist/homebrew/twitch-live-notifier.rb}"

if [[ ! -f "$source_tarball" ]]; then
  echo "source tarball not found: $source_tarball" >&2
  exit 1
fi

sha="$(sha256sum "$source_tarball" | awk '{print $1}')"
mkdir -p "$(dirname "$output")"
sed \
  -e "s/__VERSION__/$version/g" \
  -e "s/__SHA256__/$sha/g" \
  "$root/packaging/homebrew/twitch-live-notifier.rb" > "$output"
echo "$output"
