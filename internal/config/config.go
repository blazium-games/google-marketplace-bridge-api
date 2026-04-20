package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds runtime settings from the environment (and optional .env file).
type Config struct {
	HTTPAddr            string
	PostgresDSN         string
	SecurityHeaderName  string
	SecurityHeaderValue string
	WebhookInterval     time.Duration
	// WebhookNotifyDelay is how long after create before the callback may be sent (default 5h).
	WebhookNotifyDelay time.Duration
}

// Load reads configuration from environment variables, optionally loading a .env file when present.
func Load() (*Config, error) {
	_ = godotenv.Load()

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("POSTGRES_DSN is required")
	}

	name := os.Getenv("SECURITY_HEADER_NAME")
	if name == "" {
		name = "X-Instantiate-Secret"
	}
	val := os.Getenv("SECURITY_HEADER_VALUE")
	if val == "" {
		return nil, fmt.Errorf("SECURITY_HEADER_VALUE is required")
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	interval := 30 * time.Second
	if s := os.Getenv("WEBHOOK_POLL_INTERVAL"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("WEBHOOK_POLL_INTERVAL: %w", err)
		}
		interval = d
	}

	notifyDelay := 5 * time.Hour
	if s := os.Getenv("WEBHOOK_NOTIFY_DELAY"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("WEBHOOK_NOTIFY_DELAY: %w", err)
		}
		notifyDelay = d
	}

	return &Config{
		HTTPAddr:            addr,
		PostgresDSN:         dsn,
		SecurityHeaderName:  name,
		SecurityHeaderValue: val,
		WebhookInterval:     interval,
		WebhookNotifyDelay:  notifyDelay,
	}, nil
}
