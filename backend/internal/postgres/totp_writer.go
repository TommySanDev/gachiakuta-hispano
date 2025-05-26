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

// Implements user.TOTPWriter using PostgreSQL
type TOTPWriter struct {
    db *sqlx.DB
}

func NewTOTPWriter(db *sqlx.DB) *TOTPWriter {
    return &TOTPWriter{
        db: db,
    }
}

func (w *TOTPWriter) CreateTOTPSecret(ctx context.Context, totp *user.TOTPSecret) error {
    log := logger.GetLogger(zap.String("repository", "TOTPWriter"), zap.String("method", "CreateTOTPSecret"))
    
    query := `
        INSERT INTO totp_secrets (
            user_id, secret, verified, created_at, updated_at
        ) VALUES (
            :user_id, :secret, :verified, :created_at, :updated_at
        )
        RETURNING id
    `

    rows, err := w.db.NamedQueryContext(ctx, query, totp)
    if err != nil {
        log.Error("Error creating TOTP secret", zap.Error(err), zap.Uint("user_id", totp.UserID))
        return fmt.Errorf("insert TOTP secret: %w", err)
    }
    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&totp.ID)
        if err != nil {
            log.Error("Error scanning ID", zap.Error(err))
            return fmt.Errorf("scan ID: %w", err)
        }
    }
    
    return nil
}

func (w *TOTPWriter) VerifyTOTPSecret(ctx context.Context, userID uint) error {
    log := logger.GetLogger(zap.String("repository", "TOTPWriter"), zap.String("method", "VerifyTOTPSecret"))
    
    query := `
        UPDATE totp_secrets
        SET verified = true, updated_at = :updated_at
        WHERE user_id = :user_id
    `

    params := map[string]interface{}{
        "user_id":    userID,
        "updated_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
    if err != nil {
        log.Error("Error verifying TOTP secret", zap.Error(err), zap.Uint("user_id", userID))
        return fmt.Errorf("verify TOTP secret: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No TOTP secret found to verify", zap.Uint("user_id", userID))
        return user.ErrTOTPNotFound
    }

    return nil
}

func (w *TOTPWriter) DeleteTOTPSecret(ctx context.Context, userID uint) error {
    log := logger.GetLogger(zap.String("repository", "TOTPWriter"), zap.String("method", "DeleteTOTPSecret"))
    
    query := `DELETE FROM totp_secrets WHERE user_id = $1`

    result, err := w.db.ExecContext(ctx, query, userID)
    if err != nil {
        log.Error("Error deleting TOTP secret", zap.Error(err), zap.Uint("user_id", userID))
        return fmt.Errorf("delete TOTP secret: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No TOTP secret found to delete", zap.Uint("user_id", userID))
        return user.ErrTOTPNotFound
    }

    return nil
}

func (w *TOTPWriter) CreateRecoveryCodes(ctx context.Context, codes []*user.RecoveryCode) error {
    log := logger.GetLogger(zap.String("repository", "TOTPWriter"), zap.String("method", "CreateRecoveryCodes"))
    
    if len(codes) == 0 {
        return nil
    }

    query := `
        INSERT INTO recovery_codes (
            user_id, code, used, created_at, updated_at
        ) VALUES (
            :user_id, :code, :used, :created_at, :updated_at
        )
    `

    _, err := w.db.NamedExecContext(ctx, query, codes)
    if err != nil {
        log.Error("Error creating recovery codes", zap.Error(err))
        return fmt.Errorf("insert recovery codes: %w", err)
    }

    log.Info("Recovery codes created", zap.Int("count", len(codes)))
    return nil
}

func (w *TOTPWriter) UseRecoveryCode(ctx context.Context, userID uint, code string) error {
    log := logger.GetLogger(zap.String("repository", "TOTPWriter"), zap.String("method", "UseRecoveryCode"))
    
    query := `
        UPDATE recovery_codes
        SET used = true, updated_at = :updated_at
        WHERE user_id = :user_id AND code = :code AND used = false
    `

    params := map[string]interface{}{
        "user_id":    userID,
        "code":       code,
        "updated_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
    if err != nil {
        log.Error("Error using recovery code", zap.Error(err), zap.Uint("user_id", userID))
        return fmt.Errorf("use recovery code: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("Recovery code not found or already used", zap.Uint("user_id", userID))
        return user.ErrInvalidRecoveryCode
    }

    return nil
}

func (w *TOTPWriter) DeleteRecoveryCodes(ctx context.Context, userID uint) error {
    log := logger.GetLogger(zap.String("repository", "TOTPWriter"), zap.String("method", "DeleteRecoveryCodes"))
    
    query := `DELETE FROM recovery_codes WHERE user_id = $1`

    result, err := w.db.ExecContext(ctx, query, userID)
    if err != nil {
        log.Error("Error deleting recovery codes", zap.Error(err), zap.Uint("user_id", userID))
        return fmt.Errorf("delete recovery codes: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    log.Info("Recovery codes deleted", zap.Uint("user_id", userID), zap.Int64("count", rows))
    return nil
}
