package auth

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    googleOAuth2 "google.golang.org/api/oauth2/v2" // ✅ Correct import for user info retrieval

    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/utils"
)

var googleOAuthConfig = &oauth2.Config{
    ClientID:     utils.GetEnv("GOOGLE_CLIENT_ID", ""),
    ClientSecret: utils.GetEnv("GOOGLE_CLIENT_SECRET", ""),
    RedirectURL:  utils.GetEnv("REDIRECT_URL", ""),
    Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
    Endpoint:     google.Endpoint,
}

// GoogleLogin initiates the OAuth2 login process
func GoogleLogin() gin.HandlerFunc {
    return func(c *gin.Context) {
        url := googleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
        c.Redirect(http.StatusTemporaryRedirect, url)
    }
}

// GoogleCallback handles the callback from Google OAuth2
func GoogleCallback(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        code := c.Query("code")
        if code == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
            return
        }

        // Exchange code for token
        token, err := googleOAuthConfig.Exchange(context.Background(), code)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Token exchange failed"})
            return
        }

        // Use token to create a client
        client := googleOAuthConfig.Client(context.Background(), token)

        // Create OAuth2 service to fetch user info
        svc, err := googleOAuth2.New(client)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Google service"})
            return
        }

        userInfo, err := svc.Userinfo.Get().Do()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info from Google"})
            return
        }

        // Check if user exists in DB
        var user models.Consumer
        if err := db.Where("email = ?", userInfo.Email).First(&user).Error; err != nil {
            // If not, create the user
            user = models.Consumer{
                Email:    userInfo.Email,
                Name:     userInfo.Name,
                Password: "", // empty password as we use Google
                Address:  "",
                Phone:    "",
            }
            if err := db.Create(&user).Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
                return
            }
        }

        // Generate access and refresh tokens
        accessToken, expiry, err := GenerateAccessToken(user.ID.String(), "consumer")
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation error"})
            return
        }

        refreshToken := GenerateRefreshToken()
        if err := SaveToken(db, user.ID, "consumer", accessToken, refreshToken, expiry); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store tokens"})
            return
        }

        // Respond with tokens
        c.JSON(http.StatusOK, gin.H{
            "message":        "Logged in via Google",
            "access_token":   accessToken,
            "refresh_token":  refreshToken,
            "expires_at":     expiry,
            "user":           user.Name,
            "email":          user.Email,
            "login_method":   "google",
        })
    }
}
