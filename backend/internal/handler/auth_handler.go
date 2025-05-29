package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// AuthHandler handles basic authentication operations
type AuthHandler struct {
	TokenService *user.TokenService
	EmailService *user.EmailService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(tokenService *user.TokenService, emailService *user.EmailService) *AuthHandler {
	return &AuthHandler{
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

	newUser, err := user.Register(&input)
	if err != nil {
		switch err {
		case user.ErrInvalidInput:
			RespondWithError(w, http.StatusBadRequest, "Email, username, password, first name and last name are required")
		case user.ErrUserAlreadyExists:
			RespondWithError(w, http.StatusConflict, "User already exists")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		}
		return
	}

	// Send welcome email (don't fail registration if email fails)
	if h.EmailService != nil && h.EmailService.IsConfigured() {
		h.EmailService.SendWelcomeEmail(newUser.Email, newUser.Username)
		// Email errors are logged in the service, continue with registration
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

	// Authenticate user
	u, err := user.Authenticate(input.Email, input.Password)
	if err != nil {
		switch err {
		case user.ErrInvalidCredentials:
			RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		case user.ErrUserInactive:
			RespondWithError(w, http.StatusForbidden, "Account inactive")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Authentication failed")
		}
		return
	}

	// Update last login
	user.UpdateLastLogin(u.ID)

	// Generate Paseto token
	token, err := h.TokenService.GenerateToken(u, time.Hour)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	authResponse := user.AuthResponse{
		User:      u,
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
