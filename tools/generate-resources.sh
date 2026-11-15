#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "$0")/.."

for size in 16 24 32 48 64 128 256; do
	magick assets/appicon.png -resize "${size}x${size}" "winres/icon${size}.png"
done

magick winres/icon256.png -define icon:auto-resize=256,128,64,48,32,24,16 app.ico
magick assets/appicon.png -resize 256x256 web/icon.png
magick docs/streamimage.png -resize 700x web/mascot.png

winres="${GO_WINRES:-go-winres}"
if ! command -v "$winres" >/dev/null 2>&1; then
	echo "Install go-winres first:" >&2
	echo "  go install github.com/tc-hib/go-winres@v0.3.3" >&2
	exit 1
fi

"$winres" make \
	--in winres/winres.json \
	--arch amd64 \
	--out rsrc \
	--product-version "${VERSION:-2.1.0}" \
	--file-version "${VERSION:-2.1.0}"

echo "Regenerated app icon, web artwork, and rsrc_windows_amd64.syso"
