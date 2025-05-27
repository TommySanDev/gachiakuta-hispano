package comment

import "time"

type Comment struct {
    ID         uint       `json:"id" db:"id"`
    UserID     uint       `json:"user_id" db:"user_id"`
    ChapterID  uint       `json:"chapter_id" db:"chapter_id"`
    Content    string     `json:"content" db:"content"`
    CreatedAt  time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
    DeletedAt  *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// DTO for creating a comment
type CreateCommentInput struct {
    ChapterID uint   `json:"chapter_id" validate:"required"`
    Content   string `json:"content" validate:"required"`
}

// DTO for updating a comment
type UpdateCommentInput struct {
    Content *string `json:"content,omitempty"`
}

// Filter for listing comments by chapter
type ChapterCommentFilter struct {
    ChapterID      uint `json:"chapter_id"`
    IncludeDeleted bool `json:"include_deleted,omitempty"`
}

