; Inno Setup script for the per-user Windows installer.
; Build with:
;   iscc /DMyAppVersion=2.4.0 packaging\windows\TwitchLiveNotifier.iss
; The executable must exist at dist\TwitchLiveNotifier.exe first.

#ifndef MyAppVersion
  #define MyAppVersion "0.0.0"
#endif

#define MyAppName "Twitch Live to Discord"
#define MyAppPublisher "ausbitbank"
#define MyAppURL "https://github.com/ausbitbank/twitch-live-notifier"
#define MyAppExeName "TwitchLiveNotifier.exe"

[Setup]
AppId={{C8D5E6A8-7E3B-4A2E-9F5C-2B8A7C1D4E90}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}/releases
DefaultDirName={localappdata}\Programs\Twitch Live to Discord
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
OutputDir=..\..\dist\installer
OutputBaseFilename=TwitchLiveNotifier-v{#MyAppVersion}-Windows-x64-Setup
SetupIconFile=..\..\app.ico
UninstallDisplayIcon={app}\{#MyAppExeName}
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64
CloseApplications=yes
RestartApplications=no
DisableWelcomePage=no
LicenseFile=..\..\LICENSE

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "Create a desktop shortcut"; GroupDescription: "Additional shortcuts:"; Flags: unchecked

[Files]
Source: "..\..\dist\TwitchLiveNotifier.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\LICENSE"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "Launch {#MyAppName}"; Flags: nowait postinstall skipifsilent

; Settings live in %APPDATA%\TwitchLiveNotifier and are deliberately not touched
; by install or uninstall, so upgrades keep streamers, destinations, and secrets.
