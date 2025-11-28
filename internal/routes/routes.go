package routes

import (
	"chatapp/internal/config"
	"chatapp/internal/controllers"
	middleware "chatapp/internal/middlewares"
	"chatapp/internal/services"
	"chatapp/internal/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	jwtSvc := services.NewJWTService(cfg.JWTSecret)
	authCtrl := &controllers.AuthController{DB: db, JWT: jwtSvc}
	chatCtrl := &controllers.ChatController{DB: db}
	// Public
	public := r.Group("/api")
	{
		public.POST("/register", authCtrl.Register)
		public.POST("/login", authCtrl.Login)
		public.GET("/rooms", chatCtrl.ListRooms)
	}

	// Protected
	protected := r.Group("/api")
	protected.Use(middleware.JWTAuth(jwtSvc))
	{
		protected.POST("/rooms", chatCtrl.CreateRoom)
		// websocket endpoint (requires Authorization header or prior auth)
		protected.GET("/ws", func(c *gin.Context) {
			websocket.ServeWS(c, db)
		})
	}

	// health
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/login", func(c *gin.Context) { c.HTML(200, "login.html", nil) })
	r.GET("/register", func(c *gin.Context) { c.HTML(200, "register.html", nil) })
	r.GET("/dashboard", func(c *gin.Context) { c.HTML(200, "dashboard.html", nil) })
	r.GET("/chat", func(c *gin.Context) { c.HTML(200, "chat.html", nil) })

	r.GET("/", func(c *gin.Context) { c.Redirect(302, "/login") })

	return r
}
