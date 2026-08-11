# Verifies all ONTOVELA-open Go modules, SDKs, and the contract drift guard.
$env:GOWORK = "off"
$Root = Resolve-Path (Join-Path $PSScriptRoot "..")

foreach ($module in @("sdk\go", "contract")) {
    Push-Location (Join-Path $Root $module)
    go test ./...
    if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
    go vet ./...
    if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
    Pop-Location
}

foreach ($adapter in @("harmovela", "prediction", "httpwebhook", "stream", "sqlrest", "edge", "effect", "mqtt", "ros2", "opcua")) {
    Push-Location (Join-Path $Root "adapters\$adapter")
    go test ./...
    if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
    Pop-Location
}

Push-Location (Join-Path $Root "sdk\python")
python -m unittest tests.test_client -v
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location

Push-Location (Join-Path $Root "sdk\typescript")
npm test
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location

Write-Output "all checks passed"
