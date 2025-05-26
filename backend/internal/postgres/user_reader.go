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

// Implements user.Reader using PostgreSQL
type UserReader struct {
    db *sqlx.DB
}

func NewUserReader(db *sqlx.DB) *UserReader {
    return &UserReader{
        db: db,
    }
}

func (r *UserReader) GetByID(ctx context.Context, id uint) (*user.User, error) {
    log := logger.GetLogger(zap.String("repository", "UserReader"), zap.String("method", "GetByID"))
    
    query := `
        SELECT id, email, username, password_hash, first_name, last_name, role,
               is_active, email_verified, magic_link_enabled, totp_enabled,
               last_login_at, created_at, updated_at, deleted_at
        FROM users
        WHERE id = $1 AND deleted_at IS NULL
    `

    var u user.User
    err := r.db.GetContext(ctx, &u, query, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("User not found", zap.Uint("id", id))
            return nil, user.ErrUserNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &u, nil
}

func (r *UserReader) GetByEmail(ctx context.Context, email string) (*user.User, error) {
    log := logger.GetLogger(zap.String("repository", "UserReader"), zap.String("method", "GetByEmail"))
    
    query := `
        SELECT id, email, username, password_hash, first_name, last_name, role,
               is_active, email_verified, magic_link_enabled, totp_enabled,
               last_login_at, created_at, updated_at, deleted_at
        FROM users
        WHERE email = $1 AND deleted_at IS NULL
    `

    var u user.User
    err := r.db.GetContext(ctx, &u, query, email)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("User not found", zap.String("email", email))
            return nil, user.ErrUserNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &u, nil
}

func (r *UserReader) GetByUsername(ctx context.Context, username string) (*user.User, error) {
    log := logger.GetLogger(zap.String("repository", "UserReader"), zap.String("method", "GetByUsername"))
    
    query := `
        SELECT id, email, username, password_hash, first_name, last_name, role,
               is_active, email_verified, magic_link_enabled, totp_enabled,
               last_login_at, created_at, updated_at, deleted_at
        FROM users
        WHERE username = $1 AND deleted_at IS NULL
    `

    var u user.User
    err := r.db.GetContext(ctx, &u, query, username)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("User not found", zap.String("username", username))
            return nil, user.ErrUserNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &u, nil
}

func (r *UserReader) List(ctx context.Context, filter user.UserFilter) ([]*user.User, int, error) {
    log := logger.GetLogger(zap.String("repository", "UserReader"), zap.String("method", "List"))
    
    // Build query parts
    whereClauses := []string{"1=1"}
    args := []interface{}{}
    argPos := 1

    // Base where clause
    if !filter.IncludeDeleted {
        whereClauses = append(whereClauses, "deleted_at IS NULL")
    }
    
    // Apply search filter
    if filter.Search != "" {
        whereClauses = append(whereClauses, fmt.Sprintf("(email ILIKE $%d OR username ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", argPos, argPos+1, argPos+2, argPos+3))
        searchTerm := "%" + filter.Search + "%"
        args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
        argPos += 4
    }

    // Role filter
    if filter.Role != "" {
        whereClauses = append(whereClauses, fmt.Sprintf("role = $%d", argPos))
        args = append(args, filter.Role)
        argPos++
    }

    // IsActive filter
    if filter.IsActive != nil {
        whereClauses = append(whereClauses, fmt.Sprintf("is_active = $%d", argPos))
        args = append(args, *filter.IsActive)
        argPos++
    }

    // EmailVerified filter
    if filter.EmailVerified != nil {
        whereClauses = append(whereClauses, fmt.Sprintf("email_verified = $%d", argPos))
        args = append(args, *filter.EmailVerified)
        argPos++
    }

    whereClause := "WHERE " + strings.Join(whereClauses, " AND ")

    // Count total matching records
    countQuery := "SELECT COUNT(*) FROM users " + whereClause
    var total int
    err := r.db.GetContext(ctx, &total, countQuery, args...)
    if err != nil {
        log.Error("Error counting users", zap.Error(err))
        return nil, 0, fmt.Errorf("count users: %w", err)
    }

    // Main query with sorting and pagination
    query := `
        SELECT id, email, username, password_hash, first_name, last_name, role,
               is_active, email_verified, magic_link_enabled, totp_enabled,
               last_login_at, created_at, updated_at, deleted_at
        FROM users
        ` + whereClause

    // Add sorting
    validSortFields := map[string]bool{
        "email":      true,
        "username":   true,
        "first_name": true,
        "last_name":  true,
        "role":       true,
        "created_at": true,
    }
    
    if validSortFields[filter.SortBy] {
        query += fmt.Sprintf(" ORDER BY %s", filter.SortBy)
        
        if strings.ToLower(filter.SortDir) == "desc" {
            query += " DESC"
        } else {
            query += " ASC"
        }
    } else {
        query += " ORDER BY created_at DESC"
    }

    // Add pagination
    offset := (filter.Page - 1) * filter.PageSize
    query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
    args = append(args, filter.PageSize, offset)

    // Execute query
    var users []*user.User
    err = r.db.SelectContext(ctx, &users, query, args...)
    if err != nil {
        log.Error("Error querying users", zap.Error(err))
        return nil, 0, fmt.Errorf("query users: %w", err)
    }

    return users, total, nil
}

func (r *UserReader) ListByRole(ctx context.Context, role string, limit int) ([]*user.User, error) {
    log := logger.GetLogger(
        zap.String("repository", "UserReader"), 
        zap.String("method", "ListByRole"),
    )
    
    query := `
        SELECT id, email, username, password_hash, first_name, last_name, role,
               is_active, email_verified, magic_link_enabled, totp_enabled,
               last_login_at, created_at, updated_at, deleted_at
        FROM users
        WHERE role = $1 AND deleted_at IS NULL AND is_active = true
        ORDER BY created_at DESC
        LIMIT $2
    `

    var users []*user.User
    err := r.db.SelectContext(ctx, &users, query, role, limit)
    if err != nil {
        log.Error("Error querying users by role", zap.Error(err))
        return nil, fmt.Errorf("query users by role: %w", err)
    }

    return users, nil
}

func (r *UserReader) GetByToken(ctx context.Context, token string) (*user.Session, error) {
    log := logger.GetLogger(zap.String("repository", "UserReader"), zap.String("method", "GetByToken"))
    
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

func (r *UserReader) GetSessionsByUserID(ctx context.Context, userID uint) ([]*user.Session, error) {
    log := logger.GetLogger(
        zap.String("repository", "UserReader"), 
        zap.String("method", "GetSessionsByUserID"),
    )
    
    query := `
        SELECT id, user_id, token, user_agent, ip_address, expires_at, created_at, updated_at
        FROM sessions
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

    var sessions []*user.Session
    err := r.db.SelectContext(ctx, &sessions, query, userID)
    if err != nil {
        log.Error("Error querying sessions by user", zap.Error(err))
        return nil, fmt.Errorf("query sessions by user: %w", err)
    }

    return sessions, nil
}

func (r *UserReader) GetActiveSessionsByUserID(ctx context.Context, userID uint) ([]*user.Session, error) {
    log := logger.GetLogger(
        zap.String("repository", "UserReader"), 
        zap.String("method", "GetActiveSessionsByUserID"),
    )
    
    query := `
        SELECT id, user_id, token, user_agent, ip_address, expires_at, created_at, updated_at
        FROM sessions
        WHERE user_id = $1 AND expires_at > NOW()
        ORDER BY created_at DESC
    `

    var sessions []*user.Session
    err := r.db.SelectContext(ctx, &sessions, query, userID)
    if err != nil {
        log.Error("Error querying active sessions by user", zap.Error(err))
        return nil, fmt.Errorf("query active sessions by user: %w", err)
    }

    return sessions, nil
}

func (r *UserReader) GetMagicLinkByToken(ctx context.Context, token string) (*user.MagicLink, error) {
    log := logger.GetLogger(zap.String("repository", "UserReader"), zap.String("method", "GetMagicLinkByToken"))
    
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

func (r *UserReader) GetActiveMagicLinkByUserID(ctx context.Context, userID uint) (*user.MagicLink, error) {
    log := logger.GetLogger(
        zap.String("repository", "UserReader"), 
        zap.String("method", "GetActiveMagicLinkByUserID"),
    )
    
    query := `
        SELECT id, user_id, token, used, expires_at, created_at, updated_at
        FROM magic_links
        WHERE user_id = $1 AND used = false AND expires_at > NOW()
        ORDER BY created_at DESC
        LIMIT 1
    `

    var ml user.MagicLink
    err := r.db.GetContext(ctx, &ml, query, userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("Active magic link not found", zap.Uint("user_id", userID))
            return nil, user.ErrMagicLinkNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &ml, nil
}

func (r *UserReader) GetResetTokenByToken(ctx context.Context, token string) (*user.ResetToken, error) {
    log := logger.GetLogger(zap.String("repository", "UserReader"), zap.String("method", "GetResetTokenByToken"))
    
    query := `
        SELECT id, user_id, token, used, expires_at, created_at, updated_at
        FROM reset_tokens
        WHERE token = $1 AND used = false
    `

    var rt user.ResetToken
    err := r.db.GetContext(ctx, &rt, query, token)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("Reset token not found", zap.String("token_prefix", token[:8]+"..."))
            return nil, user.ErrResetTokenNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &rt, nil
}

func (r *UserReader) GetActiveResetTokenByUserID(ctx context.Context, userID uint) (*user.ResetToken, error) {
    log := logger.GetLogger(
        zap.String("repository", "UserReader"), 
        zap.String("method", "GetActiveResetTokenByUserID"),
    )
    
    query := `
        SELECT id, user_id, token, used, expires_at, created_at, updated_at
        FROM reset_tokens
        WHERE user_id = $1 AND used = false AND expires_at > NOW()
        ORDER BY created_at DESC
        LIMIT 1
    `

    var rt user.ResetToken
    err := r.db.GetContext(ctx, &rt, query, userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("Active reset token not found", zap.Uint("user_id", userID))
            return nil, user.ErrResetTokenNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &rt, nil
}

func (r *UserReader) GetTOTPByUserID(ctx context.Context, userID uint) (*user.TOTPSecret, error) {
    log := logger.GetLogger(zap.String("repository", "UserReader"), zap.String("method", "GetTOTPByUserID"))
    
    query := `
        SELECT id, user_id, secret, verified, created_at, updated_at
        FROM totp_secrets
        WHERE user_id = $1
    `

    var totp user.TOTPSecret
    err := r.db.GetContext(ctx, &totp, query, userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("TOTP secret not found", zap.Uint("user_id", userID))
            return nil, user.ErrTOTPNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &totp, nil
}

func (r *UserReader) GetUnusedRecoveryCodesByUserID(ctx context.Context, userID uint) ([]*user.RecoveryCode, error) {
    log := logger.GetLogger(
        zap.String("repository", "UserReader"), 
        zap.String("method", "GetUnusedRecoveryCodesByUserID"),
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

func (r *UserReader) GetRecoveryCodeByCode(ctx context.Context, userID uint, code string) (*user.RecoveryCode, error) {
    log := logger.GetLogger(
        zap.String("repository", "UserReader"), 
        zap.String("method", "GetRecoveryCodeByCode"),
    )
    
    query := `
        SELECT id, user_id, code, used, created_at, updated_at
        FROM recovery_codes
        WHERE user_id = $1 AND code = $2
    `

    var rc user.RecoveryCode
    err := r.db.GetContext(ctx, &rc, query, userID, code)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("Recovery code not found", zap.Uint("user_id", userID))
            return nil, user.ErrRecoveryCodeNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &rc, nil
}
