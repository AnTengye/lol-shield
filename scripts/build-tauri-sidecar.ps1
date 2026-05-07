$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..")
$bin = Join-Path $root "bin"
New-Item -ItemType Directory -Force $bin | Out-Null

$env:CGO_ENABLED = "1"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

go build -tags jsoniter -ldflags "-s -w" -o (Join-Path $bin "lol-shield-x86_64-pc-windows-msvc.exe") ./cmd/shield/main.go
go build -tags jsoniter -ldflags "-s -w" -o (Join-Path $bin "lol-shield.exe") ./cmd/shield/main.go
