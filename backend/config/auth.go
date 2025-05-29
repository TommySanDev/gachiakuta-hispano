package config

import (
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// EmailConfig contains SMTP configuration - MOVIDO AQUÍ
type EmailConfig struct {
    Host     string
    Port     string
    Username string
    Password string
    From     string
    BaseURL  string
}

// Authentication configuration settings
type AuthConfig struct {
    // ... resto igual
}

// SMTP configuration for email service
type SMTPConfig struct {
    // ... resto igual
}

// Global auth configuration
var Auth *AuthConfig
var SMTP *SMTPConfig

// GetEmailConfig creates email service configuration - SIN IMPORTAR USER
func GetEmailConfig() EmailConfig {
    if SMTP == nil {
        logger.GetLogger().Fatal("SMTP configuration not loaded")
    }
    
    return EmailConfig{
        Host:     SMTP.Host,
        Port:     SMTP.Port,
        Username: SMTP.Username,
        Password: SMTP.Password,
        From:     SMTP.From,
        BaseURL:  Auth.FrontendURL,
    }
}

// ... resto de funciones igual
