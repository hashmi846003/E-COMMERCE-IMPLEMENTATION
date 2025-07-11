package models

import (
	"errors"
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

// Consumer model
type Consumer struct {
	BaseModel
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Name     string
	Address  string
	Phone    string
}

// Supplier model with trust score
type Supplier struct {
	BaseModel
	Email      string `gorm:"unique;not null"`
	Password   string `gorm:"not null"`
	Name       string
	Company    string
	TrustScore float64
	IsTrusted  bool `gorm:"default:true"`
}

// Product model with image storage
type Product struct {
	BaseModel
	Name        string  `gorm:"not null"`
	Description string
	Price       float64 `gorm:"not null"`
	Stock       int     `gorm:"not null"`
	Image       []byte  `gorm:"type:bytea"` // Stores image as binary data
	SupplierID  uuid.UUID
	Supplier    Supplier `gorm:"foreignKey:SupplierID"`
	IsActive    bool     `gorm:"default:true"`
}

// Review model for consumer feedback
type Review struct {
	BaseModel
	Rating      int       `gorm:"check:rating >= 1 AND rating <= 5"`
	Comment     string
	ConsumerID  uuid.UUID
	Consumer    Consumer  `gorm:"foreignKey:ConsumerID"`
	ProductID   uuid.UUID
	Product     Product   `gorm:"foreignKey:ProductID"`
	SupplierID  uuid.UUID
	Supplier    Supplier  `gorm:"foreignKey:SupplierID"`
}

// Order model
type Order struct {
	BaseModel
	ConsumerID   uuid.UUID
	Consumer     Consumer    `gorm:"foreignKey:ConsumerID"`
	TotalAmount  float64    `gorm:"not null"`
	Status       string     `gorm:"default:'pending'"`
	OrderItems   []OrderItem `gorm:"foreignKey:OrderID"`
}

// OrderItem model
type OrderItem struct {
	BaseModel
	OrderID    uuid.UUID
	Order      Order   `gorm:"foreignKey:OrderID"`
	ProductID  uuid.UUID
	Product    Product `gorm:"foreignKey:ProductID"`
	Quantity   int     `gorm:"not null"`
	UnitPrice  float64 `gorm:"not null"`
}

// ScrapedProduct model for storing scraped product data
type ScrapedProduct struct {
	BaseModel
	Name         string    `gorm:"not null"`
	Description  string
	Price        float64   `gorm:"not null"`
	OriginalURL  string    `gorm:"unique;not null"`
	ImageURL     string
	ImageData    []byte    `gorm:"type:bytea"` // Stores image as binary data
	Category     string
	SourceSite   string
	LastScraped  time.Time
}

// CalculateTrustScore calculates and updates supplier's trust score
func (s *Supplier) CalculateTrustScore(db *gorm.DB) error {
	var avgRating float64
	result := db.Model(&Review{}).
		Where("supplier_id = ?", s.ID).
		Select("AVG(rating)").
		Scan(&avgRating)
	
	if result.Error != nil {
		return result.Error
	}

	s.TrustScore = avgRating
	s.IsTrusted = avgRating >= 5.0

	return db.Save(s).Error
}

// BeforeCreate hook for Product to check supplier trust score
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	var supplier Supplier
	if err := tx.First(&supplier, "id = ?", p.SupplierID).Error; err != nil {
		return err
	}

	if !supplier.IsTrusted {
		return errors.New("supplier trust score below threshold")
	}
	return nil
}

// BeforeCreate hook to set UUID for all models
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
