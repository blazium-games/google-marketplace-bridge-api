package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"google-marketplace-bridge/api/internal/apierr"
	"google-marketplace-bridge/api/internal/config"
	"google-marketplace-bridge/api/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		HTTPAddr:            ":0",
		SecurityHeaderName:  "X-Test-Secret",
		SecurityHeaderValue: "unit-test-secret",
		WebhookInterval:     time.Millisecond * 100,
		WebhookNotifyDelay:  0,
	}
}

func openTestSQLite(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func errCode(t *testing.T, rec *httptest.ResponseRecorder) int {
	t.Helper()
	var b struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("decode error body: %v body=%q", err, rec.Body.String())
	}
	return b.Code
}

func TestInstantiate_MethodNotAllowed(t *testing.T) {
	h := New(testConfig(t), openTestSQLite(t))
	req := httptest.NewRequest(http.MethodGet, "/instantiate", nil)
	rec := httptest.NewRecorder()
	h.Instantiate(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got %d", rec.Code)
	}
	if c := errCode(t, rec); c != apierr.CodeMethodNotAllowed {
		t.Fatalf("code %d", c)
	}
}

func TestInstantiate_Unauthorized(t *testing.T) {
	h := New(testConfig(t), openTestSQLite(t))
	body := `{"email":"a@b.co","contract_uid":"c1","company":"co","project":"p","webhook_url":"http://x","authorization":"z"}`
	req := httptest.NewRequest(http.MethodPost, "/instantiate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Instantiate(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
	if c := errCode(t, rec); c != apierr.CodeUnauthorized {
		t.Fatalf("code %d", c)
	}
}

func TestInstantiate_InvalidJSON(t *testing.T) {
	h := New(testConfig(t), openTestSQLite(t))
	cfg := testConfig(t)
	req := httptest.NewRequest(http.MethodPost, "/instantiate", strings.NewReader(`{`))
	req.Header.Set(cfg.SecurityHeaderName, cfg.SecurityHeaderValue)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Instantiate(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d", rec.Code)
	}
	if c := errCode(t, rec); c != apierr.CodeInvalidJSON {
		t.Fatalf("code %d", c)
	}
}

func TestInstantiate_UnknownJSONField(t *testing.T) {
	h := New(testConfig(t), openTestSQLite(t))
	cfg := testConfig(t)
	raw := `{"email":"a@b.co","contract_uid":"c1","company":"co","project":"p","webhook_url":"http://x","authorization":"z","extra":1}`
	req := httptest.NewRequest(http.MethodPost, "/instantiate", strings.NewReader(raw))
	req.Header.Set(cfg.SecurityHeaderName, cfg.SecurityHeaderValue)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Instantiate(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("got %d body=%s", rec.Code, rec.Body.String())
	}
	if c := errCode(t, rec); c != apierr.CodeJSONUnknownField {
		t.Fatalf("code %d", c)
	}
}

func TestInstantiate_Created(t *testing.T) {
	db := openTestSQLite(t)
	cfg := testConfig(t)
	h := New(cfg, db)

	payload := map[string]string{
		"email": "a@b.co", "contract_uid": "uid-99", "company": "co", "project": "p",
		"webhook_url": "https://example.com/hook", "authorization": "Bearer tok",
	}
	buf, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/instantiate", bytes.NewReader(buf))
	req.Header.Set(cfg.SecurityHeaderName, cfg.SecurityHeaderValue)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Instantiate(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
	var got instantiateResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.ID == 0 {
		t.Fatal("expected id")
	}

	var row models.Instantiate
	if err := db.First(&row, got.ID).Error; err != nil {
		t.Fatal(err)
	}
	if row.ContractUID != "uid-99" || row.WebhookAuthorization != "Bearer tok" {
		t.Fatalf("row mismatch: %+v", row)
	}
}
