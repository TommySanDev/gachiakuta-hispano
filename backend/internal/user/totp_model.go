package user

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
	"time"
	
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"github.com/TommySanDev/gachiakuta-hispano/config"
	"github.com/TommySanDev/gachiakuta-hispano/internal/logger"
	"go.uber.org/zap"
)

// TOTPSecret represents a TOTP secret for 2FA
type TOTPSecret struct {
	ID        uint      `json:"id" db:"id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	Secret    string    `json:"-" db:"secret"`                          // Hidden in JSON
	Verified  bool      `json:"verified" db:"verified"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`                      // Hidden in JSON
}

// RecoveryCode represents a backup code for 2FA recovery
type RecoveryCode struct {
	ID        uint      `json:"id" db:"id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	Code      string    `json:"-" db:"code"`                            // Hidden in JSON
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`                      // Hidden in JSON
}

// Input structures for TOTP operations
type VerifyTOTPInput struct {
	Code string `json:"code"`
}

type UseRecoveryCodeInput struct {
	Code string `json:"code"`
}

// Response structures for TOTP
type TOTPSetupResponse struct {
	Secret    string `json:"secret"`
	QRCode    []byte `json:"qr_code"`
	BackupURL string `json:"backup_url"`
}

// === TOTP SETUP OPERATIONS ===

// SetupTOTP initializes TOTP 2FA for a user
func SetupTOTP(userID uint) (*TOTPSetupResponse, error) {
	log := logger.GetLogger(zap.String("component", "totp"), zap.Uint("user_id", userID))
	log.Info("Setting up TOTP for user")

	// Get user
	u, err := GetByID(userID)
	if err != nil {
		log.Error("Failed to get user", zap.Error(err))
		return nil, ErrUserNotFound
	}

	// Check if TOTP already exists and is verified
	var existingTOTP TOTPSecret
	checkErr := config.DB.Get(&existingTOTP, "SELECT * FROM totp_secrets WHERE user_id = $1", userID)
	
	if checkErr == nil && existingTOTP.Verified {
		log.Warn("TOTP already enabled for user")
		return nil, ErrTOTPAlreadyEnabled
	}

	// Generate TOTP secret
	secret := make([]byte, 20)
	_, err = rand.Read(secret)
	if err != nil {
		log.Error("Failed to generate random secret", zap.Error(err))
		return nil, err
	}
	secretString := base32.StdEncoding.EncodeToString(secret)
	log.Info("Generated TOTP secret", zap.String("secret_length", fmt.Sprintf("%d", len(secretString))))

	// Create TOTP key for QR code
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Gachiakuta Hispano",
		AccountName: u.Email,
		Secret:      secret,
	})
	if err != nil {
		log.Error("Failed to generate TOTP key", zap.Error(err))
		return nil, err
	}

	// Generate QR code
	qrCode, err := qrcode.Encode(key.String(), qrcode.Medium, 256)
	if err != nil {
		log.Error("Failed to generate QR code", zap.Error(err))
		return nil, err
	}

	// Save or update TOTP secret (unverified)
	now := time.Now()
	totpSecret := TOTPSecret{
		UserID:    userID,
		Secret:    secretString,
		Verified:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Check if we should update or insert
	if checkErr == nil {
		// Update existing unverified TOTP
		log.Info("Updating existing TOTP secret")
		_, err = config.DB.NamedExec(`UPDATE totp_secrets SET secret = :secret, verified = :verified, 
			updated_at = :updated_at WHERE user_id = :user_id`, totpSecret)
	} else {
		// Create new TOTP secret
		log.Info("Creating new TOTP secret")
		_, err = config.DB.NamedExec(`INSERT INTO totp_secrets (user_id, secret, verified, created_at, updated_at)
			VALUES (:user_id, :secret, :verified, :created_at, :updated_at)`, totpSecret)
	}
	
	if err != nil {
		log.Error("Failed to save TOTP secret to database", zap.Error(err))
		return nil, err
	}

	log.Info("TOTP secret saved successfully")

	setupResponse := &TOTPSetupResponse{
		Secret:    secretString,
		QRCode:    qrCode,
		BackupURL: key.String(),
	}

	return setupResponse, nil
}

// VerifyAndEnableTOTP verifies and enables TOTP 2FA
func VerifyAndEnableTOTP(userID uint, code string) ([]string, error) {
	log := logger.GetLogger(zap.String("component", "totp"), zap.Uint("user_id", userID))
	log.Info("Verifying TOTP code", zap.String("code_length", fmt.Sprintf("%d", len(code))))

	if code == "" {
		return nil, ErrInvalidInput
	}

	// Get TOTP secret
	var totpSecret TOTPSecret
	err := config.DB.Get(&totpSecret, "SELECT * FROM totp_secrets WHERE user_id = $1", userID)
	if err != nil {
		log.Error("Failed to get TOTP secret", zap.Error(err))
		return nil, ErrTOTPNotFound
	}

	log.Info("Found TOTP secret", zap.Bool("verified", totpSecret.Verified))

	// Validate TOTP code
	valid := totp.Validate(code, totpSecret.Secret)
	log.Info("TOTP validation result", zap.Bool("valid", valid))
	
	if !valid {
		return nil, ErrInvalidTOTPCode
	}

	// Mark TOTP as verified and enable for user
	now := time.Now()
	_, err = config.DB.Exec("UPDATE totp_secrets SET verified = true, updated_at = $1 WHERE user_id = $2", 
		now, userID)
	if err != nil {
		log.Error("Failed to mark TOTP as verified", zap.Error(err))
		return nil, err
	}

	_, err = config.DB.Exec("UPDATE users SET totp_enabled = true, updated_at = $1 WHERE id = $2", 
		now, userID)
	if err != nil {
		log.Error("Failed to enable TOTP for user", zap.Error(err))
		return nil, err
	}

	log.Info("TOTP enabled successfully")

	// Generate recovery codes
	recoveryCodes, err := generateRecoveryCodes(userID)
	if err != nil {
		log.Error("Failed to generate recovery codes", zap.Error(err))
		// Don't fail verification for this, just return empty codes
		return []string{}, nil
	}

	log.Info("Recovery codes generated", zap.Int("count", len(recoveryCodes)))
	return recoveryCodes, nil
}

// DisableTOTP disables TOTP 2FA for a user
func DisableTOTP(userID uint) error {
	log := logger.GetLogger(zap.String("component", "totp"), zap.Uint("user_id", userID))
	log.Info("Disabling TOTP for user")

	now := time.Now()

	// Delete TOTP secret and recovery codes
	config.DB.Exec("DELETE FROM totp_secrets WHERE user_id = $1", userID)
	config.DB.Exec("DELETE FROM recovery_codes WHERE user_id = $1", userID)

	// Disable TOTP for user
	_, err := config.DB.Exec("UPDATE users SET totp_enabled = false, updated_at = $1 WHERE id = $2", 
		now, userID)
	
	if err != nil {
		log.Error("Failed to disable TOTP", zap.Error(err))
	} else {
		log.Info("TOTP disabled successfully")
	}
	
	return err
}

// === TOTP VALIDATION OPERATIONS ===

// ValidateTOTP validates a TOTP code for a user
func ValidateTOTP(userID uint, code string) error {
	log := logger.GetLogger(zap.String("component", "totp"), zap.Uint("user_id", userID))
	
	if code == "" {
		return ErrInvalidInput
	}

	// Get TOTP secret
	var totpSecret TOTPSecret
	err := config.DB.Get(&totpSecret, "SELECT * FROM totp_secrets WHERE user_id = $1", userID)
	if err != nil {
		log.Error("TOTP secret not found", zap.Error(err))
		return ErrTOTPNotFound
	}

	if !totpSecret.Verified {
		log.Warn("TOTP not verified for user")
		return ErrTOTPNotEnabled
	}

	// Validate TOTP code
	valid := totp.Validate(code, totpSecret.Secret)
	if !valid {
		log.Warn("Invalid TOTP code provided")
		return ErrInvalidTOTPCode
	}

	log.Info("TOTP code validated successfully")
	return nil
}

// === RECOVERY CODE OPERATIONS ===

// GenerateNewRecoveryCodes generates new recovery codes for 2FA
func GenerateNewRecoveryCodes(userID uint) ([]string, error) {
	log := logger.GetLogger(zap.String("component", "totp"), zap.Uint("user_id", userID))
	
	// Check if user has 2FA enabled
	var u User
	err := config.DB.Get(&u, "SELECT totp_enabled FROM users WHERE id = $1", userID)
	if err != nil || !u.TOTPEnabled {
		return nil, ErrTOTPNotEnabled
	}

	// Delete existing recovery codes
	config.DB.Exec("DELETE FROM recovery_codes WHERE user_id = $1", userID)

	// Generate new codes
	codes, err := generateRecoveryCodes(userID)
	if err != nil {
		log.Error("Failed to generate new recovery codes", zap.Error(err))
		return nil, err
	}

	log.Info("New recovery codes generated", zap.Int("count", len(codes)))
	return codes, nil
}

// UseRecoveryCode validates and uses a recovery code for 2FA bypass
func UseRecoveryCode(userID uint, code string) error {
	log := logger.GetLogger(zap.String("component", "totp"), zap.Uint("user_id", userID))
	
	if code == "" {
		return ErrInvalidInput
	}

	// Get and validate recovery code
	var recoveryCode RecoveryCode
	normalizedCode := strings.ToUpper(strings.TrimSpace(code))
	err := config.DB.Get(&recoveryCode, "SELECT * FROM recovery_codes WHERE user_id = $1 AND code = $2 AND used = false", 
		userID, normalizedCode)
	if err != nil {
		log.Warn("Invalid recovery code", zap.String("code", normalizedCode))
		return ErrInvalidRecoveryCode
	}

	// Mark recovery code as used
	_, err = config.DB.Exec("UPDATE recovery_codes SET used = true, updated_at = $1 WHERE id = $2", 
		time.Now(), recoveryCode.ID)
	if err != nil {
		log.Error("Failed to mark recovery code as used", zap.Error(err))
		return err
	}

	log.Info("Recovery code used successfully")
	return nil
}

// === HELPER FUNCTIONS ===

// generateRecoveryCodes generates 8 recovery codes for a user
func generateRecoveryCodes(userID uint) ([]string, error) {
	log := logger.GetLogger(zap.String("component", "totp"), zap.Uint("user_id", userID))
	
	codes := make([]string, 8)
	now := time.Now()

	for i := 0; i < 8; i++ {
		bytes := make([]byte, 6)
		_, err := rand.Read(bytes)
		if err != nil {
			log.Error("Failed to generate random bytes for recovery code", zap.Error(err))
			return nil, err
		}

		code := fmt.Sprintf("%02x%02x-%02x%02x-%02x%02x",
			bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5])
		codes[i] = strings.ToUpper(code)

		// Save to database
		recoveryCode := RecoveryCode{
			UserID:    userID,
			Code:      codes[i],
			Used:      false,
			CreatedAt: now,
			UpdatedAt: now,
		}

		_, err = config.DB.NamedExec(`INSERT INTO recovery_codes (user_id, code, used, created_at, updated_at)
			VALUES (:user_id, :code, :used, :created_at, :updated_at)`, recoveryCode)
		if err != nil {
			log.Error("Failed to save recovery code", zap.Error(err), zap.String("code", codes[i]))
			return nil, err
		}
	}

	log.Info("Generated recovery codes successfully", zap.Int("count", len(codes)))
	return codes, nil
}
