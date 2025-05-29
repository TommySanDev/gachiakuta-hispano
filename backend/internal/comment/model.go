package comment

import (
	"errors"
	"time"
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

// Common errors
var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrForbidden       = errors.New("action not allowed")
)
