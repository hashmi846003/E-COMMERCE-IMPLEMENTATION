package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
)

// Role constants
const (
	RoleAdmin    = "admin"
	RoleSupplier = "supplier"
	RoleConsumer = "consumer"
)

// Config for OAuth2
type Config struct {
	GoogleClientID     string
	GoogleClientSecret string
	RedirectURL        string
	DatabaseURL        string
}

type App struct {
	DB     *gorm.DB
	Router *mux.Router
	OAuth2 *oauth2.Config
}

// UserClaims for JWT-like claims (simplified)
type UserClaims struct {
	UserID   uuid.UUID
	Role     string
	Email    string
}

// Initialize application
func (app *App) Initialize(config Config) error {
	// Initialize database
	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		return err
	}
	app.DB = db

	// Auto migrate models
	err = db.AutoMigrate(
		&models.Admin{},
		&models.Supplier{},
		&models.Consumer{},
		&models.Product{},
		&models.Review{},
		&models.Order{},
		&models.OrderItem{},
		&models.ScrapedProduct{},
	)
	if err != nil {
		return err
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

	// Initialize router
	app.Router = mux.NewRouter()
	app.initializeRoutes()

	return nil
}

// Role-based middleware
func (app *App) requireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("user").(*UserClaims)
			if !ok || claims.Role != role {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// OAuth2 login handler
func (app *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	if role != RoleAdmin && role != RoleSupplier && role != RoleConsumer {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	url := app.OAuth2.AuthCodeURL(role, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// OAuth2 callback handler
func (app *App) handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	role := r.URL.Query().Get("state") // Role passed as state
	code := r.URL.Query().Get("code")

	token, err := app.OAuth2.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := app.OAuth2.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Handle user based on role
	var userID uuid.UUID
	switch role {
	case RoleAdmin:
		var admin models.Admin
		result := app.DB.Where("email = ?", userInfo.Email).First(&admin)
		if result.Error == gorm.ErrRecordNotFound {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("default-password"), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Failed to hash password", http.StatusInternalServerError)
				return
			}
			admin = models.Admin{
				Email:    userInfo.Email,
				Name:     userInfo.Name,
				Password: string(hashedPassword),
			}
			if err := app.DB.Create(&admin).Error; err != nil {
				http.Error(w, "Failed to create admin", http.StatusInternalServerError)
				return
			}
		}
		userID = admin.ID
	case RoleSupplier:
		var supplier models.Supplier
		result := app.DB.Where("email = ?", userInfo.Email).First(&supplier)
		if result.Error == gorm.ErrRecordNotFound {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("default-password"), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Failed to hash password", http.StatusInternalServerError)
				return
			}
			supplier = models.Supplier{
				Email:      userInfo.Email,
				Name:       userInfo.Name,
				Password:   string(hashedPassword),
				TrustScore: 0.0,
				IsTrusted:  false,
			}
			if err := app.DB.Create(&supplier).Error; err != nil {
				http.Error(w, "Failed to create supplier", http.StatusInternalServerError)
				return
			}
		}
		userID = supplier.ID
	case RoleConsumer:
		var consumer models.Consumer
		result := app.DB.Where("email = ?", userInfo.Email).First(&consumer)
		if result.Error == gorm.ErrRecordNotFound {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("default-password"), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Failed to hash password", http.StatusInternalServerError)
				return
			}
			consumer = models.Consumer{
				Email:    userInfo.Email,
				Name:     userInfo.Name,
				Password: string(hashedPassword),
			}
			if err := app.DB.Create(&consumer).Error; err != nil {
				http.Error(w, "Failed to create consumer", http.StatusInternalServerError)
				return
			}
		}
		userID = consumer.ID
	}

	// Set user claims in context
	claims := &UserClaims{
		UserID: userID,
		Role:   role,
		Email:  userInfo.Email,
	}
	ctx = context.WithValue(ctx, "user", claims)
	r = r.WithContext(ctx)

	// Redirect based on role
	switch role {
	case RoleAdmin:
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
	case RoleSupplier:
		http.Redirect(w, r, "/supplier/dashboard", http.StatusSeeOther)
	case RoleConsumer:
		http.Redirect(w, r, "/consumer/dashboard", http.StatusSeeOther)
	}
}

// Example role-specific handlers
func (app *App) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("user").(*UserClaims)
	fmt.Fprintf(w, "Welcome Admin %s", claims.Email)
}

func (app *App) handleSupplierDashboard(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("user").(*UserClaims)
	fmt.Fprintf(w, "Welcome Supplier %s", claims.Email)
}

func (app *App) handleConsumerDashboard(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("user").(*UserClaims)
	fmt.Fprintf(w, "Welcome Consumer %s", claims.Email)
}

// Initialize routes
func (app *App) initializeRoutes() {
	// Public routes
	app.Router.HandleFunc("/login", app.handleLogin).Methods("GET")
	app.Router.HandleFunc("/callback", app.handleCallback).Methods("GET")

	// Protected routes with role-based middleware
	app.Router.Handle("/admin/dashboard", app.requireRole(RoleAdmin)(http.HandlerFunc(app.handleAdminDashboard))).Methods("GET")
	app.Router.Handle("/supplier/dashboard", app.requireRole(RoleSupplier)(http.HandlerFunc(app.handleSupplierDashboard))).Methods("GET")
	app.Router.Handle("/consumer/dashboard", app.requireRole(RoleConsumer)(http.HandlerFunc(app.handleConsumerDashboard))).Methods("GET")
}

func main() {
	config := Config{
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:        "http://localhost:8080/callback",
		DatabaseURL:        os.Getenv("DATABASE_URL"),
	}

	app := &App{}
	if err := app.Initialize(config); err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":8080", app.Router))
}
