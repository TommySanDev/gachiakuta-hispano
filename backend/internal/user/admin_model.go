package user

import (
	"fmt"
	"strings"
	"time"
	
	"golang.org/x/crypto/bcrypt"
	"github.com/TommySanDev/gachiakuta-hispano/config"
)

// ListUsersParams parameters for listing users
type ListUsersParams struct {
	Page           int
	PageSize       int
	Search         string
	Role           string
	IncludeDeleted bool
}

// ListUsersResult result structure for paginated user list
type ListUsersResult struct {
	Data       []User `json:"data"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}

// ListUsers returns a paginated list of users (admin only)
func ListUsers(params ListUsersParams) (*ListUsersResult, error) {
	// Validate and set defaults
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	// Build query
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argPos := 1

	if !params.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	if params.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(email ILIKE $%d OR username ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", argPos, argPos+1, argPos+2, argPos+3))
		searchTerm := "%" + params.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
		argPos += 4
	}

	if params.Role != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("role = $%d", argPos))
		args = append(args, params.Role)
		argPos++
	}

	whereClause := "WHERE " + strings.Join(whereClauses, " AND ")

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM users " + whereClause
	err := config.DB.Get(&total, countQuery, args...)
	if err != nil {
		return nil, err
	}

	// Get users
	offset := (params.Page - 1) * params.PageSize
	query := fmt.Sprintf(`SELECT id, email, username, first_name, last_name, role, is_active, 
		email_verified, magic_link_enabled, totp_enabled, last_login_at, created_at, updated_at, deleted_at
		FROM users %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, whereClause, argPos, argPos+1)
	args = append(args, params.PageSize, offset)

	var users []User
	err = config.DB.Select(&users, query, args...)
	if err != nil {
		return nil, err
	}

	totalPages := (total + params.PageSize - 1) / params.PageSize

	return &ListUsersResult{
		Data:       users,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// DeletePermanently permanently deletes a user and related data (admin only)
func DeletePermanently(id uint) error {
	// Delete in transaction due to foreign key constraints
	tx, err := config.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete in proper order due to foreign key constraints
	tx.Exec("DELETE FROM recovery_codes WHERE user_id = $1", id)
	tx.Exec("DELETE FROM totp_secrets WHERE user_id = $1", id)
	tx.Exec("DELETE FROM reset_tokens WHERE user_id = $1", id)
	tx.Exec("DELETE FROM magic_links WHERE user_id = $1", id)
	
	result, err := tx.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}

	return tx.Commit()
}

// Restore restores a soft deleted user (admin only)
func Restore(id uint) error {
	result, err := config.DB.Exec("UPDATE users SET deleted_at = NULL, updated_at = $1 WHERE id = $2 AND deleted_at IS NOT NULL", 
		time.Now(), id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// Register creates a new user (public registration)
func Register(input *RegisterInput) (*User, error) {
	// Validate input
	if input.Email == "" || input.Username == "" || input.Password == "" {
		return nil, ErrInvalidInput
	}
	if input.FirstName == "" || input.LastName == "" {
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

	// Create user with default role
	now := time.Now()
	user := &User{
		Email:            input.Email,
		Username:         input.Username,
		PasswordHash:     string(hashedPassword),
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		Role:             RoleUser, // Default role for registration
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

// Authenticate validates user credentials
func Authenticate(email, password string) (*User, error) {
	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	// Get user by email
	user, err := GetByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials // Don't reveal if user exists
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
