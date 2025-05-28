package character

import (
	"errors"
	"time"
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
	ErrCharacterNotFound     = errors.New("character not found")
	ErrInvalidInput          = errors.New("invalid input")
	ErrCharacterAlreadyExists = errors.New("character already exists")
)
