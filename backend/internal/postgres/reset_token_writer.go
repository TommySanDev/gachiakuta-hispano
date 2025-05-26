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

// Implements user.ResetTokenWriter using PostgreSQL
type ResetTokenWriter struct {
    db *sqlx.DB
}

func NewResetTokenWriter(db *sqlx.DB) *ResetTokenWriter {
    return &ResetTokenWriter{
        db: db,
    }
}

func (w *ResetTokenWriter) Create(ctx context.Context, rt *user.ResetToken) error {
    log := logger.GetLogger(zap.String("repository", "ResetTokenWriter"), zap.String("method", "Create"))
    
    query := `
        INSERT INTO reset_tokens (
            user_id, token, used, expires_at, created_at, updated_at
        ) VALUES (
            :user_id, :token, :used, :expires_at, :created_at, :updated_at
        )
        RETURNING id
    `

    rows, err := w.db.NamedQueryContext(ctx, query, rt)
    if err != nil {
        log.Error("Error creating reset token", zap.Error(err), zap.Uint("user_id", rt.UserID))
        return fmt.Errorf("insert reset token: %w", err)
    }
    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&rt.ID)
        if err != nil {
            log.Error("Error scanning ID", zap.Error(err))
            return fmt.Errorf("scan ID: %w", err)
        }
    }
    
    return nil
}

func (w *ResetTokenWriter) MarkAsUsed(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "ResetTokenWriter"), zap.String("method", "MarkAsUsed"))
    
    query := `
        UPDATE reset_tokens
        SET used = true, updated_at = :updated_at
        WHERE id = :id AND used = false
    `

    params := map[string]interface{}{
        "id":         id,
        "updated_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
    if err != nil {
        log.Error("Error marking reset token as used", zap.Error(err), zap.Uint("id", id))
        return fmt.Errorf("mark reset token as used: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No reset token found to mark as used", zap.Uint("id", id))
        return user.ErrResetTokenNotFound
    }

    return nil
}

func (w *ResetTokenWriter) InvalidateByUserID(ctx context.Context, userID uint) error {
    log := logger.GetLogger(zap.String("repository", "ResetTokenWriter"), zap.String("method", "InvalidateByUserID"))
    
    query := `
        UPDATE reset_tokens
        SET used = true, updated_at = :updated_at
        WHERE user_id = :user_id AND used = false
    `

    params := map[string]interface{}{
        "user_id":    userID,
        "updated_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
    if err != nil {
        log.Error("Error invalidating reset tokens", zap.Error(err), zap.Uint("user_id", userID))
        return fmt.Errorf("invalidate reset tokens: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows > 0 {
        log.Info("Reset tokens invalidated", zap.Uint("user_id", userID), zap.Int64("count", rows))
    }

    return nil
}

func (w *ResetTokenWriter) DeleteExpired(ctx context.Context) (int, error) {
    log := logger.GetLogger(zap.String("repository", "ResetTokenWriter"), zap.String("method", "DeleteExpired"))
    
    query := `DELETE FROM reset_tokens WHERE expires_at <= NOW()`

    result, err := w.db.ExecContext(ctx, query)
    if err != nil {
        log.Error("Error deleting expired reset tokens", zap.Error(err))
        return 0, fmt.Errorf("delete expired reset tokens: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return 0, fmt.Errorf("get rows affected: %w", err)
    }

    count := int(rows)
    if count > 0 {
        log.Info("Deleted expired reset tokens", zap.Int("count", count))
    }

    return count, nil
}
