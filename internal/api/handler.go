package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/michaeltukdev/Potok/internal/middleware"
)

func StartServer() {
	r := mux.NewRouter()

	api := r.PathPrefix("/").Subrouter()
	api.Use(middleware.ApiMiddleware)

	// Vaults
	api.HandleFunc("/vaults", listVaults).Methods("GET")

	// Vaults
	// api.HandleFunc("/users/{user}/vaults", handleVaults)
	// api.HandleFunc("/users/{user}/vaults/{vault}", handlePostVault).Methods("POST")
	// api.HandleFunc("/users/{user}/vaults/{vault}", handleDeleteVault).Methods("DELETE")

	// Files
	// api.HandleFunc("/users/{user}/vaults/{vault}/files", handleListVaultFiles).Methods("GET")
	// api.HandleFunc("/users/{user}/vaults/{vault}/files/{filepath:.*}", handleDownloadFile).Methods("GET")
	// api.HandleFunc("/users/{user}/vaults/{vault}/files/{filepath:.*}", handleUploadFile).Methods("POST")

	// Authenticated user info
	// api.HandleFunc("/me", handleMe).Methods("GET")

	log.Println("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}

func extractAPIKey(r *http.Request) string {
	h := r.Header.Get("Authorization")
	return strings.TrimPrefix(h, "Bearer ")
}

// // handleVaults returns all vaults for the authenticated user.
// func handleVaults(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	vaults, err := database.FetchUserVaults(r.Header.Get("Authorization"))
// 	if err != nil {
// 		http.Error(w, "Unauthorized or error fetching vaults", http.StatusUnauthorized)
// 		return
// 	}

// 	if err := json.NewEncoder(w).Encode(vaults); err != nil {
// 		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
// 		return
// 	}
// }

// // handlePostVaults creates a vault with the data passed in by the user.
// func handlePostVault(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	user, err := database.FindByAPIKey(r.Header.Get("Authorization"))
// 	if err != nil {
// 		http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
// 		return
// 	}

// 	vars := mux.Vars(r)
// 	urlUser := vars["user"]
// 	urlVault := vars["vault"]

// 	if urlUser != user.Username {
// 		http.Error(w, "Unauthorized: user mismatch", http.StatusUnauthorized)
// 		return
// 	}

// 	if _, err := database.FetchUserVaultByName(user.Api_key, urlVault); err == nil {
// 		http.Error(w, "Vault already exists", http.StatusConflict)
// 		return
// 	}

// 	newVault, err := database.CreateVault(user.Id, urlVault)
// 	if err != nil {
// 		http.Error(w, "Failed to create vault", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusCreated)
// 	if err := json.NewEncoder(w).Encode(newVault); err != nil {
// 		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
// 		return
// 	}
// }

// // func handleDeleteVault(w http.ResponseWriter, r *http.Request) {}
// // func handleUploadVault(w http.ResponseWriter, r *http.Request)   {}
// // func handleDownloadVault(w http.ResponseWriter, r *http.Request) {}

// // handleDownloadFile takes in data from the client, and returns a specific file from the vault.
// func handleDownloadFile(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/octet-stream")

// 	user, err := database.FindByAPIKey(r.Header.Get("Authorization"))
// 	if err != nil {
// 		http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
// 		return
// 	}

// 	vars := mux.Vars(r)
// 	urlUser := vars["user"]
// 	urlVault := vars["vault"]
// 	filepathInVault := vars["filepath"]

// 	if urlUser != user.Username {
// 		http.Error(w, "Unauthorized: user mismatch", http.StatusUnauthorized)
// 		return
// 	}

// 	if _, err := database.FetchUserVaultByName(user.Api_key, urlVault); err == nil {
// 		http.Error(w, "Vault already exists", http.StatusConflict)
// 		return
// 	}

// 	fullPath := path.Join("../../data", user.Username, urlVault, filepathInVault)

// 	f, err := os.Open(fullPath)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("File not found: %s", fullPath), http.StatusNotFound)
// 		return
// 	}
// 	defer f.Close()

// 	w.WriteHeader(http.StatusOK)
// 	if _, err := io.Copy(w, f); err != nil {
// 		http.Error(w, "Failed to send file", http.StatusInternalServerError)
// 		return
// 	}
// }

// // handleUploadFile handles uploading a file to a specific vault for the authenticated user.
// func handleUploadFile(w http.ResponseWriter, r *http.Request) {
// 	user, err := database.FindByAPIKey(r.Header.Get("Authorization"))
// 	if err != nil {
// 		http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
// 		return
// 	}

// 	vars := mux.Vars(r)
// 	urlUser := vars["user"]
// 	vault := vars["vault"]
// 	filepathInVault := vars["filepath"]

// 	if urlUser != user.Username {
// 		http.Error(w, "Unauthorized: user mismatch", http.StatusUnauthorized)
// 		return
// 	}

// 	fetchedVault, err := database.FetchUserVaultByName(user.Api_key, vault)
// 	if err != nil {
// 		http.Error(w, "Vault not found or unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	fullPath := path.Join("../../data", user.Username, fetchedVault.Name, filepathInVault)

// 	if err := os.MkdirAll(path.Dir(fullPath), 0700); err != nil {
// 		http.Error(w, "Failed to create directories", http.StatusInternalServerError)
// 		return
// 	}

// 	f, err := os.Create(fullPath)
// 	if err != nil {
// 		http.Error(w, "Failed to save file", http.StatusInternalServerError)
// 		return
// 	}
// 	defer f.Close()
// 	if _, err := io.Copy(f, r.Body); err != nil {
// 		http.Error(w, "Failed to write file", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusCreated)
// }

// // handleMe returns information about the authenticated user.
// func handleMe(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	user, err := database.FindByAPIKey(r.Header.Get("Authorization"))
// 	if err != nil {
// 		http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
// 		return
// 	}

// 	resp := struct {
// 		Username string `json:"username"`
// 		ID       int    `json:"id"`
// 	}{
// 		Username: user.Username,
// 		ID:       user.Id,
// 	}

// 	if err := json.NewEncoder(w).Encode(resp); err != nil {
// 		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
// 		return
// 	}
// }

// // handleListVaultFiles returns a JSON array of all file paths in the specified vault for the authenticated user.
// func handleListVaultFiles(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	user, err := database.FindByAPIKey(r.Header.Get("Authorization"))
// 	if err != nil {
// 		http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
// 		return
// 	}

// 	vars := mux.Vars(r)
// 	urlUser := vars["user"]
// 	urlVault := vars["vault"]

// 	if urlUser != user.Username {
// 		http.Error(w, "Unauthorized: user mismatch", http.StatusUnauthorized)
// 		return
// 	}

// 	if _, err := database.FetchUserVaultByName(user.Api_key, urlVault); err == nil {
// 		http.Error(w, "Vault already exists", http.StatusConflict)
// 		return
// 	}

// 	vaultDir := path.Join("../../data", user.Username, urlVault)

// 	var files []string
// 	filepath.WalkDir(vaultDir, func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		if !d.IsDir() {
// 			rel, _ := filepath.Rel(vaultDir, path)
// 			files = append(files, rel)
// 		}
// 		return nil
// 	})

// 	json.NewEncoder(w).Encode(files)
// }
