package favorite

import (
	"errors"
	"time"
)

// Favorite represents a user's favorite entity (character, chapter, vital_instrument)
type Favorite struct {
	ID         uint      `json:"id,omitempty" db:"id"`
	UserID     uint      `json:"user_id,omitempty" db:"user_id"`
	EntityType string    `json:"entity_type" db:"entity_type"` // character, chapter, vital_instrument
	EntityID   uint      `json:"entity_id" db:"entity_id"`
	CreatedAt  time.Time `json:"created_at,omitempty" db:"created_at"`
}

// Common errors
var (
	ErrFavoriteNotFound = errors.New("favorite not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrAlreadyFavorited = errors.New("favorite already exists")
)
