package vitalinstrument

import (
	"errors"
	"time"
	
	"github.com/TommySanDev/gachiakuta-hispano/config"
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
	ErrVitalInstrumentNotFound      = errors.New("vital instrument not found")
	ErrInvalidInput                 = errors.New("invalid input")
	ErrVitalInstrumentAlreadyExists = errors.New("vital instrument already exists")
)

// GetAll retrieves all non-deleted vital instruments
func GetAll() ([]VitalInstrument, error) {
	var instruments []VitalInstrument
	err := config.DB.Select(&instruments, "SELECT * FROM vital_instruments WHERE deleted_at IS NULL ORDER BY created_at DESC")
	return instruments, err
}

// GetByID retrieves a vital instrument by ID
func GetByID(id uint) (*VitalInstrument, error) {
	var vi VitalInstrument
	err := config.DB.Get(&vi, "SELECT * FROM vital_instruments WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return nil, ErrVitalInstrumentNotFound
	}
	return &vi, nil
}

// GetByCharacterID retrieves vital instruments by character ID
func GetByCharacterID(characterID uint) ([]VitalInstrument, error) {
	var instruments []VitalInstrument
	err := config.DB.Select(&instruments, 
		"SELECT * FROM vital_instruments WHERE character_id = $1 AND deleted_at IS NULL ORDER BY name ASC", 
		characterID)
	return instruments, err
}

// Create creates a new vital instrument
func Create(vi *VitalInstrument) error {
	// Basic validation
	if vi.Name == "" {
		return ErrInvalidInput
	}
	
	// Set timestamps
	now := time.Now()
	vi.CreatedAt = now
	vi.UpdatedAt = now
	
	query := `INSERT INTO vital_instruments (
		name, main_image, description, powers, character_id,
		first_appearance, created_at, updated_at
	) VALUES (
		:name, :main_image, :description, :powers, :character_id,
		:first_appearance, :created_at, :updated_at
	) RETURNING id`
	
	rows, err := config.DB.NamedQuery(query, vi)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	if rows.Next() {
		var id uint
		rows.Scan(&id)
		vi.ID = id
	}
	
	return nil
}

// Update updates an existing vital instrument
func Update(id uint, vi *VitalInstrument) error {
	// Check if vital instrument exists
	exists, err := Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrVitalInstrumentNotFound
	}
	
	// Set ID and update timestamp
	vi.ID = id
	vi.UpdatedAt = time.Now()
	
	query := `UPDATE vital_instruments SET
		name = :name, 
		main_image = :main_image,
		description = :description,
		powers = :powers,
		character_id = :character_id,
		first_appearance = :first_appearance,
		updated_at = :updated_at
	WHERE id = :id AND deleted_at IS NULL`
	
	_, err = config.DB.NamedExec(query, vi)
	return err
}

// Delete performs soft delete on a vital instrument
func Delete(id uint) error {
	exists, err := Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrVitalInstrumentNotFound
	}
	
	now := time.Now()
	_, err = config.DB.Exec("UPDATE vital_instruments SET deleted_at = $1 WHERE id = $2", now, id)
	return err
}

// Exists checks if a vital instrument exists and is not deleted
func Exists(id uint) (bool, error) {
	var exists bool
	err := config.DB.Get(&exists, "SELECT COUNT(*) > 0 FROM vital_instruments WHERE id = $1 AND deleted_at IS NULL", id)
	return exists, err
}

// DeletePermanently permanently deletes a vital instrument (admin only)
func DeletePermanently(id uint) error {
	result, err := config.DB.Exec("DELETE FROM vital_instruments WHERE id = $1", id)
	if err != nil {
		return err
	}
	
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrVitalInstrumentNotFound
	}
	
	return nil
}

// Restore restores a soft deleted vital instrument (admin only)
func Restore(id uint) error {
	result, err := config.DB.Exec("UPDATE vital_instruments SET deleted_at = NULL, updated_at = $1 WHERE id = $2 AND deleted_at IS NOT NULL", 
		time.Now(), id)
	if err != nil {
		return err
	}
	
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrVitalInstrumentNotFound
	}
	
	return nil
}
