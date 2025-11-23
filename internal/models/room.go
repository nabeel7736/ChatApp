package models

import (
	"time"

	"gorm.io/gorm"
)

type Room struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"uniqueIndex;not null" json:"name"`
	IsPrivate bool           `json:"is_private"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
