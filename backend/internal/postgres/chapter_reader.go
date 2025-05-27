package postgres

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/chapter"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements chapter.Reader using PostgreSQL
type ChapterReader struct {
    db *sqlx.DB
}

func NewChapterReader(db *sqlx.DB) *ChapterReader {
    return &ChapterReader{db: db}
}

func (r *ChapterReader) GetByID(ctx context.Context, id uint) (*chapter.Chapter, error) {
    log := logger.GetLogger(zap.String("repository", "ChapterReader"), zap.String("method", "GetByID"))

    query := `
        SELECT id, title, number, image, created_at, updated_at, deleted_at
        FROM chapters
        WHERE id = $1 AND deleted_at IS NULL
    `

    var ch chapter.Chapter
    err := r.db.GetContext(ctx, &ch, query, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("Chapter not found", zap.Uint("id", id))
            return nil, chapter.ErrChapterNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &ch, nil
}

func (r *ChapterReader) List(ctx context.Context, filter chapter.ChapterFilter) ([]*chapter.Chapter, int, error) {
    log := logger.GetLogger(zap.String("repository", "ChapterReader"), zap.String("method", "List"))

    whereClauses := []string{"1=1"}
    args := []interface{}{}
    argPos := 1

    if !filter.IncludeDeleted {
        whereClauses = append(whereClauses, "deleted_at IS NULL")
    }

    if filter.Search != "" {
        whereClauses = append(whereClauses, fmt.Sprintf("(title ILIKE $%d)", argPos))
        args = append(args, "%"+filter.Search+"%")
        argPos++
    }

    where := "WHERE " + strings.Join(whereClauses, " AND ")

    // Total count
    countQuery := "SELECT COUNT(*) FROM chapters " + where
    var total int
    err := r.db.GetContext(ctx, &total, countQuery, args...)
    if err != nil {
        log.Error("Error counting chapters", zap.Error(err))
        return nil, 0, fmt.Errorf("count chapters: %w", err)
    }

    // Query
    query := `
        SELECT id, title, number, image, created_at, updated_at, deleted_at
        FROM chapters ` + where

    // Sorting
    validSort := map[string]bool{
        "title":      true,
        "number":     true,
        "created_at": true,
    }
    if validSort[filter.SortBy] {
        query += fmt.Sprintf(" ORDER BY %s", filter.SortBy)
        if strings.ToLower(filter.SortDir) == "asc" {
            query += " ASC"
        } else {
            query += " DESC"
        }
    } else {
        query += " ORDER BY created_at DESC"
    }

    // Pagination
    offset := (filter.Page - 1) * filter.PageSize
    query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
    args = append(args, filter.PageSize, offset)

    var chapters []*chapter.Chapter
    err = r.db.SelectContext(ctx, &chapters, query, args...)
    if err != nil {
        log.Error("Error querying chapters", zap.Error(err))
        return nil, 0, fmt.Errorf("query chapters: %w", err)
    }

    return chapters, total, nil
}

