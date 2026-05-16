package httpapi

import (
	"encoding/json"
	"net/http"
)

// errorResponse writes a JSON error response with the given status code and message.
// It does not leak internal paths or stack traces.
func errorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":    false,
		"error": message,
	})
}

// successResponse writes a 200 OK JSON success response.
func successResponse(w http.ResponseWriter, interactionID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":             true,
		"interaction_id": interactionID,
	})
}
