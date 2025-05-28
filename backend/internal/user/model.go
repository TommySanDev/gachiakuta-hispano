package user

import "time"

// User roles
const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleUser   = "user"
)

// User represents a user in the system
type User struct {
	ID                uint       `json:"id" db:"id"`
	Email             string     `json:"email" db:"email"`
	Username          string     `json:"username" db:"username"`
	PasswordHash      string     `json:"-" db:"password_hash"`           // Hidden in JSON
	FirstName         string     `json:"first_name" db:"first_name"`
	LastName          string     `json:"last_name" db:"last_name"`
	Role              string     `json:"role" db:"role"`
	IsActive          bool       `json:"is_active" db:"is_active"`
	EmailVerified     bool       `json:"email_verified" db:"email_verified"`
	MagicLinkEnabled  bool       `json:"magic_link_enabled" db:"magic_link_enabled"`
	TOTPEnabled       bool       `json:"totp_enabled" db:"totp_enabled"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"-" db:"updated_at"`               // Hidden in JSON
	DeletedAt         *time.Time `json:"-" db:"deleted_at"`               // Hidden in JSON
}

// Helper methods for User
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	return u.FirstName + " " + u.LastName
}

func (u *User) HasRole(role string) bool {
	return u.Role == role
}

func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

func (u *User) IsEditor() bool {
	return u.HasRole(RoleEditor)
}

func (u *User) CanAccess(requiredRole string) bool {
	switch requiredRole {
	case RoleUser:
		return true // All authenticated users can access user level
	case RoleEditor:
		return u.IsEditor() || u.IsAdmin()
	case RoleAdmin:
		return u.IsAdmin()
	default:
		return false
	}
}

// Input structures for API operations - Security boundary
type RegisterInput struct {
	Email             string `json:"email"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	MagicLinkEnabled  bool   `json:"magic_link_enabled"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// Admin-specific inputs
type CreateUserInput struct {
	Email             string `json:"email"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Role              string `json:"role"`
	MagicLinkEnabled  bool   `json:"magic_link_enabled"`
}

type UpdateProfileInput struct {
	Username         *string `json:"username,omitempty"`
	FirstName        *string `json:"first_name,omitempty"`
	LastName         *string `json:"last_name,omitempty"`
	MagicLinkEnabled *bool   `json:"magic_link_enabled,omitempty"`
}

type AdminUpdateUserInput struct {
	Username         *string `json:"username,omitempty"`
	FirstName        *string `json:"first_name,omitempty"`
	LastName         *string `json:"last_name,omitempty"`
	Role             *string `json:"role,omitempty"`
	IsActive         *bool   `json:"is_active,omitempty"`
	EmailVerified    *bool   `json:"email_verified,omitempty"`
	MagicLinkEnabled *bool   `json:"magic_link_enabled,omitempty"`
}
