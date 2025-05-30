package handler

import (
	"encoding/json"
	"net/http"
	
	"github.com/TommySanDev/gachiakuta-hispano/internal/vitalinstrument"
)

// VitalInstrumentHandler implementa manejadores HTTP para operaciones de instrumentos vitales
type VitalInstrumentHandler struct{}

// NewVitalInstrumentHandler crea un nuevo manejador de instrumentos vitales
func NewVitalInstrumentHandler() *VitalInstrumentHandler {
	return &VitalInstrumentHandler{}
}

// GetAll recupera todos los instrumentos vitales
func (h *VitalInstrumentHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	instruments, err := vitalinstrument.GetAll()
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
	
	vi, err := vitalinstrument.GetByID(id)
	if err != nil {
		if err == vitalinstrument.ErrVitalInstrumentNotFound {
			RespondWithError(w, http.StatusNotFound, "Vital instrument not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to fetch vital instrument")
		}
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
	
	err := vitalinstrument.Create(&vi)
	if err != nil {
		if err == vitalinstrument.ErrInvalidInput {
			RespondWithError(w, http.StatusBadRequest, "Name is required")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to create vital instrument")
		}
		return
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
	
	err = vitalinstrument.Update(id, &vi)
	if err != nil {
		if err == vitalinstrument.ErrVitalInstrumentNotFound {
			RespondWithError(w, http.StatusNotFound, "Vital instrument not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to update vital instrument")
		}
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
	
	err = vitalinstrument.Delete(id)
	if err != nil {
		if err == vitalinstrument.ErrVitalInstrumentNotFound {
			RespondWithError(w, http.StatusNotFound, "Vital instrument not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to delete vital instrument")
		}
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
	
	instruments, err := vitalinstrument.GetByCharacterID(characterID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch vital instruments")
		return
	}
	
	RespondWithJSON(w, http.StatusOK, instruments)
}

// DeletePermanently permanently deletes a vital instrument (admin only)
func (h *VitalInstrumentHandler) DeletePermanently(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	err = vitalinstrument.DeletePermanently(id)
	if err != nil {
		if err == vitalinstrument.ErrVitalInstrumentNotFound {
			RespondWithError(w, http.StatusNotFound, "Vital instrument not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to delete vital instrument permanently")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// Restore restores a soft deleted vital instrument (admin only)
func (h *VitalInstrumentHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	err = vitalinstrument.Restore(id)
	if err != nil {
		if err == vitalinstrument.ErrVitalInstrumentNotFound {
			RespondWithError(w, http.StatusNotFound, "Deleted vital instrument not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to restore vital instrument")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
