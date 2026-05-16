package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIntegrationSaveAndRead(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Save JSON content
	body := `{"feedback":"需要补充项目排期","approved":true}`
	postResp, err := http.Post(server.URL+"/interactions/int-demo-001", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	if postResp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", postResp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(postResp.Body).Decode(&result)
	postResp.Body.Close()
	if result["ok"] != true || result["interaction_id"] != "int-demo-001" {
		t.Errorf("unexpected POST response: %v", result)
	}

	// Read content back
	getResp, err := http.Get(server.URL + "/interactions/int-demo-001")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if getResp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", getResp.StatusCode)
	}
	if getResp.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", getResp.Header.Get("Content-Type"))
	}
	getBody, _ := io.ReadAll(getResp.Body)
	getResp.Body.Close()
	if string(getBody) != body {
		t.Errorf("body mismatch: expected %q, got %q", body, string(getBody))
	}
}

func TestIntegrationOverwrite(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)
	server := httptest.NewServer(mux)
	defer server.Close()

	// First save
	http.Post(server.URL+"/interactions/overwrite-int", "text/plain", strings.NewReader("first"))

	// Overwrite
	http.Post(server.URL+"/interactions/overwrite-int", "text/html", strings.NewReader("<p>second</p>"))

	// Read back - should be second content with text/html
	getResp, _ := http.Get(server.URL + "/interactions/overwrite-int")
	getBody, _ := io.ReadAll(getResp.Body)
	getResp.Body.Close()

	if string(getBody) != "<p>second</p>" {
		t.Errorf("expected overwritten content, got %q", string(getBody))
	}
	if getResp.Header.Get("Content-Type") != "text/html" {
		t.Errorf("expected Content-Type text/html, got %q", getResp.Header.Get("Content-Type"))
	}
}

func TestIntegration404(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)
	server := httptest.NewServer(mux)
	defer server.Close()

	getResp, _ := http.Get(server.URL + "/interactions/nonexistent-int")
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", getResp.StatusCode)
	}
}

func TestIntegration400(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)
	server := httptest.NewServer(mux)
	defer server.Close()

	postResp, _ := http.Post(server.URL+"/interactions/", "text/plain", strings.NewReader("test"))
	if postResp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", postResp.StatusCode)
	}
}

func TestIntegration413(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 100) // Only 100 bytes max
	server := httptest.NewServer(mux)
	defer server.Close()

	largeBody := strings.Repeat("x", 101)
	postResp, _ := http.Post(server.URL+"/interactions/large-int", "text/plain", strings.NewReader(largeBody))
	if postResp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", postResp.StatusCode)
	}
}

func TestIntegrationCORS(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test CORS on GET
	getResp, _ := http.Get(server.URL + "/interactions/any")
	if getResp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("GET: expected Access-Control-Allow-Origin: *")
	}

	// Test CORS on POST
	postResp, _ := http.Post(server.URL+"/interactions/any2", "text/plain", strings.NewReader("test"))
	if postResp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("POST: expected Access-Control-Allow-Origin: *")
	}
}

func TestIntegrationFullRoundTrip(t *testing.T) {
	store, cleanup := setupTest(t)
	defer cleanup()

	mux := NewRouter(store, 1048576)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Save HTML content
	htmlContent := "<html><body><form><input name='feedback'/></form></body></html>"
	postResp, _ := http.Post(server.URL+"/interactions/html-int", "text/html", strings.NewReader(htmlContent))
	if postResp.StatusCode != http.StatusOK {
		t.Fatalf("POST failed: %d", postResp.StatusCode)
	}

	// Read back
	getResp, _ := http.Get(server.URL + "/interactions/html-int")
	getBody, _ := io.ReadAll(getResp.Body)
	getResp.Body.Close()

	if string(getBody) != htmlContent {
		t.Errorf("HTML content not preserved:\nexpected: %q\ngot:      %q", htmlContent, string(getBody))
	}
	if getResp.Header.Get("Content-Type") != "text/html" {
		t.Errorf("expected Content-Type text/html, got %q", getResp.Header.Get("Content-Type"))
	}
}
