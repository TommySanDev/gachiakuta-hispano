package user

import "context"

// Context keys for authentication data
type ContextKey string

const (
	UserContextKey ContextKey = "user"
)

// GetUserFromContext retrieves the authenticated user from context
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(UserContextKey).(*User)
	return user, ok
}

// SetUserInContext adds user to context
func SetUserInContext(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, UserContextKey, u)
}
