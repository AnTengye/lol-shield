param(
  [string]$FrontendDir = ".\frontend",
  [string]$TargetDir = "internal\client\web\dist"
)

$ErrorActionPreference = "Stop"

$frontendPath = Resolve-Path $FrontendDir
$targetPath = Join-Path (Get-Location) $TargetDir
$distPath = Join-Path $frontendPath "dist"

if (-not (Test-Path $distPath)) {
  throw "Frontend dist not found at $distPath. Run 'pnpm build' in lol-shield-front first."
}

New-Item -ItemType Directory -Force $targetPath | Out-Null
Remove-Item -Recurse -Force (Join-Path $targetPath "*") -ErrorAction SilentlyContinue
Copy-Item -Recurse -Force (Join-Path $distPath "*") $targetPath

Write-Host "Synced frontend dist to $targetPath"
