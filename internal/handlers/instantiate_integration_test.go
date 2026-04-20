//go:build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"google-marketplace-bridge/api/internal/config"
	"google-marketplace-bridge/api/internal/models"
	"google-marketplace-bridge/api/internal/webhooks"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	tmp, err := os.MkdirTemp("", "handlers-integration")
	if err != nil {
		panic(err)
	}
	if err := os.Chdir(tmp); err != nil {
		panic(err)
	}

	if os.Getenv("POSTGRES_DSN") == "" {
		_ = os.Setenv("POSTGRES_DSN", "postgres://test:test@127.0.0.1:5433/bridge_test?sslmode=disable")
	}
	if os.Getenv("SECURITY_HEADER_VALUE") == "" {
		_ = os.Setenv("SECURITY_HEADER_VALUE", "integration-test-secret")
	}
	_ = os.Setenv("WEBHOOK_NOTIFY_DELAY", "2s")
	_ = os.Setenv("WEBHOOK_POLL_INTERVAL", "150ms")

	code := m.Run()
	_ = os.Chdir(wd)
	_ = os.RemoveAll(tmp)
	os.Exit(code)
}

func TestIntegration_WebhookCallbackWithinFiveSeconds(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	if err := models.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("TRUNCATE instantiates RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("truncate: %v", err)
	}

	var gotAuth atomic.Value
	var gotBodyMu sync.Mutex
	var gotBody string
	var webhookCalls atomic.Int32
	wh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookCalls.Add(1)
		gotAuth.Store(r.Header.Get("Authorization"))
		b, _ := io.ReadAll(r.Body)
		gotBodyMu.Lock()
		gotBody = string(b)
		gotBodyMu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer wh.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go webhooks.RunWorker(ctx, db, cfg.WebhookInterval)

	h := New(cfg, db)
	mux := http.NewServeMux()
	mux.HandleFunc("/instantiate", h.Instantiate)
	api := httptest.NewServer(mux)
	defer api.Close()

	body := map[string]string{
		"email":          "hook@example.com",
		"contract_uid":   "contract-int-1",
		"company":        "Acme",
		"project":        "P1",
		"webhook_url":    wh.URL,
		"authorization":  "Bearer callback-token",
	}
	buf, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, api.URL+"/instantiate", bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(cfg.SecurityHeaderName, cfg.SecurityHeaderValue)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("instantiate status %d", resp.StatusCode)
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if webhookCalls.Load() >= 1 {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if webhookCalls.Load() < 1 {
		t.Fatal("webhook not called within 5s")
	}

	if gotAuth.Load() != "Bearer callback-token" {
		t.Fatalf("authorization header: %v", gotAuth.Load())
	}
	var payload struct {
		ContractUID string `json:"contract_uid"`
		ConsumerURL string `json:"consumer_url"`
		AdminURL    string `json:"admin_url"`
	}
	gotBodyMu.Lock()
	bodyStr := gotBody
	gotBodyMu.Unlock()
	if err := json.Unmarshal([]byte(bodyStr), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.ContractUID != "contract-int-1" {
		t.Fatalf("payload: %s", bodyStr)
	}
	if payload.ConsumerURL == "" || payload.AdminURL == "" {
		t.Fatalf("expected consumer_url and admin_url: %s", bodyStr)
	}
	if payload.ConsumerURL == payload.AdminURL {
		t.Fatal("consumer and admin urls should differ")
	}

	var row models.Instantiate
	if err := db.Order("id DESC").First(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.WebhookDeliveredAt == nil {
		t.Fatal("expected webhook_delivered_at")
	}
}
