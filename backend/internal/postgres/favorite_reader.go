package postgres

import (
    "context"
    "fmt"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/favorite"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements favorite.Reader using PostgreSQL
type FavoriteReader struct {
    db *sqlx.DB
}

func NewFavoriteReader(db *sqlx.DB) *FavoriteReader {
    return &FavoriteReader{db: db}
}

func (r *FavoriteReader) ListByUser(ctx context.Context, userID uint) ([]*favorite.Favorite, error) {
    log := logger.GetLogger(zap.String("repository", "FavoriteReader"), zap.String("method", "ListByUser"))

    query := `
        SELECT id, user_id, entity_type, entity_id, created_at
        FROM favorites
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

    var favorites []*favorite.Favorite
    err := r.db.SelectContext(ctx, &favorites, query, userID)
    if err != nil {
        log.Error("Error listing favorites", zap.Error(err))
        return nil, fmt.Errorf("list favorites: %w", err)
    }

    return favorites, nil
}

