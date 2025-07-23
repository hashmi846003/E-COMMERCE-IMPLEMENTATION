package models

import "gorm.io/gorm"
//import "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
type Admin struct {
	BaseModel
	Email string `gorm:"unique"`
	Name  string
}

func UpsertAdmin(email, name string, db *gorm.DB) string {
	var a Admin
	db.FirstOrCreate(&a, "email = ?", email)
	a.Name = name
	db.Save(&a)
	return a.ID.String()
}
