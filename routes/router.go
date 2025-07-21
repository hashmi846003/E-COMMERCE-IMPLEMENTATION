package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/auth"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/admin/login", auth.AdminLogin(db))
		authGroup.POST("/consumer/login", auth.ConsumerLogin(db))
		authGroup.POST("/supplier/login", auth.SupplierLogin(db))
		authGroup.POST("/refresh", auth.RefreshTokenHandler(db))
		authGroup.GET("/google/login", auth.GoogleLogin())
		authGroup.GET("/google/callback", auth.GoogleCallback(db))
	}
}
