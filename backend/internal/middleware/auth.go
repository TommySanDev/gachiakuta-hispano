package middleware

import (
	"net/http"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// AuthMiddleware handles authentication and authorization using Paseto
type AuthMiddleware struct {
	DB           *sqlx.DB
	TokenService *user.TokenService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(db *sqlx.DB, tokenService *user.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		DB:           db,
		TokenService: tokenService,
	}
}

// Authenticate validates Paseto token and loads user context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token == "" {
			http.Error(w, "Token required", http.StatusUnauthorized)
			return
		}

		// Validate Paseto token
		tokenUser, err := m.TokenService.GetUserFromToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Get full user from database to ensure it's still valid
		var dbUser user.User
		err = m.DB.Get(&dbUser, "SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL", tokenUser.ID)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Check if user is still active
		if !dbUser.IsActive {
			http.Error(w, "Account inactive", http.StatusForbidden)
			return
		}

		// Add user to context (using the full DB user, not just token claims)
		ctx := user.SetUserInContext(r.Context(), &dbUser)

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole validates that user has required role
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			currentUser, ok := user.GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check role permission with hierarchy
			if !hasPermission(currentUser.Role, role) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole validates that user has any of the specified roles
func (m *AuthMiddleware) RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			currentUser, ok := user.GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user has any of the required roles
			hasAccess := false
			for _, role := range roles {
				if hasPermission(currentUser.Role, role) {
					hasAccess = true
					break
				}
			}

			if !hasAccess {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasPermission checks role permissions with hierarchy
func hasPermission(userRole, requiredRole string) bool {
	// Admin has access to everything
	if userRole == user.RoleAdmin {
		return true
	}

	// Editor has access to editor and user operations
	if userRole == user.RoleEditor && (requiredRole == user.RoleEditor || requiredRole == user.RoleUser) {
		return true
	}

	// User has access only to user operations
	if userRole == user.RoleUser && requiredRole == user.RoleUser {
		return true
	}

	return false
}
