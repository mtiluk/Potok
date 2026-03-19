package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/michaeltukdev/Potok/internal/middleware"
)

func StartServer(port string) {
	r := mux.NewRouter()

	api := r.PathPrefix("/").Subrouter()
	api.Use(middleware.ApiMiddleware)

	api.HandleFunc("/vaults", listVaults).Methods("GET")
	api.HandleFunc("/vaults/{vault}", createVault).Methods("POST")
	api.HandleFunc("/vaults/{vault}/files/{filePath:.+}", downloadFile).Methods("GET")
	api.HandleFunc("/vaults/{vault}/files", listFiles).Methods("GET")
	api.HandleFunc("/vaults/{vault}/files/{filePath:.+}", uploadFile).Methods("POST")
	api.HandleFunc("/vaults/{vault}/files/{filePath:.+}", deleteFile).Methods("DELETE")

	api.HandleFunc("/vaults/{vault}/manifest", getManifest).Methods("GET")

	api.HandleFunc("/me", handleMe).Methods("GET")

	log.Printf("Starting server on :%s\n", port)
	http.ListenAndServe(":"+port, r)
}

func extractAPIKey(r *http.Request) string {
	h := r.Header.Get("Authorization")
	return strings.TrimPrefix(h, "Bearer ")
}
