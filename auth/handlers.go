package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func AdminLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		loginUser[models.Admin](c, db, "admin")
	}
}

func ConsumerLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		loginUser[models.Consumer](c, db, "consumer")
	}
}

func SupplierLogin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		loginUser[models.Supplier](c, db, "supplier")
	}
}

func loginUser[T any](c *gin.Context, db *gorm.DB, role string) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	var user T
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// CAST manually to access methods
	var id uuid.UUID
	var password string
	switch u := any(user).(type) {
	case models.Admin:
		id = u.ID
		password = u.Password
	case models.Consumer:
		id = u.ID
		password = u.Password
	case models.Supplier:
		id = u.ID
		password = u.Password
	}

	if password != input.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}
	accessToken, exp, _ := GenerateAccessToken(id.String(), role)
	refreshToken := GenerateRefreshToken()
	SaveToken(db, id, role, accessToken, refreshToken, exp)
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_at":    exp,
	})
}
