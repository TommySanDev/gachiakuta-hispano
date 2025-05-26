package postgres

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
    "github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// Implements user session reading operations using PostgreSQL
type SessionReader struct {
    db *sqlx.DB
}

func NewSessionReader(db *sqlx.DB) *SessionReader {
    return &SessionReader{
        db: db,
    }
}

func (r *SessionReader) GetByID(ctx context.Context, id string) (*user.Session, error) {
    log := logger.GetLogger(zap.String("repository", "SessionReader"), zap.String("method", "GetByID"))
    
    query := `
        SELECT id, user_id, token, user_agent, ip_address, expires_at, created_at, updated_at
        FROM sessions
        WHERE id = $1 AND expires_at > NOW()
    `

    var s user.Session
    err := r.db.GetContext(ctx, &s, query, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("Session not found", zap.String("id", id))
            return nil, user.ErrSessionNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &s, nil
}

func (r *SessionReader) GetByToken(ctx context.Context, token string) (*user.Session, error) {
    log := logger.GetLogger(zap.String("repository", "SessionReader"), zap.String("method", "GetByToken"))
    
    query := `
        SELECT id, user_id, token, user_agent, ip_address, expires_at, created_at, updated_at
        FROM sessions
        WHERE token = $1 AND expires_at > NOW()
    `

    var s user.Session
    err := r.db.GetContext(ctx, &s, query, token)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("Session not found", zap.String("token", "***"))
            return nil, user.ErrSessionNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &s, nil
}

func (r *SessionReader) ListByUserID(ctx context.Context, userID uint, limit int) ([]*user.Session, error) {
    log := logger.GetLogger(
        zap.String("repository", "SessionReader"), 
        zap.String("method", "ListByUserID"),
        zap.Uint("userID", userID),
    )
    
    query := `
        SELECT id, user_id, token, user_agent, ip_address, expires_at, created_at, updated_at
        FROM sessions
        WHERE user_id = $1 AND expires_at > NOW()
        ORDER BY created_at DESC
        LIMIT $2
    `

    var sessions []*user.Session
    err := r.db.SelectContext(ctx, &sessions, query, userID, limit)
    if err != nil {
        log.Error("Error querying sessions by user", zap.Error(err))
        return nil, fmt.Errorf("query sessions by user: %w", err)
    }

    return sessions, nil
}

func (r *SessionReader) ListExpiredSessions(ctx context.Context, limit int) ([]*user.Session, error) {
    log := logger.GetLogger(
        zap.String("repository", "SessionReader"), 
        zap.String("method", "ListExpiredSessions"),
    )
    
    query := `
        SELECT id, user_id, token, user_agent, ip_address, expires_at, created_at, updated_at
        FROM sessions
        WHERE expires_at <= NOW()
        ORDER BY expires_at ASC
        LIMIT $1
    `

    var sessions []*user.Session
    err := r.db.SelectContext(ctx, &sessions, query, limit)
    if err != nil {
        log.Error("Error querying expired sessions", zap.Error(err))
        return nil, fmt.Errorf("query expired sessions: %w", err)
    }

    return sessions, nil
}
