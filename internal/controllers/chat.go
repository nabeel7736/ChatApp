package controllers

import (
	"chatapp/internal/models"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type ChatController struct {
	DB *gorm.DB
}

// create room
type createRoomPayload struct {
	Name string `json:"name" binding:"required,min=1"`
}

type joinRoomPayload struct {
	Code string `json:"code" binding:"required,len=6"`
}

func (cc *ChatController) CreateRoom(c *gin.Context) {
	var p createRoomPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claimsMap := claims.(jwt.MapClaims)
	userID := uint(claimsMap["user_id"].(float64))

	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	room := models.Room{
		Name:   p.Name,
		Code:   code,
		UserID: userID,
	}
	if err := cc.DB.Create(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create room"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"room": room})
}

func (cc *ChatController) ListRooms(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	claimsMap := claims.(jwt.MapClaims)
	userID := uint(claimsMap["user_id"].(float64))

	var rooms []models.Room
	cc.DB.Where("user_id = ?", userID).Find(&rooms)
	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

func (cc *ChatController) JoinRoom(c *gin.Context) {
	var p joinRoomPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code format"})
		return
	}

	var room models.Room
	if err := cc.DB.Where("code = ?", p.Code).First(&room).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"room": room})
}
