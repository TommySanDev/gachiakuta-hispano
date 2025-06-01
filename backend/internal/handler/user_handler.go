package handler

import (
	"encoding/json"
	"net/http"

	"github.com/TommySanDev/gachiakuta-hispano/internal/logger"
	"go.uber.org/zap"
	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// UserHandler handles user management and profile operations
type UserHandler struct{}

// NewUserHandler creates a new user management handler
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// === USER PROFILE OPERATIONS ===

// GetCurrentUser returns the current authenticated user
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Get fresh user data from database
	u, err := user.GetByID(authUser.ID)
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

	updatedUser, err := user.UpdateProfile(authUser.ID, &input)
	if err != nil {
		if err == user.ErrUserNotFound {
			RespondWithError(w, http.StatusNotFound, "User not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to update profile")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, updatedUser)
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

	err := user.ChangePassword(authUser.ID, input.CurrentPassword, input.NewPassword)
	if err != nil {
		switch err {
		case user.ErrInvalidInput:
			RespondWithError(w, http.StatusBadRequest, "Current password and new password are required")
		case user.ErrInvalidCredentials:
			RespondWithError(w, http.StatusBadRequest, "Current password is incorrect")
		case user.ErrUserNotFound:
			RespondWithError(w, http.StatusNotFound, "User not found")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to update password")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully",
	})
}

// === ADMIN OPERATIONS ===

// ListUsers returns a paginated list of users (admin only)
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	params := user.ListUsersParams{
		Page:           GetIntParam(r, "page", 1),
		PageSize:       GetIntParam(r, "page_size", 20),
		Search:         r.URL.Query().Get("search"),
		Role:           r.URL.Query().Get("role"),
		IncludeDeleted: r.URL.Query().Get("include_deleted") == "true",
	}

	result, err := user.ListUsers(params)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	RespondWithJSON(w, http.StatusOK, result)
}

// GetUser returns a specific user (admin only)
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	u, err := user.GetByID(id)
	if err != nil {
		if err == user.ErrUserNotFound {
			RespondWithError(w, http.StatusNotFound, "User not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to get user")
		}
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

	newUser, err := user.Create(&input)
	if err != nil {
		switch err {
		case user.ErrInvalidInput:
			RespondWithError(w, http.StatusBadRequest, "Invalid input data")
		case user.ErrUserAlreadyExists:
			RespondWithError(w, http.StatusConflict, "User already exists")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		}
		return
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

	updatedUser, err := user.AdminUpdate(id, &input)
	if err != nil {
		switch err {
		case user.ErrUserNotFound:
			RespondWithError(w, http.StatusNotFound, "User not found")
		case user.ErrInvalidInput:
			RespondWithError(w, http.StatusBadRequest, "Invalid input data")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to update user")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, updatedUser)
}

// DeleteUser soft deletes a user (admin only)
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	err = user.Delete(id)
	if err != nil {
		if err == user.ErrUserNotFound {
			RespondWithError(w, http.StatusNotFound, "User not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
		}
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

	err = user.DeletePermanently(id)
	if err != nil {
		if err == user.ErrUserNotFound {
			RespondWithError(w, http.StatusNotFound, "User not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to delete user permanently")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreUser restores a soft deleted user (admin only)
func (h *UserHandler) RestoreUser(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	err = user.Restore(id)
	if err != nil {
		if err == user.ErrUserNotFound {
			RespondWithError(w, http.StatusNotFound, "User not found or not deleted")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to restore user")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AdminResetPassword generates a temporary password for a user (admin only)
func (h *UserHandler) AdminResetPassword(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Generate temporary password
	tempPassword, err := user.AdminResetPassword(id)
	if err != nil {
		switch err {
		case user.ErrUserNotFound:
			RespondWithError(w, http.StatusNotFound, "User not found")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to reset password")
		}
		return
	}

	// Return the temporary password to admin
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":          "Password reset successfully",
		"temporary_password": tempPassword,
		"instructions":      "Please provide this temporary password to the user. They should change it on first login.",
	})
}

// AdminDisable2FA disables TOTP 2FA for any user (admin only)
func (h *UserHandler) AdminDisable2FA(w http.ResponseWriter, r *http.Request) {
	log := logger.GetLogger(zap.String("component", "admin-handler"))
	
	// Get target user ID from URL
	userID, err := GetIDParam(r)
	if err != nil {
		log.Error("Invalid user ID", zap.Error(err))
		RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get admin user from context for logging
	adminUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		log.Error("Admin user not found in context")
		RespondWithError(w, http.StatusUnauthorized, "Admin authentication required")
		return
	}

	log.Info("Admin disabling 2FA for user", 
		zap.Uint("target_user_id", userID),
		zap.Uint("admin_id", adminUser.ID),
		zap.String("admin_username", adminUser.Username))

	// Check if target user exists and has 2FA enabled
	targetUser, err := user.GetByID(userID)
	if err != nil {
		log.Error("Target user not found", zap.Error(err), zap.Uint("user_id", userID))
		RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	if !targetUser.TOTPEnabled {
		log.Warn("Target user does not have 2FA enabled", zap.Uint("user_id", userID))
		RespondWithError(w, http.StatusBadRequest, "User does not have 2FA enabled")
		return
	}

	// Disable 2FA for the target user
	err = user.AdminDisableTOTP(userID)
	if err != nil {
		log.Error("Failed to disable 2FA for user", 
			zap.Error(err), 
			zap.Uint("target_user_id", userID),
			zap.Uint("admin_id", adminUser.ID))
		RespondWithError(w, http.StatusInternalServerError, "Failed to disable 2FA")
		return
	}

	log.Info("Admin successfully disabled 2FA for user", 
		zap.Uint("target_user_id", userID),
		zap.String("target_username", targetUser.Username),
		zap.Uint("admin_id", adminUser.ID),
		zap.String("admin_username", adminUser.Username))

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "2FA disabled successfully for user",
		"user": map[string]interface{}{
			"id":       targetUser.ID,
			"username": targetUser.Username,
			"email":    targetUser.Email,
		},
		"disabled_by": map[string]interface{}{
			"id":       adminUser.ID,
			"username": adminUser.Username,
		},
	})
}
