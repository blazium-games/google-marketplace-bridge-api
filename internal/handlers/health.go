package handlers

import "net/http"

// Health responds with 200 OK for GET and HEAD (for load balancer / k8s probes).
func Health(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
