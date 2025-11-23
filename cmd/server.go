package main

import (
	"chatapp/internal/config"
	"chatapp/internal/database"
	"chatapp/internal/routes"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// load env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file loaded, relying on environment variables")
	}

	cfg := config.LoadConfig()

	// connect db and migrate
	db := database.Connect(cfg)
	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to migrate db: %v", err)
	}

	// create router
	r := routes.SetupRouter(db, cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
