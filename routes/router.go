package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/auth"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/google/login", auth.GoogleLogin())
		authGroup.GET("/google/callback", auth.GoogleCallback(db))
	}
}
