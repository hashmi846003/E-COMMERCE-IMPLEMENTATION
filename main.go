package main

import (
    "log"
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
    "github.com/gin-gonic/gin"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/utils"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/routes"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

func main() {
    utils.LoadEnv()

    dsn := utils.GetEnv("DATABASE_URL", "")
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to DB:", err)
    }

    // Migrate all role models + token
    _ = db.AutoMigrate(
        &models.Token{},
        &models.Admin{},
        &models.Consumer{},
        &models.Supplier{},
    )

    router := gin.Default()
    routes.SetupRoutes(router, db)

    // --- Supplier section: image uploading, cropping, watermark ---
    router.POST("/supplier/profile-image/:email", SupplierImageUploadHandler(db))

    port := utils.GetEnv("APP_PORT", "8080")
    log.Println("🚀 Server running on http://localhost:" + port)
    router.Run(":" + port)
}

func SupplierImageUploadHandler(db *gorm.DB) gin.HandlerFunc {
	panic("unimplemented")
}
