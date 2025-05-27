package user

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
    "github.com/TommySanDev/gachiakuta-hispano/internal/middleware"
)

type Handler struct {
    authService     *AuthService
    crudService     *CrudService
    searchService   *SearchService
    passwordService *PasswordService
    totpService     *TOTPService
}

func NewHandler(authService *AuthService, crudService *CrudService, searchService *SearchService, passwordService *PasswordService, totpService *TOTPService) *Handler {
    return &Handler{
        authService:     authService,
        crudService:     crudService,
        searchService:   searchService,
        passwordService: passwordService,
        totpService:     totpService,
    }
}

// POST /auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    var input RegisterInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    user, err := h.authService.Register(r.Context(), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        case errors.Is(err, ErrUserAlreadyExists):
            http.Error(w, "User already exists", http.StatusConflict)
        default:
            logger.Error("Error registering user", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusCreated, user)
}

// POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    var input LoginInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    authResponse, err := h.authService.Login(r.Context(), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidCredentials):
            http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        case errors.Is(err, ErrUserInactive):
            http.Error(w, "Account inactive", http.StatusForbidden)
        default:
            logger.Error("Error during login", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, authResponse)
}

// POST /auth/magic-link
func (h *Handler) RequestMagicLink(w http.ResponseWriter, r *http.Request) {
    var input MagicLinkLoginInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    err := h.authService.GenerateMagicLink(r.Context(), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrMagicLinkDisabled):
            http.Error(w, "Magic link authentication disabled", http.StatusBadRequest)
        default:
            logger.Error("Error generating magic link", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "Magic link sent to your email",
    })
}

// GET /auth/magic-link
func (h *Handler) LoginWithMagicLink(w http.ResponseWriter, r *http.Request) {
    token := r.URL.Query().Get("token")
    if token == "" {
        http.Error(w, "Token is required", http.StatusBadRequest)
        return
    }

    authResponse, err := h.authService.LoginWithMagicLink(r.Context(), token)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidMagicLink):
            http.Error(w, "Invalid magic link", http.StatusUnauthorized)
        case errors.Is(err, ErrMagicLinkExpired):
            http.Error(w, "Magic link expired", http.StatusUnauthorized)
        case errors.Is(err, ErrUserInactive):
            http.Error(w, "Account inactive", http.StatusForbidden)
        default:
            logger.Error("Error with magic link login", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, authResponse)
}

// POST /auth/forgot-password
func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
    var input ResetPasswordRequestInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    err := h.passwordService.RequestReset(r.Context(), input)
    if err != nil {
        logger.Error("Error requesting password reset", zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "If your email exists, you will receive reset instructions",
    })
}

// GET /auth/reset-password
func (h *Handler) ValidateResetToken(w http.ResponseWriter, r *http.Request) {
    token := r.URL.Query().Get("token")
    if token == "" {
        http.Error(w, "Token is required", http.StatusBadRequest)
        return
    }

    user, err := h.passwordService.ValidateResetToken(r.Context(), token)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidResetToken):
            http.Error(w, "Invalid reset token", http.StatusBadRequest)
        case errors.Is(err, ErrResetTokenExpired):
            http.Error(w, "Reset token expired", http.StatusBadRequest)
        case errors.Is(err, ErrUserInactive):
            http.Error(w, "Account inactive", http.StatusForbidden)
        default:
            logger.Error("Error validating reset token", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, map[string]interface{}{
        "valid": true,
        "email": user.Email,
    })
}

// POST /auth/reset-password
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
    var input ResetPasswordInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    err := h.passwordService.ResetPassword(r.Context(), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidResetToken):
            http.Error(w, "Invalid reset token", http.StatusBadRequest)
        case errors.Is(err, ErrResetTokenExpired):
            http.Error(w, "Reset token expired", http.StatusBadRequest)
        case errors.Is(err, ErrUserInactive):
            http.Error(w, "Account inactive", http.StatusForbidden)
        default:
            logger.Error("Error resetting password", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "Password reset successfully",
    })
}

// POST /auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
    session, ok := middleware.GetSessionFromContext(r.Context())
    if !ok {
        http.Error(w, "Session not found", http.StatusUnauthorized)
        return
    }

    err := h.authService.Logout(r.Context(), session.ID)
    if err != nil {
        logger.Error("Error during logout", zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "Logged out successfully",
    })
}

// GET /users/me
func (h *Handler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }

    respondJSON(w, http.StatusOK, user)
}

// PUT /users/me
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }

    var input UpdateUserInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    updatedUser, err := h.crudService.Update(r.Context(), user.ID, input)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            logger.Error("Error updating user profile", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, updatedUser)
}

// PUT /users/password
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }

    var input ChangePasswordInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    err := h.passwordService.ChangePassword(r.Context(), user.ID, input)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidCredentials):
            http.Error(w, "Current password is incorrect", http.StatusBadRequest)
        default:
            logger.Error("Error changing password", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "Password changed successfully",
    })
}

// POST /users/2fa/setup
func (h *Handler) Setup2FA(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }

    setupResponse, err := h.totpService.SetupTOTP(r.Context(), user.ID)
    if err != nil {
        switch {
        case errors.Is(err, ErrTOTPAlreadyEnabled):
            http.Error(w, "2FA already enabled", http.StatusConflict)
        default:
            logger.Error("Error setting up 2FA", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, setupResponse)
}

// POST /users/2fa/verify
func (h *Handler) Verify2FA(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }

    var input VerifyTOTPInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    err := h.totpService.VerifyTOTP(r.Context(), user.ID, input.Code)
    if err != nil {
        switch {
        case errors.Is(err, ErrTOTPNotEnabled):
            http.Error(w, "2FA not setup", http.StatusBadRequest)
        case errors.Is(err, ErrInvalidTOTPCode):
            http.Error(w, "Invalid verification code", http.StatusBadRequest)
        default:
            logger.Error("Error verifying 2FA", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "2FA enabled successfully",
    })
}

// DELETE /users/2fa
func (h *Handler) Disable2FA(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }

    err := h.totpService.DisableTOTP(r.Context(), user.ID)
    if err != nil {
        logger.Error("Error disabling 2FA", zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    respondJSON(w, http.StatusOK, map[string]string{
        "message": "2FA disabled successfully",
    })
}

// POST /users/2fa/recovery-codes
func (h *Handler) GenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }

    codes, err := h.totpService.GenerateRecoveryCodes(r.Context(), user.ID)
    if err != nil {
        logger.Error("Error generating recovery codes", zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    respondJSON(w, http.StatusOK, map[string]interface{}{
        "recovery_codes": codes,
        "message":        "New recovery codes generated. Save them securely.",
    })
}

// GET /admin/users
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
    filter := UserFilter{
        Search:         r.URL.Query().Get("search"),
        Role:           r.URL.Query().Get("role"),
        SortBy:         r.URL.Query().Get("sort_by"),
        SortDir:        r.URL.Query().Get("sort_dir"),
        Page:           getIntParam(r, "page", 1),
        PageSize:       getIntParam(r, "page_size", 20),
        IncludeDeleted: r.URL.Query().Get("include_deleted") == "true",
    }

    if isActiveStr := r.URL.Query().Get("is_active"); isActiveStr != "" {
        if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
            filter.IsActive = &isActive
        }
    }

    if emailVerifiedStr := r.URL.Query().Get("email_verified"); emailVerifiedStr != "" {
        if emailVerified, err := strconv.ParseBool(emailVerifiedStr); err == nil {
            filter.EmailVerified = &emailVerified
        }
    }

    users, total, err := h.searchService.List(r.Context(), filter)
    if err != nil {
        logger.Error("Error listing users", zap.Error(err))
        http.Error(w, "Error fetching users", http.StatusInternalServerError)
        return
    }

    totalPages := (total + filter.PageSize - 1) / filter.PageSize

    response := map[string]interface{}{
        "data": users,
        "meta": map[string]interface{}{
            "total":       total,
            "page":        filter.Page,
            "page_size":   filter.PageSize,
            "total_pages": totalPages,
        },
    }

    respondJSON(w, http.StatusOK, response)
}

// GET /admin/users/{id}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    user, err := h.crudService.Get(r.Context(), uint(id))
    if err != nil {
        switch {
        case errors.Is(err, ErrUserNotFound):
            http.Error(w, "User not found", http.StatusNotFound)
        default:
            logger.Error("Error fetching user", zap.Error(err), zap.String("id", idStr))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, user)
}

// POST /admin/users
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var input CreateUserInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    user, err := h.crudService.Create(r.Context(), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        case errors.Is(err, ErrUserAlreadyExists):
            http.Error(w, "User already exists", http.StatusConflict)
        default:
            logger.Error("Error creating user", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusCreated, user)
}

// PUT /admin/users/{id}
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    var input AdminUpdateUserInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    user, err := h.crudService.AdminUpdate(r.Context(), uint(id), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrUserNotFound):
            http.Error(w, "User not found", http.StatusNotFound)
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            logger.Error("Error updating user", zap.Error(err), zap.String("id", idStr))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, user)
}

// DELETE /admin/users/{id}
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    err = h.crudService.Delete(r.Context(), uint(id))
    if err != nil {
        switch {
        case errors.Is(err, ErrUserNotFound):
            http.Error(w, "User not found", http.StatusNotFound)
        default:
            logger.Error("Error deleting user", zap.Error(err), zap.String("id", idStr))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// DELETE /admin/users/{id}/permanent
func (h *Handler) DeleteUserPermanently(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    err = h.crudService.DeletePermanently(r.Context(), uint(id))
    if err != nil {
        switch {
        case errors.Is(err, ErrUserNotFound):
            http.Error(w, "User not found", http.StatusNotFound)
        default:
            logger.Error("Error permanently deleting user", zap.Error(err), zap.String("id", idStr))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// PATCH /admin/users/{id}/restore
func (h *Handler) RestoreUser(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    err = h.crudService.Restore(r.Context(), uint(id))
    if err != nil {
        switch {
        case errors.Is(err, ErrUserNotFound):
            http.Error(w, "User not found", http.StatusNotFound)
        default:
            logger.Error("Error restoring user", zap.Error(err), zap.String("id", idStr))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
