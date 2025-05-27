package auth

import "time"

// User roles
const (
    RoleAdmin  = "admin"
    RoleEditor = "editor"
    RoleUser   = "user"
)

// Context keys for authentication data
type ContextKey string

const (
    UserContextKey    ContextKey = "user"
    SessionContextKey ContextKey = "session"
)

// User represents an authenticated user entity
type User struct {
    ID                uint       `json:"id" db:"id"`
    Email             string     `json:"email" db:"email"`
    Username          string     `json:"username" db:"username"`
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
}

// Session represents an active user session
type Session struct {
    ID        string    `json:"id" db:"id"`
    UserID    uint      `json:"user_id" db:"user_id"`
    UserAgent string    `json:"user_agent" db:"user_agent"`
    IPAddress string    `json:"ip_address" db:"ip_address"`
    ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

// IsExpired checks if session has expired
func (s *Session) IsExpired() bool {
    return time.Now().After(s.ExpiresAt)
}

// FullName returns the user's full name
func (u *User) FullName() string {
    if u.FirstName == "" && u.LastName == "" {
        return u.Username
    }
    return u.FirstName + " " + u.LastName
}

// HasRole checks if user has specific role
func (u *User) HasRole(role string) bool {
    return u.Role == role
}

// IsAdmin checks if user is an admin
func (u *User) IsAdmin() bool {
    return u.HasRole(RoleAdmin)
}

// IsEditor checks if user is an editor
func (u *User) IsEditor() bool {
    return u.HasRole(RoleEditor)
}

// CanAccess checks if user can access given role level
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
