package consumer

import (
	"errors"

	"gorm.io/gorm"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

func GetConsumerByEmail(email string, db *gorm.DB) (*models.Consumer, error) {
	var consumer models.Consumer
	if err := db.Where("email = ?", email).First(&consumer).Error; err != nil {
		return nil, errors.New("consumer not found")
	}
	return &consumer, nil
}
