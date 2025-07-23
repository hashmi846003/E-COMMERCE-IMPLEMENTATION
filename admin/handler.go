package admin

import "github.com/gin-gonic/gin"

func Dashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Welcome Admin 👑"})
	}
}
