package handlers

import (
    "net/http"
   // "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

type BuyNowRequest struct {
    ConsumerID  string  `json:"consumer_id" binding:"required"`
    SupplierID  string  `json:"supplier_id" binding:"required"`
    ProductID   string  `json:"product_id" binding:"required"`
    ProductName string  `json:"product_name" binding:"required"`
    Quantity    int     `json:"quantity" binding:"required,min=1"`
    Price       float64 `json:"price" binding:"required,gt=0"`
}

func BuyNowHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req BuyNowRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Parse UUIDs
        consumerUUID, err := uuid.Parse(req.ConsumerID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consumer_id"})
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

        // Verify Consumer exists
        var consumer models.Consumer
        if err := db.First(&consumer, "id = ?", consumerUUID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Consumer not found"})
            return
        }

        // Check repository stock
        var repo models.Repository
        err = db.Where("supplier_id = ? AND product_id = ?", supplierUUID, productUUID).First(&repo).Error
        if err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "Product stock not found"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            }
            return
        }

        if repo.Quantity-req.Quantity < repo.MinQuantity {
            c.JSON(http.StatusConflict, gin.H{"error": "item not available"})
            return
        }

        // Decrement stock atomically (transaction recommended)
        tx := db.Begin()
        repo.Quantity -= req.Quantity
        if err := tx.Save(&repo).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
            return
        }

        // Create the Order record
        order := models.Order{
            ConsumerID: consumerUUID,
            TotalPrice: float64(req.Quantity) * req.Price,
            Status:     "pending",
        }
        if err := tx.Create(&order).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
            return
        }

        // Create the OrderItem record
        orderItem := models.OrderItem{
            OrderID:     order.ID,
            SupplierID:  supplierUUID,
            ProductID:   productUUID,
            ProductName: req.ProductName,
            Quantity:    req.Quantity,
            Price:       req.Price,
        }
        if err := tx.Create(&orderItem).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order item"})
            return
        }

        tx.Commit()

        c.JSON(http.StatusOK, gin.H{
            "message":    "Purchase successful",
            "order_id":   order.ID,
            "total_price": order.TotalPrice,
            "status":     order.Status,
        })
    }
}
