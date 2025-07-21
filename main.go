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
		log.Fatal("Failed to connect to DB:", err)
	}

	_ = db.AutoMigrate(&models.Admin{}, &models.Consumer{}, &models.Supplier{}, &models.Token{})

	r := gin.Default()
	routes.SetupRoutes(r, db)

	port := utils.GetEnv("APP_PORT", "8080")
	log.Printf("Server running on http://localhost:%s", port)
	r.Run(":" + port)
}
