package internal

import (
	"net/http"
	"time"
	
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	
	"github.com/TommySanDev/gachiakuta-hispano/internal/handler"
)

// RegisterRoutes configura todas las rutas de la aplicación
func RegisterRoutes(db *sqlx.DB) http.Handler {
	router := chi.NewRouter()
	
	// Middleware común
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))
	
	// CORS para desarrollo
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})
	
	// Health check endpoint
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	
	// Inicializar manejadores
	characterHandler := handler.NewCharacterHandler(db)
	vitalInstrumentHandler := handler.NewVitalInstrumentHandler(db)
	
	// Rutas de Character
	router.Route("/api/characters", func(r chi.Router) {
		r.Get("/", characterHandler.GetAll)
		r.Post("/", characterHandler.Create)
		r.Get("/{id}", characterHandler.GetByID)
		r.Put("/{id}", characterHandler.Update)
		r.Delete("/{id}", characterHandler.Delete)
	})
	
	// Rutas de VitalInstrument
	router.Route("/api/vital-instruments", func(r chi.Router) {
		r.Get("/", vitalInstrumentHandler.GetAll)
		r.Post("/", vitalInstrumentHandler.Create)
		r.Get("/{id}", vitalInstrumentHandler.GetByID)
		r.Put("/{id}", vitalInstrumentHandler.Update)
		r.Delete("/{id}", vitalInstrumentHandler.Delete)
		r.Get("/character/{id}", vitalInstrumentHandler.ListByCharacter)
	})
	
	return router
}
