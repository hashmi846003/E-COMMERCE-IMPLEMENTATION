package models

import (
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/tax"  // Add this import
    "gorm.io/gorm"
)

// --- Base Model with UUID primary key ---
type BaseModel struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// --- Section (Category: Electronics, Fashion, etc.) ---
type Section struct {
    BaseModel
    Name        string    `gorm:"unique;not null"`
    Description string
    Products    []Product // One-to-many: each section has many products
}

// --- GST Slab (for India) - now optional since we use Excel lookup ---
type GSTSlab struct {
    BaseModel
    Category      string  `gorm:"not null;uniqueIndex:category_country"`  // e.g., "Electronics"
    GSTPercent    float64 `gorm:"not null"`                               // e.g., 18.0 for 18%
    CountryCode   string  `gorm:"size:2;not null;uniqueIndex:category_country"` // "IN" for India
    EffectiveFrom time.Time
}

// --- Product ---
type Product struct {
    BaseModel
    SectionID           uuid.UUID `gorm:"type:uuid;not null;index"`
    Name                string    `gorm:"not null"`
    Description         string
    Image               []byte
    CountryCode         string  `gorm:"size:2;not null;default:'IN'"`
    GSTCategory         string  `gorm:"not null;default:''"`      // e.g. "Electronics"
    HSNCode             string  `gorm:"default:''"`               // HSN Code field for GST lookup
    CostPrice           float64 `gorm:"not null"`                 // Provided by seller
    ProfitMarginPercent float64 `gorm:"not null;default:10.0"`    // Default 10%
    GSTPercent          float64 `gorm:"not null;default:0.0"`     // Auto-applied for India
    VATPercent          float64 `gorm:"not null;default:0.0"`     // For non-India countries
    SellingPrice        float64 `gorm:"not null;default:0.0"`     // Calculated before save
}

// Auto-calculate SellingPrice and set GST/VAT percent on save using Excel lookup
func (p *Product) BeforeSave(tx *gorm.DB) (err error) {
    if p.CountryCode == "IN" {
        // Use tax package to get GST percent from Excel data
        p.GSTPercent = tax.GetGSTRateForProduct(p.GSTCategory, p.HSNCode, p.Name, p.CountryCode)
        
        // Calculate selling price including GST
        p.SellingPrice = p.CostPrice +
            (p.CostPrice * p.ProfitMarginPercent / 100.0) +
            (p.CostPrice * p.GSTPercent / 100.0)
    } else {
        // Non-Indian products use VAT
        p.SellingPrice = p.CostPrice +
            (p.CostPrice * p.ProfitMarginPercent / 100.0) +
            (p.CostPrice * p.VATPercent / 100.0)
    }
    return nil
}

// --- Admin ---
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
    Revoked  bool `gorm:"default:false"`
    Cart     Cart
}

// --- Supplier ---
type Supplier struct {
    BaseModel
    Email    string `gorm:"unique;not null"`
    Password string
    Name     string
    Company  string
    Revoked  bool   `gorm:"default:false"`
    Image    []byte
}

// --- Token ---
type Token struct {
    BaseModel
    UserID       string    `gorm:"not null"`
    Role         string
    AccessToken  string
    RefreshToken string
    Expiry       time.Time
}

// --- Cart ---
type Cart struct {
    BaseModel
    ConsumerID uuid.UUID  `gorm:"type:uuid;not null;index"`
    Items      []CartItem
}

// --- CartItem ---
type CartItem struct {
    BaseModel
    CartID      uuid.UUID `gorm:"type:uuid;not null;index"`
    SupplierID  uuid.UUID `gorm:"type:uuid"`
    ProductID   uuid.UUID `gorm:"type:uuid"`
    ProductName string
    Quantity    int     `gorm:"not null;default:1"`
    Price       float64 `gorm:"not null"`
}

// --- Repository (Inventory per Supplier/Product) ---
type Repository struct {
    BaseModel
    SupplierID  uuid.UUID `gorm:"type:uuid;not null;index"`
    ProductID   uuid.UUID `gorm:"type:uuid;not null;index"`
    ProductName string
    Quantity    int       `gorm:"not null;default:0"`
    MinQuantity int       `gorm:"not null;default:1"`
}

// --- Order ---
type Order struct {
    BaseModel
    ConsumerID uuid.UUID `gorm:"type:uuid;not null;index"`
    TotalPrice float64
    Status     string
    Items      []OrderItem
}

// --- OrderItem ---
type OrderItem struct {
    BaseModel
    OrderID     uuid.UUID `gorm:"type:uuid;not null;index"`
    SupplierID  uuid.UUID `gorm:"type:uuid"`
    ProductID   uuid.UUID `gorm:"type:uuid"`
    ProductName string
    Quantity    int
    Price       float64
}

// --- UPSERT and image update functions ---

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

func UpdateSupplierImage(email string, imageData []byte, db *gorm.DB) error {
    return db.Model(&Supplier{}).Where("email = ?", email).Update("image", imageData).Error
}

// --- Repository Logic (inventory management) ---

func (r *Repository) Increment(db *gorm.DB, amount int) error {
    if amount <= 0 {
        return fmt.Errorf("increment amount must be positive")
    }
    r.Quantity += amount
    return db.Save(r).Error
}

func (r *Repository) Decrement(db *gorm.DB, amount int) error {
    if amount <= 0 {
        return fmt.Errorf("decrement amount must be positive")
    }
    if r.Quantity-amount < r.MinQuantity {
        return fmt.Errorf("item not available")
    }
    r.Quantity -= amount
    return db.Save(r).Error
}
