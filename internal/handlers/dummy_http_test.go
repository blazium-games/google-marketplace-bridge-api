package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"google-marketplace-bridge/api/internal/models"
)

func TestDummy_CreatedIncludesScheduledTime(t *testing.T) {
	db := openTestSQLite(t)
	cfg := testConfig(t)
	h := New(cfg, db)

	payload := map[string]string{
		"email": "a@b.co", "contract_uid": "dummy-1", "company": "co", "project": "p",
		"webhook_url": "https://example.com/hook", "authorization": "Bearer tok",
	}
	buf, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/dummy", bytes.NewReader(buf))
	req.Header.Set(cfg.SecurityHeaderName, cfg.SecurityHeaderValue)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Dummy(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
	var got dummyResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !got.Success || got.Time == "" {
		t.Fatalf("response %+v", got)
	}
	at, err := time.Parse(time.RFC3339, got.Time)
	if err != nil {
		t.Fatal(err)
	}
	if delta := at.Sub(time.Now()); delta < 4*time.Minute || delta > 6*time.Minute {
		t.Fatalf("expected ~5m webhook, got delta %v", delta)
	}

	var row models.Instantiate
	if err := db.Order("id DESC").First(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.ConsumerURL == "" || row.AdminURL == "" {
		t.Fatal("expected urls in db")
	}
}
