package postgres

import (
    "context"
    "database/sql"
    "errors"
    "fmt"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
    "github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// Implements user.MagicLinkReader using PostgreSQL
type MagicLinkReader struct {
    db *sqlx.DB
}

func NewMagicLinkReader(db *sqlx.DB) *MagicLinkReader {
    return &MagicLinkReader{
        db: db,
    }
}

func (r *MagicLinkReader) GetByToken(ctx context.Context, token string) (*user.MagicLink, error) {
    log := logger.GetLogger(zap.String("repository", "MagicLinkReader"), zap.String("method", "GetByToken"))
    
    query := `
        SELECT id, user_id, token, used, expires_at, created_at, updated_at
        FROM magic_links
        WHERE token = $1 AND used = false
    `

    var ml user.MagicLink
    err := r.db.GetContext(ctx, &ml, query, token)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("Magic link not found", zap.String("token_prefix", token[:8]+"..."))
            return nil, user.ErrMagicLinkNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &ml, nil
}

func (r *MagicLinkReader) GetActiveByUserID(ctx context.Context, userID uint) ([]*user.MagicLink, error) {
    log := logger.GetLogger(
        zap.String("repository", "MagicLinkReader"), 
        zap.String("method", "GetActiveByUserID"),
    )
    
    query := `
        SELECT id, user_id, token, used, expires_at, created_at, updated_at
        FROM magic_links
        WHERE user_id = $1 AND used = false AND expires_at > NOW()
        ORDER BY created_at DESC
    `

    var links []*user.MagicLink
    err := r.db.SelectContext(ctx, &links, query, userID)
    if err != nil {
        log.Error("Error querying active magic links by user", zap.Error(err), zap.Uint("user_id", userID))
        return nil, fmt.Errorf("query active magic links: %w", err)
    }

    return links, nil
}
