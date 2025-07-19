package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/auth"
	"github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Database connection
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&models.Admin{}, &models.Consumer{}, &models.Supplier{}, &models.Token{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize auth service
	authService, err := auth.NewAuthService(db)
	if err != nil {
		log.Fatal("Failed to initialize auth service:", err)
	}

	// Set up router
	r := mux.NewRouter()

	// Login handler
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		accessToken, refreshToken, err := authService.Login(req.Email, req.Password, req.Role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		response := map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("POST")

	// Refresh handler
	r.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		newAccessToken, err := authService.RefreshToken(req.RefreshToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		response := map[string]string{
			"access_token": newAccessToken,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("POST")

	// Revoke handler
	r.Handle("/revoke", authService.AuthMiddleware("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userID").(uuid.UUID)
		if !ok {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}
		if err := authService.RevokeToken(userID); err != nil {
			http.Error(w, "Failed to revoke token", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))).Methods("POST")

	// Protected route
	r.Handle("/protected", authService.AuthMiddleware("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userID").(uuid.UUID)
		if !ok {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}
		role, ok := r.Context().Value("role").(string)
		if !ok {
			http.Error(w, "Invalid role", http.StatusUnauthorized)
			return
		}
		response := map[string]string{
			"message": "Protected resource accessed",
			"user_id": userID.String(),
			"role":    role,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))).Methods("GET")

	// Protected routes
	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.Use(authService.AuthMiddleware("admin"))
	adminRouter.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Admin dashboard"))
	}).Methods("GET")

	consumerRouter := r.PathPrefix("/consumer").Subrouter()
	consumerRouter.Use(authService.AuthMiddleware("consumer"))
	consumerRouter.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Consumer profile"))
	}).Methods("GET")

	supplierRouter := r.PathPrefix("/supplier").Subrouter()
	supplierRouter.Use(authService.AuthMiddleware("supplier"))
	supplierRouter.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Supplier inventory"))
	}).Methods("GET")

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Server failed:", err)
	}
}
