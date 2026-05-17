package httpapi

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/oisin/ai-html-feedback-bridge/internal/storage"
)

// maxContentSize is the maximum allowed request body size in bytes.
type interactionHandler struct {
	store          *storage.Store
	maxContentSize int
}

func (h *interactionHandler) handlePost(w http.ResponseWriter, r *http.Request, interactionID string) {
	// Read body with size limit (streaming, not buffering entire body first)
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain; charset=utf-8"
	}

	limitedReader := http.MaxBytesReader(w, r.Body, int64(h.maxContentSize))
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		// MaxBytesReader returns an error that contains "http: request body too large"
		errorResponse(w, http.StatusRequestEntityTooLarge, "content too large")
		return
	}

	now := time.Now().UTC()

	interaction := &storage.Interaction{
		InteractionID: interactionID,
		ContentType:   contentType,
		Content:       string(content),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := h.store.SaveInteraction(interaction); err != nil {
		log.Printf("save error: %v", err)
		errorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	successResponse(w, interactionID)
}

func (h *interactionHandler) handleGet(w http.ResponseWriter, r *http.Request, interactionID string) {
	interaction, err := h.store.GetInteraction(interactionID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			errorResponse(w, http.StatusNotFound, "interaction not found")
			return
		}
		log.Printf("get error: %v", err)
		errorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", interaction.ContentType)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(interaction.Content))
}
