package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/michaeltukdev/Potok/internal/api"
	"github.com/michaeltukdev/Potok/internal/database"
)

func main() {
	db, err := database.InitDB("potok.db")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database running...")

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}

	if err := database.RunMigrations(db, migrationsPath); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	fmt.Println("Migrations completed...")

	fmt.Println("Starting HTTP server on :8080")
	api.StartServer()
}
