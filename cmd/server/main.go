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
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./potok.db"
	}

	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database running...")

	if err := database.RunMigrations(db, migrationsPath); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	fmt.Println("Migrations completed...")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting HTTP server on :%s\n", port)
	api.StartServer(port)
}
