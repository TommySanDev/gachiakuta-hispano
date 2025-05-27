package chapter

import (
    "encoding/json"
    "net/http"
    "strconv"

    "go.uber.org/zap"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Http Helpers

// Writes JSON response to the client
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        logger.Error("Error encoding JSON response", zap.Error(err))
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
    }
}

// Extracts and parses integer parameters from the request
func getIntParam(r *http.Request, key string, defaultValue int) int {
    str := r.URL.Query().Get(key)
    if str == "" {
        return defaultValue
    }

    val, err := strconv.Atoi(str)
    if err != nil {
        return defaultValue
    }

    return val
}

