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

// Implements user session writing operations using PostgreSQL
type SessionWriter struct {
    db *sqlx.DB
}

func NewSessionWriter(db *sqlx.DB) *SessionWriter {
    return &SessionWriter{
        db: db,
    }
}

func (w *SessionWriter) Create(ctx context.Context, session *user.Session) error {
    log := logger.GetLogger(zap.String("repository", "SessionWriter"), zap.String("method", "Create"))
    
    query := `
        INSERT INTO sessions (
            id, user_id, token, user_agent, ip_address, expires_at, created_at, updated_at
        ) VALUES (
            :id, :user_id, :token, :user_agent, :ip_address, :expires_at, :created_at, :updated_at
        )
    `

    _, err := w.db.NamedExecContext(ctx, query, session)
    if err != nil {
        log.Error("Error creating session", zap.Error(err), zap.String("sessionID", session.ID))
        return fmt.Errorf("insert session: %w", err)
    }
    
    return nil
}

func (w *SessionWriter) Update(ctx context.Context, session *user.Session) error {
    log := logger.GetLogger(zap.String("repository", "SessionWriter"), zap.String("method", "Update"))
    
    query := `
        UPDATE sessions
        SET user_agent = :user_agent,
            ip_address = :ip_address,
            expires_at = :expires_at,
            updated_at = :updated_at
        WHERE id = :id
    `

    result, err := w.db.NamedExecContext(ctx, query, session)
    if err != nil {
        log.Error("Error updating session", zap.Error(err), zap.String("sessionID", session.ID))
        return fmt.Errorf("update session: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No session found to update", zap.String("sessionID", session.ID))
        return user.ErrSessionNotFound
    }

    return nil
}

func (w *SessionWriter) Delete(ctx context.Context, id string) error {
    log := logger.GetLogger(zap.String("repository", "SessionWriter"), zap.String("method", "Delete"))
    
    query := `DELETE FROM sessions WHERE id = $1`

    result, err := w.db.ExecContext(ctx, query, id)
    if err != nil {
        log.Error("Error deleting session", zap.Error(err), zap.String("sessionID", id))
        return fmt.Errorf("delete session: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No session found to delete", zap.String("sessionID", id))
        return user.ErrSessionNotFound
    }

    return nil
}

func (w *SessionWriter) DeleteByUserID(ctx context.Context, userID uint) error {
    log := logger.GetLogger(
        zap.String("repository", "SessionWriter"), 
        zap.String("method", "DeleteByUserID"),
        zap.Uint("userID", userID),
    )
    
    query := `DELETE FROM sessions WHERE user_id = $1`

    result, err := w.db.ExecContext(ctx, query, userID)
    if err != nil {
        log.Error("Error deleting sessions by user", zap.Error(err))
        return fmt.Errorf("delete sessions by user: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    log.Info("Deleted user sessions", zap.Int64("deletedSessions", rows))
    return nil
}

func (w *SessionWriter) DeleteExpiredSessions(ctx context.Context) (int64, error) {
    log := logger.GetLogger(zap.String("repository", "SessionWriter"), zap.String("method", "DeleteExpiredSessions"))
    
    query := `DELETE FROM sessions WHERE expires_at <= NOW()`

    result, err := w.db.ExecContext(ctx, query)
    if err != nil {
        log.Error("Error deleting expired sessions", zap.Error(err))
        return 0, fmt.Errorf("delete expired sessions: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return 0, fmt.Errorf("get rows affected: %w", err)
    }

    if rows > 0 {
        log.Info("Deleted expired sessions", zap.Int64("deletedSessions", rows))
    }

    return rows, nil
}
