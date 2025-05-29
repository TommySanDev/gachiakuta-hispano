package favorite

import (
	"errors"
	"strings"
	"time"
	
	"github.com/TommySanDev/gachiakuta-hispano/config"
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

// GetByUserID retrieves all favorites for a user, grouped by entity type
func GetByUserID(userID uint) (map[string][]Favorite, error) {
	var favorites []Favorite
	err := config.DB.Select(&favorites, 
		"SELECT * FROM favorites WHERE user_id = $1 ORDER BY created_at DESC", 
		userID)
	if err != nil {
		return nil, err
	}

	// Group by entity type for easier frontend consumption
	result := map[string][]Favorite{
		"characters":        {},
		"chapters":          {},
		"vital_instruments": {},
	}

	for _, fav := range favorites {
		switch fav.EntityType {
		case "character":
			result["characters"] = append(result["characters"], fav)
		case "chapter":
			result["chapters"] = append(result["chapters"], fav)
		case "vital_instrument":
			result["vital_instruments"] = append(result["vital_instruments"], fav)
		}
	}

	return result, nil
}

// Create adds a new favorite
func Create(favorite *Favorite) error {
	// Basic validation
	if favorite.EntityType == "" || favorite.EntityID == 0 {
		return ErrInvalidInput
	}

	// Normalize and validate entity type
	normalizedType := strings.ToLower(favorite.EntityType)
	if normalizedType != "character" && normalizedType != "chapter" && normalizedType != "vital_instrument" {
		return ErrInvalidInput
	}

	// Verify entity exists
	exists, err := entityExists(normalizedType, favorite.EntityID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrInvalidInput
	}

	// Check if already favorited
	alreadyExists, err := ExistsByUserAndEntity(favorite.UserID, normalizedType, favorite.EntityID)
	if err != nil {
		return err
	}
	if alreadyExists {
		return ErrAlreadyFavorited
	}

	// Set normalized type and timestamp
	favorite.EntityType = normalizedType
	favorite.CreatedAt = time.Now()

	query := `INSERT INTO favorites (user_id, entity_type, entity_id, created_at)
		VALUES (:user_id, :entity_type, :entity_id, :created_at) RETURNING id`

	rows, err := config.DB.NamedQuery(query, favorite)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		var id uint
		rows.Scan(&id)
		favorite.ID = id
	}

	return nil
}

// Delete removes a favorite
func Delete(userID uint, entityType string, entityID uint) error {
	// Basic validation
	if entityType == "" || entityID == 0 {
		return ErrInvalidInput
	}

	// Normalize entity type
	normalizedType := strings.ToLower(entityType)

	// Remove favorite
	result, err := config.DB.Exec(
		"DELETE FROM favorites WHERE user_id = $1 AND entity_type = $2 AND entity_id = $3",
		userID, normalizedType, entityID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrFavoriteNotFound
	}

	return nil
}

// ExistsByUserAndEntity checks if a favorite already exists
func ExistsByUserAndEntity(userID uint, entityType string, entityID uint) (bool, error) {
	var exists bool
	err := config.DB.Get(&exists, 
		"SELECT COUNT(*) > 0 FROM favorites WHERE user_id = $1 AND entity_type = $2 AND entity_id = $3",
		userID, entityType, entityID)
	return exists, err
}

// entityExists verifies that the referenced entity exists and is not deleted
func entityExists(entityType string, entityID uint) (bool, error) {
	var tableName string
	switch entityType {
	case "character":
		tableName = "characters"
	case "chapter":
		tableName = "chapters"
	case "vital_instrument":
		tableName = "vital_instruments"
	default:
		return false, ErrInvalidInput
	}

	var exists bool
	err := config.DB.Get(&exists, 
		"SELECT COUNT(*) > 0 FROM "+tableName+" WHERE id = $1 AND deleted_at IS NULL", 
		entityID)
	return exists, err
}
