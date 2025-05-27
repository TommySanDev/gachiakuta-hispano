package user

import (
    "github.com/go-chi/chi/v5"
    "github.com/TommySanDev/gachiakuta-hispano/internal/middleware"
)

// User routes with authentication and authorization
func (h *Handler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
    // Public authentication routes
    r.Route("/auth", func(r chi.Router) {
        r.Post("/register", h.Register)
        r.Post("/login", h.Login)
        r.Post("/magic-link", h.RequestMagicLink)
        r.Get("/magic-link", h.LoginWithMagicLink)
        r.Post("/forgot-password", h.RequestPasswordReset)
        r.Get("/reset-password", h.ValidateResetToken)
        r.Post("/reset-password", h.ResetPassword)
    })

    // Authenticated user routes
    r.Route("/users", func(r chi.Router) {
        r.Use(authMiddleware.Authenticate)

        r.Get("/me", h.GetCurrentUser)
        r.Put("/me", h.UpdateProfile)
        r.Put("/password", h.ChangePassword)

        r.Route("/2fa", func(r chi.Router) {
            r.Post("/setup", h.Setup2FA)
            r.Post("/verify", h.Verify2FA)
            r.Delete("/", h.Disable2FA)
            r.Post("/recovery-codes", h.GenerateRecoveryCodes)
        })

        r.Post("/logout", h.Logout)
    })

    // Admin routes
    r.Route("/admin/users", func(r chi.Router) {
        r.Use(authMiddleware.Authenticate)
        r.Use(authMiddleware.RequireRole(RoleAdmin))

        r.Get("/", h.ListUsers)
        r.Post("/", h.CreateUser)
        r.Get("/{id}", h.GetUser)
        r.Put("/{id}", h.UpdateUser)
        r.Delete("/{id}", h.DeleteUser)
        r.Patch("/{id}/restore", h.RestoreUser)
        r.Delete("/{id}/permanent", h.DeleteUserPermanently)
    })
}
