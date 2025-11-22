package models

import "time"

type Message struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
