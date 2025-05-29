package chapter

import (
	"errors"
	"time"
	
	"github.com/TommySanDev/gachiakuta-hispano/config"
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

// GetAll retrieves all non-deleted chapters
func GetAll() ([]Chapter, error) {
	var chapters []Chapter
	err := config.DB.Select(&chapters, "SELECT * FROM chapters WHERE deleted_at IS NULL ORDER BY created_at DESC")
	return chapters, err
}

// GetByID retrieves a chapter by ID
func GetByID(id uint) (*Chapter, error) {
	var chap Chapter
	err := config.DB.Get(&chap, "SELECT * FROM chapters WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return nil, ErrChapterNotFound
	}
	return &chap, nil
}

// GetByNumber retrieves a chapter by number
func GetByNumber(number int) (*Chapter, error) {
	if number <= 0 {
		return nil, ErrInvalidInput
	}
	
	var chap Chapter
	err := config.DB.Get(&chap, "SELECT * FROM chapters WHERE number = $1 AND deleted_at IS NULL", number)
	if err != nil {
		return nil, ErrChapterNotFound
	}
	return &chap, nil
}

// Create creates a new chapter
func Create(chap *Chapter) error {
	// Basic validation
	if chap.Title == "" {
		return ErrInvalidInput
	}
	if chap.Number <= 0 {
		return ErrInvalidInput
	}
	if chap.Image == "" {
		return ErrInvalidInput
	}
	
	// Set timestamps
	now := time.Now()
	chap.CreatedAt = now
	chap.UpdatedAt = now
	
	query := `INSERT INTO chapters (
		title, number, image, created_at, updated_at
	) VALUES (
		:title, :number, :image, :created_at, :updated_at
	) RETURNING id`
	
	rows, err := config.DB.NamedQuery(query, chap)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	if rows.Next() {
		var id uint
		rows.Scan(&id)
		chap.ID = id
	}
	
	return nil
}

// Update updates an existing chapter
func Update(id uint, chap *Chapter) error {
	// Check if chapter exists
	exists, err := Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrChapterNotFound
	}
	
	// Set ID and update timestamp
	chap.ID = id
	chap.UpdatedAt = time.Now()
	
	query := `UPDATE chapters SET
		title = :title, 
		number = :number,
		image = :image,
		updated_at = :updated_at
	WHERE id = :id AND deleted_at IS NULL`
	
	_, err = config.DB.NamedExec(query, chap)
	return err
}

// Delete performs soft delete on a chapter
func Delete(id uint) error {
	exists, err := Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrChapterNotFound
	}
	
	now := time.Now()
	_, err = config.DB.Exec("UPDATE chapters SET deleted_at = $1 WHERE id = $2", now, id)
	return err
}

// DeletePermanently permanently deletes a chapter (admin only)
func DeletePermanently(id uint) error {
	result, err := config.DB.Exec("DELETE FROM chapters WHERE id = $1", id)
	if err != nil {
		return err
	}
	
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrChapterNotFound
	}
	
	return nil
}

// Restore restores a soft deleted chapter (admin only)
func Restore(id uint) error {
	result, err := config.DB.Exec("UPDATE chapters SET deleted_at = NULL, updated_at = $1 WHERE id = $2 AND deleted_at IS NOT NULL", 
		time.Now(), id)
	if err != nil {
		return err
	}
	
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrChapterNotFound
	}
	
	return nil
}

// Exists checks if a chapter exists and is not deleted
func Exists(id uint) (bool, error) {
	var exists bool
	err := config.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM chapters WHERE id = $1 AND deleted_at IS NULL", id)
	return exists, err
}
