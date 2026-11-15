$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$Version = if ($env:VERSION) { $env:VERSION } else { "2.0.0" }
New-Item -ItemType Directory -Force -Path dist | Out-Null
Remove-Item -Recurse -Force -Path dist/docs -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path dist/docs | Out-Null

$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:GOAMD64 = "v1"

go build -trimpath -buildvcs=false `
  -ldflags "-s -w -H=windowsgui -buildid= -X main.Version=$Version" `
  -o dist/TwitchLiveNotifier.exe .

Copy-Item README.md, LICENSE -Destination dist -Force
Copy-Item docs/*.png -Destination dist/docs -Force

Get-FileHash dist/TwitchLiveNotifier.exe -Algorithm SHA256 |
  ForEach-Object { "$($_.Hash.ToLower())  dist/TwitchLiveNotifier.exe" } |
  Set-Content dist/SHA256SUMS.txt

Get-Item dist/TwitchLiveNotifier.exe | Format-List Name, Length, LastWriteTime
Get-Content dist/SHA256SUMS.txt
