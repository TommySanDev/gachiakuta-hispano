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

	// Load configuration
	_ = config.LoadAuthConfig() // Unused for now, kept for future use
	_ = config.LoadSMTPConfig()

	// Connect to database
	db := config.ConnectDB()
	defer db.Close()

	// Register routes
	router := internal.RegisterRoutes(db)

	// Start server
	log.Info("Server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("Server startup error", zap.Error(err))
	}
}

