package controllers

import (
	"chatapp/internal/db"
	"chatapp/internal/models"

	"github.com/gin-gonic/gin"
)

func GetMessages(c *gin.Context) {
	var messages []models.Message
	db.DB.Order("created_at DESC").Find(&messages)
	c.JSON(200, gin.H{"messages": messages})
}

func SaveMessage(sender string, content string) {
	message := models.Message{
		Sender:  sender,
		Content: content,
	}
	db.DB.Create(&message)
}
