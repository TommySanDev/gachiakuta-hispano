package auth

import "context"

// GetUserFromContext retrieves the authenticated user from context
func GetUserFromContext(ctx context.Context) (*User, bool) {
    userObj, ok := ctx.Value(UserContextKey).(*User)
    return userObj, ok
}

// GetSessionFromContext retrieves the active session from context
func GetSessionFromContext(ctx context.Context) (*Session, bool) {
    session, ok := ctx.Value(SessionContextKey).(*Session)
    return session, ok
}

// SetUserInContext adds user to context
func SetUserInContext(ctx context.Context, u *User) context.Context {
    return context.WithValue(ctx, UserContextKey, u)
}

// SetSessionInContext adds session to context
func SetSessionInContext(ctx context.Context, s *Session) context.Context {
    return context.WithValue(ctx, SessionContextKey, s)
}
