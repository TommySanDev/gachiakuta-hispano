package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// PasswordHandler handles magic links and password reset operations
type PasswordHandler struct {
	TokenService *user.TokenService
	EmailService *user.EmailService
}

// NewPasswordHandler creates a new password handler
func NewPasswordHandler(tokenService *user.TokenService, emailService *user.EmailService) *PasswordHandler {
	return &PasswordHandler{
		TokenService: tokenService,
		EmailService: emailService,
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

	magicLink, u, err := user.CreateMagicLink(input.Email)
	if err != nil {
		// Don't reveal if user exists - always return success
		RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "If your email exists, you will receive a magic link",
		})
		return
	}

	// Send magic link email
	if h.EmailService != nil && h.EmailService.IsConfigured() {
		err = h.EmailService.SendMagicLink(u.Email, magicLink.Token)
		if err != nil {
			// Log error but don't fail the request
			RespondWithError(w, http.StatusInternalServerError, "Failed to send magic link email")
			return
		}
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Magic link sent to your email",
	})
}

// LoginWithMagicLink authenticates user with magic link token
func (h *PasswordHandler) LoginWithMagicLink(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		RespondWithError(w, http.StatusBadRequest, "Token is required")
		return
	}

	u, err := user.ValidateAndUseMagicLink(token)
	if err != nil {
		switch err {
		case user.ErrInvalidMagicLink:
			RespondWithError(w, http.StatusUnauthorized, "Invalid magic link")
		case user.ErrMagicLinkExpired:
			RespondWithError(w, http.StatusUnauthorized, "Magic link expired")
		case user.ErrUserNotFound:
			RespondWithError(w, http.StatusUnauthorized, "User not found")
		case user.ErrUserInactive:
			RespondWithError(w, http.StatusForbidden, "Account inactive")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to process magic link")
		}
		return
	}

	// Generate Paseto token
	authToken, err := h.TokenService.GenerateToken(u, time.Hour)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	authResponse := user.AuthResponse{
		User:      u,
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

	resetToken, u, err := user.CreatePasswordResetToken(input.Email)
	if err != nil {
		// Don't reveal if user exists - always return success
		RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "If your email exists, you will receive reset instructions",
		})
		return
	}

	// Send password reset email
	if h.EmailService != nil && h.EmailService.IsConfigured() {
		err = h.EmailService.SendPasswordReset(u.Email, resetToken.Token)
		if err != nil {
			// Log error but don't fail the request
			RespondWithError(w, http.StatusInternalServerError, "Failed to send reset email")
			return
		}
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "If your email exists, you will receive reset instructions",
	})
}

// ValidateResetToken checks if a password reset token is valid
func (h *PasswordHandler) ValidateResetToken(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		RespondWithError(w, http.StatusBadRequest, "Token is required")
		return
	}

	u, err := user.ValidateResetToken(token)
	if err != nil {
		switch err {
		case user.ErrInvalidResetToken:
			RespondWithError(w, http.StatusBadRequest, "Invalid reset token")
		case user.ErrResetTokenExpired:
			RespondWithError(w, http.StatusBadRequest, "Reset token expired")
		default:
			RespondWithError(w, http.StatusBadRequest, "Invalid reset token")
		}
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

	err := user.ResetPasswordWithToken(input.Token, input.NewPassword)
	if err != nil {
		switch err {
		case user.ErrInvalidInput:
			RespondWithError(w, http.StatusBadRequest, "Token and new password are required")
		case user.ErrInvalidResetToken:
			RespondWithError(w, http.StatusBadRequest, "Invalid reset token")
		case user.ErrResetTokenExpired:
			RespondWithError(w, http.StatusBadRequest, "Reset token expired")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to reset password")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}
