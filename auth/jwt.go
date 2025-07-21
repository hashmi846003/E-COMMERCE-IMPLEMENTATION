package auth

import (
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/utils"
)

var jwtSecret = []byte(utils.GetEnv("JWT_SECRET", "secret"))

type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID, role string) (string, time.Time, error) {
	expire := time.Now().Add(15 * time.Minute)
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expire),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tStr, err := token.SignedString(jwtSecret)
	return tStr, expire, err
}

func GenerateRefreshToken() string {
	s := make([]byte, 32)
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := range s {
		s[i] = chars[rand.Intn(len(chars))]
	}
	return string(s)
}

func SaveToken(db *gorm.DB, userID uuid.UUID, role, access, refresh string, exp time.Time) error {
	token := models.Token{
		UserID:       userID,
		Role:         role,
		AccessToken:  access,
		RefreshToken: refresh,
		Expiry:       exp,
	}
	return db.Create(&token).Error
}
