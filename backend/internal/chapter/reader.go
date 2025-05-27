package chapter

import "context"

// Interfaces for reading chapter data
type ChapterGetter interface {
    GetByID(ctx context.Context, id uint) (*Chapter, error)
}

type ChapterLister interface {
    List(ctx context.Context, filter ChapterFilter) ([]*Chapter, int, error)
}

// Reader composes all read interfaces
type Reader interface {
    ChapterGetter
    ChapterLister
}

