package models

import (
	"gorm.io/gorm"
)

// ✅ No import of models (you're in models already)

type Supplier struct {
	BaseModel             // ✅ Direct access to BaseModel
	Email    string `gorm:"unique;not null"`
	Password string
	Name     string
	Company  string
}

func UpsertSupplier(email, name string, db *gorm.DB) string {
	var s Supplier
	db.FirstOrCreate(&s, "email = ?", email)
	s.Name = name
	db.Save(&s)
	return s.ID.String()
}
