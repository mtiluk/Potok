package main

import (
	"context"
	"fmt"
	"log"
	"os"

	nethttp "net/http"

	"github.com/michaeltukdev/Potok/internal/server/config"
	httpapi "github.com/michaeltukdev/Potok/internal/server/http"
	"github.com/michaeltukdev/Potok/internal/server/store"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "LoadConfig() error: %v\n", err)
		os.Exit(1)
	}

	conn, err := store.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open() error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	handler := httpapi.NewHandler(conn)
	nethttp.HandleFunc("GET /health", handler.Health)
	log.Fatal(nethttp.ListenAndServe(":3000", nil))
}
