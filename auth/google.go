package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth "google.golang.org/api/oauth2/v2" // ✅ alias Google API client
	"gorm.io/gorm"

	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/utils"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

// Returns a configured OAuth2 client for a specific role
func getOAuthConfig(role string) *oauth2.Config {
	redirectURL := utils.GetEnv(role+"_REDIRECT_URL", "")
	return &oauth2.Config{
		ClientID:     utils.GetEnv("GOOGLE_CLIENT_ID", ""),
		ClientSecret: utils.GetEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

// Starts Google OAuth2 flow for a given role
func GoogleLogin(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := getOAuthConfig(role).AuthCodeURL("state-" + role)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

// Callback handler for Google OAuth2 login
func GoogleCallback(role string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
			return
		}

		// Step 1: Exchange code for token
		token, err := getOAuthConfig(role).Exchange(context.Background(), code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token exchange failed", "detail": err.Error()})
			return
		}

		// Step 2: Use token to get user info from Google
		client := getOAuthConfig(role).Client(context.Background(), token)
		svc, err := oauth.New(client) // ✅ Correctly use the API client
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize Google service"})
			return
		}

		userInfo, err := svc.Userinfo.Get().Do()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
			return
		}

		// Step 3: Upsert user and get internal UUID
		var userID string
		switch role {
		case "ADMIN":
			userID = models.UpsertAdmin(userInfo.Email, userInfo.Name, db)

		case "SUPPLIER":
			userID = models.UpsertSupplier(userInfo.Email, userInfo.Name, db)

		case "CONSUMER":
			userID = models.UpsertConsumer(userInfo.Email, userInfo.Name, db)

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
			return
		}

		// Step 4: Generate access and refresh tokens
		accessToken, accessExp, _ := GenerateAccessToken(userID, role)
		refreshToken := GenerateRefreshToken()
		_ = SaveToken(db, userID, role, accessToken, refreshToken, accessExp)

		// Step 5: Return tokens in response
		c.JSON(http.StatusOK, gin.H{
			"message":       "Login successful",
			"user_email":    userInfo.Email,
			"user_name":     userInfo.Name,
			"user_role":     role,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_at":    accessExp,
		})
	}
}
