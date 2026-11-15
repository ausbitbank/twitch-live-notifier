#!/usr/bin/env bash
# Optional macOS signing and notarization. Both stages are skipped when the
# corresponding repository secrets are absent, so unsigned releases still work.
# Usage: packaging/macos/sign-and-notarize.sh "dist/macos-arm64/Twitch Live to Discord.app"
set -euo pipefail

app="${1:?usage: sign-and-notarize.sh <app-bundle>}"

if [[ -n "${MACOS_CERT_P12_BASE64:-}" ]]; then
  tmpdir="$(mktemp -d)"
  keychain="$tmpdir/twitch-live.keychain-db"
  p12="$tmpdir/twitch-live.p12"
  original_keychains="$(security list-keychains -d user | tr -d '"' | tr '\n' ' ')"
  cleanup() {
    security list-keychains -d user -s $original_keychains >/dev/null 2>&1 || true
    security delete-keychain "$keychain" >/dev/null 2>&1 || true
    rm -rf "$tmpdir"
  }
  trap cleanup EXIT
  echo "$MACOS_CERT_P12_BASE64" | base64 -D > "$p12"
  security create-keychain -p "" "$keychain"
  security set-keychain-settings -lut 3600 "$keychain"
  security unlock-keychain -p "" "$keychain"
  security import "$p12" -k "$keychain" -P "${MACOS_CERT_PASSWORD:-}" -T /usr/bin/codesign
  security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k "" "$keychain" >/dev/null
  security list-keychains -d user -s "$keychain" $original_keychains
  codesign --force --deep --options runtime --timestamp \
    --sign "${MACOS_SIGN_IDENTITY:?MACOS_SIGN_IDENTITY is required with MACOS_CERT_P12_BASE64}" \
    "$app"
  codesign --verify --deep --strict --verbose=2 "$app"
  echo "Signed $app"
else
  echo "MACOS_CERT_P12_BASE64 is not set; leaving $app unsigned."
fi

if [[ -n "${APPLE_ID:-}" && -n "${APPLE_TEAM_ID:-}" && -n "${APPLE_APP_PASSWORD:-}" ]]; then
  if [[ -z "${MACOS_CERT_P12_BASE64:-}" ]]; then
    echo "Notarization needs a signed app; skipping." >&2
  else
    zip_path="$(mktemp -d)/TwitchLiveNotifier.zip"
    ditto -c -k --keepParent "$app" "$zip_path"
    xcrun notarytool submit "$zip_path" \
      --apple-id "$APPLE_ID" \
      --team-id "$APPLE_TEAM_ID" \
      --password "$APPLE_APP_PASSWORD" \
      --wait
    xcrun stapler staple "$app"
    echo "Notarized and stapled $app"
  fi
else
  echo "Apple notarization secrets are not set; skipping notarization."
fi
