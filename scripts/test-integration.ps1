# Requires Docker. Starts Postgres, runs integration tests (webhook within ~5s), tears down.
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

docker compose -f docker-compose.test.yml up -d --wait
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$env:POSTGRES_DSN = "postgres://test:test@127.0.0.1:5433/bridge_test?sslmode=disable"
$env:SECURITY_HEADER_VALUE = "integration-test-secret"
$env:WEBHOOK_NOTIFY_DELAY = "2s"
$env:WEBHOOK_POLL_INTERVAL = "150ms"

$gotestsum = Get-Command gotestsum -ErrorAction SilentlyContinue
$code = 0
try {
    if ($gotestsum) {
        gotestsum --format testname -- -tags=integration -count=1 -timeout 120s ./...
    } else {
        go test -tags=integration -count=1 -timeout 120s -v ./...
    }
    $code = $LASTEXITCODE
} finally {
    docker compose -f docker-compose.test.yml down
}
exit $code
