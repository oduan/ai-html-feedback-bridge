package httpapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oduan/ai-html-feedback-bridge/internal/storage"
	_ "modernc.org/sqlite"
)

func setupTest(t *testing.T) (*storage.Store, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "httpapi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(dir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("open db failed: %v", err)
	}
	if err := storage.Migrate(db); err != nil {
		db.Close()
		os.RemoveAll(dir)
		t.Fatalf("migrate failed: %v", err)
	}
	store := &storage.Store{DB: db}
	cleanup := func() {
		db.Close()
		os.RemoveAll(dir)
	}
	return store, cleanup
}

func TestPostValidJSON(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)

	body := `{"feedback":"需要补充项目排期","approved":true}`
	req := httptest.NewRequest(http.MethodPost, "/interactions/demo-001", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result["ok"] != true {
		t.Errorf("expected ok: true, got %v", result["ok"])
	}
	if result["interaction_id"] != "demo-001" {
		t.Errorf("expected interaction_id demo-001, got %v", result["interaction_id"])
	}
}

func TestPostAndGetPreservesContentType(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)

	// POST JSON
	body := `{"key":"value"}`
	postReq := httptest.NewRequest(http.MethodPost, "/interactions/ct-test", strings.NewReader(body))
	postReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, postReq)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("POST failed: %d", w.Result().StatusCode)
	}

	// GET should return original content with original Content-Type
	getReq := httptest.NewRequest(http.MethodGet, "/interactions/ct-test", nil)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, getReq)

	resp := w2.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
	gotBody, _ := io.ReadAll(resp.Body)
	if string(gotBody) != body {
		t.Errorf("body mismatch: expected %q, got %q", body, string(gotBody))
	}
}

func TestGetNotFound(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)

	req := httptest.NewRequest(http.MethodGet, "/interactions/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["ok"] != false {
		t.Errorf("expected ok: false, got %v", result["ok"])
	}
	if result["error"] == nil {
		t.Errorf("expected error message, got nil")
	}
}

func TestInvalidInteractionID(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)

	tests := []struct {
		path string
		desc string
	}{
		{"/interactions/", "empty ID after slash"},
		{"/interactions/a%20b", "ID with space"},
		{"/interactions/%E6%B5%8B%E8%AF%95", "ID with non-ASCII"},
		{"/interactions/" + strings.Repeat("a", 129), "ID too long"},
	}

	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader("test"))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", tc.desc, resp.StatusCode)
		}
	}
}

func TestDefaultContentType(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)

	// POST without Content-Type
	body := "plain text content"
	req := httptest.NewRequest(http.MethodPost, "/interactions/no-ct", strings.NewReader(body))
	// Don't set Content-Type
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("POST failed: %d", w.Result().StatusCode)
	}

	// GET should return with default Content-Type
	getReq := httptest.NewRequest(http.MethodGet, "/interactions/no-ct", nil)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, getReq)

	resp := w2.Result()
	ct := resp.Header.Get("Content-Type")
	if ct != "text/plain; charset=utf-8" {
		t.Errorf("expected default Content-Type text/plain; charset=utf-8, got %q", ct)
	}
}

func TestBodySizeExactlyLimit(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	// Set small max content size
	maxSize := 100
	mux := NewRouter(store, maxSize)

	// Send exactly 100 bytes
	body := strings.Repeat("a", maxSize)
	req := httptest.NewRequest(http.MethodPost, "/interactions/limit-test", strings.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for exactly-limit body, got %d", resp.StatusCode)
	}

	// Verify it was saved
	getReq := httptest.NewRequest(http.MethodGet, "/interactions/limit-test", nil)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, getReq)
	getBody, _ := io.ReadAll(w2.Result().Body)
	if string(getBody) != body {
		t.Errorf("saved body mismatch")
	}
}

func TestBodySizeOverLimit(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	// Set small max content size
	maxSize := 100
	mux := NewRouter(store, maxSize)

	// Send 101 bytes (over limit)
	body := strings.Repeat("b", maxSize+1)
	req := httptest.NewRequest(http.MethodPost, "/interactions/over-test", strings.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["error"] != "content too large" {
		t.Errorf("expected error 'content too large', got %v", result["error"])
	}
}

func TestOverLimitDoesNotPolluteData(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	maxSize := 100
	mux := NewRouter(store, maxSize)

	// First save valid content
	validBody := "valid content"
	req := httptest.NewRequest(http.MethodPost, "/interactions/pollution-test", strings.NewReader(validBody))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatal("first save should succeed")
	}

	// Overlimit POST that should fail
	overBody := strings.Repeat("c", maxSize+1)
	req2 := httptest.NewRequest(http.MethodPost, "/interactions/pollution-test", strings.NewReader(overBody))
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	if w2.Result().StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 for overlimit, got %d", w2.Result().StatusCode)
	}

	// Verify original content is intact
	getReq := httptest.NewRequest(http.MethodGet, "/interactions/pollution-test", nil)
	w3 := httptest.NewRecorder()
	mux.ServeHTTP(w3, getReq)
	getBody, _ := io.ReadAll(w3.Result().Body)
	if string(getBody) != validBody {
		t.Errorf("original content should be preserved after failed overlimit POST: got %q", string(getBody))
	}
}

func TestServerErrorReturns500(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	// Close the DB to trigger a server error on subsequent operations
	store.DB.Close()

	mux := NewRouter(store, 1048576)

	req := httptest.NewRequest(http.MethodPost, "/interactions/test-001", strings.NewReader("test"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if bytes.Contains(body, []byte("stack")) || bytes.Contains(body, []byte("\\")) {
		t.Errorf("response should not leak internal details: %s", body)
	}
}
