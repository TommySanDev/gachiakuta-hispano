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

// Implements user.Writer using PostgreSQL
type UserWriter struct {
    db *sqlx.DB
}

func NewUserWriter(db *sqlx.DB) *UserWriter {
    return &UserWriter{
        db: db,
    }
}

func (w *UserWriter) Create(ctx context.Context, u *user.User) error {
    log := logger.GetLogger(zap.String("repository", "UserWriter"), zap.String("method", "Create"))
    
    query := `
        INSERT INTO users (
            email, username, password_hash, first_name, last_name, role,
            is_active, email_verified, magic_link_enabled, totp_enabled,
            created_at, updated_at
        ) VALUES (
            :email, :username, :password_hash, :first_name, :last_name, :role,
            :is_active, :email_verified, :magic_link_enabled, :totp_enabled,
            :created_at, :updated_at
        )
        RETURNING id
    `

    rows, err := w.db.NamedQueryContext(ctx, query, u)
    if err != nil {
        log.Error("Error creating user", zap.Error(err), zap.String("email", u.Email))
        return fmt.Errorf("insert user: %w", err)
    }
    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&u.ID)
        if err != nil {
            log.Error("Error scanning ID", zap.Error(err))
            return fmt.Errorf("scan ID: %w", err)
        }
    }
    
    return nil
}

func (w *UserWriter) Update(ctx context.Context, u *user.User) error {
    log := logger.GetLogger(zap.String("repository", "UserWriter"), zap.String("method", "Update"))
    
    query := `
        UPDATE users
        SET username = :username,
            password_hash = :password_hash,
            first_name = :first_name,
            last_name = :last_name,
            role = :role,
            is_active = :is_active,
            email_verified = :email_verified,
            magic_link_enabled = :magic_link_enabled,
            totp_enabled = :totp_enabled,
            updated_at = :updated_at
        WHERE id = :id AND deleted_at IS NULL
    `

    result, err := w.db.NamedExecContext(ctx, query, u)
    if err != nil {
        log.Error("Error updating user", zap.Error(err), zap.Uint("id", u.ID))
        return fmt.Errorf("update user: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No user found to update", zap.Uint("id", u.ID))
        return user.ErrUserNotFound
    }

    return nil
}

func (w *UserWriter) UpdateLastLogin(ctx context.Context, userID uint) error {
    log := logger.GetLogger(zap.String("repository", "UserWriter"), zap.String("method", "UpdateLastLogin"))
    
    query := `
        UPDATE users
        SET last_login_at = $1, updated_at = $1
        WHERE id = $2 AND deleted_at IS NULL
    `

    now := time.Now()
    result, err := w.db.ExecContext(ctx, query, now, userID)
    if err != nil {
        log.Error("Error updating last login", zap.Error(err), zap.Uint("userID", userID))
        return fmt.Errorf("update last login: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No user found to update last login", zap.Uint("userID", userID))
        return user.ErrUserNotFound
    }

    return nil
}

func (w *UserWriter) Delete(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "UserWriter"), zap.String("method", "Delete"))
    
    query := `
        UPDATE users
        SET deleted_at = :deleted_at
        WHERE id = :id AND deleted_at IS NULL
    `

    params := map[string]interface{}{
        "id":         id,
        "deleted_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
    if err != nil {
        log.Error("Error deleting user", zap.Error(err), zap.Uint("id", id))
        return fmt.Errorf("delete user: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No user found to delete", zap.Uint("id", id))
        return user.ErrUserNotFound
    }

    return nil
}

func (w *UserWriter) Restore(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "UserWriter"), zap.String("method", "Restore"))
    
    query := `
        UPDATE users
        SET deleted_at = NULL, updated_at = :updated_at
        WHERE id = :id AND deleted_at IS NOT NULL
    `

    params := map[string]interface{}{
        "id":         id,
        "updated_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
    if err != nil {
        log.Error("Error restoring user", zap.Error(err), zap.Uint("id", id))
        return fmt.Errorf("restore user: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No user found to restore", zap.Uint("id", id))
        return user.ErrUserNotFound
    }

    return nil
}

func (w *UserWriter) DeletePermanently(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "UserWriter"), zap.String("method", "DeletePermanently"))
    
    query := `DELETE FROM users WHERE id = $1`

    result, err := w.db.ExecContext(ctx, query, id)
    if err != nil {
        log.Error("Error permanently deleting user", zap.Error(err), zap.Uint("id", id))
        return fmt.Errorf("delete user permanently: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No user found to delete permanently", zap.Uint("id", id))
        return user.ErrUserNotFound
    }

    return nil
}
