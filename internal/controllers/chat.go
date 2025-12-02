// package controllers

// import (
// 	"chatapp/internal/models"
// 	"fmt"
// 	"math/rand"
// 	"net/http"
// 	"path/filepath"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang-jwt/jwt/v5"
// 	"github.com/google/uuid"
// 	"gorm.io/gorm"
// )

// type ChatController struct {
// 	DB *gorm.DB
// }

// // create room
// type createRoomPayload struct {
// 	Name string `json:"name" binding:"required,min=1"`
// }

// type joinRoomPayload struct {
// 	Code string `json:"code" binding:"required,len=6"`
// }

// func (cc *ChatController) CreateRoom(c *gin.Context) {
// 	var p createRoomPayload
// 	if err := c.ShouldBindJSON(&p); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	claims, exists := c.Get("claims")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		return
// 	}
// 	claimsMap := claims.(jwt.MapClaims)
// 	userID := uint(claimsMap["user_id"].(float64))

// 	rand.Seed(time.Now().UnixNano())
// 	code := fmt.Sprintf("%06d", rand.Intn(1000000))

// 	room := models.Room{
// 		Name:   p.Name,
// 		Code:   code,
// 		UserID: userID,
// 	}
// 	if err := cc.DB.Create(&room).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create room"})
// 		return
// 	}
// 	c.JSON(http.StatusCreated, gin.H{"room": room})
// }

// func (cc *ChatController) ListRooms(c *gin.Context) {
// 	claims, exists := c.Get("claims")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		return
// 	}
// 	claimsMap := claims.(jwt.MapClaims)
// 	userID := uint(claimsMap["user_id"].(float64))

// 	var rooms []models.Room
// 	cc.DB.Where("user_id = ?", userID).Find(&rooms)
// 	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
// }

// func (cc *ChatController) JoinRoom(c *gin.Context) {
// 	var p joinRoomPayload
// 	if err := c.ShouldBindJSON(&p); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code format"})
// 		return
// 	}

// 	var room models.Room
// 	if err := cc.DB.Where("code = ?", p.Code).First(&room).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"room": room})
// }

// func (cc *ChatController) GetMessages(c *gin.Context) {
// 	roomID := c.Query("room_id")
// 	if roomID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id is required"})
// 		return
// 	}

// 	var messages []struct {
// 		ID         uint      `json:"id"`
// 		Content    string    `json:"content"`
// 		SenderID   uint      `json:"sender_id"`
// 		Username   string    `json:"sender"`
// 		ProfilePic string    `json:"sender_pic"`
// 		CreatedAt  time.Time `json:"created_at"`
// 		MediaType  string    `json:"media_type"`
// 	}

// 	err := cc.DB.Table("messages").
// 		Select("messages.id, messages.content, messages.sender_id, users.username,users.profile_pic, messages.created_at, messages.media_type").
// 		Joins("left join users on users.id = messages.sender_id").
// 		Where("messages.room_id = ? AND messages.deleted_at IS NULL", roomID).
// 		Order("messages.created_at asc").
// 		Scan(&messages).Error

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"messages": messages})
// }

// func (cc *ChatController) DeleteMessage(c *gin.Context) {
// 	msgID := c.Param("id")
// 	claims, _ := c.Get("claims")
// 	claimsMap := claims.(jwt.MapClaims)
// 	userID := uint(claimsMap["user_id"].(float64))

// 	var msg models.Message
// 	if err := cc.DB.First(&msg, msgID).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
// 		return
// 	}

// 	if msg.SenderID != userID {
// 		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own messages"})
// 		return
// 	}

// 	// Soft delete
// 	cc.DB.Delete(&msg)
// 	c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": msg.ID})
// }

// func (cc *ChatController) UploadFile(c *gin.Context) {
// 	file, err := c.FormFile("file")
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
// 		return
// 	}

// 	ext := filepath.Ext(file.Filename)
// 	filename := uuid.New().String() + ext
// 	savePath := filepath.Join("uploads", filename)

// 	if err := c.SaveUploadedFile(file, savePath); err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
// 		return
// 	}

// 	fileURL := "/uploads/" + filename
// 	c.JSON(http.StatusOK, gin.H{"url": fileURL})
// }

package controllers

import (
	"chatapp/internal/models"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatController struct {
	DB *gorm.DB
}

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

func (cc *ChatController) GetMessages(c *gin.Context) {
	roomID := c.Query("room_id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_id is required"})
		return
	}

	var messages []struct {
		ID         uint      `json:"id"`
		Content    string    `json:"content"`
		SenderID   uint      `json:"sender_id"`
		Username   string    `json:"sender"`
		ProfilePic string    `json:"sender_pic"`
		CreatedAt  time.Time `json:"created_at"`
		MediaType  string    `json:"media_type"`
	}

	err := cc.DB.Table("messages").
		Select("messages.id, messages.content, messages.sender_id, users.username, users.profile_pic, messages.created_at, messages.media_type").
		Joins("left join users on users.id = messages.sender_id").
		Where("messages.room_id = ? AND messages.deleted_at IS NULL", roomID).
		Order("messages.created_at asc").
		Scan(&messages).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (cc *ChatController) DeleteMessage(c *gin.Context) {
	msgID := c.Param("id")
	claims, _ := c.Get("claims")
	claimsMap := claims.(jwt.MapClaims)
	userID := uint(claimsMap["user_id"].(float64))

	var msg models.Message
	if err := cc.DB.First(&msg, msgID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	if msg.SenderID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own messages"})
		return
	}

	cc.DB.Delete(&msg)
	c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": msg.ID})
}

func (cc *ChatController) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Ensure upload directory exists
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		err := os.Mkdir("uploads", 0755)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}
	}

	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext
	savePath := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	fileURL := "/uploads/" + filename
	c.JSON(http.StatusOK, gin.H{"url": fileURL})
}
