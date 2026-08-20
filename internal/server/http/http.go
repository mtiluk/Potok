package http

import (
	"net/http"

	"github.com/michaeltukdev/Potok/internal/server/store"
)

type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	http.ResponseWriter.WriteHeader(w, http.StatusOK)
	http.ResponseWriter.Write(w, []byte("OK"))
}
