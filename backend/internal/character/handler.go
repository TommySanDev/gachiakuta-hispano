package character

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

// GET /characters/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    character, err := h.crudService.Get(r.Context(), uint(id))
    if err != nil {
        switch {
        case errors.Is(err, ErrCharacterNotFound):
            http.Error(w, "Character not found", http.StatusNotFound)
        default:
            logger.Error("Error fetching character", 
                zap.Error(err), 
                zap.String("id", idStr),
            )
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, character)
}

// POST /admin/characters
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var input CreateCharacterInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    character, err := h.crudService.Create(r.Context(), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        case errors.Is(err, ErrCharacterAlreadyExists):
            http.Error(w, "Character already exists", http.StatusConflict)
        default:
            logger.Error("Error creating character", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusCreated, character)
}

// PUT /admin/characters/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    var input UpdateCharacterInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    character, err := h.crudService.Update(r.Context(), uint(id), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrCharacterNotFound):
            http.Error(w, "Character not found", http.StatusNotFound)
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            logger.Error("Error updating character", 
                zap.Error(err), 
                zap.String("id", idStr),
            )
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, character)
}

// DELETE /admin/characters/{id}
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
        case errors.Is(err, ErrCharacterNotFound):
            http.Error(w, "Character not found", http.StatusNotFound)
        default:
            logger.Error("Error deleting character", 
                zap.Error(err), 
                zap.String("id", idStr),
            )
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// DELETE /admin/characters/{id}/permanent
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
        case errors.Is(err, ErrCharacterNotFound):
            http.Error(w, "Character not found", http.StatusNotFound)
        default:
            logger.Error("Error permanently deleting character", 
                zap.Error(err), 
                zap.String("id", idStr),
            )
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// PATCH /admin/characters/{id}/restore
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
        case errors.Is(err, ErrCharacterNotFound):
            http.Error(w, "Character not found", http.StatusNotFound)
        default:
            logger.Error("Error restoring character", 
                zap.Error(err), 
                zap.String("id", idStr),
            )
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// GET /characters
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
    filter := CharacterFilter{
        Search:      r.URL.Query().Get("search"),
        Status:      r.URL.Query().Get("status"),
        Affiliation: r.URL.Query().Get("affiliation"),
        Species:     r.URL.Query().Get("species"),
        SortBy:      r.URL.Query().Get("sort_by"),
        SortDir:     r.URL.Query().Get("sort_dir"),
        Page:        getIntParam(r, "page", 1),
        PageSize:    getIntParam(r, "page_size", 20),
    }

    // Check deleted records for admin purposes
    includeDeleted := r.URL.Query().Get("include_deleted")
    filter.IncludeDeleted = includeDeleted == "true"

    characters, total, err := h.searchService.List(r.Context(), filter)
    if err != nil {
        logger.Error("Error listing characters", zap.Error(err))
        http.Error(w, "Error fetching characters", http.StatusInternalServerError)
        return
    }

    totalPages := (total + filter.PageSize - 1) / filter.PageSize

    response := map[string]interface{}{
        "data": characters,
        "meta": map[string]interface{}{
            "total":       total,
            "page":        filter.Page,
            "page_size":   filter.PageSize,
            "total_pages": totalPages,
        },
    }

    respondJSON(w, http.StatusOK, response)
}

// GET /characters/affiliation/{name}
func (h *Handler) ListByAffiliation(w http.ResponseWriter, r *http.Request) {
    affiliation := chi.URLParam(r, "name")
    limit := getIntParam(r, "limit", 10)

    characters, err := h.searchService.ListByAffiliation(r.Context(), affiliation, limit)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            logger.Error("Error listing characters by affiliation", 
                zap.Error(err),
                zap.String("affiliation", affiliation),
            )
            http.Error(w, "Error fetching characters", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, characters)
}

// GET /characters/status/{status}
func (h *Handler) ListByStatus(w http.ResponseWriter, r *http.Request) {
    status := chi.URLParam(r, "status")
    limit := getIntParam(r, "limit", 10)

    characters, err := h.searchService.ListByStatus(r.Context(), status, limit)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            logger.Error("Error listing characters by status", 
                zap.Error(err),
                zap.String("status", status),
            )
            http.Error(w, "Error fetching characters", http.StatusInternalServerError)
        }
        return
    }
    respondJSON(w, http.StatusOK, characters)
}

// GET /characters/species/{species}
func (h *Handler) ListBySpecies(w http.ResponseWriter, r *http.Request) {
   species := chi.URLParam(r, "species")
   limit := getIntParam(r, "limit", 10)

   characters, err := h.searchService.ListBySpecies(r.Context(), species, limit)
   if err != nil {
       switch {
       case errors.Is(err, ErrInvalidInput):
           http.Error(w, err.Error(), http.StatusBadRequest)
       default:
           logger.Error("Error listing characters by species", 
               zap.Error(err),
               zap.String("species", species),
           )
           http.Error(w, "Error fetching characters", http.StatusInternalServerError)
       }
       return
   }
   respondJSON(w, http.StatusOK, characters)
}
