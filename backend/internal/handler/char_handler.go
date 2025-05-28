package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/TommySanDev/gachiakuta-hispano/internal/character"
	"github.com/TommySanDev/gachiakuta-hispano/internal/logger"
	"go.uber.org/zap"
	"context"
)

// RegisterCharacterRoutes registers character-related routes to the router
func RegisterCharacterRoutes(r chi.Router, repo *character.PostgresRepository) {
	h := &CharacterHandler{repo: repo}

	r.Route("/characters", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Patch("/{id}/restore", h.Restore)
		r.Delete("/{id}/permanent", h.DeletePermanently)
	})
}

// CharacterHandler handles HTTP requests for characters
type CharacterHandler struct {
	repo *character.PostgresRepository
}

func (h *CharacterHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, err := getIDParam(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	char, err := h.repo.GetByID(ctx, id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, char)
}

func (h *CharacterHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var input character.CreateCharacterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid input")
		return
	}

	c := &character.Character{
		Name:            input.Name,
		NameJapanese:    input.NameJapanese,
		MainImage:       input.MainImage,
		Description:     input.Description,
		Species:         input.Species,
		Gender:          input.Gender,
		Age:             input.Age,
		Height:          input.Height,
		Status:          input.Status,
		Affiliation:     input.Affiliation,
		Occupation:      input.Occupation,
		BirthDate:       input.BirthDate,
		BirthPlace:      input.BirthPlace,
		Relatives:       input.Relatives,
		FirstAppearance: input.FirstAppearance,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.repo.Create(ctx, c); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, c)
}

func (h *CharacterHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, err := getIDParam(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	var input character.UpdateCharacterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid input")
		return
	}

	existing, err := h.repo.GetByID(ctx, id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	// Apply updates if fields are present
	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.NameJapanese != nil {
		existing.NameJapanese = *input.NameJapanese
	}
	if input.MainImage != nil {
		existing.MainImage = *input.MainImage
	}
	if input.Description != nil {
		existing.Description = *input.Description
	}
	if input.Species != nil {
		existing.Species = *input.Species
	}
	if input.Gender != nil {
		existing.Gender = *input.Gender
	}
	if input.Age != nil {
		existing.Age = *input.Age
	}
	if input.Height != nil {
		existing.Height = *input.Height
	}
	if input.Status != nil {
		existing.Status = *input.Status
	}
	if input.Affiliation != nil {
		existing.Affiliation = *input.Affiliation
	}
	if input.Occupation != nil {
		existing.Occupation = *input.Occupation
	}
	if input.BirthDate != nil {
		existing.BirthDate = *input.BirthDate
	}
	if input.BirthPlace != nil {
		existing.BirthPlace = *input.BirthPlace
	}
	if input.Relatives != nil {
		existing.Relatives = *input.Relatives
	}
	if input.FirstAppearance != nil {
		existing.FirstAppearance = *input.FirstAppearance
	}

	existing.UpdatedAt = time.Now()

	if err := h.repo.Update(ctx, existing); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, existing)
}

func (h *CharacterHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, err := getIDParam(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CharacterHandler) Restore(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, err := getIDParam(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	if err := h.repo.Restore(ctx, id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CharacterHandler) DeletePermanently(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	id, err := getIDParam(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid ID")
		return
	}

	if err := h.repo.DeletePermanently(ctx, id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
