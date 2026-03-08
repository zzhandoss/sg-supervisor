package control

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUIRootHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handleUI(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !strings.Contains(response.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("unexpected content type: %s", response.Header().Get("Content-Type"))
	}
	if !strings.Contains(response.Body.String(), "School Gate Control Center") {
		t.Fatalf("expected control center shell")
	}
}

func TestUIAssetHandler(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	response := httptest.NewRecorder()

	handleUI(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if !strings.Contains(response.Header().Get("Content-Type"), "application/javascript") {
		t.Fatalf("unexpected content type: %s", response.Header().Get("Content-Type"))
	}
	if !strings.Contains(response.Body.String(), "refreshStatus") {
		t.Fatalf("expected app script")
	}
}
