package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

// AddToCartHandler adds an item to the consumer's cart
func AddToCartHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        consumerIDStr := c.Param("consumer_id")
        consumerID, err := uuid.Parse(consumerIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consumer_id"})
            return
        }

        var req struct {
            SupplierID  string  `json:"supplier_id" binding:"required"`
            ProductID   string  `json:"product_id" binding:"required"`
            ProductName string  `json:"product_name" binding:"required"`
            Quantity    int     `json:"quantity" binding:"required,min=1"`
            Price       float64 `json:"price" binding:"required,gt=0"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        supplierID, err := uuid.Parse(req.SupplierID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid supplier_id"})
            return
        }
        productID, err := uuid.Parse(req.ProductID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product_id"})
            return
        }

        var cart models.Cart
        err = db.Where("consumer_id = ?", consumerID).First(&cart).Error
        if err != nil {
            if err == gorm.ErrRecordNotFound {
                cart = models.Cart{ConsumerID: consumerID}
                if err := db.Create(&cart).Error; err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
                    return
                }
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
                return
            }
        }

        var item models.CartItem
        err = db.Where("cart_id = ? AND product_id = ?", cart.ID, productID).First(&item).Error
        if err != nil {
            if err == gorm.ErrRecordNotFound {
                item = models.CartItem{
                    CartID:      cart.ID,
                    SupplierID:  supplierID,
                    ProductID:   productID,
                    ProductName: req.ProductName,
                    Quantity:    req.Quantity,
                    Price:       req.Price,
                }
                if err := db.Create(&item).Error; err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
                    return
                }
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
                return
            }
        } else {
            item.Quantity += req.Quantity
            item.Price = req.Price // optionally update price
            if err := db.Save(&item).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
                return
            }
        }

        c.JSON(http.StatusOK, gin.H{"message": "Item added to cart successfully", "item": item})
    }
}

// GetCartHandler fetches the cart and all its items for the given consumer
func GetCartHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        consumerIDStr := c.Param("consumer_id")
        consumerID, err := uuid.Parse(consumerIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consumer_id"})
            return
        }

        var cart models.Cart
        err = db.Preload("Items").Where("consumer_id = ?", consumerID).First(&cart).Error
        if err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusOK, gin.H{"cart": nil, "items": []models.CartItem{}})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            }
            return
        }

        c.JSON(http.StatusOK, gin.H{"cart": cart, "items": cart.Items})
    }
}

// RemoveFromCartHandler removes a specific item by item_id from the consumer's cart
func RemoveFromCartHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        consumerIDStr := c.Param("consumer_id")
        _, err := uuid.Parse(consumerIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consumer_id"})
            return
        }

        itemIDStr := c.Param("item_id")
        itemID, err := uuid.Parse(itemIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item_id"})
            return
        }

        var item models.CartItem
        if err := db.First(&item, "id = ?", itemID).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            }
            return
        }

        if err := db.Delete(&item).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart"})
    }
}

// ClearCartHandler deletes all items in the consumer's cart
func ClearCartHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        consumerIDStr := c.Param("consumer_id")
        consumerID, err := uuid.Parse(consumerIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consumer_id"})
            return
        }

        var cart models.Cart
        if err := db.Where("consumer_id = ?", consumerID).First(&cart).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusOK, gin.H{"message": "Cart cleared"})
                return
            }
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            return
        }

        if err := db.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart items"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Cart cleared"})
    }
}

// UpdateCartItemHandler updates quantity and/or price of a specific cart item
func UpdateCartItemHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        consumerIDStr := c.Param("consumer_id")
        _, err := uuid.Parse(consumerIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consumer_id"})
            return
        }

        itemIDStr := c.Param("item_id")
        itemID, err := uuid.Parse(itemIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item_id"})
            return
        }

        var req struct {
            Quantity *int     `json:"quantity" binding:"omitempty,min=1"`
            Price    *float64 `json:"price" binding:"omitempty,gt=0"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var item models.CartItem
        if err := db.First(&item, "id = ?", itemID).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            }
            return
        }

        // Update fields if provided
        if req.Quantity != nil {
            item.Quantity = *req.Quantity
        }
        if req.Price != nil {
            item.Price = *req.Price
        }

        if err := db.Save(&item).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Cart item updated successfully", "item": item})
    }
}
