package admin

import (
    "github.com/gin-gonic/gin"
    //"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/routes"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/auth"
)

// RegisterAdminRoutes registers the admin routes
func RegisterAdminRoutes(admin *gin.RouterGroup) {
    admin.Use(auth.AuthMiddleware("admin"))
    admin.GET("/dashboard", Dashboard())
}
