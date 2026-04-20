package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth_OK(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		method := method
		t.Run(method, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/", nil)
			Health(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("got %d", rec.Code)
			}
		})
	}
}

func TestHealth_MethodNotAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	Health(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestServeMux_HealthPatterns(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", Health)
	mux.HandleFunc("HEAD /{$}", Health)
	mux.HandleFunc("GET /health", Health)
	mux.HandleFunc("HEAD /health", Health)

	for _, path := range []string{"/", "/health"} {
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(method, path, nil)
			mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("%s %s: got %d", method, path, rec.Code)
			}
		}
	}
}
