package config

import (
    "os"
    "time"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "go.uber.org/zap"

    "github.com/joho/godotenv"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

var DB *sqlx.DB

func ConnectDB() *sqlx.DB {
    log := logger.GetLogger(zap.String("component", "database"))

    // Load .env 
    _ = godotenv.Load()

    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    host := os.Getenv("DB_HOST")
    sslmode := os.Getenv("DB_SSLMODE")

    connStr := "user=" + user + " dbname=" + dbname + " password=" + password + " host=" + host + " sslmode=" + sslmode

    var err error
    DB, err = sqlx.Connect("postgres", connStr)
    if err != nil {
        log.Fatal("Failed to connect to database", zap.Error(err))
    }

    DB.SetMaxOpenConns(25)
    DB.SetMaxIdleConns(5)
    DB.SetConnMaxLifetime(5 * time.Minute)

    log.Info("Successfully connected to database")
    return DB
}

func GetDB() *sqlx.DB {
    if DB == nil {
        ConnectDB()
    }
    return DB
}

