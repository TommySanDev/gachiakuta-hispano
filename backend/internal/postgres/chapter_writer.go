package postgres

import (
    "context"
    "fmt"
    "time"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/chapter"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements chapter.Writer using PostgreSQL
type ChapterWriter struct {
    db *sqlx.DB
}

func NewChapterWriter(db *sqlx.DB) *ChapterWriter {
    return &ChapterWriter{db: db}
}

func (w *ChapterWriter) Create(ctx context.Context, c *chapter.Chapter) error {
    log := logger.GetLogger(zap.String("repository", "ChapterWriter"), zap.String("method", "Create"))

    query := `
        INSERT INTO chapters (title, number, image, created_at, updated_at)
        VALUES (:title, :number, :image, :created_at, :updated_at)
        RETURNING id
    `

    rows, err := w.db.NamedQueryContext(ctx, query, c)
    if err != nil {
        log.Error("Error creating chapter", zap.Error(err))
        return fmt.Errorf("insert chapter: %w", err)
    }
    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&c.ID)
        if err != nil {
            log.Error("Error scanning ID", zap.Error(err))
            return fmt.Errorf("scan ID: %w", err)
        }
    }

    return nil
}

func (w *ChapterWriter) Update(ctx context.Context, c *chapter.Chapter) error {
    log := logger.GetLogger(zap.String("repository", "ChapterWriter"), zap.String("method", "Update"))

    query := `
        UPDATE chapters
        SET title = :title,
            number = :number,
            image = :image,
            updated_at = :updated_at
        WHERE id = :id AND deleted_at IS NULL
    `

    result, err := w.db.NamedExecContext(ctx, query, c)
    if err != nil {
        log.Error("Error updating chapter", zap.Error(err))
        return fmt.Errorf("update chapter: %w", err)
    }

    if rows, _ := result.RowsAffected(); rows == 0 {
        log.Warn("No chapter found to update", zap.Uint("id", c.ID))
        return chapter.ErrChapterNotFound
    }

    return nil
}

func (w *ChapterWriter) Delete(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "ChapterWriter"), zap.String("method", "Delete"))

    query := `
        UPDATE chapters
        SET deleted_at = :deleted_at
        WHERE id = :id AND deleted_at IS NULL
    `
    params := map[string]interface{}{
        "id":         id,
        "deleted_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
    if err != nil {
        log.Error("Error soft deleting chapter", zap.Error(err))
        return fmt.Errorf("delete chapter: %w", err)
    }

    if rows, _ := result.RowsAffected(); rows == 0 {
        log.Warn("No chapter found to delete", zap.Uint("id", id))
        return chapter.ErrChapterNotFound
    }

    return nil
}

func (w *ChapterWriter) DeletePermanently(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "ChapterWriter"), zap.String("method", "DeletePermanently"))

    query := `DELETE FROM chapters WHERE id = $1`
    result, err := w.db.ExecContext(ctx, query, id)
    if err != nil {
        log.Error("Error permanently deleting chapter", zap.Error(err))
        return fmt.Errorf("delete permanently: %w", err)
    }

    if rows, _ := result.RowsAffected(); rows == 0 {
        log.Warn("No chapter found to delete permanently", zap.Uint("id", id))
        return chapter.ErrChapterNotFound
    }

    return nil
}

func (w *ChapterWriter) Restore(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "ChapterWriter"), zap.String("method", "Restore"))

    query := `
        UPDATE chapters
        SET deleted_at = NULL, updated_at = :updated_at
        WHERE id = :id AND deleted_at IS NOT NULL
    `
    params := map[string]interface{}{
        "id":         id,
        "updated_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
    if err != nil {
        log.Error("Error restoring chapter", zap.Error(err))
        return fmt.Errorf("restore chapter: %w", err)
    }

    if rows, _ := result.RowsAffected(); rows == 0 {
        log.Warn("No chapter found to restore", zap.Uint("id", id))
        return chapter.ErrChapterNotFound
    }

    return nil
}

