# Builds and vets every Go module in the repository.
$env:GOWORK = "off"
$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
$Modules = @("sdk\go", "contract")
$Modules += Get-ChildItem (Join-Path $Root "adapters") -Directory | Select-Object -ExpandProperty Name | ForEach-Object { "adapters\$_" }
foreach ($module in $Modules) {
    Push-Location (Join-Path $Root $module)
    go build ./...
    if ($LASTEXITCODE -ne 0) { Pop-Location; Write-Error "build failed: $module"; exit 1 }
    go vet ./...
    if ($LASTEXITCODE -ne 0) { Pop-Location; Write-Error "vet failed: $module"; exit 1 }
    Pop-Location
}
Write-Output "all $($Modules.Count) modules build and vet clean"
