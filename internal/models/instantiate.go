package models

import (
	"time"

	"gorm.io/gorm"
)

// Instantiate is a persisted instantiation request; the webhook is delivered after WebhookNotifyAt.
type Instantiate struct {
	ID                   uint `gorm:"primaryKey"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Email                string
	ContractUID          string `gorm:"index;not null"`
	Company              string `gorm:"not null"`
	Project              string `gorm:"not null"`
	WebhookURL           string `gorm:"not null"`
	WebhookAuthorization string `gorm:"not null"` // value sent as Authorization when calling webhook_url
	// ConsumerURL / AdminURL are random placeholders included in the outbound webhook JSON.
	ConsumerURL string `gorm:"not null"`
	AdminURL    string `gorm:"not null"`
	WebhookNotifyAt      time.Time `gorm:"index;not null"`
	WebhookDeliveredAt   *time.Time
}

// AutoMigrate creates or updates the schema.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Instantiate{})
}
