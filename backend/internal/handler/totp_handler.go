package handler

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"

	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// TOTPHandler handles TOTP 2FA operations
type TOTPHandler struct {
	DB *sqlx.DB
}

// NewTOTPHandler creates a new TOTP handler
func NewTOTPHandler(db *sqlx.DB) *TOTPHandler {
	return &TOTPHandler{DB: db}
}

// Setup2FA initializes TOTP 2FA for the current user
func (h *TOTPHandler) Setup2FA(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Check if TOTP already exists and is verified
	var existingTOTP user.TOTPSecret
	err := h.DB.Get(&existingTOTP, "SELECT * FROM totp_secrets WHERE user_id = $1", authUser.ID)
	if err == nil && existingTOTP.Verified {
		RespondWithError(w, http.StatusConflict, "2FA already enabled")
		return
	}

	// Generate TOTP secret
	secret := make([]byte, 20)
	_, err = rand.Read(secret)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate secret")
		return
	}
	secretString := base32.StdEncoding.EncodeToString(secret)

	// Create TOTP key for QR code
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Gachiakuta Hispano",
		AccountName: authUser.Email,
		Secret:      secret,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate TOTP key")
		return
	}

	// Generate QR code
	qrCode, err := qrcode.Encode(key.String(), qrcode.Medium, 256)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate QR code")
		return
	}

	// Save or update TOTP secret (unverified)
	now := time.Now()
	totpSecret := user.TOTPSecret{
		UserID:    authUser.ID,
		Secret:    secretString,
		Verified:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err == nil {
		// Update existing unverified TOTP
		_, err = h.DB.NamedExec(`UPDATE totp_secrets SET secret = :secret, verified = :verified, 
			updated_at = :updated_at WHERE user_id = :user_id`, totpSecret)
	} else {
		// Create new TOTP secret
		_, err = h.DB.NamedExec(`INSERT INTO totp_secrets (user_id, secret, verified, created_at, updated_at)
			VALUES (:user_id, :secret, :verified, :created_at, :updated_at)`, totpSecret)
	}
	
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to save TOTP secret")
		return
	}

	setupResponse := user.TOTPSetupResponse{
		Secret:    secretString,
		QRCode:    qrCode,
		BackupURL: key.String(),
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

	// Basic validation
	if input.Code == "" {
		RespondWithError(w, http.StatusBadRequest, "Verification code is required")
		return
	}

	// Get TOTP secret
	var totpSecret user.TOTPSecret
	err := h.DB.Get(&totpSecret, "SELECT * FROM totp_secrets WHERE user_id = $1", authUser.ID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "2FA not setup")
		return
	}

	// Validate TOTP code
	valid := totp.Validate(input.Code, totpSecret.Secret, time.Now())
	if !valid {
		RespondWithError(w, http.StatusBadRequest, "Invalid verification code")
		return
	}

	// Mark TOTP as verified and enable for user
	now := time.Now()
	_, err = h.DB.Exec("UPDATE totp_secrets SET verified = true, updated_at = $1 WHERE user_id = $2", 
		now, authUser.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to verify TOTP")
		return
	}

	_, err = h.DB.Exec("UPDATE users SET totp_enabled = true, updated_at = $1 WHERE id = $2", 
		now, authUser.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to enable 2FA")
		return
	}

	// Generate recovery codes
	recoveryCodes, err := h.generateRecoveryCodes(authUser.ID)
	if err != nil {
		// Don't fail verification for this, just log it
		RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "2FA enabled successfully",
		})
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":        "2FA enabled successfully",
		"recovery_codes": recoveryCodes,
	})
}

// Disable2FA disables TOTP 2FA for the current user
func (h *TOTPHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	now := time.Now()

	// Delete TOTP secret and recovery codes
	h.DB.Exec("DELETE FROM totp_secrets WHERE user_id = $1", authUser.ID)
	h.DB.Exec("DELETE FROM recovery_codes WHERE user_id = $1", authUser.ID)

	// Disable TOTP for user
	_, err := h.DB.Exec("UPDATE users SET totp_enabled = false, updated_at = $1 WHERE id = $2", 
		now, authUser.ID)
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

	// Check if user has 2FA enabled
	var u user.User
	err := h.DB.Get(&u, "SELECT totp_enabled FROM users WHERE id = $1", authUser.ID)
	if err != nil || !u.TOTPEnabled {
		RespondWithError(w, http.StatusBadRequest, "2FA not enabled")
		return
	}

	// Delete existing recovery codes
	h.DB.Exec("DELETE FROM recovery_codes WHERE user_id = $1", authUser.ID)

	// Generate new codes
	codes, err := h.generateRecoveryCodes(authUser.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate recovery codes")
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

	// Basic validation
	if input.Code == "" {
		RespondWithError(w, http.StatusBadRequest, "Recovery code is required")
		return
	}

	// Get and validate recovery code
	var recoveryCode user.RecoveryCode
	err := h.DB.Get(&recoveryCode, "SELECT * FROM recovery_codes WHERE user_id = $1 AND code = $2 AND used = false", 
		authUser.ID, strings.ToUpper(strings.TrimSpace(input.Code)))
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid or used recovery code")
		return
	}

	// Mark recovery code as used
	_, err = h.DB.Exec("UPDATE recovery_codes SET used = true, updated_at = $1 WHERE id = $2", 
		time.Now(), recoveryCode.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to use recovery code")
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

	// Basic validation
	if input.Code == "" {
		RespondWithError(w, http.StatusBadRequest, "TOTP code is required")
		return
	}

	// Get TOTP secret
	var totpSecret user.TOTPSecret
	err := h.DB.Get(&totpSecret, "SELECT * FROM totp_secrets WHERE user_id = $1", authUser.ID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "2FA not enabled")
		return
	}

	if !totpSecret.Verified {
		RespondWithError(w, http.StatusBadRequest, "2FA not verified")
		return
	}

	// Validate TOTP code
	valid := totp.Validate(input.Code, totpSecret.Secret, time.Now())
	if !valid {
		RespondWithError(w, http.StatusBadRequest, "Invalid TOTP code")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "TOTP code valid",
	})
}

// Helper function to generate 8 recovery codes
func (h *TOTPHandler) generateRecoveryCodes(userID uint) ([]string, error) {
	codes := make([]string, 8)
	now := time.Now()

	for i := 0; i < 8; i++ {
		bytes := make([]byte, 6)
		_, err := rand.Read(bytes)
		if err != nil {
			return nil, err
		}

		code := fmt.Sprintf("%02x%02x-%02x%02x-%02x%02x",
			bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5])
		codes[i] = strings.ToUpper(code)

		// Save to database
		recoveryCode := user.RecoveryCode{
			UserID:    userID,
			Code:      codes[i],
			Used:      false,
			CreatedAt: now,
			UpdatedAt: now,
		}

		_, err = h.DB.NamedExec(`INSERT INTO recovery_codes (user_id, code, used, created_at, updated_at)
			VALUES (:user_id, :code, :used, :created_at, :updated_at)`, recoveryCode)
		if err != nil {
			return nil, err
		}
	}

	return codes, nil
}
