# Signing and packaging

Release tags build normal artifacts even when no signing credentials are
configured. Signing is opt-in through repository secrets so forks and local
builds keep working.

## What gets signed

| Artifact | What happens |
| --- | --- |
| `TwitchLiveNotifier.exe` | Signed when `WINDOWS_CERT_PFX_BASE64` is set. |
| Windows installer | Built after signing the executable, then signed too. |
| Windows portable ZIP | Rebuilt with the signed executable. |
| `Twitch Live to Discord.app` | Signed and notarized on a macOS runner when Apple secrets are set. |
| Linux `.deb` / AppImage | Unsigned; Linux package signing is a separate future step. |

A signed Windows build still needs SmartScreen reputation unless you use an EV
certificate or a service with an established reputation. Signing removes the
"Unknown publisher" problem first; SmartScreen warnings can clear over time.

## No-cost options

### Windows

| Option | Cost | Hardware needed | Notes |
| --- | --- | --- | --- |
| **SignPath Foundation** | Free for eligible OSS | None | Certificate issued to SignPath Foundation, private key in their HSM. Apply at <https://signpath.org/apply>. |
| **OSSign** | Free for qualifying OSS | None | Community-run initiative; signs with its own certificate. Apply at <https://ossign.org>. |
| **Self-signed certificate** | Free | None | **Not trusted** by Windows or SmartScreen; only useful for local testing. |
| **Certum Open Source Code Signing** | Discounted, not free | Smart card + reader | Hardware requirement makes it unsuitable for hosted CI. |
| **Azure Trusted Signing, SSL.com eSigner, DigiCert KeyLocker, etc.** | Paid | None/cloud | CI-friendly but not free. |

Both SignPath Foundation and OSSign decide eligibility. They generally require a
public OSI-licensed project, active maintenance, a documented download page, and
enough reputation that their certificate is not being used to vouch for unknown
software. A brand-new repository may be rejected; apply after the project has a
release history and a code-signing policy page.

The repository now includes [CODE_SIGNING.md](../CODE_SIGNING.md), which satisfies
the SignPath policy-page requirement. SignPath integration needs provider-side
project IDs and an API token; the current release workflow only implements the
PFX route.

### Applying to SignPath Foundation

Apply at <https://signpath.org/apply> with:

- Repository: `https://github.com/ausbitbank/twitch-live-notifier`
- Download page: `https://github.com/ausbitbank/twitch-live-notifier/releases`
- License: MIT
- Signing policy: link to `CODE_SIGNING.md`
- Artifacts: `TwitchLiveNotifier.exe` and the Windows installer

After acceptance, collect the SignPath organization ID, project ID, signing
policy slug, and API token. Those go into repository secrets (or a protected
`release` environment), and the Windows release job can submit the unsigned
exe/installer and download the signed result. The workflow does not currently
have those provider-specific steps because they cannot be tested before the
project is accepted.

### macOS

There is **no free substitute for an Apple Developer ID certificate**. Apple
Gatekeeper trust for distributed apps requires the paid Apple Developer Program.

Practical free paths:

- **Build from source or install through Homebrew.** Locally built binaries do
  not carry the quarantine attribute, so Gatekeeper treats them as local code.
  The included [Homebrew formula](../packaging/homebrew/README.md) installs the
  CLI and a `brew services` launch agent.
- **Ship the unsigned `.app` with instructions.** Users can right-click → Open or
  remove the quarantine attribute:
  `xattr -d com.apple.quarantine "Twitch Live to Discord.app"`.
- **Ad-hoc signing** (`codesign --sign -`) does not satisfy Gatekeeper or
  notarization for downloaded apps; it is not a public distribution solution.

## Windows

### Choose a signing route first

Modern CA/Browser Forum rules require OV and EV code-signing private keys to live
on FIPS 140-2 Level 2 / Common Criteria hardware. That means most new
certificates arrive on a USB token or in a cloud HSM, not as an exportable
`.pfx`.

| Route | Best for | How CI signs |
| --- | --- | --- |
| **Existing exportable `.pfx`** | Older certificates, some cloud CAs that let you download a PFX | GitHub Windows runner with `signtool` (current workflow). |
| **Hardware token** | You already have a USB token | Self-hosted Windows runner with the token attached; a GitHub-hosted runner cannot see it. |
| **Azure Trusted Signing** | CI-first signing without a token | Azure identity + GitHub OIDC; workflow calls the Azure signing action or `signtool` with the Trusted Signing dlib. |
| **SSL.com eSigner / DigiCert KeyLocker / Sectigo cloud** | Cloud HSM signing with a CA you already use | Workflow calls the provider's signing action/CLI. |
| **SignPath Foundation** | Eligible open-source projects, free | Upload unsigned artifact to SignPath, download signed artifact in the release job. |

The current repository workflow implements the **exportable PFX** route. If you
choose Azure Trusted Signing, SignPath, or another cloud signer, the workflow
needs a provider-specific signing step; ask and it can be wired in.

### PFX route

1. Obtain an OV or EV code-signing certificate and export it as a password-
   protected `.pfx`. Include the private key and the full chain.
2. Base64-encode it without line breaks:

   ```bash
   base64 < certificate.pfx | tr -d '\n' > certificate.pfx.b64
   ```

3. Set these repository secrets:

   | Secret | Value |
   | --- | --- |
   | `WINDOWS_CERT_PFX_BASE64` | Contents of `certificate.pfx.b64`. |
   | `WINDOWS_CERT_PASSWORD` | The `.pfx` password. |
   | `WINDOWS_TIMESTAMP_URL` | Optional RFC 3161 URL; defaults to `http://timestamp.digicert.com`. |

4. The release workflow signs `TwitchLiveNotifier.exe` before building the
   installer, then signs the installer, rebuilds the portable ZIP, and updates
   `SHA256SUMS.txt`.

For a local Linux test with an exportable PFX and `osslsigncode`:

```bash
osslsigncode sign \
  -pkcs12 certificate.pfx -pass "$PFX_PASSWORD" \
  -t http://timestamp.digicert.com -h sha256 \
  -in TwitchLiveNotifier.exe -out TwitchLiveNotifier-signed.exe
```

The Inno Setup installer is per-user (`%LOCALAPPDATA%\Programs\Twitch Live to
Discord`), creates Start Menu and optional desktop shortcuts, and never touches
`%APPDATA%\TwitchLiveNotifier`, so upgrades preserve settings and secrets.

### Hardware-token route

A GitHub-hosted runner cannot access your USB token. Use one of:

- A self-hosted Windows runner with the token attached and the signing step
  running on that runner.
- A cloud signing service that keeps the key in its HSM (recommended if you do
  not want to maintain a runner).

The current `signtool` script is easy to reuse on a self-hosted runner once the
token is visible to the Windows certificate store.

### Verify Windows signatures

On Windows:

```powershell
signtool verify /pa /v dist\TwitchLiveNotifier.exe
signtool verify /pa /v dist\installer\TwitchLiveNotifier-vX.Y.Z-Windows-x64-Setup.exe
Get-AuthenticodeSignature dist\TwitchLiveNotifier.exe | Format-List
```

On Linux with `osslsigncode` installed:

```bash
osslsigncode verify -in TwitchLiveNotifier.exe
```

## macOS

### Prerequisites

- A paid Apple Developer Program membership.
- The team ID from <https://developer.apple.com/account> → Membership.
- A **Developer ID Application** certificate. You do **not** need a Mac to create
  the certificate; you can generate the CSR and export the `.p12` on Linux.

### Generate a Developer ID Application certificate on Linux

1. Generate a private key and CSR:

   ```bash
   openssl req -new -newkey rsa:2048 -nodes \
     -keyout developerID.key \
     -out developerID.certSigningRequest \
     -subj "/emailAddress=you@example.com/CN=Your Name/C=US"
   ```

2. In the Apple Developer portal, open **Certificates, Identifiers & Profiles →
   Certificates → + → Developer ID Application**, upload the CSR, and download
   the `.cer`.
3. Convert the certificate and bundle it with the private key:

   ```bash
   openssl x509 -inform DER -in developerID.cer -out developerID.pem
   openssl pkcs12 -export \
     -inkey developerID.key \
     -in developerID.pem \
     -name "Developer ID Application: Your Name (TEAMID)" \
     -out developerID.p12
   ```

   Use a strong password and keep both the `.key` and `.p12` private.
4. Base64-encode the `.p12`:

   ```bash
   base64 < developerID.p12 | tr -d '\n' > developerID.p12.b64
   ```

5. Read the exact signing identity from the certificate:

   ```bash
   openssl x509 -inform DER -in developerID.cer -noout -subject
   ```

   Use the `CN=` value, for example
   `Developer ID Application: Your Name (TEAMID)`, as `MACOS_SIGN_IDENTITY`.

6. Create an Apple app-specific password at <https://appleid.apple.com> →
   **Sign-In and Security → App-Specific Passwords**. Label it something like
   `GitHub Actions notarization`.

### macOS repository secrets

| Secret | Value |
| --- | --- |
| `MACOS_CERT_P12_BASE64` | Contents of `developerID.p12.b64`. |
| `MACOS_CERT_PASSWORD` | Password used when exporting the `.p12`. |
| `MACOS_SIGN_IDENTITY` | `Developer ID Application: Your Name (TEAMID)`. |
| `APPLE_ID` | Apple ID used for notarization. |
| `APPLE_TEAM_ID` | Apple Developer team ID. |
| `APPLE_APP_PASSWORD` | App-specific password from the step above. |

When `MACOS_CERT_P12_BASE64` is present, a macOS runner imports the certificate
into a temporary keychain, signs the app bundles with hardened runtime and a
timestamp, notarizes them with `notarytool`, staples the ticket, and replaces the
macOS archive and checksum. If the Apple notarization secrets are missing, the
app is signed but not notarized.

### Verify macOS signatures and notarization

On a Mac:

```bash
codesign --verify --deep --strict --verbose=2 "Twitch Live to Discord.app"
spctl --assess --type execute --verbose=4 "Twitch Live to Discord.app"
xcrun stapler validate "Twitch Live to Discord.app"
```

Unsigned or un-notarized builds may require a Gatekeeper override on first launch.

## Setting the GitHub secrets

Use `gh` from this repository. For file-backed secrets:

```bash
gh secret set MACOS_CERT_P12_BASE64 < developerID.p12.b64
gh secret set WINDOWS_CERT_PFX_BASE64 < certificate.pfx.b64
```

For text secrets, let `gh` prompt so the value is not left in shell history:

```bash
gh secret set MACOS_CERT_PASSWORD
gh secret set MACOS_SIGN_IDENTITY
gh secret set APPLE_ID
gh secret set APPLE_TEAM_ID
gh secret set APPLE_APP_PASSWORD
gh secret set WINDOWS_CERT_PASSWORD
gh secret set WINDOWS_TIMESTAMP_URL
```

Check what is configured:

```bash
gh secret list
```

GitHub repository secrets are limited to 48 KB each. A normal PFX/P12 with the
leaf certificate and intermediates is well under that. If yours is larger, export
a minimal chain or fetch the certificate from a secret manager in the workflow
instead of storing it as a GitHub secret.

Recommended hardening:

- Put release secrets in a protected GitHub Environment, not just repository
  secrets, and require a reviewer for tags.
- Restrict who can push tags.
- Keep certificates and app-specific passwords in a password manager.
- Rotate the Apple app-specific password if it is ever exposed.
- Revoke and reissue a certificate immediately if its private key is exposed.

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

Linux package signing (`.deb` repository signatures and AppImage signing) is not
yet implemented.

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
