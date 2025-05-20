package character

import "context"

// Interfaces for reading character data
type CharacterGetter interface {
    GetByID(ctx context.Context, id uint) (*Character, error)
}

type CharacterLister interface {
    List(ctx context.Context, filter CharacterFilter) ([]*Character, int, error)
}

type AffiliationCharacterLister interface {
    ListByAffiliation(ctx context.Context, affiliation string, limit int) ([]*Character, error)
}

type StatusCharacterLister interface {
    ListByStatus(ctx context.Context, status string, limit int) ([]*Character, error)
}

type SpeciesCharacterLister interface {
    ListBySpecies(ctx context.Context, species string, limit int) ([]*Character, error)
}

// Reader composes all read interfaces
type Reader interface {
    CharacterGetter
    CharacterLister
    AffiliationCharacterLister
    StatusCharacterLister
    SpeciesCharacterLister
}
