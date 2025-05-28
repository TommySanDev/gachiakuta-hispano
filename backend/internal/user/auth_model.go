package user

import "time"

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
