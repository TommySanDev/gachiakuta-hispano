package main

import (
	"net/http"

	"github.com/TommySanDev/gachiakuta-hispano/config"
	"github.com/TommySanDev/gachiakuta-hispano/internal"
	"github.com/TommySanDev/gachiakuta-hispano/internal/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger.Init("development")
	log := logger.GetLogger(zap.String("component", "main"))
	log.Info("Starting Gachiakuta Hispano server")

	// Load configurations
	authConfig := config.LoadAuthConfig()
	smtpConfig := config.LoadSMTPConfig()
	
	if smtpConfig != nil {
		log.Info("SMTP configuration loaded successfully")
	} else {
		log.Warn("SMTP configuration not available - email features will be disabled")
	}

	// Connect to database
	db := config.ConnectDB()
	defer db.Close()

	// Register routes
	router := internal.RegisterRoutes(db)

	// Start server
	port := ":8080"
	log.Info("Server running", 
		zap.String("port", port),
		zap.String("paseto_configured", "yes"),
		zap.Bool("smtp_configured", smtpConfig != nil),
		zap.Int("token_expiration_hours", authConfig.TokenExpirationHours),
	)
	
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatal("Server startup error", zap.Error(err))
	}
}
