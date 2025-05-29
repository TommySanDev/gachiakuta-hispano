package user

import (
	"crypto/rand"
	"encoding/hex"
	"time"
	
	"golang.org/x/crypto/bcrypt"
	"github.com/TommySanDev/gachiakuta-hispano/config"
)

// MagicLink represents a passwordless authentication token
type MagicLink struct {
	ID        uint      `json:"id" db:"id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	Token     string    `json:"-" db:"token"`                           // Hidden in JSON
	Used      bool      `json:"used" db:"used"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`                      // Hidden in JSON
}

// ResetToken represents a password reset token
type ResetToken struct {
	ID        uint      `json:"id" db:"id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	Token     string    `json:"-" db:"token"`                           // Hidden in JSON
	Used      bool      `json:"used" db:"used"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`                      // Hidden in JSON
}

// Helper methods for MagicLink
func (ml *MagicLink) IsExpired() bool {
	return time.Now().After(ml.ExpiresAt)
}

func (ml *MagicLink) IsUsed() bool {
	return ml.Used
}

func (ml *MagicLink) IsValid() bool {
	return !ml.IsUsed() && !ml.IsExpired()
}

// Helper methods for ResetToken
func (rt *ResetToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

func (rt *ResetToken) IsUsed() bool {
	return rt.Used
}

func (rt *ResetToken) IsValid() bool {
	return !rt.IsUsed() && !rt.IsExpired()
}

// Input structures for auth operations
type MagicLinkLoginInput struct {
	Email string `json:"email"`
}

type ResetPasswordRequestInput struct {
	Email string `json:"email"`
}

type ResetPasswordInput struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// Response structures
type AuthResponse struct {
	User      *User  `json:"user"`
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
}

// === MAGIC LINK OPERATIONS ===

// CreateMagicLink creates a magic link for passwordless authentication
func CreateMagicLink(email string) (*MagicLink, *User, error) {
	// Get user by email
	u, err := GetByEmail(email)
	if err != nil {
		return nil, nil, ErrUserNotFound // Don't reveal if user exists
	}

	// Check if user is active and has magic links enabled
	if !u.IsActive || !u.MagicLinkEnabled {
		return nil, nil, ErrMagicLinkDisabled
	}

	// Generate magic link token
	token, err := generateSecureToken()
	if err != nil {
		return nil, nil, err
	}

	// Create magic link
	now := time.Now()
	magicLink := &MagicLink{
		UserID:    u.ID,
		Token:     token,
		Used:      false,
		ExpiresAt: now.Add(15 * time.Minute), // 15 minutes expiration
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `INSERT INTO magic_links (user_id, token, used, expires_at, created_at, updated_at)
		VALUES (:user_id, :token, :used, :expires_at, :created_at, :updated_at)`
	
	_, err = config.DB.NamedExec(query, magicLink)
	if err != nil {
		return nil, nil, err
	}

	return magicLink, u, nil
}

// ValidateAndUseMagicLink validates a magic link token and marks it as used
func ValidateAndUseMagicLink(token string) (*User, error) {
	if token == "" {
		return nil, ErrInvalidMagicLink
	}

	// Get magic link
	var ml MagicLink
	err := config.DB.Get(&ml, "SELECT * FROM magic_links WHERE token = $1 AND used = false", token)
	if err != nil {
		return nil, ErrInvalidMagicLink
	}

	// Check if expired
	if ml.IsExpired() {
		return nil, ErrMagicLinkExpired
	}

	// Mark as used
	_, err = config.DB.Exec("UPDATE magic_links SET used = true, updated_at = $1 WHERE id = $2", 
		time.Now(), ml.ID)
	if err != nil {
		return nil, err
	}

	// Get user
	u, err := GetByID(ml.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check if user is still active
	if !u.IsActive {
		return nil, ErrUserInactive
	}

	// Update last login
	UpdateLastLogin(u.ID)

	return u, nil
}

// === PASSWORD RESET OPERATIONS ===

// CreatePasswordResetToken creates a password reset token
func CreatePasswordResetToken(email string) (*ResetToken, *User, error) {
	// Get user by email
	u, err := GetByEmail(email)
	if err != nil {
		return nil, nil, ErrUserNotFound // Don't reveal if user exists
	}

	// Check if user is active
	if !u.IsActive {
		return nil, nil, ErrUserInactive
	}

	// Invalidate existing reset tokens
	config.DB.Exec("UPDATE reset_tokens SET used = true WHERE user_id = $1 AND used = false", u.ID)

	// Generate reset token
	token, err := generateSecureToken()
	if err != nil {
		return nil, nil, err
	}

	// Create reset token
	now := time.Now()
	resetToken := &ResetToken{
		UserID:    u.ID,
		Token:     token,
		Used:      false,
		ExpiresAt: now.Add(1 * time.Hour), // 1 hour expiration
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `INSERT INTO reset_tokens (user_id, token, used, expires_at, created_at, updated_at)
		VALUES (:user_id, :token, :used, :expires_at, :created_at, :updated_at)`
	
	_, err = config.DB.NamedExec(query, resetToken)
	if err != nil {
		return nil, nil, err
	}

	return resetToken, u, nil
}

// ValidateResetToken checks if a password reset token is valid
func ValidateResetToken(token string) (*User, error) {
	if token == "" {
		return nil, ErrInvalidResetToken
	}

	// Get reset token
	var rt ResetToken
	err := config.DB.Get(&rt, "SELECT * FROM reset_tokens WHERE token = $1 AND used = false", token)
	if err != nil {
		return nil, ErrInvalidResetToken
	}

	// Check if expired
	if rt.IsExpired() {
		return nil, ErrResetTokenExpired
	}

	// Get user to return email
	u, err := GetByID(rt.UserID)
	if err != nil {
		return nil, ErrInvalidResetToken
	}

	if !u.IsActive {
		return nil, ErrInvalidResetToken
	}

	return u, nil
}

// ResetPasswordWithToken resets user password using reset token
func ResetPasswordWithToken(token, newPassword string) error {
	if token == "" || newPassword == "" {
		return ErrInvalidInput
	}

	// Get reset token
	var rt ResetToken
	err := config.DB.Get(&rt, "SELECT * FROM reset_tokens WHERE token = $1 AND used = false", token)
	if err != nil {
		return ErrInvalidResetToken
	}

	// Check if expired
	if rt.IsExpired() {
		return ErrResetTokenExpired
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user password
	now := time.Now()
	_, err = config.DB.Exec("UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3", 
		string(hashedPassword), now, rt.UserID)
	if err != nil {
		return err
	}

	// Mark token as used
	config.DB.Exec("UPDATE reset_tokens SET used = true, updated_at = $1 WHERE id = $2", now, rt.ID)

	// Invalidate all other reset tokens for this user
	config.DB.Exec("UPDATE reset_tokens SET used = true WHERE user_id = $1 AND used = false", rt.UserID)

	return nil
}

// === HELPER FUNCTIONS ===

// generateSecureToken generates a secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
