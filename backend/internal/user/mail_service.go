package user

import (
    "fmt"
    "net/smtp"

    "go.uber.org/zap"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
    "github.com/TommySanDev/gachiakuta-hispano/config"
)

// EmailService handles email operations
type EmailService struct {
    config config.EmailConfig
    log    *zap.Logger
}

// NewEmailService creates a new email service
func NewEmailService(cfg config.EmailConfig) *EmailService {
    return &EmailService{
        config: cfg,
        log:    logger.GetLogger(zap.String("component", "email-service")),
    }
}

// IsConfigured checks if email service has valid configuration
func (es *EmailService) IsConfigured() bool {
    return es.config.Host != "" && es.config.From != ""
}

// SendMagicLink sends a magic link email to the user
func (es *EmailService) SendMagicLink(email, token string) error {
    if !es.IsConfigured() {
        es.log.Warn("Email service not configured - skipping magic link email")
        return nil
    }

    magicLinkURL := fmt.Sprintf("%s/auth/magic-link?token=%s", es.config.BaseURL, token)
    
    subject := "Your Magic Link - Gachiakuta Hispano"
    body := fmt.Sprintf(`
Hello,

Click the link below to sign in to your Gachiakuta Hispano account:

%s

This link will expire in 15 minutes for security reasons.

If you didn't request this link, you can safely ignore this email.

Best regards,
The Gachiakuta Hispano Team
`, magicLinkURL)

    return es.sendEmail(email, subject, body)
}

// SendPasswordReset sends a password reset email to the user
func (es *EmailService) SendPasswordReset(email, token string) error {
    if !es.IsConfigured() {
        es.log.Warn("Email service not configured - skipping password reset email")
        return nil
    }

    resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", es.config.BaseURL, token)
    
    subject := "Password Reset - Gachiakuta Hispano"
    body := fmt.Sprintf(`
Hello,

We received a request to reset your password for your Gachiakuta Hispano account.

Click the link below to reset your password:

%s

This link will expire in 1 hour for security reasons.

If you didn't request a password reset, you can safely ignore this email.

Best regards,
The Gachiakuta Hispano Team
`, resetURL)

    return es.sendEmail(email, subject, body)
}

// SendWelcomeEmail sends a welcome email to new users
func (es *EmailService) SendWelcomeEmail(email, username string) error {
    if !es.IsConfigured() {
        es.log.Warn("Email service not configured - skipping welcome email")
        return nil
    }

    subject := "Welcome to Gachiakuta Hispano!"
    body := fmt.Sprintf(`
Hello %s,

Welcome to Gachiakuta Hispano! Your account has been successfully created.

You can now:
- Explore character profiles and vital instruments
- Read chapter discussions
- Add your favorite characters and chapters
- Participate in community discussions

Visit our website: %s

Thank you for joining our community!

Best regards,
The Gachiakuta Hispano Team
`, username, es.config.BaseURL)

    return es.sendEmail(email, subject, body)
}

// sendEmail is the core email sending function
func (es *EmailService) sendEmail(to, subject, body string) error {
    // Build message
    message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)
    
    // Setup authentication
    var auth smtp.Auth
    if es.config.Username != "" {
        auth = smtp.PlainAuth("", es.config.Username, es.config.Password, es.config.Host)
    }
    
    // Send email
    addr := fmt.Sprintf("%s:%s", es.config.Host, es.config.Port)
    err := smtp.SendMail(addr, auth, es.config.From, []string{to}, []byte(message))
    
    if err != nil {
        es.log.Error("Failed to send email",
            zap.String("to", to),
            zap.String("subject", subject),
            zap.Error(err),
        )
        return fmt.Errorf("failed to send email: %w", err)
    }
    
    es.log.Info("Email sent successfully",
        zap.String("to", to),
        zap.String("subject", subject),
    )
    
    return nil
}
