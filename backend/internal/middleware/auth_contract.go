package middleware

import (
    "context"

    "github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

type SessionReader interface {
    GetSession(ctx context.Context, token string) (*user.Session, error)
}

