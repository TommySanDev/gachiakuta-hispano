package user

import (
    "context"
    "fmt"
    "net/smtp"

    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Configuration for SMTP email service
type EmailConfig struct {
    Host     string
    Port     string
    Username string
    Password string
    From     string
    BaseURL  string
}

// Handles email sending for authentication flows
type EmailService struct {
    config EmailConfig
}

func NewEmailService(config EmailConfig) *EmailService {
    return &EmailService{
        config: config,
    }
}

func (s *EmailService) SendMagicLink(ctx context.Context, email, token string) error {
    log := logger.GetLogger(zap.String("service", "EmailService"), zap.String("method", "SendMagicLink"))
    
    magicURL := fmt.Sprintf("%s/auth/magic-link?token=%s", s.config.BaseURL, token)
    
    subject := "Your Magic Link for Gachiakuta Hispano"
    body := fmt.Sprintf(`
Hello,

Click the link below to sign in to your account:

%s

This link will expire in 15 minutes.

If you didn't request this, please ignore this email.

Best regards,
Gachiakuta Hispano Team
`, magicURL)

    err := s.sendEmail(email, subject, body)
    if err != nil {
        log.Error("Failed to send magic link email", 
            zap.Error(err), 
            zap.String("email", email),
        )
        return fmt.Errorf("send magic link email: %w", err)
    }

    log.Info("Magic link email sent successfully", zap.String("email", email))
    return nil
}

func (s *EmailService) SendResetPassword(ctx context.Context, email, token string) error {
    log := logger.GetLogger(zap.String("service", "EmailService"), zap.String("method", "SendResetPassword"))
    
    resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", s.config.BaseURL, token)
    
    subject := "Reset Your Password - Gachiakuta Hispano"
    body := fmt.Sprintf(`
Hello,

You requested to reset your password. Click the link below to set a new password:

%s

This link will expire in 1 hour.

If you didn't request this, please ignore this email.

Best regards,
Gachiakuta Hispano Team
`, resetURL)

    err := s.sendEmail(email, subject, body)
    if err != nil {
        log.Error("Failed to send password reset email", 
            zap.Error(err), 
            zap.String("email", email),
        )
        return fmt.Errorf("send password reset email: %w", err)
    }

    log.Info("Password reset email sent successfully", zap.String("email", email))
    return nil
}

func (s *EmailService) sendEmail(to, subject, body string) error {
    auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
    
    msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)
    
    addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
    
    return smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg))
}
