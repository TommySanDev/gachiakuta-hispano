package comment

import "context"

// Interfaces for reading comment data
type CommentGetter interface {
    GetByID(ctx context.Context, id uint) (*Comment, error)
}

type CommentLister interface {
    ListByChapter(ctx context.Context, filter ChapterCommentFilter) ([]*Comment, error)
}

// Reader composes all read interfaces
type Reader interface {
    CommentGetter
    CommentLister
}

