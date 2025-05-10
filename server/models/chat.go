package models

import (
	"time"

	"gorm.io/gorm"
)


type ChatMessage struct {
	gorm.Model
	Type       string    `json:"type" gorm:"type:varchar(20)"` // chat, price, ai_response
	SenderRole string    `json:"senderRole" gorm:"type:varchar(20)"`
	Content    string    `json:"content" gorm:"type:text"`
	FairPrice  string    `json:"fair_price,omitempty" gorm:"type:varchar(20)"` // Only for AI responses
	Timestamp  time.Time `json:"timestamp" gorm:"default:CURRENT_TIMESTAMP"`
}

type PriceHistory struct {
	gorm.Model
	Date  time.Time `json:"date"`
	Price float64   `json:"price"`
	Item  string    `json:"item" gorm:"type:varchar(100)"`
}