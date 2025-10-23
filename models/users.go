package models

import (
	"time"

	"gorm.io/gorm"
)

// Users is an in-memory map to store users (username -> User)
var Users = make(map[string]User)

// Message storage
var Messages []Message
var NextMsgID = 1

type Message struct {
	ID        int       `json:"id"`
	User      string    `json:"user"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// User represents a registered user
type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Email        string         `json:"email" gorm:"unique;not null"`
	Username     string         `json:"username" gorm:"unique;not null"`
	Password     string         `json:"password" gorm:"not null"` // store hashed password
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"` // soft delete
	RefreshToken string         `json:"refresh_token,omitempty" gorm:"size:500"`
}
