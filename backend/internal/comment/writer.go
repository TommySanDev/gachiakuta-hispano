package comment

import "context"

// Interfaces for writing comment data
type CommentCreator interface {
    Create(ctx context.Context, comment *Comment) error
}

type CommentUpdater interface {
    Update(ctx context.Context, comment *Comment) error
}

type CommentDeleter interface {
    Delete(ctx context.Context, id uint) error
}

type CommentPermanentDeleter interface {
    DeletePermanently(ctx context.Context, id uint) error
}

// Writer composes all write interfaces
type Writer interface {
    CommentCreator
    CommentUpdater
    CommentDeleter
    CommentPermanentDeleter
}

// Store composes all read/write operations
type Store interface {
    Reader
    Writer
}

