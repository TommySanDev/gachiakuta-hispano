package user

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/o1egl/paseto/v2"
)

// TokenService handles Paseto token operations
type TokenService struct {
	secretKey []byte
}

// TokenClaims represents the claims inside our Paseto token
type TokenClaims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// NewTokenService creates a new token service with automatic key format detection
func NewTokenService(secretKey string) *TokenService {
	key, err := parseSecretKey(secretKey)
	if err != nil {
		// Fallback to treating as raw string (for backwards compatibility)
		key = []byte(secretKey)
	}
	
	return &TokenService{
		secretKey: key,
	}
}

// parseSecretKey detects and parses the secret key format
func parseSecretKey(secretKey string) ([]byte, error) {
	// Try to decode as base64 first
	if decoded, err := base64.StdEncoding.DecodeString(secretKey); err == nil {
		// Successful base64 decode
		if len(decoded) >= 32 {
			return decoded, nil
		}
		// Base64 decoded but too short, treat as raw
		return []byte(secretKey), fmt.Errorf("base64 key too short")
	}
	
	// Not valid base64, treat as raw string
	if len(secretKey) >= 32 {
		return []byte(secretKey), nil
	}
	
	return nil, fmt.Errorf("secret key too short (minimum 32 characters or base64 equivalent)")
}

// GenerateToken creates a new Paseto token for a user
func (ts *TokenService) GenerateToken(user *User, duration time.Duration) (string, error) {
	now := time.Now()
	
	claims := TokenClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Role:      user.Role,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(duration).Unix(),
	}

	// Create Paseto v2.local token (symmetric encryption)
	token, err := paseto.NewV2().Encrypt(ts.secretKey, claims, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// ValidateToken validates and parses a Paseto token
func (ts *TokenService) ValidateToken(tokenString string) (*TokenClaims, error) {
	var claims TokenClaims
	var footer string
	
	// Decrypt and parse Paseto token
	err := paseto.NewV2().Decrypt(tokenString, ts.secretKey, &claims, &footer)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if token is expired
	now := time.Now().Unix()
	if claims.ExpiresAt < now {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

// RefreshToken generates a new token if the current one is valid and not too old
func (ts *TokenService) RefreshToken(tokenString string, maxAge time.Duration) (string, error) {
	claims, err := ts.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Check if token is recent enough to refresh
	tokenAge := time.Since(time.Unix(claims.IssuedAt, 0))
	if tokenAge > maxAge {
		return "", fmt.Errorf("token too old to refresh")
	}

	// Create new token with same claims but new expiration
	user := &User{
		ID:       claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
		Role:     claims.Role,
	}

	return ts.GenerateToken(user, time.Hour) // New 1-hour token
}

// GetUserFromToken extracts user information from a valid token
func (ts *TokenService) GetUserFromToken(tokenString string) (*User, error) {
	claims, err := ts.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}
