#requires -version 5.0
<#
  QuickSSH Terminal installer (Windows)
  Usage:  irm https://github.com/allenbijo/QuickSSH-terminal/releases/latest/download/install.ps1 | iex
#>

$ErrorActionPreference = 'Stop'

$Repo = 'allenbijo/QuickSSH-terminal'
$Bin  = 'qsh'

function Write-Info ($msg)  { Write-Host $msg -ForegroundColor Green }
function Write-Warn2 ($msg) { Write-Host $msg -ForegroundColor Yellow }
function Write-Fail ($msg)  { Write-Host $msg -ForegroundColor Red; exit 1 }

# Detect arch.
$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    default { Write-Fail "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

# Latest version from GitHub.
try {
    $latest = Invoke-RestMethod -UseBasicParsing -Headers @{ 'User-Agent' = 'qsh-installer' } `
        -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $version = $latest.tag_name
} catch {
    Write-Fail "Failed to query GitHub: $_"
}

$asset = "$($Bin)_Windows_$arch.zip"
$url   = "https://github.com/$Repo/releases/download/$version/$asset"

Write-Info "Installing $Bin $version for Windows/$arch..."
Write-Warn2 "Source: $url"

$installDir = Join-Path $env:LOCALAPPDATA "Programs\qsh"
$tmp        = Join-Path $env:TEMP "qsh-install"
$zipPath    = Join-Path $tmp $asset

New-Item -ItemType Directory -Force -Path $installDir | Out-Null
New-Item -ItemType Directory -Force -Path $tmp        | Out-Null

try {
    Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile $zipPath

    # Extract and copy the binary.
    Expand-Archive -LiteralPath $zipPath -DestinationPath $tmp -Force
    $exe = Get-ChildItem -Path $tmp -Filter "$Bin.exe" -Recurse | Select-Object -First 1
    if (-not $exe) { Write-Fail "Did not find $Bin.exe in the archive." }
    Copy-Item -Path $exe.FullName -Destination (Join-Path $installDir "$Bin.exe") -Force
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

Write-Info "Installed $installDir\$Bin.exe"

# Add to user PATH if not already there.
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (-not ($userPath -split ';' | Where-Object { $_ -eq $installDir })) {
    [Environment]::SetEnvironmentVariable('Path', "$userPath;$installDir", 'User')
    Write-Warn2 "Added $installDir to your user PATH. Open a new terminal for this to take effect."
}

Write-Info "Run it with:  $Bin"
