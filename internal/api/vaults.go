package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
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

func createVault(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	apiKey := extractAPIKey(r)
	user, err := database.AuthenticateByKey(apiKey)
	if err != nil {
		http.Error(w, "Invalid API Key", http.StatusUnauthorized)
		return
	}

	vaultName := mux.Vars(r)["vault"]

	existing, err := database.FetchVaultByName(user.Id, vaultName)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if existing != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	newVault, err := database.CreateVault(user.Id, vaultName)
	if err != nil {
		http.Error(w, "Failed to create vault", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newVault); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// func downloadFile(w http.ResponseWriter, r *http.Request) {
// 	apiKey := extractAPIKey(r)
// 	user, err := database.AuthenticateByKey(apiKey)
// 	if err != nil {
// 		http.Error(w, "Invalid API Key", http.StatusUnauthorized)
// 		return
// 	}

// 	vaultName := mux.Vars(r)["vault"]

// 	existing, err := database.FetchVaultByName(user.Id, vaultName)
// 	if err != nil {
// 		http.Error(w, "internal error", http.StatusInternalServerError)
// 		return
// 	}

// 	if existing != nil {
// 		w.WriteHeader(http.StatusConflict)
// 		return
// 	}
// }

func uploadFile(w http.ResponseWriter, r *http.Request) {
	apiKey := extractAPIKey(r)
	user, err := database.AuthenticateByKey(apiKey)
	if err != nil {
		http.Error(w, "Invalid API Key", http.StatusUnauthorized)
		return
	}

	vaultName := mux.Vars(r)["vault"]
	filePath := mux.Vars(r)["filePath"]

	// Prevent path traversal
	cleanPath := filepath.FromSlash(filepath.Clean(filePath))
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	existing, err := database.FetchVaultByName(user.Id, vaultName)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if existing == nil {
		http.Error(w, "vault not found", http.StatusNotFound)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	dstPath := filepath.Join(
		"./uploads",
		strconv.Itoa(user.Id),
		vaultName,
		cleanPath,
	)

	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "File uploaded: %s\n", cleanPath)
}
