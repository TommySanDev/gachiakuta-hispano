package handler

import (
	"encoding/json"
	"net/http"

	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// VerificationHandler handles email verification operations
type VerificationHandler struct {
	EmailService *user.EmailService
}

// NewVerificationHandler creates a new verification handler
func NewVerificationHandler(emailService *user.EmailService) *VerificationHandler {
	return &VerificationHandler{
		EmailService: emailService,
	}
}

// VerifyEmail verifies user email with token
// POST /api/auth/verify-email
func (h *VerificationHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var input user.VerifyEmailInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if input.Token == "" {
		RespondWithError(w, http.StatusBadRequest, "Token is required")
		return
	}

	err := user.VerifyEmailWithToken(input.Token)
	if err != nil {
		switch err {
		case user.ErrInvalidToken:
			RespondWithError(w, http.StatusBadRequest, "Invalid verification token")
		case user.ErrTokenExpired:
			RespondWithError(w, http.StatusBadRequest, "Verification token expired")
		case user.ErrUserNotFound:
			RespondWithError(w, http.StatusBadRequest, "User not found")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to verify email")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Email verified successfully",
	})
}

// ResendVerification sends a new verification email
// POST /api/auth/resend-verification
func (h *VerificationHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var input user.ResendVerificationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if input.Email == "" {
		RespondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Get user to create verification token
	u, err := user.GetByEmail(input.Email)
	if err != nil {
		// Don't reveal if user exists
		RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "If your email exists, you will receive a verification link",
		})
		return
	}

	// Check if already verified
	if u.EmailVerified {
		RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Email is already verified",
		})
		return
	}

	// Create verification token
	verification, err := user.CreateEmailVerificationToken(u.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create verification token")
		return
	}

	// Send verification email
	if h.EmailService != nil && h.EmailService.IsConfigured() {
		err = h.EmailService.SendVerificationEmail(u.Email, u.Username, verification.Token)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Failed to send verification email")
			return
		}
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Verification email sent",
	})
}
