package logger
import (
    "os"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// Wrapper for zap logger with automatic initialization
var log *zap.Logger

func Init(environment string) {
    var config zap.Config
    if environment == "production" {
        config = zap.NewProductionConfig()
        config.EncoderConfig.TimeKey = "timestamp"
        config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    } else {
        config = zap.NewDevelopmentConfig()
        config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }
    var err error
    log, err = config.Build()
    if err != nil {
        log = zap.NewExample()
    }
    zap.ReplaceGlobals(log)
}

func GetLogger(fields ...zap.Field) *zap.Logger {
    if log == nil {
        Init("development")
    }
    return log.With(fields...)
}

func Info(msg string, fields ...zap.Field) {
    if log == nil {
        Init("development")
    }
    log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
    if log == nil {
        Init("development")
    }
    log.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
    if log == nil {
        Init("development")
    }
    log.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
    if log == nil {
        Init("development")
    }
    log.Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
    if log == nil {
        Init("development")
    }
    log.Fatal(msg, fields...)
    os.Exit(1)
}
