package user

import (
	"time"
	
	"github.com/TommySanDev/gachiakuta-hispano/config"
)

// EmailVerification represents an email verification token
type EmailVerification struct {
	ID        uint      `json:"id" db:"id"`
	UserID    uint      `json:"user_id" db:"user_id"`
	Token     string    `json:"-" db:"token"`                           // Hidden in JSON
	Used      bool      `json:"used" db:"used"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`                      // Hidden in JSON
}

// Input structures for email verification
type VerifyEmailInput struct {
	Token string `json:"token"`
}

type ResendVerificationInput struct {
	Email string `json:"email"`
}

// CreateEmailVerificationToken creates a verification token for a user
func CreateEmailVerificationToken(userID uint) (*EmailVerification, error) {
	// Generate verification token
	token, err := generateSecureToken()
	if err != nil {
		return nil, err
	}

	// Invalidate existing tokens
	config.DB.Exec("UPDATE email_verifications SET used = true WHERE user_id = $1 AND used = false", userID)

	// Create verification token
	now := time.Now()
	verification := &EmailVerification{
		UserID:    userID,
		Token:     token,
		Used:      false,
		ExpiresAt: now.Add(24 * time.Hour), // 24 hours expiration
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `INSERT INTO email_verifications (user_id, token, used, expires_at, created_at, updated_at)
		VALUES (:user_id, :token, :used, :expires_at, :created_at, :updated_at) RETURNING id`
	
	rows, err := config.DB.NamedQuery(query, verification)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var id uint
		rows.Scan(&id)
		verification.ID = id
	}

	return verification, nil
}

// VerifyEmailWithToken verifies an email using a verification token
func VerifyEmailWithToken(token string) error {
	if token == "" {
		return ErrInvalidInput
	}

	// Get verification token
	var verification EmailVerification
	err := config.DB.Get(&verification, "SELECT * FROM email_verifications WHERE token = $1 AND used = false", token)
	if err != nil {
		return ErrInvalidToken
	}

	// Check if expired
	if time.Now().After(verification.ExpiresAt) {
		return ErrTokenExpired
	}

	// Get user
	var user User
	err = config.DB.Get(&user, "SELECT * FROM users WHERE id = $1", verification.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	// Check if already verified
	if user.EmailVerified {
		return nil // Already verified, no error
	}

	// Mark email as verified
	now := time.Now()
	_, err = config.DB.Exec("UPDATE users SET email_verified = true, updated_at = $1 WHERE id = $2", 
		now, verification.UserID)
	if err != nil {
		return err
	}

	// Mark token as used
	config.DB.Exec("UPDATE email_verifications SET used = true, updated_at = $1 WHERE id = $2", 
		now, verification.ID)

	return nil
}

// ResendEmailVerification creates and sends a new verification email
func ResendEmailVerification(email string) error {
	// Get user by email
	user, err := GetByEmail(email)
	if err != nil {
		return ErrUserNotFound // Don't reveal if user exists
	}

	// Check if already verified
	if user.EmailVerified {
		return nil // Already verified, no error
	}

	// Create new verification token
	_, err = CreateEmailVerificationToken(user.ID)
	return err
}
