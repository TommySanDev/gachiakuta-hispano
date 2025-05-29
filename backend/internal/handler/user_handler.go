package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/jmoiron/sqlx"

	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// UserHandler handles user management and profile operations
type UserHandler struct {
	DB *sqlx.DB
}

// NewUserHandler creates a new user management handler
func NewUserHandler(db *sqlx.DB) *UserHandler {
	return &UserHandler{DB: db}
}

// GetCurrentUser returns the current authenticated user
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Get fresh user data from database
	var u user.User
	err := h.DB.Get(&u, "SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL", authUser.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	RespondWithJSON(w, http.StatusOK, u)
}

// UpdateProfile updates the current user's profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	var input user.UpdateProfileInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get current user
	var u user.User
	err := h.DB.Get(&u, "SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL", authUser.ID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Update fields if provided
	updated := false
	if input.Username != nil {
		u.Username = *input.Username
		updated = true
	}
	if input.FirstName != nil {
		u.FirstName = *input.FirstName
		updated = true
	}
	if input.LastName != nil {
		u.LastName = *input.LastName
		updated = true
	}
	if input.MagicLinkEnabled != nil {
		u.MagicLinkEnabled = *input.MagicLinkEnabled
		updated = true
	}

	if !updated {
		RespondWithJSON(w, http.StatusOK, u)
		return
	}

	u.UpdatedAt = time.Now()

	query := `UPDATE users SET username = :username, first_name = :first_name, 
		last_name = :last_name, magic_link_enabled = :magic_link_enabled, updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	_, err = h.DB.NamedExec(query, u)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	RespondWithJSON(w, http.StatusOK, u)
}

// ChangePassword changes the current user's password
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	var input user.ChangePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if input.CurrentPassword == "" || input.NewPassword == "" {
		RespondWithError(w, http.StatusBadRequest, "Current password and new password are required")
		return
	}

	// Get current user
	var u user.User
	err := h.DB.Get(&u, "SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL", authUser.ID)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.CurrentPassword))
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Current password is incorrect")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Update password
	_, err = h.DB.Exec("UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3",
		string(hashedPassword), time.Now(), u.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// Admin operations below

// ListUsers returns a paginated list of users (admin only)
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := getIntParam(r, "page", 1)
	pageSize := getIntParam(r, "page_size", 20)
	search := r.URL.Query().Get("search")
	role := r.URL.Query().Get("role")
	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	if pageSize > 100 {
		pageSize = 100
	}

	// Build query
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argPos := 1

	if !includeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(email ILIKE $%d OR username ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", argPos, argPos+1, argPos+2, argPos+3))
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
		argPos += 4
	}

	if role != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("role = $%d", argPos))
		args = append(args, role)
		argPos++
	}

	whereClause := "WHERE " + strings.Join(whereClauses, " AND ")

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM users " + whereClause
	err := h.DB.Get(&total, countQuery, args...)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to count users")
		return
	}

	// Get users
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT id, email, username, first_name, last_name, role, is_active, 
		email_verified, magic_link_enabled, totp_enabled, last_login_at, created_at, updated_at, deleted_at
		FROM users %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, whereClause, argPos, argPos+1)
	args = append(args, pageSize, offset)

	var users []user.User
	err = h.DB.Select(&users, query, args...)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	totalPages := (total + pageSize - 1) / pageSize

	response := map[string]interface{}{
		"data": users,
		"meta": map[string]interface{}{
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	}

	RespondWithJSON(w, http.StatusOK, response)
}

// GetUser returns a specific user (admin only)
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var u user.User
	err = h.DB.Get(&u, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	RespondWithJSON(w, http.StatusOK, u)
}

// CreateUser creates a new user (admin only)
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input user.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if input.Email == "" || input.Username == "" || input.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Email, username and password are required")
	}
	if input.FirstName == "" || input.LastName == "" {
		RespondWithError(w, http.StatusBadRequest, "First name and last name are required")
		return
	}

	// Validate role
	if input.Role != user.RoleAdmin && input.Role != user.RoleEditor && input.Role != user.RoleUser {
		RespondWithError(w, http.StatusBadRequest, "Invalid role")
		return
	}

	// Check if user already exists
	var exists bool
	err := h.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM users WHERE email = $1 OR username = $2", 
		input.Email, input.Username)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if exists {
		RespondWithError(w, http.StatusConflict, "User already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Create user
	now := time.Now()
	newUser := user.User{
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

	rows, err := h.DB.NamedQuery(query, newUser)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}
	defer rows.Close()

	if rows.Next() {
		var id uint
		rows.Scan(&id)
		newUser.ID = id
	}

	RespondWithJSON(w, http.StatusCreated, newUser)
}

// UpdateUser updates a user (admin only)
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var input user.AdminUpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing user
	var u user.User
	err = h.DB.Get(&u, "SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Update fields if provided
	updated := false
	if input.Username != nil {
		u.Username = *input.Username
		updated = true
	}
	if input.FirstName != nil {
		u.FirstName = *input.FirstName
		updated = true
	}
	if input.LastName != nil {
		u.LastName = *input.LastName
		updated = true
	}
	if input.Role != nil {
		if *input.Role != user.RoleAdmin && *input.Role != user.RoleEditor && *input.Role != user.RoleUser {
			RespondWithError(w, http.StatusBadRequest, "Invalid role")
			return
		}
		u.Role = *input.Role
		updated = true
	}
	if input.IsActive != nil {
		u.IsActive = *input.IsActive
		updated = true
	}
	if input.EmailVerified != nil {
		u.EmailVerified = *input.EmailVerified
		updated = true
	}
	if input.MagicLinkEnabled != nil {
		u.MagicLinkEnabled = *input.MagicLinkEnabled
		updated = true
	}

	if !updated {
		RespondWithJSON(w, http.StatusOK, u)
		return
	}

	u.UpdatedAt = time.Now()

	query := `UPDATE users SET username = :username, first_name = :first_name, last_name = :last_name,
		role = :role, is_active = :is_active, email_verified = :email_verified, 
		magic_link_enabled = :magic_link_enabled, updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	_, err = h.DB.NamedExec(query, u)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	RespondWithJSON(w, http.StatusOK, u)
}

// DeleteUser soft deletes a user (admin only)
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Check if user exists
	var exists bool
	err = h.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM users WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil || !exists {
		RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	// Soft delete
	now := time.Now()
	_, err = h.DB.Exec("UPDATE users SET deleted_at = $1 WHERE id = $2", now, id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteUserPermanently permanently deletes a user (admin only)
func (h *UserHandler) DeleteUserPermanently(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Permanently delete user and related data in transaction
	tx, err := h.DB.Beginx()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer tx.Rollback()

	// Delete in proper order due to foreign key constraints
	tx.Exec("DELETE FROM recovery_codes WHERE user_id = $1", id)
	tx.Exec("DELETE FROM totp_secrets WHERE user_id = $1", id)
	tx.Exec("DELETE FROM reset_tokens WHERE user_id = $1", id)
	tx.Exec("DELETE FROM magic_links WHERE user_id = $1", id)
	
	result, err := tx.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	tx.Commit()
	w.WriteHeader(http.StatusNoContent)
}

// RestoreUser restores a soft deleted user (admin only)
func (h *UserHandler) RestoreUser(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Restore user
	result, err := h.DB.Exec("UPDATE users SET deleted_at = NULL, updated_at = $1 WHERE id = $2 AND deleted_at IS NOT NULL", 
		time.Now(), id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to restore user")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		RespondWithError(w, http.StatusNotFound, "User not found or not deleted")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
