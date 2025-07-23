package supplier

import (
	"errors"

	"gorm.io/gorm"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

func GetSupplierByEmail(email string, db *gorm.DB) (*models.Supplier, error) {
	var supplier models.Supplier
	if err := db.Where("email = ?", email).First(&supplier).Error; err != nil {
		return nil, errors.New("supplier not found")
	}
	return &supplier, nil
}
