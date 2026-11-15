param(
    [Parameter(Mandatory = $true)]
    [string]$Version,
    [string]$ExePath = "dist\TwitchLiveNotifier.exe",
    [string]$OutputDir = "dist\installer"
)

$ErrorActionPreference = "Stop"
$root = Resolve-Path (Join-Path $PSScriptRoot "..\..")
Push-Location $root
try {
    if (-not (Test-Path $ExePath)) {
        throw "Build $ExePath before creating the installer."
    }

    function Sign-File([string]$Path) {
        $pfxBase64 = $env:WINDOWS_CERT_PFX_BASE64
        if ([string]::IsNullOrWhiteSpace($pfxBase64)) {
            Write-Host "WINDOWS_CERT_PFX_BASE64 is not set; leaving $Path unsigned."
            return
        }
        $signtoolPath = (Get-Command signtool.exe -ErrorAction SilentlyContinue).Source
        if (-not $signtoolPath) {
            $sdkRoot = "${env:ProgramFiles(x86)}\Windows Kits\10\bin"
            if (Test-Path $sdkRoot) {
                $signtoolPath = (Get-ChildItem $sdkRoot -Recurse -Filter signtool.exe |
                    Sort-Object FullName -Descending | Select-Object -First 1).FullName
            }
        }
        if (-not $signtoolPath) {
            throw "signtool.exe was not found. Install the Windows SDK or skip signing."
        }
        $pfx = Join-Path ([System.IO.Path]::GetTempPath()) ("twitch-live-" + [guid]::NewGuid().ToString("N") + ".pfx")
        try {
            [System.IO.File]::WriteAllBytes($pfx, [Convert]::FromBase64String($pfxBase64))
            $timestamp = if ($env:WINDOWS_TIMESTAMP_URL) { $env:WINDOWS_TIMESTAMP_URL } else { "http://timestamp.digicert.com" }
            & $signtoolPath sign /fd sha256 /td sha256 /tr $timestamp /f $pfx /p $env:WINDOWS_CERT_PASSWORD $Path
            if ($LASTEXITCODE -ne 0) {
                throw "Signing failed for $Path."
            }
        }
        finally {
            if (Test-Path $pfx) { Remove-Item $pfx -Force }
        }
    }

    Sign-File $ExePath

    $isccPath = $null
    $iscc = Get-Command iscc.exe -ErrorAction SilentlyContinue
    if ($iscc) {
        $isccPath = $iscc.Source
    }
    else {
        $candidate = "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe"
        if (Test-Path $candidate) {
            $isccPath = $candidate
        }
    }
    if (-not $isccPath) {
        throw "Inno Setup 6 was not found. Install it with: choco install innosetup -y"
    }

    New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
    & $isccPath "/DMyAppVersion=$Version" "packaging\windows\TwitchLiveNotifier.iss"
    if ($LASTEXITCODE -ne 0) {
        throw "Inno Setup failed."
    }

    $installer = Join-Path $OutputDir "TwitchLiveNotifier-v$Version-Windows-x64-Setup.exe"
    if (-not (Test-Path $installer)) {
        throw "Installer was not created at $installer."
    }
    Sign-File $installer
    Write-Host "Installer: $installer"
}
finally {
    Pop-Location
}
