package api

import (
	"encoding/json"
	"net/http"

	"github.com/michaeltukdev/Potok/internal/database"
)

func listVaults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	apiKey := extractAPIKey(r)
	user, err := database.AuthenticateByKey(apiKey)
	if err != nil {
		http.Error(w, "Invalid API Key", http.StatusUnauthorized)
		return
	}

	vaults, err := database.FetchUserVaults(user.Id)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(vaults)
}
