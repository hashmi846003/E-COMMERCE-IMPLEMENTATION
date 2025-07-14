package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/middleware"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config for OAuth2 and database
type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	RedirectURL        string
	DatabaseURL        string
}

type App struct {
	OAuth2 *oauth2.Config
	DB     *gorm.DB
}

// UserClaims for storing user information in context
type UserClaims struct {
	UserID uuid.UUID
	Role   string
	Email  string
}

// Initialize OAuth2 config and database
func (app *App) Initialize(config Config) error {
	// Validate required environment variables
	if config.GoogleClientID == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}
	if config.GoogleClientSecret == "" {
		return fmt.Errorf("GOOGLE_CLIENT_SECRET is required")
	}
	if config.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if config.RedirectURL == "" {
		return fmt.Errorf("REDIRECT_URL is required")
	}

	// Initialize database
	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	app.DB = db

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate Token model (and user models for authentication)
	if err := app.DB.AutoMigrate(
		&models.Token{},
		&models.Admin{},
		&models.Supplier{},
		&models.Consumer{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	// Initialize OAuth2 config
	app.OAuth2 = &oauth2.Config{
		ClientID:     config.GoogleClientID,
		ClientSecret: config.GoogleClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	return nil
}

// OAuth2 login handler
func (app *App) handleLogin(c *gin.Context) {
	role := c.Query("role")
	if role != "admin" && role != "supplier" && role != "consumer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	url := app.OAuth2.AuthCodeURL(role, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// OAuth2 callback handler
func (app *App) handleCallback(c *gin.Context) {
	ctx := c.Request.Context()
	role := c.Query("state") // Role passed as state
	code := c.Query("code")

	// Exchange code for OAuth2 token
	token, err := app.OAuth2.Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	// Get user info from Google
	client := app.OAuth2.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	// Check database for user role and ID
	var userID uuid.UUID
	var actualRole string
	switch role {
	case "admin":
		var admin models.Admin
		if err := app.DB.Where("email = ?", userInfo.Email).First(&admin).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin not found"})
			return
		}
		userID = admin.ID
		actualRole = "admin"
	case "supplier":
		var supplier models.Supplier
		if err := app.DB.Where("email = ?", userInfo.Email).First(&supplier).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Supplier not found"})
			return
		}
		userID = supplier.ID
		actualRole = "supplier"
	case "consumer":
		var consumer models.Consumer
		if err := app.DB.Where("email = ?", userInfo.Email).First(&consumer).Error; err != nil {
			// Create new consumer if not found
			consumer = models.Consumer{
				Email: userInfo.Email,
				Name:  userInfo.Name,
			}
			if err := app.DB.Create(&consumer).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create consumer"})
				return
			}
		}
		userID = consumer.ID
		actualRole = "consumer"
	}

	// Store token in database
	tokenRecord := models.Token{
		UserID:       userID,
		Role:         actualRole,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}
	if err := app.DB.Create(&tokenRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store token"})
		return
	}

	// Store user claims in context
	claims := &UserClaims{
		UserID: userID,
		Role:   actualRole,
		Email:  userInfo.Email,
	}
	ctx = context.WithValue(ctx, "user", claims)
	c.Request = c.Request.WithContext(ctx)

	// Return token to client
	c.JSON(http.StatusOK, gin.H{
		"access_token": token.AccessToken,
		"user_id":      userID,
		"role":         actualRole,
	})
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: Error loading .env file:", err)
	}

	config := Config{
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:        os.Getenv("REDIRECT_URL"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
	}

	app := &App{}
	if err := app.Initialize(config); err != nil {
		panic(fmt.Sprintf("Failed to initialize app: %v", err))
	}

	r := gin.Default()

	// Public routes
	r.GET("/login", app.handleLogin)
	r.GET("/callback", app.handleCallback)

	// Protected route (example)
	r.GET("/protected", middleware.AuthenticateToken(app.DB), middleware.RequireRole("consumer"), func(c *gin.Context) {
		claims := c.Request.Context().Value("user").(*UserClaims)
		c.JSON(http.StatusOK, gin.H{"message": "Access granted", "email": claims.Email})
	})

	r.Run(":8080")
}
