package models

import "time"

type ChatMessage struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	SenderRole string    `json:"sender_role"`
	Content    string    `json:"content"`
	Type       string    `json:"type"`
	Role       string    `json:"role"`
	Price      float64   `json:"price"`
	SentAt     time.Time `json:"sent_at" gorm:"autoCreateTime"`
}
