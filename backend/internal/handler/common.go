package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/TommySanDev/gachiakuta-hispano/internal/logger"
	"go.uber.org/zap"
)

// GetIDParam extracts and validates the ID path parameter
func GetIDParam(r *http.Request) (uint, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, err
	}
	return uint(id), nil
}

// RespondWithJSON writes a JSON response with status code
func RespondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// RespondWithError logs and sends a JSON error response
func RespondWithError(w http.ResponseWriter, status int, message string) {
	logger.GetLogger(zap.String("layer", "handler"), zap.String("error", message)).Error("handler error")
	RespondWithJSON(w, status, map[string]string{"error": message})
}

