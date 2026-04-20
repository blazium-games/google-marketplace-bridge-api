package handlers

import (
	"testing"

	"google-marketplace-bridge/api/internal/apierr"
)

func TestValidateInstantiate_AllRequired(t *testing.T) {
	base := InstantiateRequest{
		Email:         "a@b.co",
		ContractUID:   "u1",
		Company:       "co",
		Project:       "pr",
		WebhookURL:    "https://example.com/h",
		Authorization: "tok",
	}
	if code := validateInstantiate(&base); code != 0 {
		t.Fatalf("got code %d", code)
	}

	tests := []struct {
		name     string
		mut      func(*InstantiateRequest)
		wantCode int
	}{
		{"email", func(r *InstantiateRequest) { r.Email = "" }, apierr.CodeMissingEmail},
		{"contract_uid", func(r *InstantiateRequest) { r.ContractUID = "" }, apierr.CodeMissingContractUID},
		{"company", func(r *InstantiateRequest) { r.Company = "" }, apierr.CodeMissingCompany},
		{"project", func(r *InstantiateRequest) { r.Project = "" }, apierr.CodeMissingProject},
		{"webhook_url", func(r *InstantiateRequest) { r.WebhookURL = "" }, apierr.CodeMissingWebhookURL},
		{"authorization", func(r *InstantiateRequest) { r.Authorization = "" }, apierr.CodeMissingAuthorization},
		{"whitespace email", func(r *InstantiateRequest) { r.Email = "   " }, apierr.CodeMissingEmail},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := base
			tc.mut(&req)
			if got := validateInstantiate(&req); got != tc.wantCode {
				t.Fatalf("got %d want %d", got, tc.wantCode)
			}
		})
	}
}
