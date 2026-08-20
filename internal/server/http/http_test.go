package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)

	response := httptest.NewRecorder()
	handler := NewHandler(nil)
	handler.Health(response, req)

	if response.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, response.Code)
	}

	if response.Body.String() != "OK" {
		t.Errorf("expected body %s, got %s", "OK", response.Body.String())
	}
}
