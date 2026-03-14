package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

func downloadFile(w http.ResponseWriter, r *http.Request) {
	apiKey := extractAPIKey(r)
	user, err := database.AuthenticateByKey(apiKey)
	if err != nil {
		http.Error(w, "Invalid API Key", http.StatusUnauthorized)
		return
	}

	vaultName := mux.Vars(r)["vault"]
	filePath := mux.Vars(r)["filePath"]

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

	absPath := filepath.Join(
		"./uploads",
		strconv.Itoa(user.Id),
		vaultName,
		cleanPath,
	)

	file, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, file)
}

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

func listFiles(w http.ResponseWriter, r *http.Request) {
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

	if existing == nil {
		http.Error(w, "vault not found", http.StatusNotFound)
		return
	}

	root := filepath.Join("./uploads", strconv.Itoa(user.Id), vaultName)

	var files []string
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			skip := map[string]bool{
				".potok": true,
				".git":   true,
			}
			if skip[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		files = append(files, filepath.ToSlash(rel))
		return nil
	})

	if err != nil {
		if os.IsNotExist(err) {
			json.NewEncoder(w).Encode([]string{})
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if files == nil {
		files = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vault := vars["vault"]
	filePath := vars["filePath"]

	apiKey := extractAPIKey(r)
	user, err := database.AuthenticateByKey(apiKey)
	if err != nil {
		http.Error(w, "Invalid API Key", http.StatusUnauthorized)
		return
	}

	fullPath := filepath.Join("uploads", fmt.Sprintf("%d", user.Id), vault, filePath)

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	safeBase, _ := filepath.Abs(filepath.Join("uploads", fmt.Sprintf("%d", user.Id), vault))
	if !strings.HasPrefix(absPath, safeBase) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	if err := os.Remove(fullPath); err != nil {
		http.Error(w, "failed to delete file", http.StatusInternalServerError)
		return
	}

	dir := filepath.Dir(fullPath)
	for dir != safeBase {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			break
		}
		os.Remove(dir)
		dir = filepath.Dir(dir)
	}

	w.WriteHeader(http.StatusNoContent)
}

func getManifest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vault := vars["vault"]

	apiKey := extractAPIKey(r)
	user, err := database.AuthenticateByKey(apiKey)
	if err != nil {
		http.Error(w, "Invalid API Key", http.StatusUnauthorized)
		return
	}

	vaultRoot := filepath.Join(
		"uploads",
		fmt.Sprintf("%d", user.Id),
		vault,
	)

	if _, err := os.Stat(vaultRoot); os.IsNotExist(err) {
		http.Error(w, "vault not found", http.StatusNotFound)
		return
	}

	type FileInfo struct {
		Size    int64  `json:"size"`
		ModTime string `json:"mod_time"`
	}

	manifest := make(map[string]FileInfo)

	err = filepath.WalkDir(vaultRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".potok" {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(vaultRoot, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		info, err := d.Info()
		if err != nil {
			return nil
		}

		manifest[rel] = FileInfo{
			Size:    info.Size(),
			ModTime: info.ModTime().UTC().Format(time.RFC3339),
		}
		return nil
	})

	if err != nil {
		http.Error(w, "failed to walk vault", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(manifest)
}
