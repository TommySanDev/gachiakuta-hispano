package context

import "context"

type ctxKey string

const (
    sessionKey ctxKey = "session"
    userKey    ctxKey = "user"
)

func GetUserFromContext(ctx context.Context) (interface{}, bool) {
    u, ok := ctx.Value(userKey).(interface{})
    return u, ok
}

func GetSessionFromContext(ctx context.Context) (interface{}, bool) {
    s, ok := ctx.Value(sessionKey).(interface{})
    return s, ok
}

func SetUserInContext(ctx context.Context, u interface{}) context.Context {
    return context.WithValue(ctx, userKey, u)
}

func SetSessionInContext(ctx context.Context, s interface{}) context.Context {
    return context.WithValue(ctx, sessionKey, s)
}

