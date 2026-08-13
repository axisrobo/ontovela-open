# Packages the ONTOVELA developer binary with a README into a zip.
$env:GOWORK = "off"
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$CoreBackend = Join-Path (Split-Path $RepoRoot -Parent) "ONTOVELA\backend"
$Version = if ($env:ONTOVELA_VERSION) { $env:ONTOVELA_VERSION } else { "dev" }
$Dist = Join-Path $RepoRoot "dist"
New-Item -ItemType Directory -Force -Path $Dist | Out-Null
$ZipName = Join-Path $Dist "ontovela-$Version-windows-amd64.zip"
$Temp = Join-Path $Dist "ontovela-$Version"
New-Item -ItemType Directory -Force -Path $Temp | Out-Null

if (Test-Path (Join-Path $CoreBackend "go.mod")) {
    Push-Location $CoreBackend
    if ($env:LD_FLAGS) { go build "-ldflags=$env:LD_FLAGS" -o (Join-Path $Temp "ontovela.exe") ./cmd/ontovela }
    else { go build -o (Join-Path $Temp "ontovela.exe") ./cmd/ontovela }
    if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
    Pop-Location
} else {
    Write-Error "core backend not found at $CoreBackend"
    exit 1
}
Copy-Item (Join-Path $RepoRoot "README.md") (Join-Path $Temp "README.md") -ErrorAction SilentlyContinue
Copy-Item (Join-Path $RepoRoot "docs\quickstart.md") (Join-Path $Temp "quickstart.md") -ErrorAction SilentlyContinue

Compress-Archive -Path (Join-Path $Temp "*") -DestinationPath $ZipName -Force
Remove-Item $Temp -Recurse -Force
Write-Output "packaged $ZipName"
