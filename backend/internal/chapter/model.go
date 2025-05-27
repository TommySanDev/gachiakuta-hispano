package chapter

import "time"

type Chapter struct {
    ID        uint       `json:"id" db:"id"`
    Title     string     `json:"title" db:"title"`
    Number    int        `json:"number" db:"number"`
    Image     string     `json:"image" db:"image"`
    CreatedAt time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Dto for creating and updating a chapter
type CreateChapterInput struct {
    Title  string `json:"title" validate:"required"`
    Number int    `json:"number" validate:"required"`
    Image  string `json:"image" validate:"required"`
}

type UpdateChapterInput struct {
    Title  *string `json:"title,omitempty"`
    Number *int    `json:"number,omitempty"`
    Image  *string `json:"image,omitempty"`
}

// Parameters for filtering chapters
type ChapterFilter struct {
    Search         string `json:"search"`
    SortBy         string `json:"sort_by,omitempty"`
    SortDir        string `json:"sort_dir,omitempty"`
    Page           int    `json:"page"`
    PageSize       int    `json:"page_size"`
    IncludeDeleted bool   `json:"include_deleted,omitempty"`
}

