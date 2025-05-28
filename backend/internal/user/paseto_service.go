package user

import (
	"fmt"
	"time"

	"github.com/vk-rv/pvx"
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

// NewTokenService creates a new token service with secret key
func NewTokenService(secretKey string) *TokenService {
	return &TokenService{
		secretKey: []byte(secretKey),
	}
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

	// Create Paseto v4.local token (symmetric encryption)
	token, err := pvx.NewPV4Local().Encrypt(ts.secretKey, claims)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// ValidateToken validates and parses a Paseto token
func (ts *TokenService) ValidateToken(tokenString string) (*TokenClaims, error) {
	var claims TokenClaims
	
	// Decrypt and parse Paseto token
	err := pvx.NewPV4Local().Decrypt(tokenString, ts.secretKey).ScanClaims(&claims)
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
