package httpapi

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/oduan/ai-html-feedback-bridge/internal/storage"
)

var validInteractionID = regexp.MustCompile(`^[a-zA-Z0-9_.\-]{1,128}$`)

// NewRouter creates an http.Handler that routes requests to the appropriate handlers.
func NewRouter(store *storage.Store, maxContentSize int) http.Handler {
	h := &interactionHandler{
		store:          store,
		maxContentSize: maxContentSize,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/interactions/", func(w http.ResponseWriter, r *http.Request) {
		// Extract interaction_id from path: /interactions/{interaction_id}
		path := strings.TrimPrefix(r.URL.Path, "/interactions/")
		if path == "" || !validInteractionID.MatchString(path) {
			errorResponse(w, http.StatusBadRequest, "invalid interaction_id")
			return
		}

		switch r.Method {
		case http.MethodPost:
			h.handlePost(w, r, path)
		case http.MethodGet:
			h.handleGet(w, r, path)
		case http.MethodOptions:
			w.Header().Set("Allow", "GET, POST, OPTIONS")
			w.WriteHeader(http.StatusNoContent)
		default:
			errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	return withCORS(mux)
}

// withCORS wraps a handler with CORS headers allowing any origin.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		// Explicitly not setting Access-Control-Allow-Credentials
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
