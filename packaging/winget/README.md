# winget manifests

The release workflow builds the Inno Setup installer and renders these
manifests with the installer's SHA-256. They are uploaded as
`TwitchLiveNotifier-vX.Y.Z-winget.zip` on the GitHub release.

To submit to the winget community repository:

1. Download the winget zip from the release.
2. Copy the manifests into a `winget-pkgs` fork under
   `manifests/a/ausbitbank/TwitchLiveNotifier/<version>/`.
3. Validate with `winget validate --manifest <directory>` on Windows.
4. Open a pull request against `microsoft/winget-pkgs`.

To render locally:

```bash
packaging/winget/render.sh 2.4.0 dist/installer/TwitchLiveNotifier-v2.4.0-Windows-x64-Setup.exe
```
