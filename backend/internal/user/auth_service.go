package user

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"

    "golang.org/x/crypto/bcrypt"
    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Handles core authentication operations
type AuthService struct {
    userReader    UserReader
    userWriter    UserWriter
    sessionWriter SessionWriter
    mlReader      MagicLinkReader
    mlWriter      MagicLinkWriter
    emailService  *EmailService
}

func NewAuthService(
    userReader UserReader,
    userWriter UserWriter,
    sessionWriter SessionWriter,
    mlReader MagicLinkReader,
    mlWriter MagicLinkWriter,
    emailService *EmailService,
) *AuthService {
    return &AuthService{
        userReader:    userReader,
        userWriter:    userWriter,
        sessionWriter: sessionWriter,
        mlReader:      mlReader,
        mlWriter:      mlWriter,
        emailService:  emailService,
    }
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*User, error) {
    log := logger.GetLogger(zap.String("service", "AuthService"), zap.String("method", "Register"))
    
    // Check if user already exists
    existingUser, err := s.userReader.GetByEmail(ctx, input.Email)
    if err != nil && err != ErrUserNotFound {
        log.Error("Error checking existing user", zap.Error(err))
        return nil, fmt.Errorf("check existing user: %w", err)
    }
    
    if existingUser != nil {
        return nil, ErrUserAlreadyExists
    }
    
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        log.Error("Error hashing password", zap.Error(err))
        return nil, fmt.Errorf("hash password: %w", err)
    }
    
    // Create user
    now := time.Now()
    user := &User{
        Email:            input.Email,
        Username:         input.Username,
        PasswordHash:     string(hashedPassword),
        FirstName:        input.FirstName,
        LastName:         input.LastName,
        Role:             RoleUser,
        IsActive:         true,
        EmailVerified:    false,
        MagicLinkEnabled: input.MagicLinkEnabled,
        TOTPEnabled:      false,
        CreatedAt:        now,
        UpdatedAt:        now,
    }
    
    err = s.userWriter.Create(ctx, user)
    if err != nil {
        log.Error("Error creating user", zap.Error(err))
        return nil, fmt.Errorf("create user: %w", err)
    }
    
    log.Info("User registered successfully", zap.String("email", input.Email))
    return user, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
    log := logger.GetLogger(zap.String("service", "AuthService"), zap.String("method", "Login"))
    
    // Get user by email
    user, err := s.userReader.GetByEmail(ctx, input.Email)
    if err != nil {
        if err == ErrUserNotFound {
            return nil, ErrInvalidCredentials
        }
        log.Error("Error getting user", zap.Error(err))
        return nil, fmt.Errorf("get user: %w", err)
    }
    
    // Check if user is active
    if !user.IsActive {
        return nil, ErrUserInactive
    }
    
    // Verify password
    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
    if err != nil {
        return nil, ErrInvalidCredentials
    }
    
    // Update last login
    now := time.Now()
    user.LastLoginAt = &now
    user.UpdatedAt = now
    
    err = s.userWriter.Update(ctx, user)
    if err != nil {
        log.Warn("Failed to update last login time", zap.Error(err))
        // Don't fail login for this
    }
    
    // Create session
    session, tokens, err := s.createSession(ctx, user.ID)
    if err != nil {
        log.Error("Error creating session", zap.Error(err))
        return nil, fmt.Errorf("create session: %w", err)
    }
    
    log.Info("User logged in successfully", zap.String("email", input.Email))
    
    return &AuthResponse{
        User:         user,
        Token:        tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
        ExpiresIn:    3600, // 1 hour
    }, nil
}

func (s *AuthService) GenerateMagicLink(ctx context.Context, input MagicLinkLoginInput) error {
    log := logger.GetLogger(zap.String("service", "AuthService"), zap.String("method", "GenerateMagicLink"))
    
    // Get user by email
    user, err := s.userReader.GetByEmail(ctx, input.Email)
    if err != nil {
        if err == ErrUserNotFound {
            // Don't reveal if user exists
            log.Info("Magic link requested for non-existent email", zap.String("email", input.Email))
            return nil
        }
        log.Error("Error getting user", zap.Error(err))
        return fmt.Errorf("get user: %w", err)
    }
    
    // Check if user is active and magic link enabled
    if !user.IsActive {
        return nil // Don't reveal user status
    }
    
    if !user.MagicLinkEnabled {
        return ErrMagicLinkDisabled
    }
    
    // Generate secure token
    token, err := s.generateSecureToken()
    if err != nil {
        log.Error("Error generating token", zap.Error(err))
        return fmt.Errorf("generate token: %w", err)
    }
    
    // Create magic link
    now := time.Now()
    magicLink := &MagicLink{
        UserID:    user.ID,
        Token:     token,
        Used:      false,
        ExpiresAt: now.Add(15 * time.Minute),
        CreatedAt: now,
        UpdatedAt: now,
    }
    
    err = s.mlWriter.Create(ctx, magicLink)
    if err != nil {
        log.Error("Error creating magic link", zap.Error(err))
        return fmt.Errorf("create magic link: %w", err)
    }
    
    // Send email
    err = s.emailService.SendMagicLink(ctx, user.Email, token)
    if err != nil {
        log.Error("Error sending magic link email", zap.Error(err))
        return fmt.Errorf("send magic link: %w", err)
    }
    
    log.Info("Magic link generated and sent", zap.String("email", input.Email))
    return nil
}

func (s *AuthService) LoginWithMagicLink(ctx context.Context, token string) (*AuthResponse, error) {
    log := logger.GetLogger(zap.String("service", "AuthService"), zap.String("method", "LoginWithMagicLink"))
    
    // Get magic link by token
    magicLink, err := s.mlReader.GetByToken(ctx, token)
    if err != nil {
        if err == ErrMagicLinkNotFound {
            return nil, ErrInvalidMagicLink
        }
        log.Error("Error getting magic link", zap.Error(err))
        return nil, fmt.Errorf("get magic link: %w", err)
    }
    
    // Check if expired
    if time.Now().After(magicLink.ExpiresAt) {
        return nil, ErrMagicLinkExpired
    }
    
    // Mark as used
    err = s.mlWriter.MarkAsUsed(ctx, magicLink.ID)
    if err != nil {
        log.Error("Error marking magic link as used", zap.Error(err))
        return nil, fmt.Errorf("mark magic link as used: %w", err)
    }
    
    // Get user
    user, err := s.userReader.GetByID(ctx, magicLink.UserID)
    if err != nil {
        log.Error("Error getting user", zap.Error(err))
        return nil, fmt.Errorf("get user: %w", err)
    }
    
    // Check if user is active
    if !user.IsActive {
        return nil, ErrUserInactive
    }
    
    // Update last login
    now := time.Now()
    user.LastLoginAt = &now
    user.UpdatedAt = now
    
    err = s.userWriter.Update(ctx, user)
    if err != nil {
        log.Warn("Failed to update last login time", zap.Error(err))
    }
    
    // Create session
    session, tokens, err := s.createSession(ctx, user.ID)
    if err != nil {
        log.Error("Error creating session", zap.Error(err))
        return nil, fmt.Errorf("create session: %w", err)
    }
    
    log.Info("User logged in with magic link", zap.Uint("user_id", user.ID))
    
    return &AuthResponse{
        User:         user,
        Token:        tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
        ExpiresIn:    3600,
    }, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
    log := logger.GetLogger(zap.String("service", "AuthService"), zap.String("method", "Logout"))
    
    err := s.sessionWriter.Delete(ctx, sessionID)
    if err != nil {
        log.Error("Error deleting session", zap.Error(err))
        return fmt.Errorf("delete session: %w", err)
    }
    
    return nil
}

// Helper method to generate secure random token
func (s *AuthService) generateSecureToken() (string, error) {
    bytes := make([]byte, 32)
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}

// Helper method to create session and tokens
func (s *AuthService) createSession(ctx context.Context, userID uint) (*Session, *TokenPair, error) {
    // Generate session ID and tokens
    sessionID, err := s.generateSecureToken()
    if err != nil {
        return nil, nil, fmt.Errorf("generate session ID: %w", err)
    }
    
    accessToken, err := s.generateSecureToken()
    if err != nil {
        return nil, nil, fmt.Errorf("generate access token: %w", err)
    }
    
    refreshToken, err := s.generateSecureToken()
    if err != nil {
        return nil, nil, fmt.Errorf("generate refresh token: %w", err)
    }
    
    // Create session
    now := time.Now()
    session := &Session{
        ID:        sessionID,
        UserID:    userID,
        Token:     accessToken,
        ExpiresAt: now.Add(time.Hour),
        CreatedAt: now,
        UpdatedAt: now,
    }
    
    err = s.sessionWriter.Create(ctx, session)
    if err != nil {
        return nil, nil, fmt.Errorf("create session: %w", err)
    }
    
    tokens := &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }
    
    return session, tokens, nil
}

// Helper struct for token management
type TokenPair struct {
    AccessToken  string
    RefreshToken string
}
