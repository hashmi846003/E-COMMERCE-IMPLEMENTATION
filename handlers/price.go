package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/tax"  // Add this import
    "gorm.io/gorm"

    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

// GetPriceHandler returns price and related info for a product specified by product_id
func GetPriceHandler(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        productIDStr := c.Param("product_id")
        productID, err := uuid.Parse(productIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product_id"})
            return
        }

        var product models.Product
        if err := db.First(&product, "id = ?", productID).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
            }
            return
        }

        // Get detailed GST information from Excel lookup
        gstDetails := tax.GetGSTRateDetails(product.HSNCode, product.GSTCategory, product.Name)

        // Return comprehensive pricing info
        response := gin.H{
            "product_id":            product.ID,
            "name":                  product.Name,
            "description":           product.Description,
            "hsn_code":              product.HSNCode,
            "gst_category":          product.GSTCategory,
            "cost_price":            product.CostPrice,
            "profit_margin_percent": product.ProfitMarginPercent,
            "gst_percent":           product.GSTPercent,
            "vat_percent":           product.VATPercent,
            "selling_price":         product.SellingPrice,
            "country_code":          product.CountryCode,
            "gst_details":           gstDetails,
        }

        c.JSON(http.StatusOK, response)
    }
}
