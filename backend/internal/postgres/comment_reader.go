package postgres

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/comment"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements comment.Reader using PostgreSQL
type CommentReader struct {
    db *sqlx.DB
}

func NewCommentReader(db *sqlx.DB) *CommentReader {
    return &CommentReader{db: db}
}

func (r *CommentReader) GetByID(ctx context.Context, id uint) (*comment.Comment, error) {
    log := logger.GetLogger(zap.String("repository", "CommentReader"), zap.String("method", "GetByID"))

    query := `
        SELECT id, user_id, chapter_id, content, created_at, updated_at, deleted_at
        FROM comments
        WHERE id = $1
    `

    var c comment.Comment
    err := r.db.GetContext(ctx, &c, query, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, comment.ErrCommentNotFound
        }
        log.Error("Error getting comment", zap.Error(err))
        return nil, fmt.Errorf("get comment: %w", err)
    }

    return &c, nil
}

func (r *CommentReader) ListByChapter(ctx context.Context, filter comment.ChapterCommentFilter) ([]*comment.Comment, error) {
    log := logger.GetLogger(zap.String("repository", "CommentReader"), zap.String("method", "ListByChapter"))

    query := `
        SELECT id, user_id, chapter_id, content, created_at, updated_at, deleted_at
        FROM comments
        WHERE chapter_id = $1
    `
    if !filter.IncludeDeleted {
        query += " AND deleted_at IS NULL"
    }
    query += " ORDER BY created_at ASC"

    var comments []*comment.Comment
    err := r.db.SelectContext(ctx, &comments, query, filter.ChapterID)
    if err != nil {
        log.Error("Error listing comments", zap.Error(err))
        return nil, fmt.Errorf("list comments: %w", err)
    }

    return comments, nil
}

