# Code signing policy

Free code signing provided by [SignPath.io](https://about.signpath.io),
certificate by [SignPath Foundation](https://signpath.org).

This policy covers the Windows release artifacts:

- `TwitchLiveNotifier.exe`
- `TwitchLiveNotifier-vX.Y.Z-Windows-x64-Setup.exe`
- `TwitchLiveNotifier-vX.Y.Z-Windows-x64.zip` (contains the signed executable)

## Release roles

- **Committers and reviewers:** [repository contributors](https://github.com/ausbitbank/twitch-live-notifier/graphs/contributors)
- **Approvers:** [repository owner](https://github.com/ausbitbank)
- **Signing approver:** repository owner (`ausbitbank`)

Every signing request is manually approved by the signing approver after the
release tag and CI build have been reviewed. Builds are produced from this
public repository by the release workflow; changes to build scripts and workflow
files receive normal code review.

## What is signed

Only artifacts built from this repository by the release workflow are signed.
No third-party or proprietary binaries are included in the signed artifacts.

## Privacy

This program does not include analytics or telemetry. It connects to:

- Twitch's public web GraphQL endpoint to read channel live status.
- The Discord webhook URLs the user configures, to post alerts.
- `api.github.com` for the optional update check, when enabled.

No streamer list, webhook URL, template, or settings data is sent to the
maintainers, SignPath, or any analytics service. Settings and webhook secrets are
stored locally. See [SECURITY.md](SECURITY.md) and the privacy section of
[README.md](README.md) for details.

## Uninstallation

The Windows installer registers an uninstaller in **Settings → Apps → Installed
apps**. Uninstalling removes the application files but deliberately leaves the
user's `%APPDATA%\TwitchLiveNotifier` settings and secrets so a reinstall keeps
the configuration. Users can delete that folder manually to remove all data.

## macOS and Linux

Apple Developer ID signing and notarization require a paid Apple Developer
Program membership and are not covered by the SignPath Foundation certificate.
Linux packages are currently unsigned. See [docs/signing.md](docs/signing.md).
