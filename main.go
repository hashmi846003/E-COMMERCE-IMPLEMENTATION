package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/handlers"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/routes"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/utils"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/tax"  // Add this import
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    utils.LoadEnv()

    // Initialize GST tax system - MUST be called before database operations
    log.Println("Loading GST tax rates from Excel...")
    if err := tax.Initialize(); err != nil {
        log.Fatalf("Failed to initialize tax system: %v", err)
    }
    log.Println("GST tax rates loaded successfully")

    dsn := utils.GetEnv("DATABASE_URL", "")
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to DB:", err)
    }

    // --- Migrate all models, including Repository ---
    _ = db.AutoMigrate(
        &models.Token{},
        &models.Admin{},
        &models.Consumer{},
        &models.Supplier{},
        &models.Cart{},
        &models.CartItem{},
        &models.Repository{}, // Inventory/in-stock model
        &models.Section{},    // Add missing models
        &models.GSTSlab{},
        &models.Product{},
        &models.Order{},
        &models.OrderItem{},
    )

    router := gin.Default()
    routes.SetupRoutes(router, db)

    // --- Supplier: image uploading and fetching endpoints ---
    router.POST("/supplier/profile-image/:email", handlers.SupplierImageUploadHandler(db))
    router.GET("/supplier/profile-image/:email", handlers.SupplierImageFetchHandler(db))

    // --- Cart Endpoints ---
    router.POST("/cart/:consumer_id/items", handlers.AddToCartHandler(db))
    router.GET("/cart/:consumer_id", handlers.GetCartHandler(db))
    router.DELETE("/cart/:consumer_id/items/:item_id", handlers.RemoveFromCartHandler(db))
    router.DELETE("/cart/:consumer_id/items", handlers.ClearCartHandler(db))
    router.PUT("/cart/:consumer_id/items/:item_id", handlers.UpdateCartItemHandler(db))
    
    // --- "Buy Now" Endpoint ---
    router.POST("/buy-now", handlers.BuyNowHandler(db))

    // --- Repository/Inventory Endpoints ---
    router.POST("/repository/add", handlers.AddStockHandler(db))
    router.POST("/repository/purchase", handlers.PurchaseItemHandler(db))
    router.GET("/repository/:supplier_id/:product_id", handlers.GetRepositoryHandler(db))
    
    // Price Endpoint (GET product price by product_id)
    router.GET("/price/:product_id", handlers.GetPriceHandler(db))

    // --- GST Rate Lookup Endpoints ---
    router.GET("/tax/hsn/:hsn", func(c *gin.Context) {
        hsn := c.Param("hsn")
        result := tax.GetGSTRateByHSN(hsn)
        if result == nil {
            c.JSON(404, gin.H{"error": "GST rate not found for HSN: " + hsn})
            return
        }
        c.JSON(200, gin.H{"gst_rate": result})
    })
    
    router.GET("/tax/category/:category", func(c *gin.Context) {
        category := c.Param("category")
        result := tax.GetGSTRateByCategory(category)
        if result == nil {
            c.JSON(404, gin.H{"error": "GST rate not found for category: " + category})
            return
        }
        c.JSON(200, gin.H{"gst_rate": result})
    })

    router.GET("/tax/search", func(c *gin.Context) {
        hsn := c.Query("hsn")
        category := c.Query("category")
        description := c.Query("description")
        
        details := tax.GetGSTRateDetails(hsn, category, description)
        c.JSON(200, details)
    })

    router.GET("/tax/all", func(c *gin.Context) {
        rates := tax.GetAllGSTRates()
        c.JSON(200, gin.H{"total_rates": len(rates), "rates": rates})
    })

    router.GET("/tax/search/:keyword", func(c *gin.Context) {
        keyword := c.Param("keyword")
        results := tax.SearchGSTRates(keyword)
        c.JSON(200, gin.H{"results": results})
    })

    port := utils.GetEnv("APP_PORT", "8080")
    log.Println("🚀 Server running on http://localhost:" + port)
    router.Run(":" + port)
}
