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
	api.HandleFunc("/vaults/{vault}", createVault).Methods("POST")
	api.HandleFunc("/vaults/{vault}/files/{filePath:.+}", downloadFile).Methods("GET")
	api.HandleFunc("/vaults/{vault}/files", listFiles).Methods("GET")
	api.HandleFunc("/vaults/{vault}/files/{filePath:.+}", uploadFile).Methods("POST")
	api.HandleFunc("/vaults/{vault}/files/{filePath:.+}", deleteFile).Methods("DELETE")

	// Authenticated user info
	api.HandleFunc("/me", handleMe).Methods("GET")

	log.Println("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}

func extractAPIKey(r *http.Request) string {
	h := r.Header.Get("Authorization")
	return strings.TrimPrefix(h, "Bearer ")
}
