package models

import (
	"time"
)

type Token struct {
	BaseModel
	UserID       string `gorm:"not null"`
	Role         string
	AccessToken  string
	RefreshToken string // Store your own issued refresh token
	Expiry       time.Time
}
