package postgres

import (
    "context"
    "fmt"
    "time"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
    "github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// Implements user.MagicLinkWriter using PostgreSQL
type MagicLinkWriter struct {
    db *sqlx.DB
}

func NewMagicLinkWriter(db *sqlx.DB) *MagicLinkWriter {
    return &MagicLinkWriter{
        db: db,
    }
}

func (w *MagicLinkWriter) Create(ctx context.Context, ml *user.MagicLink) error {
    log := logger.GetLogger(zap.String("repository", "MagicLinkWriter"), zap.String("method", "Create"))
    
    query := `
        INSERT INTO magic_links (
            user_id, token, used, expires_at, created_at, updated_at
        ) VALUES (
            :user_id, :token, :used, :expires_at, :created_at, :updated_at
        )
        RETURNING id
    `

    rows, err := w.db.NamedQueryContext(ctx, query, ml)
    if err != nil {
        log.Error("Error creating magic link", zap.Error(err), zap.Uint("user_id", ml.UserID))
        return fmt.Errorf("insert magic link: %w", err)
    }
    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&ml.ID)
        if err != nil {
            log.Error("Error scanning ID", zap.Error(err))
            return fmt.Errorf("scan ID: %w", err)
        }
    }
    
    return nil
}

func (w *MagicLinkWriter) MarkAsUsed(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "MagicLinkWriter"), zap.String("method", "MarkAsUsed"))
    
    query := `
        UPDATE magic_links
        SET used = true, updated_at = :updated_at
        WHERE id = :id AND used = false
    `

    params := map[string]interface{}{
        "id":         id,
        "updated_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
    if err != nil {
        log.Error("Error marking magic link as used", zap.Error(err), zap.Uint("id", id))
        return fmt.Errorf("mark magic link as used: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No magic link found to mark as used", zap.Uint("id", id))
        return user.ErrMagicLinkNotFound
    }

    return nil
}

func (w *MagicLinkWriter) DeleteExpired(ctx context.Context) (int, error) {
    log := logger.GetLogger(zap.String("repository", "MagicLinkWriter"), zap.String("method", "DeleteExpired"))
    
    query := `DELETE FROM magic_links WHERE expires_at <= NOW()`

    result, err := w.db.ExecContext(ctx, query)
    if err != nil {
        log.Error("Error deleting expired magic links", zap.Error(err))
        return 0, fmt.Errorf("delete expired magic links: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return 0, fmt.Errorf("get rows affected: %w", err)
    }

    count := int(rows)
    if count > 0 {
        log.Info("Deleted expired magic links", zap.Int("count", count))
    }

    return count, nil
}
