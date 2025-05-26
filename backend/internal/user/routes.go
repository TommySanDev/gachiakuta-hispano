package user

import (
    "github.com/go-chi/chi/v5"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/middleware"
)

// User routes with authentication and authorization
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
    // Public authentication routes
    r.Route("/auth", func(r chi.Router) {
        r.Post("/register", h.Register)                    // POST /auth/register
        r.Post("/login", h.Login)                          // POST /auth/login
        r.Post("/magic-link", h.RequestMagicLink)          // POST /auth/magic-link
        r.Get("/magic-link", h.LoginWithMagicLink)         // GET /auth/magic-link
        r.Post("/forgot-password", h.RequestPasswordReset) // POST /auth/forgot-password
        r.Get("/reset-password", h.ValidateResetToken)     // GET /auth/reset-password
        r.Post("/reset-password", h.ResetPassword)         // POST /auth/reset-password
    })

    // Authenticated user routes
    r.Route("/users", func(r chi.Router) {
        r.Use(authMiddleware.Authenticate)
        
        r.Get("/me", h.GetCurrentUser)                     // GET /users/me
        r.Put("/me", h.UpdateProfile)                      // PUT /users/me
        r.Put("/password", h.ChangePassword)               // PUT /users/password
        
        // 2FA routes
        r.Route("/2fa", func(r chi.Router) {
            r.Post("/setup", h.Setup2FA)                   // POST /users/2fa/setup
            r.Post("/verify", h.Verify2FA)                 // POST /users/2fa/verify
            r.Delete("/", h.Disable2FA)                    // DELETE /users/2fa
            r.Post("/recovery-codes", h.GenerateRecoveryCodes) // POST /users/2fa/recovery-codes
        })
        
        r.Post("/logout", h.Logout)                        // POST /users/logout
    })

    // Admin routes for user management
    r.Route("/admin/users", func(r chi.Router) {
        r.Use(authMiddleware.Authenticate)
        r.Use(authMiddleware.RequireRole(RoleAdmin))
        
        r.Get("/", h.ListUsers)                            // GET /admin/users
        r.Post("/", h.CreateUser)                          // POST /admin/users
        r.Get("/{id}", h.GetUser)                          // GET /admin/users/{id}
        r.Put("/{id}", h.UpdateUser)                       // PUT /admin/users/{id}
        r.Delete("/{id}", h.DeleteUser)                    // DELETE /admin/users/{id}
        r.Patch("/{id}/restore", h.RestoreUser)            // PATCH /admin/users/{id}/restore
        r.Delete("/{id}/permanent", h.DeleteUserPermanently) // DELETE /admin/users/{id}/permanent
    })
}
