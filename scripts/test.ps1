# Unit tests: optional gotestsum (go install github.com/gotestyourself/gotestsum@latest)
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

$gotestsum = Get-Command gotestsum -ErrorAction SilentlyContinue
if ($gotestsum) {
    gotestsum --format testname -- -count=1 ./...
} else {
    go test -count=1 ./...
}
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
