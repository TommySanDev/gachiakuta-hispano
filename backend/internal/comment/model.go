package comment

import (
	"errors"
	"time"
	
	"github.com/TommySanDev/gachiakuta-hispano/config"
)

// Comment represents a user comment on a chapter
type Comment struct {
	ID        uint       `json:"id,omitempty" db:"id"`
	UserID    uint       `json:"user_id,omitempty" db:"user_id"`
	ChapterID uint       `json:"chapter_id" db:"chapter_id"`
	Content   string     `json:"content" db:"content"`
	CreatedAt time.Time  `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time  `json:"-" db:"updated_at"`        // Hidden in JSON
	DeletedAt *time.Time `json:"-" db:"deleted_at"`        // Hidden in JSON
}

// CommentWithUser represents a comment with user information
type CommentWithUser struct {
	Comment
	Username  string `json:"username" db:"username"`
	FirstName string `json:"first_name" db:"first_name"`
	LastName  string `json:"last_name" db:"last_name"`
}

// Common errors
var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrForbidden       = errors.New("action not allowed")
)

// GetByChapterID retrieves comments for a specific chapter with user info
func GetByChapterID(chapterID uint, includeDeleted bool) ([]CommentWithUser, error) {
	query := `SELECT c.*, u.username, u.first_name, u.last_name 
		FROM comments c 
		JOIN users u ON c.user_id = u.id 
		WHERE c.chapter_id = $1`
	
	if !includeDeleted {
		query += " AND c.deleted_at IS NULL"
	}
	
	query += " ORDER BY c.created_at ASC"

	var comments []CommentWithUser
	err := config.DB.Select(&comments, query, chapterID)
	return comments, err
}

// GetByID retrieves a comment by ID
func GetByID(id uint) (*Comment, error) {
	var comment Comment
	err := config.DB.Get(&comment, "SELECT * FROM comments WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return nil, ErrCommentNotFound
	}
	return &comment, nil
}

// Create creates a new comment
func Create(comment *Comment) error {
	// Basic validation
	if comment.ChapterID == 0 || comment.Content == "" {
		return ErrInvalidInput
	}
	
	// Verify chapter exists
	var chapterExists bool
	err := config.DB.Get(&chapterExists, "SELECT COUNT(*) > 0 FROM chapters WHERE id = $1 AND deleted_at IS NULL", comment.ChapterID)
	if err != nil || !chapterExists {
		return ErrInvalidInput
	}
	
	// Set timestamps
	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now
	
	query := `INSERT INTO comments (user_id, chapter_id, content, created_at, updated_at)
		VALUES (:user_id, :chapter_id, :content, :created_at, :updated_at) RETURNING id`

	rows, err := config.DB.NamedQuery(query, comment)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		var id uint
		rows.Scan(&id)
		comment.ID = id
	}

	return nil
}

// Update updates an existing comment
func Update(id uint, userID uint, content string) (*Comment, error) {
	// Get existing comment
	existingComment, err := GetByID(id)
	if err != nil {
		return nil, err
	}
	
	// Check ownership (business logic - could be moved to handler if needed)
	if existingComment.UserID != userID {
		return nil, ErrForbidden
	}
	
	// Update only if content is provided
	if content == "" {
		return existingComment, nil
	}
	
	existingComment.Content = content
	existingComment.UpdatedAt = time.Now()

	query := `UPDATE comments SET content = :content, updated_at = :updated_at 
		WHERE id = :id AND deleted_at IS NULL`

	_, err = config.DB.NamedExec(query, existingComment)
	if err != nil {
		return nil, err
	}

	return existingComment, nil
}

// Delete performs soft delete on a comment
func Delete(id uint, userID uint, isAdminOrEditor bool) error {
	// Get comment to check ownership
	existingComment, err := GetByID(id)
	if err != nil {
		return err
	}

	// Check ownership (only owner, editor or admin can delete)
	if existingComment.UserID != userID && !isAdminOrEditor {
		return ErrForbidden
	}

	// Soft delete
	now := time.Now()
	_, err = config.DB.Exec("UPDATE comments SET deleted_at = $1 WHERE id = $2", now, id)
	return err
}

// DeletePermanently permanently deletes a comment (admin only)
func DeletePermanently(id uint) error {
	result, err := config.DB.Exec("DELETE FROM comments WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrCommentNotFound
	}

	return nil
}

// Exists checks if a comment exists and is not deleted
func Exists(id uint) (bool, error) {
	var exists bool
	err := config.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM comments WHERE id = $1 AND deleted_at IS NULL", id)
	return exists, err
}
