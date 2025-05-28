package favorite

import (
    "encoding/json"
    "errors"
    "net/http"

    "github.com/TommySanDev/gachiakuta-hispano/internal/auth"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
    "go.uber.org/zap"
)

// Handler provides HTTP handlers for favorite operations
type Handler struct {
    crudService *CrudService
}

// NewHandler creates a new favorite handler
func NewHandler(crudService *CrudService) *Handler {
    return &Handler{crudService: crudService}
}

// POST /favorites
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
    user, ok := auth.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var input CreateFavoriteInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    fav, err := h.crudService.Add(r.Context(), user.ID, input)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        case errors.Is(err, ErrAlreadyFavorited):
            http.Error(w, "Already favorited", http.StatusConflict)
        default:
            logger.Error("Error adding favorite", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusCreated, fav)
}

// DELETE /favorites
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
    user, ok := auth.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var input DeleteFavoriteInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    if err := h.crudService.Remove(r.Context(), user.ID, input); err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        case errors.Is(err, ErrFavoriteNotFound):
            http.Error(w, "Favorite not found", http.StatusNotFound)
        default:
            logger.Error("Error removing favorite", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// GET /favorites
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
    user, ok := auth.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    favorites, err := h.crudService.List(r.Context(), user.ID)
    if err != nil {
        logger.Error("Error listing favorites", zap.Error(err))
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Optional: split by type for profile UI
    result := map[string][]*Favorite{
        "characters":        {},
        "chapters":          {},
        "vital_instruments": {},
    }

    for _, fav := range favorites {
        switch fav.EntityType {
        case "character":
            result["characters"] = append(result["characters"], fav)
        case "chapter":
            result["chapters"] = append(result["chapters"], fav)
        case "vital_instrument":
            result["vital_instruments"] = append(result["vital_instruments"], fav)
        }
    }

    respondJSON(w, http.StatusOK, result)
}
