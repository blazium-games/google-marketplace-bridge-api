package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"google-marketplace-bridge/api/internal/apierr"
	"google-marketplace-bridge/api/internal/config"
	"google-marketplace-bridge/api/internal/models"

	"gorm.io/gorm"
)

// Handler serves HTTP endpoints.
type Handler struct {
	cfg *config.Config
	db  *gorm.DB
}

// New returns a Handler.
func New(cfg *config.Config, db *gorm.DB) *Handler {
	return &Handler{cfg: cfg, db: db}
}

// InstantiateRequest is the JSON body for POST /instantiate.
type InstantiateRequest struct {
	Email         string `json:"email"`
	ContractUID   string `json:"contract_uid"`
	Company       string `json:"company"`
	Project       string `json:"project"`
	WebhookURL    string `json:"webhook_url"`
	Authorization string `json:"authorization"`
}

type instantiateResponse struct {
	Success bool `json:"success"`
}

// readInstantiateRequest parses POST /instantiate-style bodies (shared with /dummy).
func (h *Handler) readInstantiateRequest(w http.ResponseWriter, r *http.Request) (*InstantiateRequest, bool) {
	if r.Method != http.MethodPost {
		apierr.Write(w, http.StatusMethodNotAllowed, apierr.CodeMethodNotAllowed)
		return nil, false
	}

	if r.Header.Get(h.cfg.SecurityHeaderName) != h.cfg.SecurityHeaderValue {
		apierr.Write(w, http.StatusUnauthorized, apierr.CodeUnauthorized)
		return nil, false
	}

	var req InstantiateRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		code := apierr.CodeInvalidJSON
		if strings.Contains(err.Error(), "unknown field") {
			code = apierr.CodeJSONUnknownField
		}
		apierr.Write(w, http.StatusBadRequest, code)
		return nil, false
	}

	if code := validateInstantiate(&req); code != 0 {
		apierr.Write(w, http.StatusBadRequest, code)
		return nil, false
	}

	return &req, true
}

// Instantiate handles POST /instantiate.
func (h *Handler) Instantiate(w http.ResponseWriter, r *http.Request) {
	req, ok := h.readInstantiateRequest(w, r)
	if !ok {
		return
	}

	consumerURL, adminURL := GenerateConsumerAdminURLs()
	notifyAt := time.Now().Add(h.cfg.WebhookNotifyDelay)
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
	_ = json.NewEncoder(w).Encode(instantiateResponse{Success: true})
}

// validateInstantiate returns 0 if valid, otherwise an apierr code for the first failing field.
func validateInstantiate(req *InstantiateRequest) int {
	checks := []struct {
		val  string
		code int
	}{
		{req.Email, apierr.CodeMissingEmail},
		{req.ContractUID, apierr.CodeMissingContractUID},
		{req.Company, apierr.CodeMissingCompany},
		{req.Project, apierr.CodeMissingProject},
		{req.WebhookURL, apierr.CodeMissingWebhookURL},
		{req.Authorization, apierr.CodeMissingAuthorization},
	}
	for _, c := range checks {
		if strings.TrimSpace(c.val) == "" {
			return c.code
		}
	}
	return 0
}
