package user

import (
    "fmt"
    "net/smtp"
    "strings"
    "time"

    "go.uber.org/zap"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
    "github.com/TommySanDev/gachiakuta-hispano/config"  // ← Importar config
)

// EmailService handles email operations
type EmailService struct {
    config config.EmailConfig  // ← Usar config.EmailConfig
    log    *zap.Logger
}

// NewEmailService creates a new email service
func NewEmailService(cfg config.EmailConfig) *EmailService {  // ← Cambiar tipo
    return &EmailService{
        config: cfg,
        log:    logger.GetLogger(zap.String("component", "email-service")),
    }
}

// ... resto del código igual, solo cambiar el tipo
