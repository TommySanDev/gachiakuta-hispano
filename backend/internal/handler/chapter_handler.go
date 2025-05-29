package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"github.com/TommySanDev/gachiakuta-hispano/internal/chapter"
)

// ChapterHandler implements HTTP handlers for chapter operations
type ChapterHandler struct{}

// NewChapterHandler creates a new chapter handler
func NewChapterHandler() *ChapterHandler {
	return &ChapterHandler{}
}

// GetAll retrieves all chapters
func (h *ChapterHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	chapters, err := chapter.GetAll()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch chapters")
		return
	}
	
	RespondWithJSON(w, http.StatusOK, chapters)
}

// GetByID retrieves a chapter by ID
func (h *ChapterHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	chap, err := chapter.GetByID(id)
	if err != nil {
		if err == chapter.ErrChapterNotFound {
			RespondWithError(w, http.StatusNotFound, "Chapter not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to fetch chapter")
		}
		return
	}
	
	RespondWithJSON(w, http.StatusOK, chap)
}

// GetByNumber retrieves a chapter by number
func (h *ChapterHandler) GetByNumber(w http.ResponseWriter, r *http.Request) {
	numberStr := r.URL.Query().Get("number")
	if numberStr == "" {
		RespondWithError(w, http.StatusBadRequest, "Chapter number is required")
		return
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid chapter number")
		return
	}
	
	chap, err := chapter.GetByNumber(number)
	if err != nil {
		if err == chapter.ErrChapterNotFound {
			RespondWithError(w, http.StatusNotFound, "Chapter not found")
		} else if err == chapter.ErrInvalidInput {
			RespondWithError(w, http.StatusBadRequest, "Invalid chapter number")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to fetch chapter")
		}
		return
	}
	
	RespondWithJSON(w, http.StatusOK, chap)
}

// Create creates a new chapter
func (h *ChapterHandler) Create(w http.ResponseWriter, r *http.Request) {
	var chap chapter.Chapter
	
	if err := json.NewDecoder(r.Body).Decode(&chap); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	err := chapter.Create(&chap)
	if err != nil {
		if err == chapter.ErrInvalidInput {
			RespondWithError(w, http.StatusBadRequest, "Title, number and image are required")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to create chapter")
		}
		return
	}
	
	RespondWithJSON(w, http.StatusCreated, chap)
}

// Update updates an existing chapter
func (h *ChapterHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	var chap chapter.Chapter
	if err := json.NewDecoder(r.Body).Decode(&chap); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	err = chapter.Update(id, &chap)
	if err != nil {
		if err == chapter.ErrChapterNotFound {
			RespondWithError(w, http.StatusNotFound, "Chapter not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to update chapter")
		}
		return
	}
	
	RespondWithJSON(w, http.StatusOK, chap)
}

// Delete deletes a chapter (soft delete)
func (h *ChapterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	err = chapter.Delete(id)
	if err != nil {
		if err == chapter.ErrChapterNotFound {
			RespondWithError(w, http.StatusNotFound, "Chapter not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to delete chapter")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// DeletePermanently permanently deletes a chapter (admin only)
func (h *ChapterHandler) DeletePermanently(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	err = chapter.DeletePermanently(id)
	if err != nil {
		if err == chapter.ErrChapterNotFound {
			RespondWithError(w, http.StatusNotFound, "Chapter not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to delete chapter permanently")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// Restore restores a soft deleted chapter (admin only)
func (h *ChapterHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	
	err = chapter.Restore(id)
	if err != nil {
		if err == chapter.ErrChapterNotFound {
			RespondWithError(w, http.StatusNotFound, "Deleted chapter not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to restore chapter")
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
