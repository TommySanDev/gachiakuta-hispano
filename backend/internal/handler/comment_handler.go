package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	
	"github.com/TommySanDev/gachiakuta-hispano/internal/comment"
	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// CommentHandler implements HTTP handlers for comment operations
type CommentHandler struct {
	DB *sqlx.DB
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(db *sqlx.DB) *CommentHandler {
	return &CommentHandler{DB: db}
}

// ListByChapter retrieves comments for a specific chapter
// GET /api/comments/chapter/{id}
func (h *CommentHandler) ListByChapter(w http.ResponseWriter, r *http.Request) {
	chapterIDStr := chi.URLParam(r, "id")
	chapterID, err := strconv.ParseUint(chapterIDStr, 10, 32)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid chapter ID")
		return
	}

	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	query := `SELECT c.*, u.username, u.first_name, u.last_name 
		FROM comments c 
		JOIN users u ON c.user_id = u.id 
		WHERE c.chapter_id = $1`
	
	if !includeDeleted {
		query += " AND c.deleted_at IS NULL"
	}
	
	query += " ORDER BY c.created_at ASC"

	type CommentWithUser struct {
		comment.Comment
		Username  string `json:"username" db:"username"`
		FirstName string `json:"first_name" db:"first_name"`
		LastName  string `json:"last_name" db:"last_name"`
	}

	var comments []CommentWithUser
	err = h.DB.Select(&comments, query, uint(chapterID))
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch comments")
		return
	}

	RespondWithJSON(w, http.StatusOK, comments)
}

// Create creates a new comment
// POST /api/comments
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	var newComment comment.Comment
	if err := json.NewDecoder(r.Body).Decode(&newComment); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if newComment.ChapterID == 0 || newComment.Content == "" {
		RespondWithError(w, http.StatusBadRequest, "Chapter ID and content are required")
		return
	}

	// Verify chapter exists
	var chapterExists bool
	err := h.DB.Get(&chapterExists, "SELECT COUNT(*) > 0 FROM chapters WHERE id = $1 AND deleted_at IS NULL", newComment.ChapterID)
	if err != nil || !chapterExists {
		RespondWithError(w, http.StatusBadRequest, "Chapter not found")
		return
	}

	// Set fields from context and timestamps
	now := time.Now()
	newComment.UserID = authUser.ID
	newComment.CreatedAt = now
	newComment.UpdatedAt = now

	query := `INSERT INTO comments (user_id, chapter_id, content, created_at, updated_at)
		VALUES (:user_id, :chapter_id, :content, :created_at, :updated_at) RETURNING id`

	rows, err := h.DB.NamedQuery(query, newComment)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}
	defer rows.Close()

	if rows.Next() {
		var id uint
		rows.Scan(&id)
		newComment.ID = id
	}

	RespondWithJSON(w, http.StatusCreated, newComment)
}

// Update updates an existing comment
// PUT /api/comments/{id}
func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var updateData comment.Comment
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing comment
	var existingComment comment.Comment
	err = h.DB.Get(&existingComment, "SELECT * FROM comments WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Comment not found")
		return
	}

	// Check ownership (only owner or admin can update)
	if existingComment.UserID != authUser.ID && !authUser.IsAdmin() {
		RespondWithError(w, http.StatusForbidden, "Not allowed to update this comment")
		return
	}

	// Update only non-empty fields
	updated := false
	if updateData.Content != "" {
		existingComment.Content = updateData.Content
		updated = true
	}

	if !updated {
		RespondWithJSON(w, http.StatusOK, existingComment)
		return
	}

	existingComment.UpdatedAt = time.Now()

	query := `UPDATE comments SET content = :content, updated_at = :updated_at 
		WHERE id = :id AND deleted_at IS NULL`

	_, err = h.DB.NamedExec(query, existingComment)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update comment")
		return
	}

	RespondWithJSON(w, http.StatusOK, existingComment)
}

// Delete deletes a comment (soft delete)
// DELETE /api/comments/{id}
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	authUser, ok := user.GetUserFromContext(r.Context())
	if !ok {
		RespondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Get comment to check ownership
	var existingComment comment.Comment
	err = h.DB.Get(&existingComment, "SELECT * FROM comments WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "Comment not found")
		return
	}

	// Check ownership (only owner, editor or admin can delete)
	if existingComment.UserID != authUser.ID && !authUser.IsEditor() && !authUser.IsAdmin() {
		RespondWithError(w, http.StatusForbidden, "Not allowed to delete this comment")
		return
	}

	// Soft delete
	now := time.Now()
	_, err = h.DB.Exec("UPDATE comments SET deleted_at = $1 WHERE id = $2", now, id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeletePermanently permanently deletes a comment (admin only)
// DELETE /api/admin/comments/{id}/permanent
func (h *CommentHandler) DeletePermanently(w http.ResponseWriter, r *http.Request) {
	id, err := GetIDParam(r)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Permanently delete
	result, err := h.DB.Exec("DELETE FROM comments WHERE id = $1", id)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to permanently delete comment")
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		RespondWithError(w, http.StatusNotFound, "Comment not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
