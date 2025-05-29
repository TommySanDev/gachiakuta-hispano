package internal

import (
	"net/http"
	"time"
	
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	
	"github.com/TommySanDev/gachiakuta-hispano/config"
	"github.com/TommySanDev/gachiakuta-hispano/internal/handler"
	mw "github.com/TommySanDev/gachiakuta-hispano/internal/middleware"
	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
)

// RegisterRoutes configura todas las rutas de la aplicación
func RegisterRoutes(db *sqlx.DB) http.Handler {
	router := chi.NewRouter()
	
	// Middleware común
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mw.RequestLogger)
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
	
	// Initialize services
	tokenService := user.NewTokenService(config.Auth.PasetoSecret)
	emailService := user.NewEmailService(config.GetEmailConfig())
	authMW := mw.NewAuthMiddleware(db, tokenService)
	
	// Initialize handlers
	characterHandler := handler.NewCharacterHandler(db)
	vitalInstrumentHandler := handler.NewVitalInstrumentHandler(db)
	chapterHandler := handler.NewChapterHandler(db)
	commentHandler := handler.NewCommentHandler(db)
	favoriteHandler := handler.NewFavoriteHandler(db)
	authHandler := handler.NewAuthHandler(db, tokenService, emailService)
	userHandler := handler.NewUserHandler(db)
	passwordHandler := handler.NewPasswordHandler(db, tokenService, emailService)
	totpHandler := handler.NewTOTPHandler(db)
	
	// === PUBLIC ROUTES (no authentication required) ===
	
	// Authentication routes
	router.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		
		// Magic link routes
		r.Post("/magic-link", passwordHandler.RequestMagicLink)
		r.Get("/magic-link", passwordHandler.LoginWithMagicLink)
		
		// Password reset routes
		r.Post("/reset-password", passwordHandler.RequestPasswordReset)
		r.Get("/reset-password/validate", passwordHandler.ValidateResetToken)
		r.Post("/reset-password/confirm", passwordHandler.ResetPassword)
	})
	
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
	
	// Chapter routes
	router.Route("/api/chapters", func(r chi.Router) {
		r.Get("/", chapterHandler.GetAll)
		r.Get("/{id}", chapterHandler.GetByID)
		r.Get("/number", chapterHandler.GetByNumber)
	})
	
	// Comment routes (public read)
	router.Get("/api/comments/chapter/{id}", commentHandler.ListByChapter)
	
	// === AUTHENTICATED ROUTES ===
	router.Route("/api", func(r chi.Router) {
		r.Use(authMW.Authenticate)
		
		// User profile routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/me", userHandler.GetCurrentUser)
			r.Put("/me", userHandler.UpdateProfile)
			r.Post("/me/change-password", userHandler.ChangePassword)
		})
		
		// 2FA routes
		r.Route("/2fa", func(r chi.Router) {
			r.Post("/setup", totpHandler.Setup2FA)
			r.Post("/verify", totpHandler.Verify2FA)
			r.Delete("/disable", totpHandler.Disable2FA)
			r.Post("/recovery-codes", totpHandler.GenerateRecoveryCodes)
			r.Post("/recovery-codes/use", totpHandler.UseRecoveryCode)
			r.Post("/validate", totpHandler.ValidateTOTP)
		})
		
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
		
		r.Route("/chapters", func(r chi.Router) {
			r.Post("/", chapterHandler.Create)
			r.Put("/{id}", chapterHandler.Update)
			r.Delete("/{id}", chapterHandler.Delete)
			r.Patch("/{id}/restore", chapterHandler.Restore)
		})
	})
	
	// === ADMIN ROUTES (admin only) ===
	router.Route("/api/admin", func(r chi.Router) {
		r.Use(authMW.Authenticate)
		r.Use(authMW.RequireRole(user.RoleAdmin))
		
		// User management
		r.Route("/users", func(r chi.Router) {
			r.Get("/", userHandler.ListUsers)
			r.Get("/{id}", userHandler.GetUser)
			r.Post("/", userHandler.CreateUser)
			r.Put("/{id}", userHandler.UpdateUser)
			r.Delete("/{id}", userHandler.DeleteUser)
			r.Delete("/{id}/permanent", userHandler.DeleteUserPermanently)
			r.Patch("/{id}/restore", userHandler.RestoreUser)
		})
		
		// Permanent deletion operations
		r.Delete("/comments/{id}/permanent", commentHandler.DeletePermanently)
		r.Delete("/chapters/{id}/permanent", chapterHandler.DeletePermanently)
	})
	
	return router
}
