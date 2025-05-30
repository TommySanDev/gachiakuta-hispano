package handler

import (
	"encoding/json"
	"net/http"
	
	"github.com/TommySanDev/gachiakuta-hispano/internal/character"
)

// CharacterHandler implements HTTP handlers for character operations
type CharacterHandler struct{}

// NewCharacterHandler creates a new character handler
func NewCharacterHandler() *CharacterHandler {
	return &CharacterHandler{}
}

// GetAll retrieves all characters
func (h *CharacterHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	characters, err := character.GetAll()
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
	
	char, err := character.GetByID(id)
	if err != nil {
		if err == character.ErrCharacterNotFound {
			RespondWithError(w, http.StatusNotFound, "Character not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to fetch character")
		}
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
	
	err := character.Create(&char)
	if err != nil {
		if err == character.ErrInvalidInput {
			RespondWithError(w, http.StatusBadRequest, "Name is required")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to create character")
		}
		return
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
	
	err = character.Update(id, &char)
	if err != nil {
		if err == character.ErrCharacterNotFound {
			RespondWithError(w, http.StatusNotFound, "Character not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to update character")
		}
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
	
	err = character.Delete(id)
	if err != nil {
		if err == character.ErrCharacterNotFound {
			RespondWithError(w, http.StatusNotFound, "Character not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to delete character")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// DeletePermanently permanently deletes a character (admin only)
func (h *CharacterHandler) DeletePermanently(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	err = character.DeletePermanently(id)
	if err != nil {
		if err == character.ErrCharacterNotFound {
			RespondWithError(w, http.StatusNotFound, "Character not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to delete character permanently")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// Restore restores a soft deleted character (admin only)
func (h *CharacterHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	err = character.Restore(id)
	if err != nil {
		if err == character.ErrCharacterNotFound {
			RespondWithError(w, http.StatusNotFound, "Deleted character not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to restore character")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
