package character

import "time"

type Character struct {
    ID              uint      `json:"id" db:"id"`
    Name            string    `json:"name" db:"name"`
    NameJapanese    string    `json:"name_japanese" db:"name_japanese"`
    MainImage       string    `json:"main_image" db:"main_image"`
    Description     string    `json:"description" db:"description"`
    Species         string    `json:"species" db:"species"`
    Gender          string    `json:"gender" db:"gender"`
    Age             int       `json:"age" db:"age"`
    Height          string    `json:"height" db:"height"`
    Status          string    `json:"status" db:"status"`
    Affiliation     string    `json:"affiliation" db:"affiliation"`
    Occupation      string    `json:"occupation" db:"occupation"`
    BirthDate       string    `json:"birth_date" db:"birth_date"`
    BirthPlace      string    `json:"birth_place" db:"birth_place"`
    Relatives       string    `json:"relatives" db:"relatives"`
    FirstAppearance int       `json:"first_appearance" db:"first_appearance"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
    DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Dto for creating and updating a character
type CreateCharacterInput struct {
    Name            string `json:"name" validate:"required"`
    NameJapanese    string `json:"name_japanese"`
    MainImage       string `json:"main_image" validate:"required"`
    Description     string `json:"description" validate:"required"`
    Species         string `json:"species"`
    Gender          string `json:"gender"`
    Age             int    `json:"age"`
    Height          string `json:"height"`
    Status          string `json:"status" validate:"required"`
    Affiliation     string `json:"affiliation"`
    Occupation      string `json:"occupation"`
    BirthDate       string `json:"birth_date"`
    BirthPlace      string `json:"birth_place"`
    Relatives       string `json:"relatives"`
    FirstAppearance int    `json:"first_appearance" validate:"required"`
}

type UpdateCharacterInput struct {
    Name            *string `json:"name,omitempty"`
    NameJapanese    *string `json:"name_japanese,omitempty"`
    MainImage       *string `json:"main_image,omitempty"`
    Description     *string `json:"description,omitempty"`
    Species         *string `json:"species,omitempty"`
    Gender          *string `json:"gender,omitempty"`
    Age             *int    `json:"age,omitempty"`
    Height          *string `json:"height,omitempty"`
    Status          *string `json:"status,omitempty"`
    Affiliation     *string `json:"affiliation,omitempty"`
    Occupation      *string `json:"occupation,omitempty"`
    BirthDate       *string `json:"birth_date,omitempty"`
    BirthPlace      *string `json:"birth_place,omitempty"`
    Relatives       *string `json:"relatives,omitempty"`
    FirstAppearance *int    `json:"first_appearance,omitempty"`
}

// Parameters for filtering characters
type CharacterFilter struct {
    Search          string  `json:"search"`
    Status          string  `json:"status,omitempty"`
    Affiliation     string  `json:"affiliation,omitempty"`
    Species         string  `json:"species,omitempty"`
    SortBy          string  `json:"sort_by,omitempty"`
    SortDir         string  `json:"sort_dir,omitempty"`
    Page            int     `json:"page"`
    PageSize        int     `json:"page_size"`
    IncludeDeleted  bool    `json:"include_deleted,omitempty"`
}
