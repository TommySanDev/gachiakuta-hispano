package favorite

import "time"

type Favorite struct {
    ID         uint      `json:"id" db:"id"`
    UserID     uint      `json:"user_id" db:"user_id"`
    EntityType string    `json:"entity_type" db:"entity_type"` // character, chapter, vital_instrument
    EntityID   uint      `json:"entity_id" db:"entity_id"`
    CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Dto for creating a favorite
type CreateFavoriteInput struct {
    EntityType string `json:"entity_type" validate:"required"` // must match allowed types
    EntityID   uint   `json:"entity_id" validate:"required"`
}

// Dto for removing a favorite
type DeleteFavoriteInput struct {
    EntityType string `json:"entity_type" validate:"required"`
    EntityID   uint   `json:"entity_id" validate:"required"`
}

