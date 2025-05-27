package user

import "context"

// Interfaces for writing user data
type UserCreator interface {
    Create(ctx context.Context, user *User) error
}

type UserUpdater interface {
    Update(ctx context.Context, user *User) error
}

type UserDeleter interface {
    Delete(ctx context.Context, id uint) error
}

type UserRestorer interface {
    Restore(ctx context.Context, id uint) error
}

type UserPermanentDeleter interface {
    DeletePermanently(ctx context.Context, id uint) error
}

type LastLoginUpdater interface {
    UpdateLastLogin(ctx context.Context, userID uint) error
}

// Session interfaces
type SessionCreator interface {
    Create(ctx context.Context, session *Session) error
}

type SessionUpdater interface {
    Update(ctx context.Context, session *Session) error
}

type SessionDeleter interface {
    Delete(ctx context.Context, id string) error
    DeleteByUserID(ctx context.Context, userID uint) error
    DeleteExpiredSessions(ctx context.Context) (int64, error)
}

// Magic link interfaces
type MagicLinkCreator interface {
    Create(ctx context.Context, ml *MagicLink) error
}

type MagicLinkUpdater interface {
    MarkAsUsed(ctx context.Context, id uint) error
}

type MagicLinkDeleter interface {
    DeleteExpired(ctx context.Context) (int, error)
}

// Reset token interfaces
type ResetTokenCreator interface {
    Create(ctx context.Context, rt *ResetToken) error
}

type ResetTokenUpdater interface {
    MarkAsUsed(ctx context.Context, id uint) error
    InvalidateByUserID(ctx context.Context, userID uint) error
}

type ResetTokenDeleter interface {
    DeleteExpired(ctx context.Context) (int, error)
}

// TOTP interfaces
type TOTPCreator interface {
    CreateTOTPSecret(ctx context.Context, totp *TOTPSecret) error
    CreateRecoveryCodes(ctx context.Context, codes []*RecoveryCode) error
}

type TOTPUpdater interface {
    VerifyTOTPSecret(ctx context.Context, userID uint) error
    UseRecoveryCode(ctx context.Context, userID uint, code string) error
}

type TOTPDeleter interface {
    DeleteTOTPSecret(ctx context.Context, userID uint) error
    DeleteRecoveryCodes(ctx context.Context, userID uint) error
}

// Composite writer interfaces
type UserWriter interface {
    UserCreator
    UserUpdater
    UserDeleter
    UserRestorer
    UserPermanentDeleter
    LastLoginUpdater
}

type SessionWriter interface {
    SessionCreator
    SessionUpdater
    SessionDeleter
}

type MagicLinkWriter interface {
    MagicLinkCreator
    MagicLinkUpdater
    MagicLinkDeleter
}

type ResetTokenWriter interface {
    ResetTokenCreator
    ResetTokenUpdater
    ResetTokenDeleter
}

type TOTPWriter interface {
    TOTPCreator
    TOTPUpdater
    TOTPDeleter
}

// Writer composes all write interfaces
type Writer interface {
    UserWriter
    SessionWriter
    MagicLinkWriter
    ResetTokenWriter
    TOTPWriter
}

// Store composes all read and write operations
type UserStore interface {
    UserGetter
    UserLister
    UserWriter
}

type SessionStore interface {
    SessionGetter
    SessionWriter
}

type MagicLinkStore interface {
    MagicLinkGetter
    MagicLinkWriter
}

type ResetTokenStore interface {
    ResetTokenGetter
    ResetTokenWriter
}

type TOTPStore interface {
    TOTPGetter
    RecoveryCodeGetter
    TOTPWriter
}

// Store composes all read and write operations
type Store interface {
    Reader
    Writer
}
