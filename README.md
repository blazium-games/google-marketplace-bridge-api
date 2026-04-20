# Google Marketplace Bridge API

Go HTTP service for marketplace instantiation: `POST /instantiate`, `POST /dummy` (same body; 5-minute webhook delay), `GET`/`HEAD` `/` and `/health` for probes, PostgreSQL persistence, and delayed webhook callbacks.

See **`doc.md`** for request/response details.

## Quick start

1. Copy `example.env` to `.env` and set `POSTGRES_DSN`, `SECURITY_HEADER_VALUE`, and optional tuning.
2. Run: `go run .` or build with `go build -o api.exe .`

## Tests

- Unit: `go test ./...` or `make test` / `scripts/test.ps1`
- Integration (Docker): `make test-integration` or `scripts/test-integration.ps1`

## Error codes

See `reference.md`.
