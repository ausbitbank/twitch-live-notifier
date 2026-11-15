# Signing and packaging

Release tags build the normal artifacts even when no signing credentials are
configured. Signing is opt-in through repository secrets so forks and local
builds keep working.

## Windows

The release workflow builds `TwitchLiveNotifier.exe`, signs it when configured,
creates the Inno Setup installer, signs the installer, rebuilds the portable ZIP,
and updates `SHA256SUMS.txt`.

Repository secrets:

| Secret | Purpose |
| --- | --- |
| `WINDOWS_CERT_PFX_BASE64` | Base64-encoded `.pfx` code-signing certificate. |
| `WINDOWS_CERT_PASSWORD` | Password for the `.pfx`. |
| `WINDOWS_TIMESTAMP_URL` | Optional RFC 3161 timestamp server. Defaults to `http://timestamp.digicert.com`. |

To create the base64 value on Linux/macOS:

```bash
base64 < certificate.pfx | tr -d '\n' > certificate.pfx.b64
```

Use an OV or EV code-signing certificate. A newly issued certificate still needs
reputation before SmartScreen stops warning; EV certificates usually clear
faster. The executable and installer are both signed with SHA-256 and a
timestamp.

The Inno Setup installer is per-user (`%LOCALAPPDATA%\Programs\Twitch Live to
Discord`), creates Start Menu and optional desktop shortcuts, and never touches
`%APPDATA%\TwitchLiveNotifier`, so upgrades preserve settings and secrets.

## macOS

The macOS archive contains both raw binaries and `Twitch Live to Discord.app`
bundles for Intel and Apple Silicon. The app bundle starts the dashboard, opens
it in the browser, and keeps running until the app is quit.

Repository secrets:

| Secret | Purpose |
| --- | --- |
| `MACOS_CERT_P12_BASE64` | Base64-encoded Developer ID Application certificate. |
| `MACOS_CERT_PASSWORD` | Password for the `.p12`. |
| `MACOS_SIGN_IDENTITY` | Exact codesign identity, for example `Developer ID Application: Name (TEAMID)`. |
| `APPLE_ID` | Apple ID used for notarization. |
| `APPLE_TEAM_ID` | Apple Developer team ID. |
| `APPLE_APP_PASSWORD` | App-specific password for notarization. |

When `MACOS_CERT_P12_BASE64` is present, a macOS runner signs and notarizes the
app bundles and replaces the macOS archive and checksum. If the Apple
notarization secrets are missing, the app is signed but not notarized.

## Linux

The release publishes:

- `TwitchLiveNotifier-vX.Y.Z-Linux-x64.deb`
- `TwitchLiveNotifier-vX.Y.Z-Linux-x64.AppImage`
- the raw Linux binary and a `.tar.gz` containing the binary and Linux packages.

Install the deb:

```bash
sudo apt install ./TwitchLiveNotifier-vX.Y.Z-Linux-x64.deb
systemctl --user enable --now twitch-live-notifier
```

The AppImage is a single executable; mark it executable and run it. The deb
installs a desktop entry and `twitch-live-notifier-launch`, which starts the
server and opens the dashboard. The systemd user service runs `--serve` only.

## winget

The Windows job renders manifests with the installer SHA-256 and uploads
`TwitchLiveNotifier-vX.Y.Z-winget.zip`. See
[packaging/winget/README.md](../packaging/winget/README.md) for the submission
steps.

## Update checks

The app checks the latest GitHub release once a day and from the Settings card.
It reports the current version, latest version, download link, and release notes
link. It does **not** silently replace its own binary. A self-replacing updater
should only be enabled after signed artifacts and signature verification are in
place.
