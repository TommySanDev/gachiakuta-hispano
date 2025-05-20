package vitalinstrument

import "github.com/go-chi/chi/v5"

// Vital Instrumen routes
func (h *Handler) RegisterRoutes(r chi.Router) {
    // Public routes
    r.Route("/vital-instruments", func(r chi.Router) {
        r.Get("/", h.List)                           // GET /vital-instruments
        r.Get("/{id}", h.Get)                        // GET /vital-instruments/123
        r.Get("/character/{id}", h.ListByCharacter)  // GET /vital-instruments/character/123
    })

    // Admin routes
    r.Route("/admin/vital-instruments", func(r chi.Router) {
        r.Post("/", h.Create)                        // POST /admin/vital-instruments
        r.Put("/{id}", h.Update)                     // PUT /admin/vital-instruments/123
        r.Delete("/{id}", h.Delete)                  // DELETE /admin/vital-instruments/123
        r.Patch("/{id}/restore", h.Restore)          // PATCH /admin/vital-instruments/123/restore
        r.Delete("/{id}/permanent", h.DeletePermanently) // DELETE /admin/vital-instruments/123/permanent
    })
}
