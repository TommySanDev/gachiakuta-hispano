package user

import "errors"

// General errors
var (
    ErrInvalidInput       = errors.New("invalid input")
    ErrInvalidToken       = errors.New("invalid token")
    ErrTokenExpired       = errors.New("token expired")
    ErrUnauthorized       = errors.New("unauthorized")
    ErrInsufficientRole   = errors.New("insufficient role")
)

// User errors
var (
    ErrUserNotFound       = errors.New("user not found")
    ErrUserAlreadyExists  = errors.New("user already exists")
    ErrUserInactive       = errors.New("user inactive")
    ErrEmailAlreadyExists = errors.New("email already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
)

// Session errors
var (
    ErrSessionNotFound = errors.New("session not found")
    ErrSessionExpired  = errors.New("session expired")
)

// Magic link errors
var (
    ErrMagicLinkNotFound  = errors.New("magic link not found")
    ErrMagicLinkExpired   = errors.New("magic link expired")
    ErrMagicLinkUsed      = errors.New("magic link already used")
    ErrMagicLinkDisabled  = errors.New("magic link disabled")
    ErrInvalidMagicLink   = errors.New("invalid magic link")
)

// Reset token errors
var (
    ErrResetTokenNotFound = errors.New("reset token not found")
    ErrResetTokenExpired  = errors.New("reset token expired")
    ErrResetTokenUsed     = errors.New("reset token already used")
    ErrInvalidResetToken  = errors.New("invalid reset token")
)

// TOTP errors
var (
    ErrTOTPNotFound       = errors.New("TOTP not found")
    ErrTOTPNotEnabled     = errors.New("totp not enabled")
    ErrTOTPAlreadyEnabled = errors.New("totp already enabled")
    ErrInvalidTOTPCode    = errors.New("invalid totp code")
)

// Recovery code errors
var (
    ErrRecoveryCodeNotFound  = errors.New("recovery code not found")
    ErrRecoveryCodeUsed      = errors.New("recovery code already used")
    ErrInvalidRecoveryCode   = errors.New("invalid recovery code")
)
