package internal

import (
	"net/http"
	"time"
	
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	
	"github.com/TommySanDev/gachiakuta-hispano/config"
	"github.com/TommySanDev/gachiakuta-hispano/internal/handler"
	"github.com/TommySanDev/gachiakuta-hispano/internal/middleware"
	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
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
	
	// Initialize auth middleware
	tokenService := user.NewTokenService(config.Auth.JWTSecret)
	authMW := middleware.NewAuthMiddleware(db, tokenService)
	
	// Initialize handlers
	characterHandler := handler.NewCharacterHandler(db)
	vitalInstrumentHandler := handler.NewVitalInstrumentHandler(db)
	commentHandler := handler.NewCommentHandler(db)
	favoriteHandler := handler.NewFavoriteHandler(db)
	
	// === PUBLIC ROUTES (no authentication required) ===
	
	// Character routes
	router.Route("/api/characters", func(r chi.Router) {
		r.Get("/", characterHandler.GetAll)
		r.Get("/{id}", characterHandler.GetByID)
	})
	
	// VitalInstrument routes
	router.Route("/api/vital-instruments", func(r chi.Router) {
		r.Get("/", vitalInstrumentHandler.GetAll)
		r.Get("/{id}", vitalInstrumentHandler.GetByID)
		r.Get("/character/{id}", vitalInstrumentHandler.ListByCharacter)
	})
	
	// Comment routes (public read)
	router.Get("/api/comments/chapter/{id}", commentHandler.ListByChapter)
	
	// === AUTHENTICATED ROUTES ===
	router.Route("/api", func(r chi.Router) {
		r.Use(authMW.Authenticate)
		
		// Comment operations (authenticated users)
		r.Route("/comments", func(r chi.Router) {
			r.Post("/", commentHandler.Create)         // POST /comments
			r.Put("/{id}", commentHandler.Update)      // PUT /comments/{id}
			r.Delete("/{id}", commentHandler.Delete)   // DELETE /comments/{id}
		})
		
		// Favorite operations (authenticated users)
		r.Route("/favorites", func(r chi.Router) {
			r.Get("/", favoriteHandler.List)     // GET /favorites
			r.Post("/", favoriteHandler.Add)     // POST /favorites
			r.Delete("/", favoriteHandler.Remove) // DELETE /favorites
		})
	})
	
	// === EDITOR ROUTES (editor + admin) ===
	router.Route("/api/editor", func(r chi.Router) {
		r.Use(authMW.Authenticate)
		r.Use(authMW.RequireRole(user.RoleEditor))
		
		// Content management
		r.Route("/characters", func(r chi.Router) {
			r.Post("/", characterHandler.Create)
			r.Put("/{id}", characterHandler.Update)
			r.Delete("/{id}", characterHandler.Delete)
		})
		
		r.Route("/vital-instruments", func(r chi.Router) {
			r.Post("/", vitalInstrumentHandler.Create)
			r.Put("/{id}", vitalInstrumentHandler.Update)
			r.Delete("/{id}", vitalInstrumentHandler.Delete)
		})
	})
	
	// === ADMIN ROUTES (admin only) ===
	router.Route("/api/admin", func(r chi.Router) {
		r.Use(authMW.Authenticate)
		r.Use(authMW.RequireRole(user.RoleAdmin))
		
		// Permanent deletion operations
		r.Delete("/comments/{id}/permanent", commentHandler.DeletePermanently)
	})
	
	return router
}
