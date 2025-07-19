package main

import (
    "encoding/json"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/auth"
    "github.com/hashmi846003/E-COMMERCE-IMPLEMENTATION/models"
)

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    Role     string `json:"role"`
}

type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
}

func main() {
    // Initialize database
    db, err := gorm.Open(postgres.Open("your-dsn"), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Auto-migrate models
    db.AutoMigrate(&models.Admin{}, &models.Consumer{}, &models.Supplier{}, &models.Token{})

    // Initialize auth service
    authService, err := auth.NewAuthService(db)
    if err != nil {
        log.Fatal("Failed to initialize auth service:", err)
    }

    // Initialize router
    r := mux.NewRouter()

    // OAuth 2.0 endpoints
    r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
        var req LoginRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        accessToken, refreshToken, err := authService.Login(req.Email, req.Password, req.Role)
        if err != nil {
            http.Error(w, err.Error(), http.StatusUnauthorized)
            return
        }

        response := TokenResponse{
            AccessToken:  accessToken,
            RefreshToken: refreshToken,
            TokenType:    "Bearer",
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }).Methods("POST")

    r.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            RefreshToken string `json:"refresh_token"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        accessToken, err := authService.RefreshToken(req.RefreshToken)
        if err != nil {
            http.Error(w, err.Error(), http.StatusUnauthorized)
            return
        }

        response := TokenResponse{
            AccessToken:  accessToken,
            RefreshToken: req.RefreshToken,
            TokenType:    "Bearer",
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }).Methods("POST")

    r.HandleFunc("/revoke", authService.AuthMiddleware("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := r.Context().Value("userID").(uuid.UUID)
        if err := authService.RevokeToken(userID); err != nil {
            http.Error(w, "Failed to revoke token", http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
    }))).Methods("POST")

    // Example protected route
    r.HandleFunc("/protected", authService.AuthMiddleware("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := r.Context().Value("userID").(uuid.UUID)
        role := r.Context().Value("role").(string)
        response := map[string]string{
            "message": "Protected resource accessed",
            "user_id": userID.String(),
            "role":    role,
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }))).Methods("GET")

    // Start server
    log.Fatal(http.ListenAndServe(":8080", r))
}
