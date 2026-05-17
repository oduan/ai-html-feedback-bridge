package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSHeadersOnPOST(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)

	req := httptest.NewRequest(http.MethodPost, "/interactions/cors-test", strings.NewReader("test"))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
	if resp.Header.Get("Access-Control-Allow-Credentials") != "" {
		t.Errorf("should not set Access-Control-Allow-Credentials, got %q", resp.Header.Get("Access-Control-Allow-Credentials"))
	}
}

func TestCORSHeadersOnGET(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)

	req := httptest.NewRequest(http.MethodGet, "/interactions/cors-get", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestOPTIONSReturnsSuccess(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)

	req := httptest.NewRequest(http.MethodOptions, "/interactions/any-id", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", resp.StatusCode)
	}
}

func TestOPTIONSHeaders(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)

	req := httptest.NewRequest(http.MethodOptions, "/interactions/any-id", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
	allowMethods := resp.Header.Get("Access-Control-Allow-Methods")
	if !strings.Contains(allowMethods, "GET") || !strings.Contains(allowMethods, "POST") || !strings.Contains(allowMethods, "OPTIONS") {
		t.Errorf("expected Allow-Methods to include GET/POST/OPTIONS, got %q", allowMethods)
	}
	if resp.Header.Get("Access-Control-Allow-Headers") != "Content-Type" {
		t.Errorf("expected Allow-Headers: Content-Type, got %q", resp.Header.Get("Access-Control-Allow-Headers"))
	}
}

func TestOPTIONSOnRootInteractions(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)

	// OPTIONS on /interactions/ (without ID) should still return valid CORS
	req := httptest.NewRequest(http.MethodOptions, "/interactions/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}
