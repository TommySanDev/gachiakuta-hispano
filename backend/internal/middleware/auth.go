package middleware

import (
    "net/http"
    "strings"

    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/auth"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Reader interfaces for authentication middleware
type SessionReader interface {
    GetByToken(ctx context.Context, token string) (*auth.Session, error)
}

type UserReader interface {
    GetByID(ctx context.Context, userID uint) (*auth.User, error)
}

// Middleware for authentication and authorization
type AuthMiddleware struct {
    sessionReader SessionReader
    userReader    UserReader
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(sessionReader SessionReader, userReader UserReader) *AuthMiddleware {
    return &AuthMiddleware{
        sessionReader: sessionReader,
        userReader:    userReader,
    }
}

// Authenticate validates JWT token and loads user context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log := logger.GetLogger(zap.String("middleware", "Auth"), zap.String("method", "Authenticate"))
        
        // Extract token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            log.Debug("No authorization header provided")
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }

        // Check Bearer token format
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            log.Debug("Invalid authorization header format")
            http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
            return
        }

        token := parts[1]
        if token == "" {
            log.Debug("Empty token provided")
            http.Error(w, "Token required", http.StatusUnauthorized)
            return
        }

        // Get session by token
        session, err := m.sessionReader.GetByToken(r.Context(), token)
        if err != nil {
            if err == auth.ErrSessionNotFound {
                log.Debug("Invalid token provided")
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }
            log.Error("Error validating session", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        // Check if session is expired
        if session.IsExpired() {
            log.Debug("Session expired", zap.String("session_id", session.ID))
            http.Error(w, "Session expired", http.StatusUnauthorized)
            return
        }

        // Get user
        user, err := m.userReader.GetByID(r.Context(), session.UserID)
        if err != nil {
            if err == auth.ErrUserNotFound {
                log.Debug("User not found for session", zap.Uint("user_id", session.UserID))
                http.Error(w, "User not found", http.StatusUnauthorized)
                return
            }
            log.Error("Error getting user", zap.Error(err))
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            return
        }

        // Check if user is active
        if !user.IsActive {
            log.Debug("Inactive user attempted access", zap.Uint("user_id", user.ID))
            http.Error(w, "Account inactive", http.StatusForbidden)
            return
        }

        // Add user and session to context
        ctx := auth.SetUserInContext(r.Context(), user)
        ctx = auth.SetSessionInContext(ctx, session)

        // Continue to next handler
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// RequireRole validates that user has required role
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            log := logger.GetLogger(zap.String("middleware", "Auth"), zap.String("method", "RequireRole"))
            
            // Get user from context
            user, ok := auth.GetUserFromContext(r.Context())
            if !ok {
                log.Error("User not found in context")
                http.Error(w, "Authentication required", http.StatusUnauthorized)
                return
            }

            // Check role permission
            if !m.hasPermission(user.Role, role) {
                log.Debug("Insufficient permissions", 
                    zap.String("user_role", user.Role), 
                    zap.String("required_role", role),
                    zap.Uint("user_id", user.ID),
                )
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
            log := logger.GetLogger(zap.String("middleware", "Auth"), zap.String("method", "RequireAnyRole"))
            
            // Get user from context
            user, ok := auth.GetUserFromContext(r.Context())
            if !ok {
                log.Error("User not found in context")
                http.Error(w, "Authentication required", http.StatusUnauthorized)
                return
            }

            // Check if user has any of the required roles
            hasPermission := false
            for _, role := range roles {
                if m.hasPermission(user.Role, role) {
                    hasPermission = true
                    break
                }
            }

            if !hasPermission {
                log.Debug("Insufficient permissions for any required role", 
                    zap.String("user_role", user.Role), 
                    zap.Strings("required_roles", roles),
                    zap.Uint("user_id", user.ID),
                )
                http.Error(w, "Insufficient permissions", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Require2FA validates 2FA when enabled (prepared for Phase 3)
func (m *AuthMiddleware) Require2FA(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log := logger.GetLogger(zap.String("middleware", "Auth"), zap.String("method", "Require2FA"))
        
        // Get user from context
        user, ok := auth.GetUserFromContext(r.Context())
        if !ok {
            log.Error("User not found in context")
            http.Error(w, "Authentication required", http.StatusUnauthorized)
            return
        }

        // Check if 2FA is enabled and verified (Phase 3 implementation)
        if user.TOTPEnabled {
            // TODO: Implement 2FA verification in Phase 3
            log.Debug("2FA verification required but not yet implemented", zap.Uint("user_id", user.ID))
            // For now, allow access - will be implemented in Phase 3
        }

        next.ServeHTTP(w, r)
    })
}

// Helper method to check role permissions with hierarchy
func (m *AuthMiddleware) hasPermission(userRole, requiredRole string) bool {
    // Admin has access to everything
    if userRole == auth.RoleAdmin {
        return true
    }

    // Editor has access to editor and user operations
    if userRole == auth.RoleEditor && (requiredRole == auth.RoleEditor || requiredRole == auth.RoleUser) {
        return true
    }

    // User has access only to user operations
    if userRole == auth.RoleUser && requiredRole == auth.RoleUser {
        return true
    }

    return false
}
