package favorite

import "context"

// Interfaces for writing favorite data
type FavoriteCreator interface {
    Create(ctx context.Context, fav *Favorite) error
}

type FavoriteDeleter interface {
    Delete(ctx context.Context, userID uint, entityType string, entityID uint) error
}

// Writer composes all write interfaces
type Writer interface {
    FavoriteCreator
    FavoriteDeleter
}

// Store composes all read/write operations
type Store interface {
    Reader
    Writer
}

