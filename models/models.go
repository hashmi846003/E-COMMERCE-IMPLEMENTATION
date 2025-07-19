package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel for common fields
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Admin model
type Admin struct {
	BaseModel
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Name     string
}

func (a *Admin) GetID() uuid.UUID    { return a.ID }
func (a *Admin) GetEmail() string    { return a.Email }
func (a *Admin) GetPassword() string { return a.Password }
func (a *Admin) GetRole() string     { return "admin" }

// Consumer model
type Consumer struct {
	BaseModel
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Name     string
	Address  string
	Phone    string
}

func (c *Consumer) GetID() uuid.UUID    { return c.ID }
func (c *Consumer) GetEmail() string    { return c.Email }
func (c *Consumer) GetPassword() string { return c.Password }
func (c *Consumer) GetRole() string     { return "consumer" }

// Supplier model
type Supplier struct {
	BaseModel
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Name     string
	Company  string
}

func (s *Supplier) GetID() uuid.UUID    { return s.ID }
func (s *Supplier) GetEmail() string    { return s.Email }
func (s *Supplier) GetPassword() string { return s.Password }
func (s *Supplier) GetRole() string     { return "supplier" }

// Token model
type Token struct {
	BaseModel
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	Role         string    `gorm:"type:varchar(20);not null"`
	AccessToken  string    `gorm:"type:text;not null"`
	RefreshToken string    `gorm:"type:text"`
	Expiry       time.Time
}
