package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

// AddStockHandler increments stock quantity for a supplier's product
func AddStockHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req struct {
            SupplierID  string  `json:"supplier_id" binding:"required"`
            ProductID   string  `json:"product_id" binding:"required"`
            ProductName string  `json:"product_name" binding:"required"`
            Amount      int     `json:"amount" binding:"required,gt=0"`
            MinQuantity int     `json:"min_quantity"` // optional; default 1 if missing or zero
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        supplierUUID, err := uuid.Parse(req.SupplierID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid supplier_id"})
            return
        }

        productUUID, err := uuid.Parse(req.ProductID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product_id"})
            return
        }

        // Find existing repository record
        var repo models.Repository
        err = db.Where("supplier_id = ? AND product_id = ?", supplierUUID, productUUID).First(&repo).Error
        if err != nil {
            if err == gorm.ErrRecordNotFound {
                // Create new repo record
                if req.MinQuantity <= 0 {
                    req.MinQuantity = 1 // default minimum quantity
                }
                repo = models.Repository{
                    SupplierID:  supplierUUID,
                    ProductID:   productUUID,
                    ProductName: req.ProductName,
                    Quantity:    req.Amount,
                    MinQuantity: req.MinQuantity,
                }
                if err := db.Create(&repo).Error; err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create repository record"})
                    return
                }
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
                return
            }
        } else {
            // Increment existing quantity
            repo.Quantity += req.Amount
            // Update min quantity if provided and valid
            if req.MinQuantity > 0 {
                repo.MinQuantity = req.MinQuantity
            }
            if err := db.Save(&repo).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update repository stock"})
                return
            }
        }

        c.JSON(http.StatusOK, gin.H{"message": "Stock incremented successfully", "repository": repo})
    }
}

// PurchaseItemHandler decrements stock quantity when a purchase is made
func PurchaseItemHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req struct {
            SupplierID string `json:"supplier_id" binding:"required"`
            ProductID  string `json:"product_id" binding:"required"`
            Amount     int    `json:"amount" binding:"required,gt=0"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        supplierUUID, err := uuid.Parse(req.SupplierID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid supplier_id"})
            return
        }

        productUUID, err := uuid.Parse(req.ProductID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product_id"})
            return
        }

        var repo models.Repository
        err = db.Where("supplier_id = ? AND product_id = ?", supplierUUID, productUUID).First(&repo).Error
        if err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "Repository record not found"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            }
            return
        }

        if repo.Quantity-req.Amount < repo.MinQuantity {
            c.JSON(http.StatusConflict, gin.H{"error": "item not available"})
            return
        }

        repo.Quantity -= req.Amount
        if err := db.Save(&repo).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update repository stock"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Purchase successful", "repository": repo})
    }
}

// GetRepositoryHandler returns the current inventory details for a given supplier and product
func GetRepositoryHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        supplierIDStr := c.Param("supplier_id")
        productIDStr := c.Param("product_id")

        supplierUUID, err := uuid.Parse(supplierIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid supplier_id"})
            return
        }

        productUUID, err := uuid.Parse(productIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product_id"})
            return
        }

        var repo models.Repository
        err = db.Where("supplier_id = ? AND product_id = ?", supplierUUID, productUUID).First(&repo).Error
        if err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "Repository record not found"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            }
            return
        }

        c.JSON(http.StatusOK, gin.H{"repository": repo})
    }
}
