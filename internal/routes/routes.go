package routes

import (
	"chatapp/internal/controllers"
	"chatapp/internal/websocket"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, hub *websocket.Hub) {
	r.GET("/", func(ctx *gin.Context) {
		ctx.File("templates/index.html")
	})
	r.GET("/messages", controllers.GetMessages)

	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(hub, c.Writer, c.Request)
	})
}
