package models

import (
	//"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
	"gorm.io/gorm"
)

type Consumer struct {
	BaseModel
	Email    string `gorm:"unique;not null"`
	Password string
	Name     string
	Address  string
	Phone    string
}

func UpsertConsumer(email, name string, db *gorm.DB) string {
	var c Consumer
	db.FirstOrCreate(&c, "email = ?", email)
	c.Name = name
	db.Save(&c)
	return c.ID.String()
}
