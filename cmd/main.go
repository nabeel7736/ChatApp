package main

import (
	"chatapp/internal/db"
	"chatapp/internal/models"
	"chatapp/internal/routes"
	"chatapp/internal/websocket"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	// 1. Connect DB
	db.ConnectDB()

	// 2. Auto migrate
	// db.DB.AutoMigrate(&models.Message{})

	// migrate
	if err := Migrate(db.DB); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	// 3. WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// 4. Start Gin Router
	r := gin.Default()
	routes.RegisterRoutes(r, hub)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("server running on %s", addr)
	r.Run(addr)
}
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.Message{})
}
