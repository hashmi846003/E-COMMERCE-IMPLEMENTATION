package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/handlers"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/routes"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	router.POST("/supplier/profile-image/:email", handlers.SupplierImageUploadHandler(db))
	router.GET("/supplier/profile-image/:email", handlers.SupplierImageFetchHandler(db))


	port := utils.GetEnv("APP_PORT", "8080")
	log.Println("🚀 Server running on http://localhost:" + port)
	router.Run(":" + port)
}
