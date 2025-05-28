package internal

import (
	"net/http"
	"time"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/TommySanDev/gachiakuta-hispano/internal/character"
)

// RegisterRoutes sets up all application routes
func RegisterRoutes(db *sqlx.DB) http.Handler {
	router := chi.NewRouter()
	// Common middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))
	// Health check endpoint
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	
	// Character routes 
	characterController := character.NewCharacterController(db)
	router.Route("/characters", func(r chi.Router) {
		r.Get("/", characterController.GetAll)
		r.Post("/", characterController.Create)
		r.Get("/{id}", characterController.GetByID)
		r.Put("/{id}", characterController.Update)
		r.Delete("/{id}", characterController.Delete)
	})
	
	return router
}
