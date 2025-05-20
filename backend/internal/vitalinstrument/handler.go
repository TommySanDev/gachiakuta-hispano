package vitalinstrument

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"
    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

type Handler struct {
    crudService   *CrudService
    searchService *SearchService
}

func NewHandler(crudService *CrudService, searchService *SearchService) *Handler {
    return &Handler{
        crudService:   crudService,
        searchService: searchService,
    }
}

// GET /vital-instruments/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    instrument, err := h.crudService.Get(r.Context(), uint(id))
    if err != nil {
        switch {
        case errors.Is(err, ErrVitalInstrumentNotFound):
            http.Error(w, "Vital instrument not found", http.StatusNotFound)
        default:
            logger.Error("Error fetching vital instrument", 
                zap.Error(err), 
                zap.String("id", idStr),
            )
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, instrument)
}

// POST /admin/vital-instruments
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var input CreateVitalInstrumentInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    instrument, err := h.crudService.Create(r.Context(), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        case errors.Is(err, ErrVitalInstrumentAlreadyExists):
            http.Error(w, "Vital instrument already exists", http.StatusConflict)
        default:
            logger.Error("Error creating vital instrument", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusCreated, instrument)
}

// PUT /admin/vital-instruments/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    var input UpdateVitalInstrumentInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    instrument, err := h.crudService.Update(r.Context(), uint(id), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrVitalInstrumentNotFound):
            http.Error(w, "Vital instrument not found", http.StatusNotFound)
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            logger.Error("Error updating vital instrument", 
                zap.Error(err), 
                zap.String("id", idStr),
            )
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, instrument)
}

// DELETE /admin/vital-instruments/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    err = h.crudService.Delete(r.Context(), uint(id))
    if err != nil {
        switch {
        case errors.Is(err, ErrVitalInstrumentNotFound):
            http.Error(w, "Vital instrument not found", http.StatusNotFound)
        default:
            logger.Error("Error deleting vital instrument", 
                zap.Error(err), 
                zap.String("id", idStr),
            )
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// DELETE /admin/vital-instruments/{id}/permanent
func (h *Handler) DeletePermanently(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    err = h.crudService.DeletePermanently(r.Context(), uint(id))
    if err != nil {
        switch {
        case errors.Is(err, ErrVitalInstrumentNotFound):
            http.Error(w, "Vital instrument not found", http.StatusNotFound)
        default:
            logger.Error("Error permanently deleting vital instrument", 
                zap.Error(err), 
                zap.String("id", idStr),
            )
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// PATCH /admin/vital-instruments/{id}/restore
func (h *Handler) Restore(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    err = h.crudService.Restore(r.Context(), uint(id))
    if err != nil {
        switch {
        case errors.Is(err, ErrVitalInstrumentNotFound):
            http.Error(w, "Vital instrument not found", http.StatusNotFound)
        default:
            logger.Error("Error restoring vital instrument", 
                zap.Error(err), 
                zap.String("id", idStr),
            )
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// GET /vital-instruments
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
    filter := VitalInstrumentFilter{
        Search:      r.URL.Query().Get("search"),
        SortBy:      r.URL.Query().Get("sort_by"),
        SortDir:     r.URL.Query().Get("sort_dir"),
        Page:        getIntParam(r, "page", 1),
        PageSize:    getIntParam(r, "page_size", 20),
    }

    // Check if character_id is provided
    if characterIDStr := r.URL.Query().Get("character_id"); characterIDStr != "" {
        if characterID, err := strconv.ParseUint(characterIDStr, 10, 32); err == nil {
            id := uint(characterID)
            filter.CharacterID = &id
        }
    }

    // Check deleted records for admin purposes
    includeDeleted := r.URL.Query().Get("include_deleted")
    filter.IncludeDeleted = includeDeleted == "true"

    instruments, total, err := h.searchService.List(r.Context(), filter)
    if err != nil {
        logger.Error("Error listing vital instruments", zap.Error(err))
        http.Error(w, "Error fetching vital instruments", http.StatusInternalServerError)
        return
    }

    totalPages := (total + filter.PageSize - 1) / filter.PageSize

    response := map[string]interface{}{
        "data": instruments,
        "meta": map[string]interface{}{
            "total":       total,
            "page":        filter.Page,
            "page_size":   filter.PageSize,
            "total_pages": totalPages,
        },
    }

    respondJSON(w, http.StatusOK, response)
}

// GET /vital-instruments/character/{id}
func (h *Handler) ListByCharacter(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    limit := getIntParam(r, "limit", 10)

    instruments, err := h.searchService.ListByCharacter(r.Context(), uint(id), limit)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            logger.Error("Error listing vital instruments by character", 
                zap.Error(err),
                zap.String("character_id", idStr),
            )
            http.Error(w, "Error fetching vital instruments", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, instruments)
}
