package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/jmoiron/sqlx"

	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// AuthHandler handles basic authentication operations
type AuthHandler struct {
	DB           *sqlx.DB
	TokenService *user.TokenService
	EmailService *user.EmailService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *sqlx.DB, tokenService *user.TokenService, emailService *user.EmailService) *AuthHandler {
	return &AuthHandler{
		DB:           db,
		TokenService: tokenService,
		EmailService: emailService,
	}
}

// Register creates a new user account
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input user.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if input.Email == "" || input.Username == "" || input.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Email, username and password are required")
		return
	}
	if input.FirstName == "" || input.LastName == "" {
		RespondWithError(w, http.StatusBadRequest, "First name and last name are required")
		return
	}

	// Check if user already exists
	var exists bool
	err := h.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM users WHERE email = $1 OR username = $2", 
		input.Email, input.Username)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if exists {
		RespondWithError(w, http.StatusConflict, "User already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Create user
	now := time.Now()
	newUser := user.User{
		Email:            input.Email,
		Username:         input.Username,
		PasswordHash:     string(hashedPassword),
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		Role:             user.RoleUser,
		IsActive:         true,
		EmailVerified:    false,
		MagicLinkEnabled: input.MagicLinkEnabled,
		TOTPEnabled:      false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	query := `INSERT INTO users (
		email, username, password_hash, first_name, last_name, role,
		is_active, email_verified, magic_link_enabled, totp_enabled,
		created_at, updated_at
	) VALUES (
		:email, :username, :password_hash, :first_name, :last_name, :role,
		:is_active, :email_verified, :magic_link_enabled, :totp_enabled,
		:created_at, :updated_at
	) RETURNING id`

	rows, err := h.DB.NamedQuery(query, newUser)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}
	defer rows.Close()

	if rows.Next() {
		var id uint
		rows.Scan(&id)
		newUser.ID = id
	}

	// Send welcome email
	if h.EmailService != nil && h.EmailService.IsConfigured() {
		// Don't fail registration if email fails, just log it
		err = h.EmailService.SendWelcomeEmail(newUser.Email, newUser.Username)
		if err != nil {
			// Email error is logged in the service, continue with registration
		}
	}

	RespondWithJSON(w, http.StatusCreated, newUser)
}

// Login authenticates a user with email and password
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input user.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if input.Email == "" || input.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Get user by email
	var u user.User
	err := h.DB.Get(&u, "SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL", input.Email)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check if user is active
	if !u.IsActive {
		RespondWithError(w, http.StatusForbidden, "Account inactive")
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.Password))
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Update last login
	h.DB.Exec("UPDATE users SET last_login_at = $1, updated_at = $1 WHERE id = $2", time.Now(), u.ID)

	// Generate Paseto token
	token, err := h.TokenService.GenerateToken(&u, time.Hour)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	authResponse := user.AuthResponse{
		User:      &u,
		Token:     token,
		ExpiresIn: 3600, // 1 hour in seconds
	}

	RespondWithJSON(w, http.StatusOK, authResponse)
}

// Logout invalidates the current session (with Paseto, this is just a client-side action)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// With Paseto tokens, logout is primarily client-side (remove token)
	// Server-side logout would require a blacklist, but for simplicity we'll keep it stateless
	
	// We could implement token blacklisting here if needed in the future
	// For now, just return success - client should discard the token
	
	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}
