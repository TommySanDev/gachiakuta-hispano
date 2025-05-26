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
    CreateSession(ctx context.Context, session *Session) error
}

type SessionUpdater interface {
    UpdateSession(ctx context.Context, session *Session) error
}

type SessionDeleter interface {
    DeleteSession(ctx context.Context, token string) error
    DeleteUserSessions(ctx context.Context, userID uint) error
    DeleteExpiredSessions(ctx context.Context) error
}

// Interfaces for writing magic link data
type MagicLinkCreator interface {
    CreateMagicLink(ctx context.Context, link *MagicLink) error
}

type MagicLinkUpdater interface {
    MarkMagicLinkUsed(ctx context.Context, token string) error
}

type MagicLinkDeleter interface {
    DeleteExpiredMagicLinks(ctx context.Context) error
    DeleteUserMagicLinks(ctx context.Context, userID uint) error
}

// Interfaces for writing reset token data
type ResetTokenCreator interface {
    CreateResetToken(ctx context.Context, token *ResetToken) error
}

type ResetTokenUpdater interface {
    MarkResetTokenUsed(ctx context.Context, token string) error
}

type ResetTokenDeleter interface {
    DeleteExpiredResetTokens(ctx context.Context) error
    DeleteUserResetTokens(ctx context.Context, userID uint) error
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
    MarkRecoveryCodeUsed(ctx context.Context, userID uint, code string) error
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
