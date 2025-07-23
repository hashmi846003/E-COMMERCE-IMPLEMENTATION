package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

func RefreshTokenHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}

		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token required"})
			return
		}

		var token models.Token
		if err := db.Where("refresh_token = ?", req.RefreshToken).First(&token).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		if time.Now().After(token.Expiry.Add(7 * 24 * time.Hour)) { // optional: set refresh token expiry
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
			return
		}

		// Create a new access token
		newAccess, newExp, err := GenerateAccessToken(token.UserID, token.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new token"})
			return
		}

		token.AccessToken = newAccess
		token.Expiry = newExp
		db.Save(&token)

		c.JSON(http.StatusOK, gin.H{
			"access_token": newAccess,
			"expires_at":   newExp,
		})
	}
}
