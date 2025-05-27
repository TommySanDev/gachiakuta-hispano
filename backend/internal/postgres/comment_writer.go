package postgres

import (
    "context"
    "fmt"
    "time"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/comment"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements comment.Writer using PostgreSQL
type CommentWriter struct {
    db *sqlx.DB
}

func NewCommentWriter(db *sqlx.DB) *CommentWriter {
    return &CommentWriter{db: db}
}

func (w *CommentWriter) Create(ctx context.Context, c *comment.Comment) error {
    log := logger.GetLogger(zap.String("repository", "CommentWriter"), zap.String("method", "Create"))

    query := `
        INSERT INTO comments (user_id, chapter_id, content, created_at, updated_at)
        VALUES (:user_id, :chapter_id, :content, :created_at, :updated_at)
        RETURNING id
    `

    rows, err := w.db.NamedQueryContext(ctx, query, c)
    if err != nil {
        log.Error("Error creating comment", zap.Error(err))
        return fmt.Errorf("insert comment: %w", err)
    }
    defer rows.Close()

    if rows.Next() {
        if err := rows.Scan(&c.ID); err != nil {
            return fmt.Errorf("scan ID: %w", err)
        }
    }

    return nil
}

func (w *CommentWriter) Update(ctx context.Context, c *comment.Comment) error {
    log := logger.GetLogger(zap.String("repository", "CommentWriter"), zap.String("method", "Update"))

    query := `
        UPDATE comments
        SET content = :content,
            updated_at = :updated_at
        WHERE id = :id
    `
    _, err := w.db.NamedExecContext(ctx, query, c)
    if err != nil {
        log.Error("Error updating comment", zap.Error(err))
        return fmt.Errorf("update comment: %w", err)
    }

    return nil
}

func (w *CommentWriter) Delete(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "CommentWriter"), zap.String("method", "Delete"))

    query := `
        UPDATE comments
        SET deleted_at = $1
        WHERE id = $2 AND deleted_at IS NULL
    `
    _, err := w.db.ExecContext(ctx, query, time.Now(), id)
    if err != nil {
        log.Error("Error soft deleting comment", zap.Error(err))
        return fmt.Errorf("soft delete comment: %w", err)
    }

    return nil
}

func (w *CommentWriter) DeletePermanently(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "CommentWriter"), zap.String("method", "DeletePermanently"))

    query := `DELETE FROM comments WHERE id = $1`
    _, err := w.db.ExecContext(ctx, query, id)
    if err != nil {
        log.Error("Error permanently deleting comment", zap.Error(err))
        return fmt.Errorf("delete comment permanently: %w", err)
    }

    return nil
}

