# Verified auto-update

The app can check GitHub releases, download the correct build for the current
platform, verify it, install it in place, and restart. Settings live outside the
binary, so streamers, destinations, templates, and encrypted secrets survive an
update.

## Checks

- The app checks the latest GitHub release after startup and every 24 hours
  when **Settings → Check GitHub for updates** is enabled.
- **Check now** runs a manual check.
- A check only reports status; nothing is downloaded until you choose
  **Install update**.
- Update checks work offline and can be disabled. The rest of the app keeps
  working without GitHub.

## What gets downloaded

The release publishes an `update-manifest.json` next to the normal artifacts.
It maps each platform to one asset and its SHA-256:

```json
{
  "version": "2.5.0",
  "assets": {
    "windows-amd64": {
      "name": "TwitchLiveNotifier.exe",
      "url": "https://github.com/.../TwitchLiveNotifier.exe",
      "sha256": "…",
      "size": 12345678
    }
  }
}
```

If the manifest is missing, the app falls back to the release asset list and
`SHA256SUMS.txt` — unless a signing key is pinned, in which case self-update
requires a valid signed manifest and fails closed. The asset is chosen by OS,
architecture, and install mode. For example, a Windows portable copy gets the
raw `.exe`, an AppImage gets the `.AppImage`, and a macOS `.app` gets the
signed/notarized macOS archive.

## Install flow

1. **Download** to a temporary folder on the same filesystem as the app.
2. **Verify** the SHA-256 checksum. A missing or mismatched checksum aborts the
   install.
3. **Preflight** the downloaded binary with `--version`.
4. **Apply**:
   - Windows per-user install or portable: keep the old `.exe` as `.old`, move
     the new one into place, and relaunch.
   - Linux raw binary: replace the binary and relaunch.
   - Linux AppImage: replace the AppImage file and relaunch.
   - macOS `.app`: extract the archive, then a small helper waits for the app to
     exit, swaps the whole bundle, and relaunches it.
5. **Restart** with the same command-line arguments. The old binary or bundle is
   kept as `.old` for rollback.

The dashboard shows download progress, checksum verification, install status,
and a "restarting" state. After the app comes back, the page reconnects and
shows the new version.

## Platform notes

| Install | Behavior |
| --- | --- |
| Windows installer / portable ZIP | In-place update in the per-user install folder |
| Linux raw binary (writable folder) | In-place update |
| Linux AppImage | In-place update of the AppImage file |
| macOS `.app` | Bundle swap with relaunch; code signature stays inside the bundle |
| Linux `.deb` / system package | Guided: download manually or use the package manager |
| Read-only location / `/usr`, `/opt` | Guided: move the app or update with the package manager |
| Docker | Guided: `docker compose pull` and `docker compose up -d` |

If the app is in a non-writable location, the install button is replaced by a
manual download link and a clear explanation. The app never writes to a system
location without permission.

## Security

- Downloads use HTTPS from the GitHub release.
- The SHA-256 checksum comes from the same release and is verified before the
  swap. This protects against truncated or corrupted downloads.
- The code also supports an optional pinned Ed25519 public key
  (`updateSigningPublicKeyHex`) for `update-manifest.json`. When a key is pinned,
  a release without a valid signature is rejected. This is stronger than a
  checksum because it does not rely on the same release host.
- Windows Authenticode and macOS notarization remain separate release-signing
  features. Unsigned development builds still work.

## Rollback

The previous binary or app bundle is retained as `<name>.old`. If a new version
fails to start, replace the current file with the `.old` copy. The app also
verifies the new binary before swapping, which catches the most common bad
download or wrong-architecture case before the restart.

## Docker

Docker images are immutable and should be updated by pulling a new tag:

```bash
docker compose pull
docker compose up -d
```

The app detects a container and disables in-place self-update, showing the same
command in the dashboard.

## For maintainers

- The release workflow builds and uploads `update-manifest.json` after the
  Windows and macOS artifact jobs finish.
- If you pin `updateSigningPublicKeyHex`, publish
  `update-manifest.json.sig` with an Ed25519 signature over the exact manifest
  bytes. Without a pinned key, unsigned manifests are accepted and the checksum
  is still enforced.
- To enable optional signing, generate a key once:

  ```bash
  openssl genpkey -algorithm ED25519 -out update-signing-key.pem
  # Public key as 32-byte hex for update_key.go:
  openssl pkey -in update-signing-key.pem -pubout -outform DER \
    | tail -c 32 | xxd -p -c 64
  ```

  Put the private PEM in the `UPDATE_SIGNING_KEY` repository secret and set the
  public hex in `updateSigningPublicKeyHex`. The release workflow signs
  `update-manifest.json` when the secret is present; the updater requires a valid
  signature when a public key is pinned.
- Keep asset names stable. The updater prefers the manifest but falls back to
  `SHA256SUMS.txt` and platform-aware asset matching.
