package auth

import (
	"time"
	"math/rand"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/utils"
	"gorm.io/gorm"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

var jwtSecret = []byte(utils.GetEnv("JWT_SECRET", "secret"))

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID, role string) (string, time.Time, error) {
	exp := time.Now().Add(15 * time.Minute)
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(jwtSecret)
	return s, exp, err
}

func GenerateRefreshToken() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 64)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func SaveToken(db *gorm.DB, userID, role, access, refresh string, exp time.Time) error {
	token := models.Token{
		UserID:      userID,
		Role:        role,
		AccessToken: access,
		RefreshToken: refresh,
		Expiry:      exp,
	}
	return db.Create(&token).Error
}
