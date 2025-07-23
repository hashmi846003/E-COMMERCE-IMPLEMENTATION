package consumer

import (
	"github.com/gin-gonic/gin"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/auth"
)

func RegisterConsumerRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/consumer")
	g.Use(auth.AuthMiddleware("consumer"))
	g.GET("/home", Home())
}
