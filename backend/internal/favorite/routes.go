package favorite

import "github.com/go-chi/chi/v5"

// Favorite routes
func (h *Handler) RegisterRoutes(r chi.Router) {
    r.Route("/favorites", func(r chi.Router) {
        r.Get("/", h.List)     // GET    /favorites
        r.Post("/", h.Add)     // POST   /favorites
        r.Delete("/", h.Remove) // DELETE /favorites
    })
}

