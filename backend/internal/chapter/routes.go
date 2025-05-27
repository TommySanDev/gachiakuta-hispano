package chapter

import "github.com/go-chi/chi/v5"

// Chapter routes
func (h *Handler) RegisterRoutes(r chi.Router) {
    // Public routes
    r.Route("/chapters", func(r chi.Router) {
        r.Get("/", h.List)
        r.Get("/{id}", h.Get)
    })

    // Admin routes
    r.Route("/admin/chapters", func(r chi.Router) {
        r.Post("/", h.Create)
        r.Put("/{id}", h.Update)
        r.Delete("/{id}", h.Delete)
        r.Patch("/{id}/restore", h.Restore)
        r.Delete("/{id}/permanent", h.DeletePermanently)
    })
}

