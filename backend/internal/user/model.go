package user

import "time"

// User roles
const (
    RoleAdmin  = "admin"
    RoleEditor = "editor"
    RoleUser   = "user"
)

type User struct {
    ID                uint       `json:"id" db:"id"`
    Email             string     `json:"email" db:"email"`
    Username          string     `json:"username" db:"username"`
    PasswordHash      string     `json:"-" db:"password_hash"`
    FirstName         string     `json:"first_name" db:"first_name"`
    LastName          string     `json:"last_name" db:"last_name"`
    Role              string     `json:"role" db:"role"`
    IsActive          bool       `json:"is_active" db:"is_active"`
    EmailVerified     bool       `json:"email_verified" db:"email_verified"`
    MagicLinkEnabled  bool       `json:"magic_link_enabled" db:"magic_link_enabled"`
    TOTPEnabled       bool       `json:"totp_enabled" db:"totp_enabled"`
    LastLoginAt       *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
    CreatedAt         time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
    DeletedAt         *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type Session struct {
    ID        string    `json:"id" db:"id"`
    UserID    uint      `json:"user_id" db:"user_id"`
    Token     string    `json:"-" db:"token"`
    UserAgent string    `json:"user_agent" db:"user_agent"`
    IPAddress string    `json:"ip_address" db:"ip_address"`
    ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type MagicLink struct {
    ID        uint      `json:"id" db:"id"`
    UserID    uint      `json:"user_id" db:"user_id"`
    Token     string    `json:"-" db:"token"`
    Used      bool      `json:"used" db:"used"`
    ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type ResetToken struct {
    ID        uint      `json:"id" db:"id"`
    UserID    uint      `json:"user_id" db:"user_id"`
    Token     string    `json:"-" db:"token"`
    Used      bool      `json:"used" db:"used"`
    ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type TOTPSecret struct {
    ID        uint      `json:"id" db:"id"`
    UserID    uint      `json:"user_id" db:"user_id"`
    Secret    string    `json:"-" db:"secret"`
    Verified  bool      `json:"verified" db:"verified"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type RecoveryCode struct {
    ID        uint      `json:"id" db:"id"`
    UserID    uint      `json:"user_id" db:"user_id"`
    Code      string    `json:"-" db:"code"`
    Used      bool      `json:"used" db:"used"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Dto for user registration
type RegisterInput struct {
    Email             string `json:"email" validate:"required,email"`
    Username          string `json:"username" validate:"required"`
    Password          string `json:"password" validate:"required,min=8"`
    FirstName         string `json:"first_name" validate:"required"`
    LastName          string `json:"last_name" validate:"required"`
    MagicLinkEnabled  bool   `json:"magic_link_enabled"`
}

// Dto for user login
type LoginInput struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

// Dto for magic link login
type MagicLinkLoginInput struct {
    Email string `json:"email" validate:"required,email"`
}

// Dto for creating users by admin
type CreateUserInput struct {
    Email             string `json:"email" validate:"required,email"`
    Username          string `json:"username" validate:"required"`
    Password          string `json:"password" validate:"required,min=8"`
    FirstName         string `json:"first_name" validate:"required"`
    LastName          string `json:"last_name" validate:"required"`
    Role              string `json:"role" validate:"required"`
    MagicLinkEnabled  bool   `json:"magic_link_enabled"`
}

// Dto for updating user profile
type UpdateUserInput struct {
    Username         *string `json:"username,omitempty"`
    FirstName        *string `json:"first_name,omitempty"`
    LastName         *string `json:"last_name,omitempty"`
    MagicLinkEnabled *bool   `json:"magic_link_enabled,omitempty"`
}

// Dto for admin user updates
type AdminUpdateUserInput struct {
    Username         *string `json:"username,omitempty"`
    FirstName        *string `json:"first_name,omitempty"`
    LastName         *string `json:"last_name,omitempty"`
    Role             *string `json:"role,omitempty"`
    IsActive         *bool   `json:"is_active,omitempty"`
    EmailVerified    *bool   `json:"email_verified,omitempty"`
    MagicLinkEnabled *bool   `json:"magic_link_enabled,omitempty"`
}

// Dto for password change
type ChangePasswordInput struct {
    CurrentPassword string `json:"current_password" validate:"required"`
    NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// Dto for password reset request
type ResetPasswordRequestInput struct {
    Email string `json:"email" validate:"required,email"`
}

// Dto for password reset
type ResetPasswordInput struct {
    Token       string `json:"token" validate:"required"`
    NewPassword string `json:"new_password" validate:"required,min=8"`
}

// Parameters for filtering users
type UserFilter struct {
    Search         string `json:"search"`
    Role           string `json:"role,omitempty"`
    IsActive       *bool  `json:"is_active,omitempty"`
    EmailVerified  *bool  `json:"email_verified,omitempty"`
    SortBy         string `json:"sort_by,omitempty"`
    SortDir        string `json:"sort_dir,omitempty"`
    Page           int    `json:"page"`
    PageSize       int    `json:"page_size"`
    IncludeDeleted bool   `json:"include_deleted,omitempty"`
}

// Response for authentication
type AuthResponse struct {
    User         *User  `json:"user"`
    Token        string `json:"token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}
