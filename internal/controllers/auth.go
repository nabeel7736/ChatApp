package controllers

import (
	"chatapp/internal/models"
	"chatapp/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthController struct {
	DB  *gorm.DB
	JWT *services.JWTService
}

type registerPayload struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register endpoint
func (a *AuthController) Register(c *gin.Context) {
	var payload registerPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.User
	if err := a.DB.Where("email = ? OR username = ?", payload.Email, payload.Username).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		return
	}

	hashed, err := services.HashPassword(payload.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}

	user := models.User{
		UUID:     uuid.New().String(),
		Username: payload.Username,
		Email:    payload.Email,
		Password: hashed,
	}

	if err := a.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully. Please login.",
		"user":    gin.H{"id": user.ID, "username": user.Username, "email": user.Email},
	})
}

// Login endpoint
func (a *AuthController) Login(c *gin.Context) {
	var payload loginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := a.DB.Where("email = ?", payload.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !services.CheckPasswordHash(user.Password, payload.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := a.JWT.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  gin.H{"id": user.ID, "username": user.Username, "email": user.Email, "profile_pic": user.ProfilePic},
		"token": token,
	})
}

func (a *AuthController) UpdateProfile(c *gin.Context) {
	claims, _ := c.Get("claims")
	userID := uint(claims.(jwt.MapClaims)["user_id"].(float64))

	var payload struct {
		ProfilePic string `json:"profile_pic" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.DB.Model(&models.User{}).Where("id = ?", userID).Update("profile_pic", payload.ProfilePic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated", "profile_pic": payload.ProfilePic})
}
