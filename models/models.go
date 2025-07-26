package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

// --- Base Model with UUID primary key ---
type BaseModel struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// --- Admin (controls supplier and consumer accounts) ---
type Admin struct {
    BaseModel
    Email string `gorm:"unique"`
    Name  string
}

// --- Consumer ---
type Consumer struct {
    BaseModel
    Email    string `gorm:"unique;not null"`
    Password string
    Name     string
    Address  string
    Phone    string
    Revoked  bool   `gorm:"default:false"` // If true, account is blacklisted
}

// --- Supplier (now with image) ---
type Supplier struct {
    BaseModel
    Email    string `gorm:"unique;not null"`
    Password string
    Name     string
    Company  string
    Revoked  bool   `gorm:"default:false"` // If true, account is blacklisted
    Image    []byte // Store image/blob directly in database
}

// --- Token (for auth/session management) ---
type Token struct {
    BaseModel
    UserID       string    `gorm:"not null"` // references Admin/Consumer/Supplier
    Role         string
    AccessToken  string
    RefreshToken string
    Expiry       time.Time
}

// --- UPSERT functions (create or update by email) ---
func UpsertAdmin(email, name string, db *gorm.DB) string {
    var a Admin
    db.FirstOrCreate(&a, "email = ?", email)
    a.Name = name
    db.Save(&a)
    return a.ID.String()
}

func UpsertConsumer(email, name string, db *gorm.DB) string {
    var c Consumer
    db.FirstOrCreate(&c, "email = ?", email)
    c.Name = name
    db.Save(&c)
    return c.ID.String()
}

func UpsertSupplier(email, name string, db *gorm.DB) string {
    var s Supplier
    db.FirstOrCreate(&s, "email = ?", email)
    s.Name = name
    db.Save(&s)
    return s.ID.String()
}

// --- Admin Controls for Revoke/Unrevoke ---
func (a *Admin) RevokeSupplier(email string, db *gorm.DB) error {
    return db.Model(&Supplier{}).Where("email = ?", email).Update("revoked", true).Error
}

func (a *Admin) UnrevokeSupplier(email string, db *gorm.DB) error {
    return db.Model(&Supplier{}).Where("email = ?", email).Update("revoked", false).Error
}

func (a *Admin) RevokeConsumer(email string, db *gorm.DB) error {
    return db.Model(&Consumer{}).Where("email = ?", email).Update("revoked", true).Error
}

func (a *Admin) UnrevokeConsumer(email string, db *gorm.DB) error {
    return db.Model(&Consumer{}).Where("email = ?", email).Update("revoked", false).Error
}

// --- Utility: Update supplier's profile image ---
func UpdateSupplierImage(email string, imageData []byte, db *gorm.DB) error {
    return db.Model(&Supplier{}).Where("email = ?", email).Update("image", imageData).Error
}
