$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$dist = Join-Path $root "dist"
$stage = Join-Path $dist "ogit-linux-amd64"
$binary = Join-Path $stage "ogit"
$zipPath = Join-Path $dist "ogit-linux-amd64.zip"

New-Item -ItemType Directory -Force -Path $stage | Out-Null
Remove-Item -Force -ErrorAction SilentlyContinue $binary, $zipPath

$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

go build -trimpath -ldflags="-s -w" -o $binary .

if (Test-Path $zipPath) {
  Remove-Item -Force $zipPath
}

Compress-Archive -Path $binary -DestinationPath $zipPath
Write-Host "Built $zipPath"
