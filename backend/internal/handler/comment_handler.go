package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	
	"github.com/TommySanDev/gachiakuta-hispano/internal/comment"
	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// CommentHandler implements HTTP handlers for comment operations
type CommentHandler struct{}

// NewCommentHandler creates a new comment handler
func NewCommentHandler() *CommentHandler {
	return &CommentHandler{}
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

	comments, err := comment.GetByChapterID(uint(chapterID), includeDeleted)
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

	// Set user ID from context
	newComment.UserID = authUser.ID

	err := comment.Create(&newComment)
	if err != nil {
		if err == comment.ErrInvalidInput {
			RespondWithError(w, http.StatusBadRequest, "Chapter ID and content are required")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to create comment")
		}
		return
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

	var updateData struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedComment, err := comment.Update(id, authUser.ID, updateData.Content)
	if err != nil {
		if err == comment.ErrCommentNotFound {
			RespondWithError(w, http.StatusNotFound, "Comment not found")
		} else if err == comment.ErrForbidden {
			RespondWithError(w, http.StatusForbidden, "Not allowed to update this comment")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to update comment")
		}
		return
	}

	RespondWithJSON(w, http.StatusOK, updatedComment)
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

	// Check if user is editor or admin
	isAdminOrEditor := authUser.IsEditor() || authUser.IsAdmin()

	err = comment.Delete(id, authUser.ID, isAdminOrEditor)
	if err != nil {
		if err == comment.ErrCommentNotFound {
			RespondWithError(w, http.StatusNotFound, "Comment not found")
		} else if err == comment.ErrForbidden {
			RespondWithError(w, http.StatusForbidden, "Not allowed to delete this comment")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to delete comment")
		}
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

	err = comment.DeletePermanently(id)
	if err != nil {
		if err == comment.ErrCommentNotFound {
			RespondWithError(w, http.StatusNotFound, "Comment not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Failed to permanently delete comment")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
