package postgres

import (
    "context"
    "fmt"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/favorite"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements favorite.Writer using PostgreSQL
type FavoriteWriter struct {
    db *sqlx.DB
}

func NewFavoriteWriter(db *sqlx.DB) *FavoriteWriter {
    return &FavoriteWriter{db: db}
}

func (w *FavoriteWriter) Create(ctx context.Context, f *favorite.Favorite) error {
    log := logger.GetLogger(zap.String("repository", "FavoriteWriter"), zap.String("method", "Create"))

    query := `
        INSERT INTO favorites (user_id, entity_type, entity_id, created_at)
        VALUES (:user_id, :entity_type, :entity_id, :created_at)
        ON CONFLICT (user_id, entity_type, entity_id) DO NOTHING
        RETURNING id
    `

    rows, err := w.db.NamedQueryContext(ctx, query, f)
    if err != nil {
        log.Error("Error creating favorite", zap.Error(err))
        return fmt.Errorf("insert favorite: %w", err)
    }
    defer rows.Close()

    if rows.Next() {
        if err := rows.Scan(&f.ID); err != nil {
            log.Error("Error scanning ID", zap.Error(err))
            return fmt.Errorf("scan ID: %w", err)
        }
    }

    return nil
}

func (w *FavoriteWriter) Delete(ctx context.Context, userID uint, entityType string, entityID uint) error {
    log := logger.GetLogger(zap.String("repository", "FavoriteWriter"), zap.String("method", "Delete"))

    query := `
        DELETE FROM favorites
        WHERE user_id = $1 AND entity_type = $2 AND entity_id = $3
    `

    result, err := w.db.ExecContext(ctx, query, userID, entityType, entityID)
    if err != nil {
        log.Error("Error deleting favorite", zap.Error(err))
        return fmt.Errorf("delete favorite: %w", err)
    }

    if rows, _ := result.RowsAffected(); rows == 0 {
        log.Warn("Favorite not found", zap.Uint("user_id", userID), zap.String("type", entityType), zap.Uint("entity_id", entityID))
        return favorite.ErrFavoriteNotFound
    }

    return nil
}

