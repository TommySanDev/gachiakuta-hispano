package chapter

import (
	"errors"
	"time"
)

// Chapter represents a chapter in the Gachiakuta manga
type Chapter struct {
	ID        uint       `json:"id" db:"id"`
	Title     string     `json:"title" db:"title"`
	Number    int        `json:"number" db:"number"`
	Image     string     `json:"image" db:"image"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"-" db:"updated_at"`        // Hidden in JSON
	DeletedAt *time.Time `json:"-" db:"deleted_at"`        // Hidden in JSON
}

// Common errors
var (
	ErrChapterNotFound      = errors.New("chapter not found")
	ErrInvalidInput         = errors.New("invalid input")
	ErrChapterAlreadyExists = errors.New("chapter already exists")
)
