package user

import (
    "context"

    "github.com/TommySanDev/gachiakuta-hispano/internal/auth"
)

// UserAdapter adapts User to auth.User
type UserAdapter struct {
    reader UserGetter
}

// NewUserAdapter creates a new user adapter
func NewUserAdapter(reader UserGetter) *UserAdapter {
    return &UserAdapter{reader: reader}
}

// GetByID adapts User to auth.User
func (a *UserAdapter) GetByID(ctx context.Context, userID uint) (*auth.User, error) {
    user, err := a.reader.GetByID(ctx, userID)
    if err != nil {
        return nil, mapError(err)
    }

    return &auth.User{
        ID:                user.ID,
        Email:             user.Email,
        Username:          user.Username,
        FirstName:         user.FirstName,
        LastName:          user.LastName,
        Role:              user.Role,
        IsActive:          user.IsActive,
        EmailVerified:     user.EmailVerified,
        MagicLinkEnabled:  user.MagicLinkEnabled,
        TOTPEnabled:       user.TOTPEnabled,
        LastLoginAt:       user.LastLoginAt,
        CreatedAt:         user.CreatedAt,
        UpdatedAt:         user.UpdatedAt,
    }, nil
}

// SessionAdapter adapts Session to auth.Session
type SessionAdapter struct {
    reader SessionGetter
}

// NewSessionAdapter creates a new session adapter
func NewSessionAdapter(reader SessionGetter) *SessionAdapter {
    return &SessionAdapter{reader: reader}
}

// GetByToken adapts Session to auth.Session
func (a *SessionAdapter) GetByToken(ctx context.Context, token string) (*auth.Session, error) {
    session, err := a.reader.GetByToken(ctx, token)
    if err != nil {
        return nil, mapError(err)
    }

    return &auth.Session{
        ID:        session.ID,
        UserID:    session.UserID,
        UserAgent: session.UserAgent,
        IPAddress: session.IPAddress,
        ExpiresAt: session.ExpiresAt,
    }, nil
}

// Helper to map errors between packages
func mapError(err error) error {
    switch err {
    case ErrUserNotFound:
        return auth.ErrUserNotFound
    case ErrSessionNotFound:
        return auth.ErrSessionNotFound
    case ErrSessionExpired:
        return auth.ErrSessionExpired
    default:
        return err
    }
}


