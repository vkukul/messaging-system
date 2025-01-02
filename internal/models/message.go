package models

import (
	"time"
)

type Message struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	To        string    `json:"to" gorm:"not null"`
	Content   string    `json:"content" gorm:"not null;size:160"`
	Sent      bool      `json:"sent" gorm:"default:false"`
	SentAt    time.Time `json:"sent_at,omitempty"`
	MessageID string    `json:"message_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
