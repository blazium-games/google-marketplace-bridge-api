.PHONY: test test-integration docker-up docker-down

# Unit tests (no Docker). Uses gotestsum when installed: go install github.com/gotestyourself/gotestsum@latest
test:
	@command -v gotestsum >/dev/null 2>&1 && gotestsum --format testname -- -count=1 ./... || go test -count=1 ./...

docker-up:
	docker compose -f docker-compose.test.yml up -d --wait

docker-down:
	docker compose -f docker-compose.test.yml down

# Integration tests (PostgreSQL via Docker). Callback is asserted within ~5s (2s delay + poll).
test-integration: docker-up
	POSTGRES_DSN=postgres://test:test@127.0.0.1:5433/bridge_test?sslmode=disable \
		SECURITY_HEADER_VALUE=integration-test-secret \
		WEBHOOK_NOTIFY_DELAY=2s \
		WEBHOOK_POLL_INTERVAL=150ms \
		sh -c 'command -v gotestsum >/dev/null 2>&1 && gotestsum --format testname -- -tags=integration -count=1 -timeout 120s ./... || go test -tags=integration -count=1 -timeout 120s -v ./...'
	$(MAKE) docker-down
