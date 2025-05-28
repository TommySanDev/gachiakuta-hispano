package vitalinstrument

import (
	"errors"
	"time"
)

// VitalInstrument representa un instrumento vital en el universo Gachiakuta
type VitalInstrument struct {
	ID              uint       `json:"id" db:"id"`
	Name            string     `json:"name" db:"name"`
	MainImage       string     `json:"main_image" db:"main_image"`
	Description     string     `json:"description" db:"description"`
	Powers          string     `json:"powers" db:"powers"`
	CharacterID     *uint      `json:"character_id,omitempty" db:"character_id"` // Propietario actual
	FirstAppearance int        `json:"first_appearance" db:"first_appearance"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"-" db:"updated_at"`        // Oculto en JSON
	DeletedAt       *time.Time `json:"-" db:"deleted_at"`        // Oculto en JSON
}

// Errores comunes
var (
	ErrVitalInstrumentNotFound     = errors.New("vital instrument not found")
	ErrInvalidInput                = errors.New("invalid input")
	ErrVitalInstrumentAlreadyExists = errors.New("vital instrument already exists")
)
