package user

import "context"

// Interfaces for reading user data
type UserGetter interface {
    GetByID(ctx context.Context, id uint) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    GetByUsername(ctx context.Context, username string) (*User, error)
}

type UserLister interface {
    List(ctx context.Context, filter UserFilter) ([]*User, int, error)
    ListByRole(ctx context.Context, role string, limit int) ([]*User, error)
}

// Interfaces for reading session data
type SessionGetter interface {
    GetByToken(ctx context.Context, token string) (*Session, error)
    GetSessionsByUserID(ctx context.Context, userID uint) ([]*Session, error)
    GetActiveSessionsByUserID(ctx context.Context, userID uint) ([]*Session, error)
}

// Interfaces for reading magic link data - Fixed to include GetByToken method
type MagicLinkGetter interface {
    GetMagicLinkByToken(ctx context.Context, token string) (*MagicLink, error)
    GetActiveMagicLinkByUserID(ctx context.Context, userID uint) (*MagicLink, error)
    GetByToken(ctx context.Context, token string) (*MagicLink, error) // Added this method
}

// Interfaces for reading reset token data - Fixed to include GetByToken method
type ResetTokenGetter interface {
    GetResetTokenByToken(ctx context.Context, token string) (*ResetToken, error)
    GetActiveResetTokenByUserID(ctx context.Context, userID uint) (*ResetToken, error)
    GetByToken(ctx context.Context, token string) (*ResetToken, error) // Added this method
}

// Interfaces for reading TOTP data - Fixed to include GetByUserID method
type TOTPGetter interface {
    GetTOTPByUserID(ctx context.Context, userID uint) (*TOTPSecret, error)
    GetByUserID(ctx context.Context, userID uint) (*TOTPSecret, error) // Added this method
}

type RecoveryCodeGetter interface {
    GetUnusedRecoveryCodesByUserID(ctx context.Context, userID uint) ([]*RecoveryCode, error)
    GetRecoveryCodeByCode(ctx context.Context, userID uint, code string) (*RecoveryCode, error)
}

// Composite interfaces for easier use
type UserReader interface {
    UserGetter
    UserLister
}

type SessionReader interface {
    SessionGetter
}

type MagicLinkReader interface {
    MagicLinkGetter
}

type ResetTokenReader interface {
    ResetTokenGetter
}

type TOTPReader interface {
    TOTPGetter
    RecoveryCodeGetter
}

// Reader composes all read interfaces
type Reader interface {
    UserGetter
    UserLister
    SessionGetter
    MagicLinkGetter
    ResetTokenGetter
    TOTPGetter
    RecoveryCodeGetter
}
