package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret string
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: secret}
}

func (s *JWTService) GenerateToken(userID uint, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *JWTService) ParseToken(tokenStr string) (jwt.MapClaims, error) {
	tkn, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidId
}
