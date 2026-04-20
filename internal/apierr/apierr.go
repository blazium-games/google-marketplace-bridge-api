package apierr

import (
	"encoding/json"
	"net/http"
)

// Error response bodies use only a numeric code (see reference.md).
type body struct {
	Code int `json:"code"`
}

// Write sends a JSON body {"code": <code>} with the given HTTP status.
func Write(w http.ResponseWriter, status int, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body{Code: code})
}
