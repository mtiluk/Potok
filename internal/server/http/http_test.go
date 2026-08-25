package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/michaeltukdev/Potok/internal/server/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	ctx := context.Background()

	s, err := store.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })

	if err := s.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return s
}

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

func TestRegisterEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		seed       string
		input      string
		wantStatus int
	}{
		{
			name:       "valid",
			input:      `{"email":"a@example.com","password":"hunter2"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "malformed json",
			input:      `{"email":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing email",
			input:      `{"password":"hunter2"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "duplicate email",
			seed:       `{"email":"a@example.com","password":"hunter2"}`,
			input:      `{"email":"a@example.com","password":"different"}`,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "duplicate email different case",
			seed:       `{"email":"a@example.com","password":"hunter2"}`,
			input:      `{"email":"A@Example.com","password":"hunter2"}`,
			wantStatus: http.StatusConflict,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHandler(newTestStore(t))

			post := func(body string) *httptest.ResponseRecorder {
				req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()
				handler.Register(rec, req)
				return rec
			}

			if test.seed != "" {
				if rec := post(test.seed); rec.Code != http.StatusCreated {
					t.Fatalf("seed request failed: %d (body: %s)", rec.Code, rec.Body.String())
				}
			}

			response := post(test.input)
			if response.Code != test.wantStatus {
				t.Errorf("expected status code %d, got %d (body: %s)",
					test.wantStatus, response.Code, response.Body.String())
			}
		})
	}
}
