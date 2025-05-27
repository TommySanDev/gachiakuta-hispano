package chapter

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

// GET /chapters/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    chapter, err := h.crudService.Get(r.Context(), uint(id))
    if err != nil {
        switch {
        case errors.Is(err, ErrChapterNotFound):
            http.Error(w, "Chapter not found", http.StatusNotFound)
        default:
            logger.Error("Error fetching chapter", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, chapter)
}

// POST /admin/chapters
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var input CreateChapterInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    chapter, err := h.crudService.Create(r.Context(), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        case errors.Is(err, ErrChapterAlreadyExists):
            http.Error(w, "Chapter already exists", http.StatusConflict)
        default:
            logger.Error("Error creating chapter", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusCreated, chapter)
}

// PUT /admin/chapters/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        http.Error(w, "Invalid ID format", http.StatusBadRequest)
        return
    }

    var input UpdateChapterInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    chapter, err := h.crudService.Update(r.Context(), uint(id), input)
    if err != nil {
        switch {
        case errors.Is(err, ErrChapterNotFound):
            http.Error(w, "Chapter not found", http.StatusNotFound)
        case errors.Is(err, ErrInvalidInput):
            http.Error(w, err.Error(), http.StatusBadRequest)
        default:
            logger.Error("Error updating chapter", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    respondJSON(w, http.StatusOK, chapter)
}

// DELETE /admin/chapters/{id}
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
        case errors.Is(err, ErrChapterNotFound):
            http.Error(w, "Chapter not found", http.StatusNotFound)
        default:
            logger.Error("Error deleting chapter", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// DELETE /admin/chapters/{id}/permanent
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
        case errors.Is(err, ErrChapterNotFound):
            http.Error(w, "Chapter not found", http.StatusNotFound)
        default:
            logger.Error("Error permanently deleting chapter", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// PATCH /admin/chapters/{id}/restore
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
        case errors.Is(err, ErrChapterNotFound):
            http.Error(w, "Chapter not found", http.StatusNotFound)
        default:
            logger.Error("Error restoring chapter", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

// GET /chapters
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
    filter := ChapterFilter{
        Search:         r.URL.Query().Get("search"),
        SortBy:         r.URL.Query().Get("sort_by"),
        SortDir:        r.URL.Query().Get("sort_dir"),
        Page:           getIntParam(r, "page", 1),
        PageSize:       getIntParam(r, "page_size", 20),
        IncludeDeleted: r.URL.Query().Get("include_deleted") == "true",
    }

    chapters, total, err := h.searchService.List(r.Context(), filter)
    if err != nil {
        logger.Error("Error listing chapters", zap.Error(err))
        http.Error(w, "Error fetching chapters", http.StatusInternalServerError)
        return
    }

    totalPages := (total + filter.PageSize - 1) / filter.PageSize

    response := map[string]interface{}{
        "data": chapters,
        "meta": map[string]interface{}{
            "total":       total,
            "page":        filter.Page,
            "page_size":   filter.PageSize,
            "total_pages": totalPages,
        },
    }

    respondJSON(w, http.StatusOK, response)
}

