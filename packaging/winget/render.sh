#!/usr/bin/env bash
# Render winget manifests for a built Windows installer.
# Usage: packaging/winget/render.sh <version> <installer.exe> [output-dir]
set -euo pipefail

version="${1:?usage: render.sh <version> <installer.exe> [output-dir]}"
installer="${2:?usage: render.sh <version> <installer.exe> [output-dir]}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
outdir="${3:-$root/dist/winget}"

if [[ ! -f "$installer" ]]; then
  echo "installer not found: $installer" >&2
  exit 1
fi

mkdir -p "$outdir"
sha="$(sha256sum "$installer" | awk '{print toupper($1)}')"
for template in "$root"/packaging/winget/manifests/*.yaml; do
  name="$(basename "$template")"
  sed \
    -e "s/__VERSION__/$version/g" \
    -e "s/__SHA256__/$sha/g" \
    "$template" > "$outdir/$name"
done

cat > "$outdir/README.txt" <<EOF
winget manifests for Twitch Live to Discord $version

Generated from the signed-or-unsigned installer:
  $(basename "$installer")
SHA256: $sha

These are ready to copy into a winget-pkgs pull request under:
  manifests/a/ausbitbank/TwitchLiveNotifier/$version/
EOF

echo "$outdir"
