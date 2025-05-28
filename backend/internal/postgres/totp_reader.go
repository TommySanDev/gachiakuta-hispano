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

// Implements user.TOTPReader using PostgreSQL
type TOTPReader struct {
    db *sqlx.DB
}

func NewTOTPReader(db *sqlx.DB) *TOTPReader {
    return &TOTPReader{
        db: db,
    }
}

func (r *TOTPReader) GetByUserID(ctx context.Context, userID uint) (*user.TOTPSecret, error) {
    return r.GetTOTPByUserID(ctx, userID)
}

func (r *TOTPReader) GetRecoveryCodes(ctx context.Context, userID uint) ([]*user.RecoveryCode, error) {
    log := logger.GetLogger(
        zap.String("repository", "TOTPReader"), 
        zap.String("method", "GetRecoveryCodes"),
    )
    
    query := `
        SELECT id, user_id, code, used, created_at, updated_at
        FROM recovery_codes
        WHERE user_id = $1 AND used = false
        ORDER BY created_at ASC
    `

    var codes []*user.RecoveryCode
    err := r.db.SelectContext(ctx, &codes, query, userID)
    if err != nil {
        log.Error("Error querying recovery codes", zap.Error(err), zap.Uint("user_id", userID))
        return nil, fmt.Errorf("query recovery codes: %w", err)
    }

    return codes, nil
}
