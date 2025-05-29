package character

import (
	"errors"
	"time"
	
	"github.com/TommySanDev/gachiakuta-hispano/config"
)

// Character represents a character in the Gachiakuta universe
type Character struct {
	ID              uint       `json:"id" db:"id"`
	Name            string     `json:"name" db:"name"`
	NameJapanese    string     `json:"name_japanese" db:"name_japanese"`
	MainImage       string     `json:"main_image" db:"main_image"`
	Description     string     `json:"description" db:"description"`
	Species         string     `json:"species" db:"species"`
	Gender          string     `json:"gender" db:"gender"`
	Age             int        `json:"age" db:"age"`
	Height          string     `json:"height" db:"height"`
	Status          string     `json:"status" db:"status"`
	Affiliation     string     `json:"affiliation" db:"affiliation"`
	Occupation      string     `json:"occupation" db:"occupation"`
	BirthDate       string     `json:"birth_date" db:"birth_date"`
	BirthPlace      string     `json:"birth_place" db:"birth_place"`
	Relatives       string     `json:"relatives" db:"relatives"`
	FirstAppearance int        `json:"first_appearance" db:"first_appearance"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"-" db:"updated_at"`        // Hidden in JSON
	DeletedAt       *time.Time `json:"-" db:"deleted_at"`        // Hidden in JSON
}

// Common errors
var (
	ErrCharacterNotFound      = errors.New("character not found")
	ErrInvalidInput           = errors.New("invalid input")
	ErrCharacterAlreadyExists = errors.New("character already exists")
)

// GetAll retrieves all non-deleted characters
func GetAll() ([]Character, error) {
	var characters []Character
	err := config.DB.Select(&characters, "SELECT * FROM characters WHERE deleted_at IS NULL ORDER BY created_at DESC")
	return characters, err
}

// GetByID retrieves a character by ID
func GetByID(id uint) (*Character, error) {
	var char Character
	err := config.DB.Get(&char, "SELECT * FROM characters WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return nil, ErrCharacterNotFound
	}
	return &char, nil
}

// Create creates a new character
func Create(char *Character) error {
	// Basic validation
	if char.Name == "" {
		return ErrInvalidInput
	}
	
	// Set timestamps
	now := time.Now()
	char.CreatedAt = now
	char.UpdatedAt = now
	
	query := `INSERT INTO characters (
		name, name_japanese, main_image, description, species, gender, age,
		height, status, affiliation, occupation, birth_date, birth_place,
		relatives, first_appearance, created_at, updated_at
	) VALUES (
		:name, :name_japanese, :main_image, :description, :species, :gender, :age,
		:height, :status, :affiliation, :occupation, :birth_date, :birth_place,
		:relatives, :first_appearance, :created_at, :updated_at
	) RETURNING id`
	
	rows, err := config.DB.NamedQuery(query, char)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	if rows.Next() {
		var id uint
		rows.Scan(&id)
		char.ID = id
	}
	
	return nil
}

// Update updates an existing character
func Update(id uint, char *Character) error {
	// Check if character exists
	exists, err := Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCharacterNotFound
	}
	
	// Set ID and update timestamp
	char.ID = id
	char.UpdatedAt = time.Now()
	
	query := `UPDATE characters SET
		name = :name, 
		name_japanese = :name_japanese,
		main_image = :main_image,
		description = :description,
		species = :species,
		gender = :gender,
		age = :age,
		height = :height,
		status = :status,
		affiliation = :affiliation,
		occupation = :occupation,
		birth_date = :birth_date,
		birth_place = :birth_place,
		relatives = :relatives,
		first_appearance = :first_appearance,
		updated_at = :updated_at
	WHERE id = :id AND deleted_at IS NULL`
	
	_, err = config.DB.NamedExec(query, char)
	return err
}

// Delete performs soft delete on a character
func Delete(id uint) error {
	exists, err := Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCharacterNotFound
	}
	
	now := time.Now()
	_, err = config.DB.Exec("UPDATE characters SET deleted_at = $1 WHERE id = $2", now, id)
	return err
}

// Exists checks if a character exists and is not deleted
func Exists(id uint) (bool, error) {
	var exists bool
	err := config.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM characters WHERE id = $1 AND deleted_at IS NULL", id)
	return exists, err
}
