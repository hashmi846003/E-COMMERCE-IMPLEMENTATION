package auth

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleOAuth2 "google.golang.org/api/oauth2/v2"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/utils"
)

var googleOAuthConfig *oauth2.Config

// Lazy GetGoogleOAuthConfig ensures env vars are loaded
func GetGoogleOAuthConfig() *oauth2.Config {
	if googleOAuthConfig == nil {
		googleOAuthConfig = &oauth2.Config{
			ClientID:     utils.GetEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: utils.GetEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  utils.GetEnv("REDIRECT_URL", ""),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/userinfo.email",
			},
			Endpoint: google.Endpoint,
		}
	}
	return googleOAuthConfig
}

// GoogleLogin → redirects to Google login page
func GoogleLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := GetGoogleOAuthConfig().AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		fmt.Println("[DEBUG] Redirecting to:", url)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// GoogleCallback → handles Google callback response
func GoogleCallback(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code parameter"})
			return
		}

		token, err := GetGoogleOAuthConfig().Exchange(context.Background(), code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token exchange failed", "details": err.Error()})
			return
		}

		client := GetGoogleOAuthConfig().Client(context.Background(), token)

		svc, err := googleOAuth2.New(client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Google service"})
			return
		}

		userInfo, err := svc.Userinfo.Get().Do()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info", "details": err.Error()})
			return
		}

		// Check/Create user (consumer)
		var user models.Consumer
		if err := db.Where("email = ?", userInfo.Email).First(&user).Error; err != nil {
			// No user found → create new
			user = models.Consumer{
				Email:    userInfo.Email,
				Name:     userInfo.Name,
				Password: "", // Google login doesn't use password
				Address:  "",
				Phone:    "",
			}
			if err := db.Create(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new user"})
				return
			}
		}

		// Create tokens
		accessToken, expiresAt, err := GenerateAccessToken(user.ID.String(), "consumer")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
			return
		}

		refreshToken := GenerateRefreshToken()
		if err := SaveToken(db, user.ID, "consumer", accessToken, refreshToken, expiresAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Logged in via Google",
			"user_email":    user.Email,
			"user_name":     user.Name,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_at":    expiresAt,
			"login_method":  "google",
		})
	}
}
