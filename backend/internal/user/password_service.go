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

// Handles password recovery and reset operations
type PasswordService struct {
    userReader       UserGetter
    userWriter       UserWriter
    resetTokenReader ResetTokenGetter
    resetTokenWriter ResetTokenWriter
    emailService     *EmailService
}

func NewPasswordService(
    userReader UserGetter,
    userWriter UserWriter,
    resetTokenReader ResetTokenGetter,
    resetTokenWriter ResetTokenWriter,
    emailService *EmailService,
) *PasswordService {
    return &PasswordService{
        userReader:       userReader,
        userWriter:       userWriter,
        resetTokenReader: resetTokenReader,
        resetTokenWriter: resetTokenWriter,
        emailService:     emailService,
    }
}

func (s *PasswordService) RequestReset(ctx context.Context, input ResetPasswordRequestInput) error {
    log := logger.GetLogger(zap.String("service", "PasswordService"), zap.String("method", "RequestReset"))
    
    // Get user by email
    user, err := s.userReader.GetByEmail(ctx, input.Email)
    if err != nil {
        if err == ErrUserNotFound {
            // Don't reveal if user exists
            log.Info("Password reset requested for non-existent email", zap.String("email", input.Email))
            return nil
        }
        log.Error("Error getting user", zap.Error(err))
        return fmt.Errorf("get user: %w", err)
    }
    
    // Check if user is active
    if !user.IsActive {
        // Don't reveal user status
        log.Info("Password reset requested for inactive user", zap.Uint("user_id", user.ID))
        return nil
    }
    
    // Invalidate existing reset tokens for security
    err = s.resetTokenWriter.InvalidateByUserID(ctx, user.ID)
    if err != nil {
        log.Error("Error invalidating existing reset tokens", zap.Error(err))
        return fmt.Errorf("invalidate existing tokens: %w", err)
    }
    
    // Generate secure token
    token, err := s.generateSecureToken()
    if err != nil {
        log.Error("Error generating token", zap.Error(err))
        return fmt.Errorf("generate token: %w", err)
    }
    
    // Create reset token
    now := time.Now()
    resetToken := &ResetToken{
        UserID:    user.ID,
        Token:     token,
        Used:      false,
        ExpiresAt: now.Add(time.Hour), // 1 hour expiration
        CreatedAt: now,
        UpdatedAt: now,
    }
    
    err = s.resetTokenWriter.Create(ctx, resetToken)
    if err != nil {
        log.Error("Error creating reset token", zap.Error(err))
        return fmt.Errorf("create reset token: %w", err)
    }
    
    // Send email
    err = s.emailService.SendResetPassword(ctx, user.Email, token)
    if err != nil {
        log.Error("Error sending reset password email", zap.Error(err))
        return fmt.Errorf("send reset email: %w", err)
    }
    
    log.Info("Password reset requested and sent", zap.String("email", input.Email))
    return nil
}

func (s *PasswordService) ValidateResetToken(ctx context.Context, token string) (*User, error) {
    log := logger.GetLogger(zap.String("service", "PasswordService"), zap.String("method", "ValidateResetToken"))
    
    // Get reset token
    resetToken, err := s.resetTokenReader.GetByToken(ctx, token)
    if err != nil {
        if err == ErrResetTokenNotFound {
            return nil, ErrInvalidResetToken
        }
        log.Error("Error getting reset token", zap.Error(err))
        return nil, fmt.Errorf("get reset token: %w", err)
    }
    
    // Check if expired
    if resetToken.IsExpired() {
        return nil, ErrResetTokenExpired
    }
    
    // Get user
    user, err := s.userReader.GetByID(ctx, resetToken.UserID)
    if err != nil {
        log.Error("Error getting user for reset token", zap.Error(err))
        return nil, fmt.Errorf("get user: %w", err)
    }
    
    // Check if user is active
    if !user.IsActive {
        return nil, ErrUserInactive
    }
    
    return user, nil
}

func (s *PasswordService) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
    log := logger.GetLogger(zap.String("service", "PasswordService"), zap.String("method", "ResetPassword"))
    
    // Validate token and get user
    user, err := s.ValidateResetToken(ctx, input.Token)
    if err != nil {
        return err // Error already formatted
    }
    
    // Get reset token to mark as used
    resetToken, err := s.resetTokenReader.GetByToken(ctx, input.Token)
    if err != nil {
        log.Error("Error getting reset token for marking used", zap.Error(err))
        return fmt.Errorf("get reset token: %w", err)
    }
    
    // Hash new password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
    if err != nil {
        log.Error("Error hashing new password", zap.Error(err))
        return fmt.Errorf("hash password: %w", err)
    }
    
    // Update user password
    user.PasswordHash = string(hashedPassword)
    user.UpdatedAt = time.Now()
    
    err = s.userWriter.Update(ctx, user)
    if err != nil {
        log.Error("Error updating user password", zap.Error(err))
        return fmt.Errorf("update password: %w", err)
    }
    
    // Mark token as used
    err = s.resetTokenWriter.MarkAsUsed(ctx, resetToken.ID)
    if err != nil {
        log.Error("Error marking reset token as used", zap.Error(err))
        // Don't fail the password reset for this
    }
    
    // Invalidate all other reset tokens for this user
    err = s.resetTokenWriter.InvalidateByUserID(ctx, user.ID)
    if err != nil {
        log.Warn("Error invalidating other reset tokens", zap.Error(err))
        // Don't fail the password reset for this
    }
    
    log.Info("Password reset successfully", zap.Uint("user_id", user.ID))
    return nil
}

func (s *PasswordService) ChangePassword(ctx context.Context, userID uint, input ChangePasswordInput) error {
    log := logger.GetLogger(zap.String("service", "PasswordService"), zap.String("method", "ChangePassword"))
    
    // Get user
    user, err := s.userReader.GetByID(ctx, userID)
    if err != nil {
        log.Error("Error getting user", zap.Error(err))
        return fmt.Errorf("get user: %w", err)
    }
    
    // Verify current password
    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword))
    if err != nil {
        return ErrInvalidCredentials
    }
    
    // Hash new password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
    if err != nil {
        log.Error("Error hashing new password", zap.Error(err))
        return fmt.Errorf("hash password: %w", err)
    }
    
    // Update password
    user.PasswordHash = string(hashedPassword)
    user.UpdatedAt = time.Now()
    
    err = s.userWriter.Update(ctx, user)
    if err != nil {
        log.Error("Error updating password", zap.Error(err))
        return fmt.Errorf("update password: %w", err)
    }
    
    log.Info("Password changed successfully", zap.Uint("user_id", userID))
    return nil
}

// Helper method to generate secure random token
func (s *PasswordService) generateSecureToken() (string, error) {
    bytes := make([]byte, 32)
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}
