package auth

import "errors"

var (
    ErrUnauthorized     = errors.New("unauthorized")
    ErrInsufficientRole = errors.New("insufficient role")
    ErrUserNotFound     = errors.New("user not found")
    ErrSessionNotFound  = errors.New("session not found")
    ErrSessionExpired   = errors.New("session expired")
)
