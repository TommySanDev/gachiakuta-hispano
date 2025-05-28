package handler

import (
	"encoding/json"
	"net/http"
	"time"
	
	"github.com/jmoiron/sqlx"
	"github.com/TommySanDev/gachiakuta-hispano/internal/vitalinstrument"
)

// VitalInstrumentHandler implementa manejadores HTTP para operaciones de instrumentos vitales
type VitalInstrumentHandler struct {
	DB *sqlx.DB
}

// NewVitalInstrumentHandler crea un nuevo manejador de instrumentos vitales
func NewVitalInstrumentHandler(db *sqlx.DB) *VitalInstrumentHandler {
	return &VitalInstrumentHandler{DB: db}
}

// GetAll recupera todos los instrumentos vitales
func (h *VitalInstrumentHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	var instruments []vitalinstrument.VitalInstrument
	
	err := h.DB.Select(&instruments, "SELECT * FROM vital_instruments WHERE deleted_at IS NULL ORDER BY created_at DESC")
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch vital instruments")
		return
	}
	
	RespondWithJSON(w, http.StatusOK, instruments)
}

// GetByID recupera un instrumento vital por ID
func (h *VitalInstrumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	var vi vitalinstrument.VitalInstrument
	err = h.DB.Get(&vi, "SELECT * FROM vital_instruments WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Vital instrument not found")
		return
	}
	
	RespondWithJSON(w, http.StatusOK, vi)
}

// Create crea un nuevo instrumento vital
func (h *VitalInstrumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var vi vitalinstrument.VitalInstrument
	
	if err := json.NewDecoder(r.Body).Decode(&vi); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Validación básica
	if vi.Name == "" {
		RespondWithError(w, http.StatusBadRequest, "Name is required")
		return
	}
	
	// Establecer marcas de tiempo
	now := time.Now()
	vi.CreatedAt = now
	vi.UpdatedAt = now
	
	query := `INSERT INTO vital_instruments (
		name, main_image, description, powers, character_id,
		first_appearance, created_at, updated_at
	) VALUES (
		:name, :main_image, :description, :powers, :character_id,
		:first_appearance, :created_at, :updated_at
	) RETURNING id`
	
	rows, err := h.DB.NamedQuery(query, vi)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create vital instrument")
		return
	}
	defer rows.Close()
	
	if rows.Next() {
		var id uint
		rows.Scan(&id)
		vi.ID = id
	}
	
	RespondWithJSON(w, http.StatusCreated, vi)
}

// Update actualiza un instrumento vital existente
func (h *VitalInstrumentHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	var vi vitalinstrument.VitalInstrument
	if err := json.NewDecoder(r.Body).Decode(&vi); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Verificar si el instrumento vital existe
	var exists bool
	err = h.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM vital_instruments WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil || !exists {
		RespondWithError(w, http.StatusNotFound, "Vital instrument not found")
		return
	}
	
	// Establecer ID y actualizar marca de tiempo
	vi.ID = id
	vi.UpdatedAt = time.Now()
	
	query := `UPDATE vital_instruments SET
		name = :name, 
		main_image = :main_image,
		description = :description,
		powers = :powers,
		character_id = :character_id,
		first_appearance = :first_appearance,
		updated_at = :updated_at
	WHERE id = :id AND deleted_at IS NULL`
	
	_, err = h.DB.NamedExec(query, vi)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update vital instrument")
		return
	}
	
	RespondWithJSON(w, http.StatusOK, vi)
}

// Delete elimina un instrumento vital (eliminación lógica)
func (h *VitalInstrumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	// Verificar si el instrumento vital existe
	var exists bool
	err = h.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM vital_instruments WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil || !exists {
		RespondWithError(w, http.StatusNotFound, "Vital instrument not found")
		return
	}
	
	// Eliminación lógica
	now := time.Now()
	_, err = h.DB.Exec("UPDATE vital_instruments SET deleted_at = $1 WHERE id = $2", now, id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete vital instrument")
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// ListByCharacter recupera los instrumentos vitales por personaje
func (h *VitalInstrumentHandler) ListByCharacter(w http.ResponseWriter, r *http.Request) {
	characterID, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid character ID")
		return
	}
	
	var instruments []vitalinstrument.VitalInstrument
	err = h.DB.Select(&instruments, 
		"SELECT * FROM vital_instruments WHERE character_id = $1 AND deleted_at IS NULL ORDER BY name ASC", 
		characterID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch vital instruments")
		return
	}
	
	RespondWithJSON(w, http.StatusOK, instruments)
}
