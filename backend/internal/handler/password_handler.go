package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/jmoiron/sqlx"

	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// PasswordHandler handles magic links and password reset operations
type PasswordHandler struct {
	DB           *sqlx.DB
	TokenService *user.TokenService
}

// NewPasswordHandler creates a new password handler
func NewPasswordHandler(db *sqlx.DB, tokenService *user.TokenService) *PasswordHandler {
	return &PasswordHandler{
		DB:           db,
		TokenService: tokenService,
	}
}

// RequestMagicLink sends a magic link to user's email
func (h *PasswordHandler) RequestMagicLink(w http.ResponseWriter, r *http.Request) {
	var input user.MagicLinkLoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if input.Email == "" {
		RespondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Get user by email
	var u user.User
	err := h.DB.Get(&u, "SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL", input.Email)
	if err != nil {
		// Don't reveal if user exists
		RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "If your email exists, you will receive a magic link",
		})
		return
	}

	// Check if user is active and has magic links enabled
	if !u.IsActive || !u.MagicLinkEnabled {
		RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "If your email exists, you will receive a magic link",
		})
		return
	}

	// Generate magic link token
	token, err := generateSecureToken()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate magic link")
		return
	}

	// Create magic link
	now := time.Now()
	magicLink := user.MagicLink{
		UserID:    u.ID,
		Token:     token,
		Used:      false,
		ExpiresAt: now.Add(15 * time.Minute), // 15 minutes expiration
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `INSERT INTO magic_links (user_id, token, used, expires_at, created_at, updated_at)
		VALUES (:user_id, :token, :used, :expires_at, :created_at, :updated_at)`
	
	_, err = h.DB.NamedExec(query, magicLink)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create magic link")
		return
	}

	// TODO: Send email with magic link
	// For now, return success message (in production, remove the token from response)
	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Magic link sent to your email",
		"token":   token, // Remove this in production
	})
}

// LoginWithMagicLink authenticates user with magic link token
func (h *PasswordHandler) LoginWithMagicLink(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		RespondWithError(w, http.StatusBadRequest, "Token is required")
		return
	}

	// Get magic link
	var ml user.MagicLink
	err := h.DB.Get(&ml, "SELECT * FROM magic_links WHERE token = $1 AND used = false", token)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Invalid magic link")
		return
	}

	// Check if expired
	if ml.IsExpired() {
		RespondWithError(w, http.StatusUnauthorized, "Magic link expired")
		return
	}

	// Mark as used
	_, err = h.DB.Exec("UPDATE magic_links SET used = true, updated_at = $1 WHERE id = $2", 
		time.Now(), ml.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to process magic link")
		return
	}

	// Get user
	var u user.User
	err = h.DB.Get(&u, "SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL", ml.UserID)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Check if user is still active
	if !u.IsActive {
		RespondWithError(w, http.StatusForbidden, "Account inactive")
		return
	}

	// Update last login
	h.DB.Exec("UPDATE users SET last_login_at = $1, updated_at = $1 WHERE id = $2", time.Now(), u.ID)

	// Generate Paseto token
	authToken, err := h.TokenService.GenerateToken(&u, time.Hour)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	authResponse := user.AuthResponse{
		User:      &u,
		Token:     authToken,
		ExpiresIn: 3600, // 1 hour
	}

	RespondWithJSON(w, http.StatusOK, authResponse)
}

// RequestPasswordReset sends a password reset token to user's email
func (h *PasswordHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var input user.ResetPasswordRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if input.Email == "" {
		RespondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Get user by email
	var u user.User
	err := h.DB.Get(&u, "SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL", input.Email)
	if err != nil {
		// Don't reveal if user exists
		RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "If your email exists, you will receive reset instructions",
		})
		return
	}

	// Check if user is active
	if !u.IsActive {
		RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "If your email exists, you will receive reset instructions",
		})
		return
	}

	// Invalidate existing reset tokens
	h.DB.Exec("UPDATE reset_tokens SET used = true WHERE user_id = $1 AND used = false", u.ID)

	// Generate reset token
	token, err := generateSecureToken()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate reset token")
		return
	}

	// Create reset token
	now := time.Now()
	resetToken := user.ResetToken{
		UserID:    u.ID,
		Token:     token,
		Used:      false,
		ExpiresAt: now.Add(1 * time.Hour), // 1 hour expiration
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `INSERT INTO reset_tokens (user_id, token, used, expires_at, created_at, updated_at)
		VALUES (:user_id, :token, :used, :expires_at, :created_at, :updated_at)`
	
	_, err = h.DB.NamedExec(query, resetToken)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create reset token")
		return
	}

	// TODO: Send email with reset link
	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "If your email exists, you will receive reset instructions",
		"token":   token, // Remove this in production
	})
}

// ValidateResetToken checks if a password reset token is valid
func (h *PasswordHandler) ValidateResetToken(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		RespondWithError(w, http.StatusBadRequest, "Token is required")
		return
	}

	// Get reset token
	var rt user.ResetToken
	err := h.DB.Get(&rt, "SELECT * FROM reset_tokens WHERE token = $1 AND used = false", token)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid reset token")
		return
	}

	// Check if expired
	if rt.IsExpired() {
		RespondWithError(w, http.StatusBadRequest, "Reset token expired")
		return
	}

	// Get user to return email
	var u user.User
	err = h.DB.Get(&u, "SELECT email FROM users WHERE id = $1 AND deleted_at IS NULL AND is_active = true", 
		rt.UserID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid reset token")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"valid": true,
		"email": u.Email,
	})
}

// ResetPassword resets user password using reset token
func (h *PasswordHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input user.ResetPasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if input.Token == "" || input.NewPassword == "" {
		RespondWithError(w, http.StatusBadRequest, "Token and new password are required")
		return
	}

	// Get reset token
	var rt user.ResetToken
	err := h.DB.Get(&rt, "SELECT * FROM reset_tokens WHERE token = $1 AND used = false", input.Token)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid reset token")
		return
	}

	// Check if expired
	if rt.IsExpired() {
		RespondWithError(w, http.StatusBadRequest, "Reset token expired")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Update user password
	now := time.Now()
	_, err = h.DB.Exec("UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3", 
		string(hashedPassword), now, rt.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	// Mark token as used
	h.DB.Exec("UPDATE reset_tokens SET used = true, updated_at = $1 WHERE id = $2", now, rt.ID)

	// Invalidate all other reset tokens for this user
	h.DB.Exec("UPDATE reset_tokens SET used = true WHERE user_id = $1 AND used = false", rt.UserID)

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}

// Helper function to generate secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
