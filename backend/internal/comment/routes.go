package comment

import "github.com/go-chi/chi/v5"

// Comment routes
func (h *Handler) RegisterRoutes(r chi.Router) {
    // Public access
    r.Get("/comments/chapter/{id}", h.ListByChapter)

    // Authenticated user actions
    r.Route("/comments", func(r chi.Router) {
        r.Post("/", h.Create)         // POST /comments
        r.Put("/{id}", h.Update)      // PUT /comments/{id}
        r.Delete("/{id}", h.Delete)   // DELETE /comments/{id}
    })

    // Admin-only permanent deletion
    r.Delete("/admin/comments/{id}/permanent", h.DeletePermanently)
}

