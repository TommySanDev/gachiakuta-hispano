package vitalinstrument

import "time"

type VitalInstrument struct {
    ID              uint      `json:"id" db:"id"`
    Name            string    `json:"name" db:"name"`
    MainImage       string    `json:"main_image" db:"main_image"`
    Description     string    `json:"description" db:"description"`
    Powers          string    `json:"powers" db:"powers"`
    CharacterID     *uint     `json:"character_id,omitempty" db:"character_id"` // Current owner
    FirstAppearance int       `json:"first_appearance" db:"first_appearance"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
    DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Dto for creating and updating a vital instrument
type CreateVitalInstrumentInput struct {
    Name            string `json:"name" validate:"required"`
    MainImage       string `json:"main_image" validate:"required"`
    Description     string `json:"description" validate:"required"`
    Powers          string `json:"powers"`
    CharacterID     *uint  `json:"character_id,omitempty"`
    FirstAppearance int    `json:"first_appearance" validate:"required"`
}

type UpdateVitalInstrumentInput struct {
    Name            *string `json:"name,omitempty"`
    MainImage       *string `json:"main_image,omitempty"`
    Description     *string `json:"description,omitempty"`
    Powers          *string `json:"powers,omitempty"`
    CharacterID     *uint   `json:"character_id,omitempty"`
    FirstAppearance *int    `json:"first_appearance,omitempty"`
}

// Parameters for filtering vital instruments
type VitalInstrumentFilter struct {
    Search          string `json:"search"`
    CharacterID     *uint  `json:"character_id,omitempty"`
    SortBy          string `json:"sort_by,omitempty"`
    SortDir         string `json:"sort_dir,omitempty"`
    Page            int    `json:"page"`
    PageSize        int    `json:"page_size"`
    IncludeDeleted  bool   `json:"include_deleted,omitempty"`
}
