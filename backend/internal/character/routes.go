package character

import "github.com/go-chi/chi/v5"

// Character routes
func (h *Handler) RegisterRoutes(r chi.Router) {
    // Public routes
    r.Route("/characters", func(r chi.Router) {
        r.Get("/", h.List)                
        r.Get("/{id}", h.Get)             
        r.Get("/affiliation/{name}", h.ListByAffiliation) 
        r.Get("/status/{status}", h.ListByStatus)         
        r.Get("/species/{species}", h.ListBySpecies)      
    })

    // Admin routes
    r.Route("/admin/characters", func(r chi.Router) {
        r.Post("/", h.Create)             
        r.Put("/{id}", h.Update)          
        r.Delete("/{id}", h.Delete)       
        r.Patch("/{id}/restore", h.Restore) 
        r.Delete("/{id}/permanent", h.DeletePermanently) 
    })
}
