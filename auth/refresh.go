package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	//"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

func RefreshTokenHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		var token models.Token
		if err := db.Where("access_token = ? AND refresh_token = ?", req.AccessToken, req.RefreshToken).First(&token).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		if time.Now().After(token.Expiry) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			return
		}

		newAccess, newExp, _ := GenerateAccessToken(token.UserID.String(), token.Role)
		newRefresh := GenerateRefreshToken()
		token.AccessToken = newAccess
		token.RefreshToken = newRefresh
		token.Expiry = newExp
		db.Save(&token)

		c.JSON(http.StatusOK, gin.H{
			"access_token":  newAccess,
			"refresh_token": newRefresh,
			"expires_at":    newExp,
		})
	}
}
