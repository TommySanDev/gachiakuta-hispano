package user

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"go.uber.org/zap"
	"github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// EmailConfig contains SMTP configuration
type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	BaseURL  string
}

// EmailService handles email operations
type EmailService struct {
	config EmailConfig
	log    *zap.Logger
}

// NewEmailService creates a new email service
func NewEmailService(config EmailConfig) *EmailService {
	return &EmailService{
		config: config,
		log:    logger.GetLogger(zap.String("component", "email-service")),
	}
}

// SendMagicLink sends a magic link to user's email
func (s *EmailService) SendMagicLink(email, token string) error {
	subject := "Tu enlace mágico - Gachiakuta Hispano"
	magicURL := fmt.Sprintf("%s/auth/magic-link?token=%s", s.config.BaseURL, token)
	
	body := fmt.Sprintf(`
¡Hola!

Has solicitado un enlace mágico para iniciar sesión en Gachiakuta Hispano.

Haz clic en el siguiente enlace para acceder a tu cuenta:
%s

Este enlace expirará en 15 minutos por seguridad.

Si no solicitaste este enlace, puedes ignorar este mensaje.

Saludos,
El equipo de Gachiakuta Hispano
`, magicURL)

	return s.sendEmail(email, subject, body)
}

// SendPasswordReset sends a password reset email
func (s *EmailService) SendPasswordReset(email, token string) error {
	subject := "Restablece tu contraseña - Gachiakuta Hispano"
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", s.config.BaseURL, token)
	
	body := fmt.Sprintf(`
¡Hola!

Has solicitado restablecer tu contraseña en Gachiakuta Hispano.

Haz clic en el siguiente enlace para crear una nueva contraseña:
%s

Este enlace expirará en 1 hora por seguridad.

Si no solicitaste este cambio, puedes ignorar este mensaje.

Saludos,
El equipo de Gachiakuta Hispano
`, resetURL)

	return s.sendEmail(email, subject, body)
}

// SendWelcomeEmail sends a welcome email to new users
func (s *EmailService) SendWelcomeEmail(email, username string) error {
	subject := "¡Bienvenido a Gachiakuta Hispano!"
	
	body := fmt.Sprintf(`
¡Hola %s!

¡Bienvenido a Gachiakuta Hispano! Tu cuenta ha sido creada exitosamente.

Ahora puedes:
- Explorar información sobre personajes y instrumentos vitales
- Comentar en los capítulos del manga
- Guardar tus elementos favoritos
- Y mucho más

¡Esperamos que disfrutes de tu experiencia en nuestra comunidad!

Saludos,
El equipo de Gachiakuta Hispano

---
Visita: %s
`, username, s.config.BaseURL)

	return s.sendEmail(email, subject, body)
}

// SendEmailVerification sends an email verification link
func (s *EmailService) SendEmailVerification(email, token string) error {
	subject := "Verifica tu email - Gachiakuta Hispano"
	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.config.BaseURL, token)
	
	body := fmt.Sprintf(`
¡Hola!

Para completar el registro en Gachiakuta Hispano, necesitamos verificar tu dirección de email.

Haz clic en el siguiente enlace para verificar tu cuenta:
%s

Este enlace expirará en 24 horas.

Si no creaste una cuenta en Gachiakuta Hispano, puedes ignorar este mensaje.

Saludos,
El equipo de Gachiakuta Hispano
`, verifyURL)

	return s.sendEmail(email, subject, body)
}

// sendEmail sends an email using SMTP
func (s *EmailService) sendEmail(to, subject, body string) error {
	// Validate email service is configured
	if s.config.Host == "" || s.config.Username == "" || s.config.Password == "" {
		s.log.Warn("Email service not configured, skipping email send",
			zap.String("to", to),
			zap.String("subject", subject))
		return nil
	}

	// Create SMTP authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Prepare message
	msg := s.formatMessage(s.config.From, to, subject, body)

	// Send email
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	err := smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg))
	
	if err != nil {
		s.log.Error("Failed to send email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err))
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.log.Info("Email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject))
	
	return nil
}

// formatMessage formats an email message with proper headers
func (s *EmailService) formatMessage(from, to, subject, body string) string {
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["Date"] = time.Now().Format(time.RFC822)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=utf-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	return msg.String()
}

// IsConfigured returns true if email service is properly configured
func (s *EmailService) IsConfigured() bool {
	return s.config.Host != "" && s.config.Username != "" && s.config.Password != "" && s.config.From != ""
}
