package main

import (
<<<<<<< HEAD
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/routes"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
=======
    "log"
    "gorm.io/gorm"
    "gorm.io/driver/postgres"
    "github.com/gin-gonic/gin"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/utils"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/routes"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
>>>>>>> abdf1899fb3e30543537140a955278abc1c5e4d8
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

<<<<<<< HEAD
	// --- Supplier section: image uploading, cropping, watermark ---
	router.POST("/supplier/profile-image/:email", SupplierImageUploadHandler(db))

	port := utils.GetEnv("APP_PORT", "8080")
	log.Println("🚀 Server running on http://localhost:" + port)
	router.Run(":" + port)
=======
    // --- Supplier section: image uploading, cropping, watermark ---
    router.POST("/supplier/profile-image/:email", SupplierImageUploadHandler(db))

    port := utils.GetEnv("APP_PORT", "8080")
    log.Println("🚀 Server running on http://localhost:" + port)
    router.Run(":" + port)
}

func SupplierImageUploadHandler(db *gorm.DB) gin.HandlerFunc {
	panic("unimplemented")
>>>>>>> abdf1899fb3e30543537140a955278abc1c5e4d8
}

func SupplierImageUploadHandler(db *gorm.DB) gin.HandlerFunc {
	panic("unimplemented")
}
