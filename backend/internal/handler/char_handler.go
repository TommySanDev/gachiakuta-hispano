package character

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// Implements HTTP handlers for character operations
type CharacterController struct {
	DB *sqlx.DB
}

// Creates a new character controller
func NewCharacterController(db *sqlx.DB) *CharacterController {
	return &CharacterController{DB: db}
}

// Retrieves all characters
func (c *CharacterController) GetAll(w http.ResponseWriter, r *http.Request) {
	var characters []Character
	
	err := c.DB.Select(&characters, "SELECT * FROM characters WHERE deleted_at IS NULL")
	if err != nil {
		http.Error(w, "Failed to fetch characters", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(characters)
}

// Retrieves a character by ID
func (c *CharacterController) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var character Character
	err = c.DB.Get(&character, "SELECT * FROM characters WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil {
		http.Error(w, "Character not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(character)
}

// Creates a new character
func (c *CharacterController) Create(w http.ResponseWriter, r *http.Request) {
	var character Character
	
	if err := json.NewDecoder(r.Body).Decode(&character); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Set creation time
	now := time.Now()
	character.CreatedAt = now
	character.UpdatedAt = now
	
	query := `INSERT INTO characters (
		name, name_japanese, main_image, description, species, gender, age,
		height, status, affiliation, occupation, birth_date, birth_place,
		relatives, first_appearance, created_at, updated_at
	) VALUES (
		:name, :name_japanese, :main_image, :description, :species, :gender, :age,
		:height, :status, :affiliation, :occupation, :birth_date, :birth_place,
		:relatives, :first_appearance, :created_at, :updated_at
	) RETURNING id`
	
	rows, err := c.DB.NamedQuery(query, character)
	if err != nil {
		http.Error(w, "Failed to create character", http.StatusInternalServerError)
		return
	}
	
	if rows.Next() {
		var id uint
		rows.Scan(&id)
		character.ID = id
	}
	rows.Close()
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(character)
}

// Updates an existing character
func (c *CharacterController) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var character Character
	if err := json.NewDecoder(r.Body).Decode(&character); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Check if character exists
	var exists bool
	err = c.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM characters WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil || !exists {
		http.Error(w, "Character not found", http.StatusNotFound)
		return
	}
	
	// Set ID and update time
	character.ID = uint(id)
	character.UpdatedAt = time.Now()
	
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
	
	_, err = c.DB.NamedExec(query, character)
	if err != nil {
		http.Error(w, "Failed to update character", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(character)
}

// Deletes a character (soft delete)
func (c *CharacterController) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	// Check if character exists
	var exists bool
	err = c.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM characters WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil || !exists {
		http.Error(w, "Character not found", http.StatusNotFound)
		return
	}
	
	// Soft delete
	now := time.Now()
	_, err = c.DB.Exec("UPDATE characters SET deleted_at = ? WHERE id = ?", now, id)
	if err != nil {
		http.Error(w, "Failed to delete character", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
