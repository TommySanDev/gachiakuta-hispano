package comment

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"
    "github.com/TommySanDev/gachiakuta-hispano/internal/auth"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
    "go.uber.org/zap"
)

// Handler provides HTTP handlers for comment operations
type Handler struct {
    crudService *CrudService
}

// NewHandler creates a new comment handler
func NewHandler(crudService *CrudService) *Handler {
    return &Handler{crudService: crudService}
}

// GET /comments/chapter/{id}
func (h *Handler) ListByChapter(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid chapter ID", http.StatusBadRequest)
        return
    }

    includeDeleted := r.URL.Query().Get("include_deleted") == "true"

    comments, err := h.crudService.ListByChapter(r.Context(), uint(id), includeDeleted)
    if err != nil {
        logger.Error("Error listing comments", zap.Error(err))
        http.Error(w, "Error fetching comments", http.StatusInternalServerError)
        return
    }

    respondJSON(w, http.StatusOK, comments)
}

// POST /comments
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    user, ok := auth.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var input CreateCommentInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }

    comment, err := h.crudService.Create(r.Context(), user.ID, input)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            logger.Error("Error creating comment", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusCreated, comment)
}

// PUT /comments/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
    user, ok := auth.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    var input UpdateCommentInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }

    comment, err := h.crudService.Update(r.Context(), user.ID, uint(id), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrForbidden):
            http.Error(w, err.Error(), http.StatusForbidden)
        default:
            logger.Error("Error updating comment", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, comment)
}

// DELETE /comments/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
    user, ok := auth.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    err = h.crudService.Delete(r.Context(), user.ID, uint(id))
    if err != nil {
        switch {
        case errors.Is(err, ErrForbidden):
            http.Error(w, "Forbidden", http.StatusForbidden)
        default:
            logger.Error("Error deleting comment", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// DELETE /admin/comments/{id}/permanent
func (h *Handler) DeletePermanently(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    err = h.crudService.DeletePermanently(r.Context(), uint(id))
    if err != nil {
        logger.Error("Error permanently deleting comment", zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
