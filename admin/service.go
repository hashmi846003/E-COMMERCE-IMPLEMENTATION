package admin

import (
	"errors"

	"gorm.io/gorm"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models" 
)

// GetAdminByEmail looks up an admin by email
func GetAdminByEmail(email string, db *gorm.DB) (*models.Admin, error) { 
	var admin models.Admin

	if err := db.Where("email = ?", email).First(&admin).Error; err != nil {
		return nil, errors.New("admin not found")
	}
	return &admin, nil
}
