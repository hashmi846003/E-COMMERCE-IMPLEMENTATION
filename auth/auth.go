package auth

import (
    "context"
    "errors"
    "os"
    "time"
    "github.com/joho/godotenv"
    "github.com/google/uuid"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
    "net/http"
)

type AuthService struct {
    db            *gorm.DB
    jwtSecret     string
    accessExpiry  time.Duration
    refreshExpiry time.Duration
}

type Claims struct {
    UserID uuid.UUID `json:"user_id"`
    Role   string    `json:"role"`
    jwt.RegisteredClaims
}

// NewAuthService initializes AuthService with configuration from .env
func NewAuthService(db *gorm.DB) (*AuthService, error) {
    err := godotenv.Load()
    if err != nil {
        return nil, errors.New("failed to load .env file")
    }

    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        return nil, errors.New("JWT_SECRET not set in .env")
    }

    accessExpiry, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_EXPIRY"))
    if err != nil {
        accessExpiry = 15 * time.Minute
    }

    refreshExpiry, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_EXPIRY"))
    if err != nil {
        refreshExpiry = 7 * 24 * time.Hour
    }

    return &AuthService{
        db:            db,
        jwtSecret:     jwtSecret,
        accessExpiry:  accessExpiry,
        refreshExpiry: refreshExpiry,
    }, nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// CheckPasswordHash verifies a password against its hash
func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// GenerateTokens creates access and refresh tokens
func (s *AuthService) GenerateTokens(userID uuid.UUID, role string) (string, string, error) {
    accessClaims := Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   userID.String(),
        },
    }

    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenStr, err := accessToken.SignedString([]byte(s.jwtSecret))
    if err != nil {
        return "", "", err
    }

    refreshClaims := Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   userID.String(),
        },
    }

    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshTokenStr, err := refreshToken.SignedString([]byte(s.jwtSecret))
    if err != nil {
        return "", "", err
    }

    token := models.Token{
        BaseModel:    models.BaseModel{ID: uuid.New()},
        UserID:       userID,
        Role:         role,
        AccessToken:  accessTokenStr,
        RefreshToken: refreshTokenStr,
        Expiry:       time.Now().Add(s.accessExpiry),
    }

    if err := s.db.Create(&token).Error; err != nil {
        return "", "", err
    }

    return accessTokenStr, refreshTokenStr, nil
}

// VerifyToken validates a JWT token
func (s *AuthService) VerifyToken(tokenStr string) (*Claims, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
        return []byte(s.jwtSecret), nil
    })

    if err != nil {
        return nil, err
    }

    if !token.Valid {
        return nil, errors.New("invalid token")
    }

    var storedToken models.Token
    if err := s.db.Where("access_token = ? AND expiry > ?", tokenStr, time.Now()).First(&storedToken).Error; err != nil {
        return nil, errors.New("token not found or expired")
    }

    return claims, nil
}

// RefreshToken generates new access token using refresh token
func (s *AuthService) RefreshToken(refreshTokenStr string) (string, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(refreshTokenStr, claims, func(token *jwt.Token) (interface{}, error) {
        return []byte(s.jwtSecret), nil
    })

    if err != nil || !token.Valid {
        return "", errors.New("invalid refresh token")
    }

    var storedToken models.Token
    if err := s.db.Where("refresh_token = ? AND expiry > ?", refreshTokenStr, time.Now()).First(&storedToken).Error; err != nil {
        return "", errors.New("invalid or expired refresh token")
    }

    newAccessClaims := Claims{
        UserID: claims.UserID,
        Role:   claims.Role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   claims.UserID.String(),
        },
    }

    newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newAccessClaims)
    newAccessTokenStr, err := newAccessToken.SignedString([]byte(s.jwtSecret))
    if err != nil {
        return "", err
    }

    storedToken.AccessToken = newAccessTokenStr
    storedToken.Expiry = time.Now().Add(s.accessExpiry)
    if err := s.db.Save(&storedToken).Error; err != nil {
        return "", err
    }

    return newAccessTokenStr, nil
}

// RevokeToken revokes a user's tokens
func (s *AuthService) RevokeToken(userID uuid.UUID) error {
    return s.db.Where("user_id = ?", userID).Delete(&models.Token{}).Error
}

// Login authenticates user and generates tokens
func (s *AuthService) Login(email, password, role string) (string, string, error) {
    var userID uuid.UUID
    var storedPassword string

    switch role {
    case "admin":
        var admin models.Admin
        if err := s.db.Where("email = ?", email).First(&admin).Error; err != nil {
            return "", "", errors.New("user not found")
        }
        userID = admin.GetID()
        storedPassword = admin.GetPassword()
    case "consumer":
        var consumer models.Consumer
        if err := s.db.Where("email = ?", email).First(&consumer).Error; err != nil {
            return "", "", errors.New("user not found")
        }
        userID = consumer.GetID()
        storedPassword = consumer.GetPassword()
    case "supplier":
        var supplier models.Supplier
        if err := s.db.Where("email = ?", email).First(&supplier).Error; err != nil {
            return "", "", errors.New("user not found")
        }
        userID = supplier.GetID()
        storedPassword = supplier.GetPassword()
    default:
        return "", "", errors.New("invalid role")
    }

    if !CheckPasswordHash(password, storedPassword) {
        return "", "", errors.New("invalid credentials")
    }

    return s.GenerateTokens(userID, role)
}

// AuthMiddleware validates JWT token for protected routes
func (s *AuthService) AuthMiddleware(requiredRole string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" || len(authHeader) < 7 || authHeader[:7] != "Bearer " {
                http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
                return
            }

            tokenStr := authHeader[7:]
            claims, err := s.VerifyToken(tokenStr)
            if err != nil {
                http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
                return
            }

            if requiredRole != "" && claims.Role != requiredRole {
                http.Error(w, "Insufficient permissions", http.StatusForbidden)
                return
            }

            ctx := context.WithValue(r.Context(), "userID", claims.UserID)
            ctx = context.WithValue(ctx, "role", claims.Role)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
