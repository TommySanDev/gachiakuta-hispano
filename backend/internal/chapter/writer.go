package chapter

import "context"

// Interfaces for writing chapter data
type ChapterCreator interface {
    Create(ctx context.Context, chapter *Chapter) error
}

type ChapterUpdater interface {
    Update(ctx context.Context, chapter *Chapter) error
}

type ChapterDeleter interface {
    Delete(ctx context.Context, id uint) error
}

type ChapterRestorer interface {
    Restore(ctx context.Context, id uint) error
}

type ChapterPermanentDeleter interface {
    DeletePermanently(ctx context.Context, id uint) error
}

// Writer composes all write interfaces
type Writer interface {
    ChapterCreator
    ChapterUpdater
    ChapterDeleter
    ChapterRestorer
    ChapterPermanentDeleter
}

// Store composes all read and write operations
type Store interface {
    Reader
    Writer
}

