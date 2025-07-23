package routes

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/admin"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/supplier"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/consumer"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/auth"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
    authGroup := r.Group("/auth/google")
    {
        authGroup.GET("/login/admin", auth.GoogleLogin("ADMIN"))
        authGroup.GET("/login/supplier", auth.GoogleLogin("SUPPLIER"))
        authGroup.GET("/login/consumer", auth.GoogleLogin("CONSUMER"))

        authGroup.GET("/callback/admin", auth.GoogleCallback("ADMIN", db))
        authGroup.GET("/callback/supplier", auth.GoogleCallback("SUPPLIER", db))
        authGroup.GET("/callback/consumer", auth.GoogleCallback("CONSUMER", db))
    }

    // JWT Refresh Route
    r.POST("/auth/refresh", auth.RefreshTokenHandler(db))

    // Role-based API Routes
    api := r.Group("/api")
    {
        admin.RegisterAdminRoutes(api)     // ✅ Now resolved
        supplier.RegisterSupplierRoutes(api)
        consumer.RegisterConsumerRoutes(api)
    }
}
