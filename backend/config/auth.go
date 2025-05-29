package config

import (
    "os"
    "strconv"
    "time"

    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
    "github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// Authentication configuration settings
type AuthConfig struct {
    // Paseto settings (reemplaza JWT)
    PasetoSecret       string
    TokenExpirationHours int
    RefreshTokenDays   int
    
    // Session settings
    SessionTimeout     time.Duration
    MaxSessionsPerUser int
    
    // Magic link settings
    MagicLinkExpiration time.Duration
    MagicLinkEnabled    bool
    
    // Password reset settings
    ResetTokenExpiration time.Duration
    
    // Rate limiting
    LoginAttempts       int
    LoginWindowMinutes  int
    
    // Base URLs
    FrontendURL string
    BackendURL  string
}

// SMTP configuration for email service
type SMTPConfig struct {
    Host     string
    Port     string
    Username string
    Password string
    From     string
    TLS      bool
}

// Global auth configuration
var Auth *AuthConfig
var SMTP *SMTPConfig

// LoadAuthConfig initializes authentication configuration from environment
func LoadAuthConfig() *AuthConfig {
    log := logger.GetLogger(zap.String("component", "auth-config"))
    
    // Load Paseto settings (reemplaza JWT)
    pasetoSecret := getEnvOrDefault("PASETO_SECRET", "")
    if pasetoSecret == "" {
        log.Fatal("PASETO_SECRET environment variable is required")
    }
    
    tokenExpiration := getEnvAsIntOrDefault("TOKEN_EXPIRATION_HOURS", 24)
    refreshTokenDays := getEnvAsIntOrDefault("REFRESH_TOKEN_DAYS", 30)
    
    // Load session settings
    sessionTimeoutMinutes := getEnvAsIntOrDefault("SESSION_TIMEOUT_MINUTES", 60)
    maxSessionsPerUser := getEnvAsIntOrDefault("MAX_SESSIONS_PER_USER", 5)
    
    // Load magic link settings
    magicLinkMinutes := getEnvAsIntOrDefault("MAGIC_LINK_EXPIRATION_MINUTES", 15)
    magicLinkEnabled := getEnvAsBoolOrDefault("MAGIC_LINK_ENABLED", true)
    
    // Load password reset settings
    resetTokenMinutes := getEnvAsIntOrDefault("RESET_TOKEN_EXPIRATION_MINUTES", 60)
    
    // Load rate limiting settings
    loginAttempts := getEnvAsIntOrDefault("MAX_LOGIN_ATTEMPTS", 5)
    loginWindowMinutes := getEnvAsIntOrDefault("LOGIN_WINDOW_MINUTES", 15)
    
    // Load URLs
    frontendURL := getEnvOrDefault("FRONTEND_URL", "http://localhost:3000")
    backendURL := getEnvOrDefault("BACKEND_URL", "http://localhost:8080")
    
    config := &AuthConfig{
        PasetoSecret:         pasetoSecret,
        TokenExpirationHours: tokenExpiration,
        RefreshTokenDays:     refreshTokenDays,
        SessionTimeout:       time.Duration(sessionTimeoutMinutes) * time.Minute,
        MaxSessionsPerUser:   maxSessionsPerUser,
        MagicLinkExpiration:  time.Duration(magicLinkMinutes) * time.Minute,
        MagicLinkEnabled:     magicLinkEnabled,
        ResetTokenExpiration: time.Duration(resetTokenMinutes) * time.Minute,
        LoginAttempts:        loginAttempts,
        LoginWindowMinutes:   loginWindowMinutes,
        FrontendURL:          frontendURL,
        BackendURL:           backendURL,
    }
    
    Auth = config
    log.Info("Authentication configuration loaded successfully",
        zap.Int("token_expiration_hours", tokenExpiration),
        zap.Bool("magic_link_enabled", magicLinkEnabled),
        zap.Int("session_timeout_minutes", sessionTimeoutMinutes),
    )
    
    return config
}

// LoadSMTPConfig initializes SMTP configuration from environment
func LoadSMTPConfig() *SMTPConfig {
    log := logger.GetLogger(zap.String("component", "smtp-config"))
    
    host := getEnvOrDefault("SMTP_HOST", "")
    port := getEnvOrDefault("SMTP_PORT", "587")
    username := getEnvOrDefault("SMTP_USERNAME", "")
    password := getEnvOrDefault("SMTP_PASSWORD", "")
    from := getEnvOrDefault("SMTP_FROM", "")
    tls := getEnvAsBoolOrDefault("SMTP_TLS", true)
    
    if host == "" || username == "" || password == "" || from == "" {
        log.Warn("SMTP configuration incomplete - email features will be disabled")
        return nil
    }
    
    config := &SMTPConfig{
        Host:     host,
        Port:     port,
        Username: username,
        Password: password,
        From:     from,
        TLS:      tls,
    }
    
    SMTP = config
    log.Info("SMTP configuration loaded successfully",
        zap.String("host", host),
        zap.String("port", port),
        zap.Bool("tls", tls),
    )
    
    return config
}

// GetEmailConfig creates email service configuration
func GetEmailConfig() user.EmailConfig {
    if SMTP == nil {
        logger.GetLogger().Fatal("SMTP configuration not loaded")
    }
    
    return user.EmailConfig{
        Host:     SMTP.Host,
        Port:     SMTP.Port,
        Username: SMTP.Username,
        Password: SMTP.Password,
        From:     SMTP.From,
        BaseURL:  Auth.FrontendURL,
    }
}

// Helper functions for environment variable parsing
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}
