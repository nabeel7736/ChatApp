package main

import (
	"chatapp/internal/db"
	"chatapp/internal/models"
	"chatapp/internal/routes"
	websocket "chatapp/internal/websoaket"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Connect DB
	db.ConnectDB()

	// 2. Auto migrate
	db.DB.AutoMigrate(&models.Message{})

	// 3. WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// 4. Start Gin Router
	r := gin.Default()
	routes.RegisterRoutes(r, hub)

	r.Run(":8080")
}
