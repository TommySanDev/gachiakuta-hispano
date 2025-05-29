package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	
	"github.com/jmoiron/sqlx"
	"github.com/TommySanDev/gachiakuta-hispano/internal/chapter"
)

// ChapterHandler implements HTTP handlers for chapter operations
type ChapterHandler struct {
	DB *sqlx.DB
}

// NewChapterHandler creates a new chapter handler
func NewChapterHandler(db *sqlx.DB) *ChapterHandler {
	return &ChapterHandler{DB: db}
}

// GetAll retrieves all chapters
func (h *ChapterHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	var chapters []chapter.Chapter
	
	err := h.DB.Select(&chapters, "SELECT * FROM chapters WHERE deleted_at IS NULL ORDER BY created_at DESC")
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
	
	var chap chapter.Chapter
	err = h.DB.Get(&chap, "SELECT * FROM chapters WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Chapter not found")
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
	if err != nil || number <= 0 {
		RespondWithError(w, http.StatusBadRequest, "Invalid chapter number")
		return
	}
	
	var chap chapter.Chapter
	err = h.DB.Get(&chap, "SELECT * FROM chapters WHERE number = $1 AND deleted_at IS NULL", number)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Chapter not found")
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
	
	// Basic validation
	if chap.Title == "" {
		RespondWithError(w, http.StatusBadRequest, "Title is required")
		return
	}
	if chap.Number <= 0 {
		RespondWithError(w, http.StatusBadRequest, "Number must be positive")
		return
	}
	if chap.Image == "" {
		RespondWithError(w, http.StatusBadRequest, "Image is required")
		return
	}
	
	// Set timestamps
	now := time.Now()
	chap.CreatedAt = now
	chap.UpdatedAt = now
	
	query := `INSERT INTO chapters (
		title, number, image, created_at, updated_at
	) VALUES (
		:title, :number, :image, :created_at, :updated_at
	) RETURNING id`
	
	rows, err := h.DB.NamedQuery(query, chap)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create chapter")
		return
	}
	defer rows.Close()
	
	if rows.Next() {
		var id uint
		rows.Scan(&id)
		chap.ID = id
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
	
	// Check if chapter exists
	var exists bool
	err = h.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM chapters WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil || !exists {
		RespondWithError(w, http.StatusNotFound, "Chapter not found")
		return
	}
	
	// Set ID and update timestamp
	chap.ID = id
	chap.UpdatedAt = time.Now()
	
	query := `UPDATE chapters SET
		title = :title, 
		number = :number,
		image = :image,
		updated_at = :updated_at
	WHERE id = :id AND deleted_at IS NULL`
	
	_, err = h.DB.NamedExec(query, chap)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update chapter")
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
	
	// Check if chapter exists
	var exists bool
	err = h.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM chapters WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil || !exists {
		RespondWithError(w, http.StatusNotFound, "Chapter not found")
		return
	}
	
	// Soft delete
	now := time.Now()
	_, err = h.DB.Exec("UPDATE chapters SET deleted_at = $1 WHERE id = $2", now, id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete chapter")
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
	
	// Delete permanently
	result, err := h.DB.Exec("DELETE FROM chapters WHERE id = $1", id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete chapter permanently")
		return
	}
	
	rows, _ := result.RowsAffected()
	if rows == 0 {
		RespondWithError(w, http.StatusNotFound, "Chapter not found")
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
	
	// Restore chapter
	result, err := h.DB.Exec("UPDATE chapters SET deleted_at = NULL, updated_at = $1 WHERE id = $2 AND deleted_at IS NOT NULL", 
		time.Now(), id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to restore chapter")
		return
	}
	
	rows, _ := result.RowsAffected()
	if rows == 0 {
		RespondWithError(w, http.StatusNotFound, "Deleted chapter not found")
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}
