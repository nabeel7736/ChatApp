package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	RoomID    uint           `gorm:"index" json:"room_id"`
	SenderID  uint           `gorm:"index" json:"sender_id"`
	Content   string         `gorm:"type:text" json:"content"`
	MediaType string         `gorm:"default:'text'" json:"media_type"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
