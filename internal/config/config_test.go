package config

import (
	"os"
	"testing"
	"time"
)

func chdirWithoutDotenv(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
}

func TestLoad_WebhookNotifyDelayDefault(t *testing.T) {
	chdirWithoutDotenv(t)
	t.Setenv("POSTGRES_DSN", "postgres://u:p@localhost:9/db?sslmode=disable")
	t.Setenv("SECURITY_HEADER_VALUE", "secret")
	t.Setenv("WEBHOOK_NOTIFY_DELAY", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WebhookNotifyDelay != 5*time.Hour {
		t.Fatalf("got %v", cfg.WebhookNotifyDelay)
	}
}

func TestLoad_CustomDurations(t *testing.T) {
	chdirWithoutDotenv(t)
	t.Setenv("POSTGRES_DSN", "postgres://u:p@localhost:9/db?sslmode=disable")
	t.Setenv("SECURITY_HEADER_VALUE", "secret")
	t.Setenv("WEBHOOK_NOTIFY_DELAY", "3s")
	t.Setenv("WEBHOOK_POLL_INTERVAL", "500ms")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WebhookNotifyDelay != 3*time.Second {
		t.Fatalf("notify: %v", cfg.WebhookNotifyDelay)
	}
	if cfg.WebhookInterval != 500*time.Millisecond {
		t.Fatalf("poll: %v", cfg.WebhookInterval)
	}
}
