# Homebrew formula (free macOS path)

Apple Developer ID signing and notarization require a paid Apple Developer
Program membership. The free alternative is to build from source with Homebrew;
locally built binaries do not carry the quarantine attribute that triggers
Gatekeeper for downloaded apps.

This directory contains a formula template and renderer. A released formula can
live in an `ausbitbank/homebrew-tap` repository or be installed directly from a
checkout:

```bash
brew install --build-from-source ./packaging/homebrew/twitch-live-notifier.rb
brew services start twitch-live-notifier
open http://127.0.0.1:8787/
```

Render a formula for a release after downloading the GitHub source tarball:

```bash
curl -L -o /tmp/twitch-live-notifier.tar.gz \
  https://github.com/ausbitbank/twitch-live-notifier/archive/refs/tags/v2.4.0.tar.gz
packaging/homebrew/render.sh 2.4.0 /tmp/twitch-live-notifier.tar.gz
```

Copy the rendered file into a tap repository as
`Formula/twitch-live-notifier.rb` and update the SHA-256 on each release.

The formula installs a CLI binary and a `brew services` launch agent. It does not
install a `.app` bundle. The dashboard opens in the browser at
`http://127.0.0.1:8787/`.

For users who specifically want the `.app`, an unsigned bundle can be downloaded
and opened with **right-click → Open** or after removing the quarantine
attribute:

```bash
xattr -d com.apple.quarantine "Twitch Live to Discord.app"
```
