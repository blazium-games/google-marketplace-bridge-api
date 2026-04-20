package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"google-marketplace-bridge/api/internal/models"

	"gorm.io/gorm"
)

// CallbackPayload is the JSON sent to webhook_url after the delay.
type CallbackPayload struct {
	ContractUID string `json:"contract_uid"`
	ConsumerURL string `json:"consumer_url"`
	AdminURL    string `json:"admin_url"`
}

// RunWorker periodically finds due records and POSTs to their webhook_url.
// It runs one pass immediately so the first delivery is not delayed by a full poll interval.
func RunWorker(ctx context.Context, db *gorm.DB, interval time.Duration) {
	client := &http.Client{Timeout: 30 * time.Second}
	processDue(ctx, db, client)

	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			processDue(ctx, db, client)
		}
	}
}

func processDue(ctx context.Context, db *gorm.DB, client *http.Client) {
	now := time.Now()
	var rows []models.Instantiate
	if err := db.Where("webhook_delivered_at IS NULL AND webhook_notify_at <= ?", now).
		Order("id ASC").
		Limit(50).
		Find(&rows).Error; err != nil {
		log.Printf("webhook worker: query: %v", err)
		return
	}

	for i := range rows {
		if err := deliverOne(ctx, db, client, &rows[i]); err != nil {
			log.Printf("webhook worker: id=%d: %v", rows[i].ID, err)
		}
	}
}

func deliverOne(ctx context.Context, db *gorm.DB, client *http.Client, row *models.Instantiate) error {
	body, err := json.Marshal(CallbackPayload{
		ContractUID: row.ContractUID,
		ConsumerURL: row.ConsumerURL,
		AdminURL:    row.AdminURL,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, row.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", row.WebhookAuthorization)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("webhook worker: id=%d: callback returned status %d", row.ID, resp.StatusCode)
		return &httpStatusError{code: resp.StatusCode}
	}

	now := time.Now()
	res := db.Model(&models.Instantiate{}).
		Where("id = ? AND webhook_delivered_at IS NULL", row.ID).
		Update("webhook_delivered_at", now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil
	}
	return nil
}

type httpStatusError struct {
	code int
}

func (e *httpStatusError) Error() string {
	return "webhook returned non-success status"
}
