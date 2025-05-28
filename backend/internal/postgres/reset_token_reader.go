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

// Implements user.ResetTokenReader using PostgreSQL
type ResetTokenReader struct {
    db *sqlx.DB
}

func NewResetTokenReader(db *sqlx.DB) *ResetTokenReader {
    return &ResetTokenReader{
        db: db,
    }
}

func (r *ResetTokenReader) GetByToken(ctx context.Context, token string) (*user.ResetToken, error) {
    return r.GetResetTokenByToken(ctx, token)
}

func (r *ResetTokenReader) GetActiveByUserID(ctx context.Context, userID uint) ([]*user.ResetToken, error) {
    log := logger.GetLogger(
        zap.String("repository", "ResetTokenReader"), 
        zap.String("method", "GetActiveByUserID"),
    )
    
    query := `
        SELECT id, user_id, token, used, expires_at, created_at, updated_at
        FROM reset_tokens
        WHERE user_id = $1 AND used = false AND expires_at > NOW()
        ORDER BY created_at DESC
    `

    var tokens []*user.ResetToken
    err := r.db.SelectContext(ctx, &tokens, query, userID)
    if err != nil {
        log.Error("Error querying active reset tokens by user", zap.Error(err), zap.Uint("user_id", userID))
        return nil, fmt.Errorf("query active reset tokens: %w", err)
    }

    return tokens, nil
}
