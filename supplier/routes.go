package supplier

import (
	"github.com/gin-gonic/gin"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/auth"
)

func RegisterSupplierRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/supplier")
	g.Use(auth.AuthMiddleware("supplier"))
	g.GET("/portal", Portal())
}
