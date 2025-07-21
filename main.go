package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/routes"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/utils"
)

func main() {
	utils.LoadEnv()

	dsn := utils.GetEnv("DATABASE_URL", "")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect database:", err)
	}

	_ = db.AutoMigrate(&models.Consumer{}, &models.Token{}) // Add others as needed

	r := gin.Default()
	routes.SetupRoutes(r, db)

	port := utils.GetEnv("APP_PORT", "8080")
	log.Println("✅ Server running at http://localhost:" + port)
	log.Fatal(r.Run(":" + port))
}
