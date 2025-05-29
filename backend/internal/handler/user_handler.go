package handler

import (
	"encoding/json"
	"net/http"

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
