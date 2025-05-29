package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	
	"github.com/TommySanDev/gachiakuta-hispano/internal/favorite"
	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// FavoriteHandler implements HTTP handlers for favorite operations
type FavoriteHandler struct {
	DB *sqlx.DB
}

// NewFavoriteHandler creates a new favorite handler
func NewFavoriteHandler(db *sqlx.DB) *FavoriteHandler {
	return &FavoriteHandler{DB: db}
}

// List retrieves all favorites for the current user
// GET /api/favorites
func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	var favorites []favorite.Favorite
	err := h.DB.Select(&favorites, 
		"SELECT * FROM favorites WHERE user_id = $1 ORDER BY created_at DESC", 
		authUser.ID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch favorites")
		return
	}

	// Group by entity type for easier frontend consumption
	result := map[string][]favorite.Favorite{
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

	RespondWithJSON(w, http.StatusOK, result)
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

	// Basic validation
	if newFavorite.EntityType == "" || newFavorite.EntityID == 0 {
		RespondWithError(w, http.StatusBadRequest, "Entity type and ID are required")
		return
	}

	// Normalize and validate entity type
	normalizedType := strings.ToLower(newFavorite.EntityType)
	if normalizedType != "character" && normalizedType != "chapter" && normalizedType != "vital_instrument" {
		RespondWithError(w, http.StatusBadRequest, "Invalid entity type")
		return
	}

	// Verify entity exists
	var exists bool
	var tableName string
	switch normalizedType {
	case "character":
		tableName = "characters"
	case "chapter":
		tableName = "chapters"
	case "vital_instrument":
		tableName = "vital_instruments"
	}

	err := h.DB.Get(&exists, 
		"SELECT COUNT(*) > 0 FROM "+tableName+" WHERE id = $1 AND deleted_at IS NULL", 
		newFavorite.EntityID)
	if err != nil || !exists {
		RespondWithError(w, http.StatusBadRequest, "Entity not found")
		return
	}

	// Check if already favorited
	var alreadyExists bool
	err = h.DB.Get(&alreadyExists, 
		"SELECT COUNT(*) > 0 FROM favorites WHERE user_id = $1 AND entity_type = $2 AND entity_id = $3",
		authUser.ID, normalizedType, newFavorite.EntityID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if alreadyExists {
		RespondWithError(w, http.StatusConflict, "Already favorited")
		return
	}

	// Set fields from context and normalize
	newFavorite.UserID = authUser.ID
	newFavorite.EntityType = normalizedType
	newFavorite.CreatedAt = time.Now()

	query := `INSERT INTO favorites (user_id, entity_type, entity_id, created_at)
		VALUES (:user_id, :entity_type, :entity_id, :created_at) RETURNING id`

	rows, err := h.DB.NamedQuery(query, newFavorite)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create favorite")
		return
	}
	defer rows.Close()

	if rows.Next() {
		var id uint
		rows.Scan(&id)
		newFavorite.ID = id
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

	var deleteData favorite.Favorite
	if err := json.NewDecoder(r.Body).Decode(&deleteData); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if deleteData.EntityType == "" || deleteData.EntityID == 0 {
		RespondWithError(w, http.StatusBadRequest, "Entity type and ID are required")
		return
	}

	// Normalize entity type
	normalizedType := strings.ToLower(deleteData.EntityType)

	// Remove favorite
	result, err := h.DB.Exec(
		"DELETE FROM favorites WHERE user_id = $1 AND entity_type = $2 AND entity_id = $3",
		authUser.ID, normalizedType, deleteData.EntityID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to remove favorite")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		RespondWithError(w, http.StatusNotFound, "Favorite not found")
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
