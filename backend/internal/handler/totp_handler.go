package handler

import (
	"encoding/json"
	"net/http"

	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// TOTPHandler handles TOTP 2FA operations
type TOTPHandler struct{}

// NewTOTPHandler creates a new TOTP handler
func NewTOTPHandler() *TOTPHandler {
	return &TOTPHandler{}
}

// Setup2FA initializes TOTP 2FA for the current user
func (h *TOTPHandler) Setup2FA(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	setupResponse, err := user.SetupTOTP(authUser.ID)
	if err != nil {
		switch err {
		case user.ErrUserNotFound:
			RespondWithError(w, http.StatusUnauthorized, "User not found")
		case user.ErrTOTPAlreadyEnabled:
			RespondWithError(w, http.StatusConflict, "2FA already enabled")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to setup 2FA")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, setupResponse)
}

// Verify2FA verifies and enables TOTP 2FA
func (h *TOTPHandler) Verify2FA(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	var input user.VerifyTOTPInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	recoveryCodes, err := user.VerifyAndEnableTOTP(authUser.ID, input.Code)
	if err != nil {
		switch err {
		case user.ErrInvalidInput:
			RespondWithError(w, http.StatusBadRequest, "Verification code is required")
		case user.ErrTOTPNotFound:
			RespondWithError(w, http.StatusBadRequest, "2FA not setup")
		case user.ErrInvalidTOTPCode:
			RespondWithError(w, http.StatusBadRequest, "Invalid verification code")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to verify 2FA")
		}
		return
	}

	response := map[string]interface{}{
		"message": "2FA enabled successfully",
	}
	
	if len(recoveryCodes) > 0 {
		response["recovery_codes"] = recoveryCodes
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// Disable2FA disables TOTP 2FA for the current user
func (h *TOTPHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	err := user.DisableTOTP(authUser.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to disable 2FA")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "2FA disabled successfully",
	})
}

// GenerateRecoveryCodes generates new recovery codes for 2FA
func (h *TOTPHandler) GenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	codes, err := user.GenerateNewRecoveryCodes(authUser.ID)
	if err != nil {
		switch err {
		case user.ErrTOTPNotEnabled:
			RespondWithError(w, http.StatusBadRequest, "2FA not enabled")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to generate recovery codes")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"recovery_codes": codes,
		"message":        "New recovery codes generated. Save them securely.",
	})
}

// UseRecoveryCode validates and uses a recovery code for 2FA bypass
func (h *TOTPHandler) UseRecoveryCode(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	var input user.UseRecoveryCodeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := user.UseRecoveryCode(authUser.ID, input.Code)
	if err != nil {
		switch err {
		case user.ErrInvalidInput:
			RespondWithError(w, http.StatusBadRequest, "Recovery code is required")
		case user.ErrInvalidRecoveryCode:
			RespondWithError(w, http.StatusBadRequest, "Invalid or used recovery code")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to use recovery code")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Recovery code used successfully",
	})
}

// ValidateTOTP validates a TOTP code for an authenticated user
func (h *TOTPHandler) ValidateTOTP(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	var input user.VerifyTOTPInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := user.ValidateTOTP(authUser.ID, input.Code)
	if err != nil {
		switch err {
		case user.ErrInvalidInput:
			RespondWithError(w, http.StatusBadRequest, "TOTP code is required")
		case user.ErrTOTPNotFound:
			RespondWithError(w, http.StatusBadRequest, "2FA not enabled")
		case user.ErrTOTPNotEnabled:
			RespondWithError(w, http.StatusBadRequest, "2FA not verified")
		case user.ErrInvalidTOTPCode:
			RespondWithError(w, http.StatusBadRequest, "Invalid TOTP code")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to validate TOTP")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "TOTP code valid",
	})
}
