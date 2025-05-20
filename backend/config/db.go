package config

import (
    "time"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

var DB *sqlx.DB

// Establishes database connection and returns it
func ConnectDB() *sqlx.DB {
    log := logger.GetLogger(zap.String("component", "database"))
    connStr := "user=postgres dbname=gachiakuta_fan_db password=postgres host=localhost sslmode=disable"
    
    var err error
    DB, err = sqlx.Connect("postgres", connStr)
    if err != nil {
        log.Fatal("Failed to connect to database", zap.Error(err))
    }
    
    // Configure connection pool
    DB.SetMaxOpenConns(25)
    DB.SetMaxIdleConns(5)
    DB.SetConnMaxLifetime(5 * time.Minute)
    
    log.Info("Successfully connected to database")
    return DB
}

// Returns existing database connection or creates a new one
func GetDB() *sqlx.DB {
    if DB == nil {
        ConnectDB()
    }
    return DB
}
