# Starts a real ONTOVELA core against local PostgreSQL for live source-class
# verification, applies embedded migrations at startup, and serves until Ctrl+C.
#
# Usage:
#   .\scripts\run-live-core.ps1 [-PgDsn <dsn>] [-Addr <host:port>] [-CoreDir <core-backend-path>]
param(
    [string]$PgDsn = "postgres://ontovela:ontovela@localhost:5432/ontovela?sslmode=disable",
    [string]$Addr = ":8080",
    [string]$CoreDir = (Join-Path $PSScriptRoot "..\..\ONTOVELA\backend")
)

$ErrorActionPreference = "Continue"
$env:GOWORK = "off"

if (-not (Test-Path (Join-Path $CoreDir "go.mod"))) {
    Write-Error "Core backend not found at $CoreDir; pass -CoreDir <path>"
    exit 1
}

Write-Host "== ONTOVELA live core =="
Write-Host "DSN:      $PgDsn"
Write-Host "Addr:     $Addr"
Write-Host "Core dir: $CoreDir"

Push-Location $CoreDir
try {
    Write-Host "== Applying migrations =="
    Write-Host "== Starting core on $Addr =="
    go run ./cmd/ontovela -migrate -pg-dsn $PgDsn -addr $Addr
    if ($LASTEXITCODE -ne 0) { Write-Error "core exited with error $LASTEXITCODE"; exit 1 }
} finally {
    Pop-Location
}
