package webhooks

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"google-marketplace-bridge/api/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Instantiate{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestProcessDue_DeliversAndMarksDelivered(t *testing.T) {
	db := openTestDB(t)

	var gotMethod string
	var gotAuth string
	var gotPayload CallbackPayload

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	row := models.Instantiate{
		Email:                "a@b.co",
		ContractUID:          "c-webhook",
		Company:              "co",
		Project:              "p",
		WebhookURL:           srv.URL,
		WebhookAuthorization: "Bearer unit",
		ConsumerURL:          "https://consumer.pogr-bridge.invalid/a",
		AdminURL:             "https://admin.pogr-bridge.invalid/b",
		WebhookNotifyAt:      time.Now().Add(-time.Minute),
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	processDue(context.Background(), db, client)

	if gotMethod != http.MethodPost {
		t.Fatalf("method %q", gotMethod)
	}
	if gotAuth != "Bearer unit" {
		t.Fatalf("auth %q", gotAuth)
	}
	if gotPayload.ContractUID != "c-webhook" {
		t.Fatalf("payload %+v", gotPayload)
	}
	if gotPayload.ConsumerURL != row.ConsumerURL || gotPayload.AdminURL != row.AdminURL {
		t.Fatalf("urls %+v", gotPayload)
	}

	var updated models.Instantiate
	if err := db.First(&updated, row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if updated.WebhookDeliveredAt == nil {
		t.Fatal("expected WebhookDeliveredAt")
	}
}

func TestDeliverOne_Non2xxDoesNotMarkDelivered(t *testing.T) {
	db := openTestDB(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	row := models.Instantiate{
		Email:                "a@b.co",
		ContractUID:          "c-fail",
		Company:              "co",
		Project:              "p",
		WebhookURL:           srv.URL,
		WebhookAuthorization: "tok",
		ConsumerURL:          "https://consumer.pogr-bridge.invalid/x",
		AdminURL:             "https://admin.pogr-bridge.invalid/y",
		WebhookNotifyAt:      time.Now().Add(-time.Second),
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	err := deliverOne(context.Background(), db, client, &row)
	if err == nil {
		t.Fatal("expected error")
	}

	var updated models.Instantiate
	_ = db.First(&updated, row.ID)
	if updated.WebhookDeliveredAt != nil {
		t.Fatal("should not mark delivered on 500")
	}
}
