package user

import (
	"time"
	
	"golang.org/x/crypto/bcrypt"
	"github.com/TommySanDev/gachiakuta-hispano/config"
)

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

// CRUD Operations

// GetByID retrieves a user by ID
func GetByID(id uint) (*User, error) {
	var u User
	err := config.DB.Get(&u, "SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &u, nil
}

// GetByEmail retrieves a user by email
func GetByEmail(email string) (*User, error) {
	var u User
	err := config.DB.Get(&u, "SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL", email)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &u, nil
}

// Create creates a new user
func Create(input *CreateUserInput) (*User, error) {
	// Validate input
	if input.Email == "" || input.Username == "" || input.Password == "" {
		return nil, ErrInvalidInput
	}
	if input.FirstName == "" || input.LastName == "" {
		return nil, ErrInvalidInput
	}

	// Validate role
	if input.Role != RoleAdmin && input.Role != RoleEditor && input.Role != RoleUser {
		return nil, ErrInvalidInput
	}

	// Check if user already exists
	exists, err := ExistsByEmailOrUsername(input.Email, input.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	now := time.Now()
	user := &User{
		Email:            input.Email,
		Username:         input.Username,
		PasswordHash:     string(hashedPassword),
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		Role:             input.Role,
		IsActive:         true,
		EmailVerified:    false,
		MagicLinkEnabled: input.MagicLinkEnabled,
		TOTPEnabled:      false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	query := `INSERT INTO users (
		email, username, password_hash, first_name, last_name, role,
		is_active, email_verified, magic_link_enabled, totp_enabled,
		created_at, updated_at
	) VALUES (
		:email, :username, :password_hash, :first_name, :last_name, :role,
		:is_active, :email_verified, :magic_link_enabled, :totp_enabled,
		:created_at, :updated_at
	) RETURNING id`

	rows, err := config.DB.NamedQuery(query, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var id uint
		rows.Scan(&id)
		user.ID = id
	}

	return user, nil
}

// UpdateProfile updates user profile (self-service)
func UpdateProfile(userID uint, input *UpdateProfileInput) (*User, error) {
	// Get current user
	user, err := GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	updated := false
	if input.Username != nil {
		user.Username = *input.Username
		updated = true
	}
	if input.FirstName != nil {
		user.FirstName = *input.FirstName
		updated = true
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
		updated = true
	}
	if input.MagicLinkEnabled != nil {
		user.MagicLinkEnabled = *input.MagicLinkEnabled
		updated = true
	}

	if !updated {
		return user, nil
	}

	user.UpdatedAt = time.Now()

	query := `UPDATE users SET username = :username, first_name = :first_name, 
		last_name = :last_name, magic_link_enabled = :magic_link_enabled, updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	_, err = config.DB.NamedExec(query, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// AdminUpdate updates user (admin operations)
func AdminUpdate(userID uint, input *AdminUpdateUserInput) (*User, error) {
	// Get existing user
	user, err := GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	updated := false
	if input.Username != nil {
		user.Username = *input.Username
		updated = true
	}
	if input.FirstName != nil {
		user.FirstName = *input.FirstName
		updated = true
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
		updated = true
	}
	if input.Role != nil {
		if *input.Role != RoleAdmin && *input.Role != RoleEditor && *input.Role != RoleUser {
			return nil, ErrInvalidInput
		}
		user.Role = *input.Role
		updated = true
	}
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
		updated = true
	}
	if input.EmailVerified != nil {
		user.EmailVerified = *input.EmailVerified
		updated = true
	}
	if input.MagicLinkEnabled != nil {
		user.MagicLinkEnabled = *input.MagicLinkEnabled
		updated = true
	}

	if !updated {
		return user, nil
	}

	user.UpdatedAt = time.Now()

	query := `UPDATE users SET username = :username, first_name = :first_name, last_name = :last_name,
		role = :role, is_active = :is_active, email_verified = :email_verified, 
		magic_link_enabled = :magic_link_enabled, updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	_, err = config.DB.NamedExec(query, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword changes user password
func ChangePassword(userID uint, currentPassword, newPassword string) error {
	if currentPassword == "" || newPassword == "" {
		return ErrInvalidInput
	}

	// Get current user
	user, err := GetByID(userID)
	if err != nil {
		return err
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	_, err = config.DB.Exec("UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3",
		string(hashedPassword), time.Now(), userID)
	return err
}

// Delete performs soft delete
func Delete(id uint) error {
	exists, err := Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}

	now := time.Now()
	_, err = config.DB.Exec("UPDATE users SET deleted_at = $1 WHERE id = $2", now, id)
	return err
}

// Helper functions

// ExistsByEmailOrUsername checks if user exists by email or username
func ExistsByEmailOrUsername(email, username string) (bool, error) {
	var exists bool
	err := config.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM users WHERE email = $1 OR username = $2", 
		email, username)
	return exists, err
}

// Exists checks if a user exists and is not deleted
func Exists(id uint) (bool, error) {
	var exists bool
	err := config.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM users WHERE id = $1 AND deleted_at IS NULL", id)
	return exists, err
}

// UpdateLastLogin updates the last login timestamp
func UpdateLastLogin(userID uint) error {
	_, err := config.DB.Exec("UPDATE users SET last_login_at = $1, updated_at = $1 WHERE id = $2", 
		time.Now(), userID)
	return err
}
