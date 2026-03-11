package api

import (
	"encoding/json"
	"net/http"

	"github.com/michaeltukdev/Potok/internal/database"
)

func handleMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	apiKey := extractAPIKey(r)
	user, err := database.AuthenticateByKey(apiKey)
	if err != nil {
		http.Error(w, "Invalid API Key", http.StatusUnauthorized)
		return
	}

	resp := struct {
		Username string `json:"username"`
		ID       int    `json:"id"`
	}{
		Username: user.Username,
		ID:       user.Id,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
