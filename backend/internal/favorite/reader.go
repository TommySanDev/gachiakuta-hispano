package favorite

import "context"

// Interfaces for reading favorite data
type FavoriteLister interface {
    ListByUser(ctx context.Context, userID uint) ([]*Favorite, error)
}

// Reader composes all read interfaces
type Reader interface {
    FavoriteLister
}

