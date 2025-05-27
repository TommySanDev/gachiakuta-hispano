package middleware

import (
    "net/http"
    "time"

    "github.com/go-chi/chi/v5/middleware"
    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// RequestLogger logs HTTP request details
func RequestLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
        start := time.Now()
        
        defer func() {
            duration := time.Since(start)
            requestID := middleware.GetReqID(r.Context())
            status := ww.Status()
            
            fields := []zap.Field{
                zap.String("method", r.Method),
                zap.String("path", r.URL.Path),
                zap.String("query", r.URL.RawQuery),
                zap.Int("status", status),
                zap.Int("bytes", ww.BytesWritten()),
                zap.Duration("duration", duration),
                zap.String("remote_addr", r.RemoteAddr),
                zap.String("user_agent", r.UserAgent()),
                zap.String("request_id", requestID),
            }
            
            msg := "HTTP Request"
            
            switch {
            case status >= 500:
                logger.Error(msg, fields...)
            case status >= 400:
                logger.Warn(msg, fields...)
            default:
                logger.Info(msg, fields...)
            }
        }()
        
        next.ServeHTTP(ww, r)
    })
}
