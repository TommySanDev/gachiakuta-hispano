package user

import (
    "encoding/json"
    "net/http"
    "strconv"
    "time"

    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Http helpers

// Writes JSON response to the client
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        logger.Error("Error encoding JSON response", zap.Error(err))
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
    }
}

// Extracts and parses integer parameters from the request
func getIntParam(r *http.Request, key string, defaultValue int) int {
    str := r.URL.Query().Get(key)
    if str == "" {
        return defaultValue
    }

    val, err := strconv.Atoi(str)
    if err != nil {
        return defaultValue
    }

    return val
}

// Extracts and parses boolean parameters from the request
func getBoolParam(r *http.Request, key string, defaultValue bool) bool {
    str := r.URL.Query().Get(key)
    if str == "" {
        return defaultValue
    }

    val, err := strconv.ParseBool(str)
    if err != nil {
        return defaultValue
    }

    return val
}

// Session helper methods

// IsExpired checks if session has expired
func (s *Session) IsExpired() bool {
    return time.Now().After(s.ExpiresAt)
}

// TimeUntilExpiry returns duration until session expires
func (s *Session) TimeUntilExpiry() time.Duration {
    if s.IsExpired() {
        return 0
    }
    return time.Until(s.ExpiresAt)
}

// MagicLink helper methods

// IsExpired checks if magic link has expired
func (ml *MagicLink) IsExpired() bool {
    return time.Now().After(ml.ExpiresAt)
}

// IsUsed checks if magic link has been used
func (ml *MagicLink) IsUsed() bool {
    return ml.Used
}

// IsValid checks if magic link is valid (not used and not expired)
func (ml *MagicLink) IsValid() bool {
    return !ml.IsUsed() && !ml.IsExpired()
}

// ResetToken helper methods

// IsExpired checks if reset token has expired
func (rt *ResetToken) IsExpired() bool {
    return time.Now().After(rt.ExpiresAt)
}

// IsUsed checks if reset token has been used
func (rt *ResetToken) IsUsed() bool {
    return rt.Used
}

// IsValid checks if reset token is valid (not used and not expired)
func (rt *ResetToken) IsValid() bool {
    return !rt.IsUsed() && !rt.IsExpired()
}

// User helper methods

// FullName returns the user's full name
func (u *User) FullName() string {
    if u.FirstName == "" && u.LastName == "" {
        return u.Username
    }
    return u.FirstName + " " + u.LastName
}

// HasRole checks if user has specific role
func (u *User) HasRole(role string) bool {
    return u.Role == role
}

// IsAdmin checks if user is an admin
func (u *User) IsAdmin() bool {
    return u.HasRole(RoleAdmin)
}

// IsEditor checks if user is an editor
func (u *User) IsEditor() bool {
    return u.HasRole(RoleEditor)
}

// CanAccess checks if user can access given role level
func (u *User) CanAccess(requiredRole string) bool {
    switch requiredRole {
    case RoleUser:
        return true // All authenticated users can access user level
    case RoleEditor:
        return u.IsEditor() || u.IsAdmin()
    case RoleAdmin:
        return u.IsAdmin()
    default:
        return false
    }
}
