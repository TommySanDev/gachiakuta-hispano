package user

import "context"

// Interfaces for writing user data
type UserCreator interface {
    Create(ctx context.Context, user *User) error
}

type UserUpdater interface {
    Update(ctx context.Context, user *User) error
    UpdateLastLogin(ctx context.Context, userID uint) error
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

// Interfaces for writing session data
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

// Interfaces for writing magic link data
type MagicLinkCreator interface {
    Create(ctx context.Context, link *MagicLink) error
}

type MagicLinkUpdater interface {
    MarkAsUsed(ctx context.Context, id uint) error
}

type MagicLinkDeleter interface {
    DeleteExpired(ctx context.Context) (int, error)
}

// Interfaces for writing reset token data
type ResetTokenCreator interface {
    Create(ctx context.Context, token *ResetToken) error
}

type ResetTokenUpdater interface {
    MarkAsUsed(ctx context.Context, id uint) error
    InvalidateByUserID(ctx context.Context, userID uint) error
}

type ResetTokenDeleter interface {
    DeleteExpired(ctx context.Context) (int, error)
}

// Interfaces for writing TOTP data
type TOTPCreator interface {
    CreateTOTPSecret(ctx context.Context, secret *TOTPSecret) error
}

type TOTPUpdater interface {
    VerifyTOTPSecret(ctx context.Context, userID uint) error
}

type TOTPDeleter interface {
    DeleteTOTPSecret(ctx context.Context, userID uint) error
}

// Interfaces for writing recovery codes
type RecoveryCodeCreator interface {
    CreateRecoveryCodes(ctx context.Context, codes []*RecoveryCode) error
}

type RecoveryCodeUpdater interface {
    UseRecoveryCode(ctx context.Context, userID uint, code string) error
}

type RecoveryCodeDeleter interface {
    DeleteRecoveryCodes(ctx context.Context, userID uint) error
}

// Writer composes all write interfaces
type Writer interface {
    UserCreator
    UserUpdater
    UserDeleter
    UserRestorer
    UserPermanentDeleter
    SessionCreator
    SessionUpdater
    SessionDeleter
    MagicLinkCreator
    MagicLinkUpdater
    MagicLinkDeleter
    ResetTokenCreator
    ResetTokenUpdater
    ResetTokenDeleter
    TOTPCreator
    TOTPUpdater
    TOTPDeleter
    RecoveryCodeCreator
    RecoveryCodeUpdater
    RecoveryCodeDeleter
}

// Store composes all read and write operations
type Store interface {
    Reader
    Writer
}
