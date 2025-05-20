package character

import "context"

// Interfaces for writing character data
type CharacterCreator interface {
    Create(ctx context.Context, character *Character) error
}

type CharacterUpdater interface {
    Update(ctx context.Context, character *Character) error
}

type CharacterDeleter interface {
    Delete(ctx context.Context, id uint) error
}

type CharacterRestorer interface {
    Restore(ctx context.Context, id uint) error
}

type CharacterPermanentDeleter interface {
    DeletePermanently(ctx context.Context, id uint) error
}

// Writer composes all write interfaces
type Writer interface {
    CharacterCreator
    CharacterUpdater
    CharacterDeleter
    CharacterRestorer
    CharacterPermanentDeleter
}

// Store composes all read and write operations
type Store interface {
    Reader
    Writer
}
