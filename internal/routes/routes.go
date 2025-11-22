package routes

import (
	"chatapp/internal/controllers"
	websocket "chatapp/internal/websoaket"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, hub *websocket.Hub) {
	r.GET("/messages", controllers.GetMessages)

	r.GET("/ws", func(c *gin.Context) {
		websocket.ServeWs(hub, c.Writer, c.Request)
	})
}
