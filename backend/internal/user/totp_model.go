package user

import "time"

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
