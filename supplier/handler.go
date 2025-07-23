package supplier

import "github.com/gin-gonic/gin"

func Portal() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome Supplier 🧰",
		})
	}
}
