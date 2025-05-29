package handler

import (
	"encoding/json"
	"net/http"

	"github.com/TommySanDev/gachiakuta-hispano/internal/favorite"
	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// FavoriteHandler implements HTTP handlers for favorite operations
type FavoriteHandler struct{}

// NewFavoriteHandler creates a new favorite handler
func NewFavoriteHandler() *FavoriteHandler {
	return &FavoriteHandler{}
}

// List retrieves all favorites for the current user
// GET /api/favorites
func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	favorites, err := favorite.GetByUserID(authUser.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch favorites")
		return
	}

	RespondWithJSON(w, http.StatusOK, favorites)
}

// Add adds a new favorite
// POST /api/favorites
func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	var newFavorite favorite.Favorite
	if err := json.NewDecoder(r.Body).Decode(&newFavorite); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set user ID from context
	newFavorite.UserID = authUser.ID

	err := favorite.Create(&newFavorite)
	if err != nil {
		switch err {
		case favorite.ErrInvalidInput:
			RespondWithError(w, http.StatusBadRequest, "Invalid entity type or entity not found")
		case favorite.ErrAlreadyFavorited:
			RespondWithError(w, http.StatusConflict, "Already favorited")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to create favorite")
		}
		return
	}

	RespondWithJSON(w, http.StatusCreated, newFavorite)
}

// Remove removes a favorite
// DELETE /api/favorites
func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	var deleteData struct {
		EntityType string `json:"entity_type"`
		EntityID   uint   `json:"entity_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&deleteData); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := favorite.Delete(authUser.ID, deleteData.EntityType, deleteData.EntityID)
	if err != nil {
		switch err {
		case favorite.ErrInvalidInput:
			RespondWithError(w, http.StatusBadRequest, "Entity type and ID are required")
		case favorite.ErrFavoriteNotFound:
			RespondWithError(w, http.StatusNotFound, "Favorite not found")
		default:
			RespondWithError(w, http.StatusInternalServerError, "Failed to remove favorite")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListByEntity retrieves favorites for a specific entity (helper endpoint)
// GET /api/favorites/{entity_type}/{entity_id}
func (h *FavoriteHandler) ListByEntity(w http.ResponseWriter, r *http.Request) {
	// This could be useful for showing "X users favorited this" count
	// Implementation depends on your specific needs
	RespondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Feature not implemented yet",
	})
}
