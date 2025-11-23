package controllers

import (
	"chatapp/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ChatController struct {
	DB *gorm.DB
}

// create room
type createRoomPayload struct {
	Name string `json:"name" binding:"required,min=1"`
}

func (cc *ChatController) CreateRoom(c *gin.Context) {
	var p createRoomPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	room := models.Room{Name: p.Name}
	if err := cc.DB.Create(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create room"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"room": room})
}

func (cc *ChatController) ListRooms(c *gin.Context) {
	var rooms []models.Room
	cc.DB.Find(&rooms)
	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}
