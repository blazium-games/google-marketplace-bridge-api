package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"google-marketplace-bridge/api/internal/apierr"
	"google-marketplace-bridge/api/internal/models"
)

type dummyResponse struct {
	Success bool   `json:"success"`
	Time    string `json:"time"` // RFC3339: when the webhook callback is scheduled (now + 5 minutes)
}

// Dummy handles POST /dummy — same request contract as /instantiate, but the callback runs after 5 minutes
// and the JSON response includes the scheduled webhook time.
func (h *Handler) Dummy(w http.ResponseWriter, r *http.Request) {
	req, ok := h.readInstantiateRequest(w, r)
	if !ok {
		return
	}

	consumerURL, adminURL := GenerateConsumerAdminURLs()
	notifyAt := time.Now().Add(5 * time.Minute)
	rec := models.Instantiate{
		Email:                strings.TrimSpace(req.Email),
		ContractUID:          strings.TrimSpace(req.ContractUID),
		Company:              strings.TrimSpace(req.Company),
		Project:              strings.TrimSpace(req.Project),
		WebhookURL:           strings.TrimSpace(req.WebhookURL),
		WebhookAuthorization: strings.TrimSpace(req.Authorization),
		ConsumerURL:          consumerURL,
		AdminURL:             adminURL,
		WebhookNotifyAt:      notifyAt,
	}

	if err := h.db.Create(&rec).Error; err != nil {
		apierr.Write(w, http.StatusInternalServerError, apierr.CodePersistFailed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dummyResponse{
		Success: true,
		Time:    notifyAt.UTC().Format(time.RFC3339),
	})
}
