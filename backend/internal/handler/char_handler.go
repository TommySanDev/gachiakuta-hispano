package handler

import (
	"encoding/json"
	"net/http"
	"time"
	
	"github.com/jmoiron/sqlx"
	"github.com/TommySanDev/gachiakuta-hispano/internal/character"
)

// CharacterHandler implements HTTP handlers for character operations
type CharacterHandler struct {
	DB *sqlx.DB
}

// NewCharacterHandler creates a new character handler
func NewCharacterHandler(db *sqlx.DB) *CharacterHandler {
	return &CharacterHandler{DB: db}
}

// GetAll retrieves all characters
func (h *CharacterHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	var characters []character.Character
	
	err := h.DB.Select(&characters, "SELECT * FROM characters WHERE deleted_at IS NULL ORDER BY created_at DESC")
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch characters")
		return
	}
	
	RespondWithJSON(w, http.StatusOK, characters)
}

// GetByID retrieves a character by ID
func (h *CharacterHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	var char character.Character
	err = h.DB.Get(&char, "SELECT * FROM characters WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Character not found")
		return
	}
	
	RespondWithJSON(w, http.StatusOK, char)
}

// Create creates a new character
func (h *CharacterHandler) Create(w http.ResponseWriter, r *http.Request) {
	var char character.Character
	
	if err := json.NewDecoder(r.Body).Decode(&char); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Basic validation
	if char.Name == "" {
		RespondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}
	
	// Set timestamps
	now := time.Now()
	char.CreatedAt = now
	char.UpdatedAt = now
	
	query := `INSERT INTO characters (
		name, name_japanese, main_image, description, species, gender, age,
		height, status, affiliation, occupation, birth_date, birth_place,
		relatives, first_appearance, created_at, updated_at
	) VALUES (
		:name, :name_japanese, :main_image, :description, :species, :gender, :age,
		:height, :status, :affiliation, :occupation, :birth_date, :birth_place,
		:relatives, :first_appearance, :created_at, :updated_at
	) RETURNING id`
	
	rows, err := h.DB.NamedQuery(query, char)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create character")
		return
	}
	defer rows.Close()
	
	if rows.Next() {
		var id uint
		rows.Scan(&id)
		char.ID = id
	}
	
	RespondWithJSON(w, http.StatusCreated, char)
}

// Update updates an existing character
func (h *CharacterHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	var char character.Character
	if err := json.NewDecoder(r.Body).Decode(&char); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Check if character exists
	var exists bool
	err = h.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM characters WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil || !exists {
		RespondWithError(w, http.StatusNotFound, "Character not found")
		return
	}
	
	// Set ID and update timestamp
	char.ID = id
	char.UpdatedAt = time.Now()
	
	query := `UPDATE characters SET
		name = :name, 
		name_japanese = :name_japanese,
		main_image = :main_image,
		description = :description,
		species = :species,
		gender = :gender,
		age = :age,
		height = :height,
		status = :status,
		affiliation = :affiliation,
		occupation = :occupation,
		birth_date = :birth_date,
		birth_place = :birth_place,
		relatives = :relatives,
		first_appearance = :first_appearance,
		updated_at = :updated_at
	WHERE id = :id AND deleted_at IS NULL`
	
	_, err = h.DB.NamedExec(query, char)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update character")
		return
	}
	
	RespondWithJSON(w, http.StatusOK, char)
}

// Delete deletes a character (soft delete)
func (h *CharacterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	// Check if character exists
	var exists bool
	err = h.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM characters WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil || !exists {
		RespondWithError(w, http.StatusNotFound, "Character not found")
		return
	}
	
	// Soft delete
	now := time.Now()
	_, err = h.DB.Exec("UPDATE characters SET deleted_at = $1 WHERE id = $2", now, id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete character")
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
