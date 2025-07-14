package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

// UserClaims for storing user information in request context
type UserClaims struct {
	UserID uuid.UUID
	Role   string
	Email  string
}

// AuthenticateToken validates OAuth2 tokens against the database
func AuthenticateToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract and validate Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || len(authHeader) < 7 || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			c.Abort()
			return
		}
		tokenString := authHeader[7:]

		// Check token in database
		var tokenRecord models.Token
		if err := db.Where("access_token = ? AND expiry > ?", tokenString, time.Now()).First(&tokenRecord).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user claims in context
		claims := &UserClaims{
			UserID: tokenRecord.UserID,
			Role:   tokenRecord.Role,
		}
		ctx := context.WithValue(c.Request.Context(), "user", claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// RequireRole enforces role-based access control
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Request.Context().Value("user").(*UserClaims)
		if !ok || claims.Role != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}
