package vitalinstrument

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks the CrudService for testing
type MockCrudService struct {
	mock.Mock
}

func (m *MockCrudService) Get(ctx context.Context, id uint) (*VitalInstrument, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*VitalInstrument), args.Error(1)
}

func (m *MockCrudService) Create(ctx context.Context, input CreateVitalInstrumentInput) (*VitalInstrument, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*VitalInstrument), args.Error(1)
}

func (m *MockCrudService) Update(ctx context.Context, id uint, input UpdateVitalInstrumentInput) (*VitalInstrument, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*VitalInstrument), args.Error(1)
}

func (m *MockCrudService) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCrudService) DeletePermanently(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCrudService) Restore(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Mocks the SearchService for testing
type MockSearchService struct {
	mock.Mock
}

func (m *MockSearchService) List(ctx context.Context, filter VitalInstrumentFilter) ([]*VitalInstrument, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*VitalInstrument), args.Int(1), args.Error(2)
}

func (m *MockSearchService) ListByCharacter(ctx context.Context, characterID uint, limit int) ([]*VitalInstrument, error) {
	args := m.Called(ctx, characterID, limit)
	return args.Get(0).([]*VitalInstrument), args.Error(1)
}

func TestHandler_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/vital-instruments/1", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		characterID := uint(2)
		instrument := &VitalInstrument{
			ID:          1,
			Name:        "3R",
			Description: "A vital instrument",
			CharacterID: &characterID,
		}
		mockCrudService.On("Get", mock.Anything, uint(1)).Return(instrument, nil)
		
		// Execute
		handler.Get(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Parse response
		var response VitalInstrument
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, instrument.ID, response.ID)
		assert.Equal(t, instrument.Name, response.Name)
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("invalid_id", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request with invalid ID
		req := httptest.NewRequest(http.MethodGet, "/vital-instruments/invalid", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Execute
		handler.Get(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid ID format")
	})
	
	t.Run("instrument_not_found", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/vital-instruments/999", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Get", mock.Anything, uint(999)).Return(nil, ErrVitalInstrumentNotFound)
		
		// Execute
		handler.Get(w, req)
		
		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Vital instrument not found")
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("internal_error", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/vital-instruments/1", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Get", mock.Anything, uint(1)).Return(nil, errors.New("database error"))
		
		// Execute
		handler.Get(w, req)
		
		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
		
		mockCrudService.AssertExpectations(t)
	})
}

func TestHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test input
		characterID := uint(1)
		input := CreateVitalInstrumentInput{
			Name:            "3R",
			MainImage:       "/images/3r.jpg",
			Description:     "A vital instrument from Gachiakuta",
			Powers:          "Creation of vital instruments",
			CharacterID:     &characterID,
			FirstAppearance: 1,
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPost, "/admin/vital-instruments", bytes.NewBuffer(inputJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Setup mock expectations
		instrument := &VitalInstrument{
			ID:              1,
			Name:            input.Name,
			MainImage:       input.MainImage,
			Description:     input.Description,
			Powers:          input.Powers,
			CharacterID:     input.CharacterID,
			FirstAppearance: input.FirstAppearance,
		}
		mockCrudService.On("Create", mock.Anything, input).Return(instrument, nil)
		
		// Execute
		handler.Create(w, req)
		
		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)
		
		// Parse response
		var response VitalInstrument
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, instrument.ID, response.ID)
		assert.Equal(t, instrument.Name, response.Name)
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("invalid_request_body", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request with invalid JSON
		req := httptest.NewRequest(http.MethodPost, "/admin/vital-instruments", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Execute
		handler.Create(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request payload")
	})
	
	t.Run("validation_error", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test input with missing required fields
		input := CreateVitalInstrumentInput{
			Name: "3R", // Missing other required fields
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPost, "/admin/vital-instruments", bytes.NewBuffer(inputJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Setup mock expectations
		mockCrudService.On("Create", mock.Anything, input).Return(nil, ErrInvalidInput)
		
		// Execute
		handler.Create(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("instrument_already_exists", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test input
		characterID := uint(1)
		input := CreateVitalInstrumentInput{
			Name:            "3R",
			MainImage:       "/images/3r.jpg",
			Description:     "A vital instrument from Gachiakuta",
			Powers:          "Creation of vital instruments",
			CharacterID:     &characterID,
			FirstAppearance: 1,
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPost, "/admin/vital-instruments", bytes.NewBuffer(inputJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Setup mock expectations
		mockCrudService.On("Create", mock.Anything, input).Return(nil, ErrVitalInstrumentAlreadyExists)
		
		// Execute
		handler.Create(w, req)
		
		// Assert
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "Vital instrument already exists")
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("internal_error", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test input
		characterID := uint(1)
		input := CreateVitalInstrumentInput{
			Name:            "3R",
			MainImage:       "/images/3r.jpg",
			Description:     "A vital instrument from Gachiakuta",
			Powers:          "Creation of vital instruments",
			CharacterID:     &characterID,
			FirstAppearance: 1,
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPost, "/admin/vital-instruments", bytes.NewBuffer(inputJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Setup mock expectations
		mockCrudService.On("Create", mock.Anything, input).Return(nil, errors.New("database error"))
		
		// Execute
		handler.Create(w, req)
		
		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
		
		mockCrudService.AssertExpectations(t)
	})
}

func TestHandler_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/vital-instruments?page=1&page_size=10", nil)
		w := httptest.NewRecorder()
		
		// Setup mock expectations
		characterID1 := uint(1)
		characterID2 := uint(2)
		instruments := []*VitalInstrument{
			{ID: 1, Name: "3R", CharacterID: &characterID1},
			{ID: 2, Name: "Another Instrument", CharacterID: &characterID2},
		}
		mockSearchService.On("List", mock.Anything, mock.AnythingOfType("vitalinstrument.VitalInstrumentFilter")).
			Return(instruments, len(instruments), nil).
			Run(func(args mock.Arguments) {
				// Verify filter parameters
				filter := args.Get(1).(VitalInstrumentFilter)
				assert.Equal(t, 1, filter.Page)
				assert.Equal(t, 10, filter.PageSize)
			})
		
		// Execute
		handler.List(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Parse response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		// Check data
		assert.Contains(t, response, "data")
		assert.Contains(t, response, "meta")
		
		data, ok := response["data"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, data, 2)
		
		meta, ok := response["meta"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(2), meta["total"])
		assert.Equal(t, float64(1), meta["page"])
		
		mockSearchService.AssertExpectations(t)
	})
	
	t.Run("with_filter_params", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request with filter params
		req := httptest.NewRequest(http.MethodGet, "/vital-instruments?search=3R&character_id=1&sort_by=name&sort_dir=asc", nil)
		w := httptest.NewRecorder()
		
		// Setup mock expectations
		characterID := uint(1)
		instruments := []*VitalInstrument{
			{ID: 1, Name: "3R", CharacterID: &characterID},
		}
		mockSearchService.On("List", mock.Anything, mock.AnythingOfType("vitalinstrument.VitalInstrumentFilter")).
			Return(instruments, len(instruments), nil).
			Run(func(args mock.Arguments) {
				// Verify filter parameters
				filter := args.Get(1).(VitalInstrumentFilter)
				assert.Equal(t, "3R", filter.Search)
				assert.NotNil(t, filter.CharacterID)
				assert.Equal(t, characterID, *filter.CharacterID)
				assert.Equal(t, "name", filter.SortBy)
				assert.Equal(t, "asc", filter.SortDir)
			})
		
		// Execute
		handler.List(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		mockSearchService.AssertExpectations(t)
	})
	
	t.Run("list_error", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/vital-instruments", nil)
		w := httptest.NewRecorder()
		
		// Setup mock expectations
		mockSearchService.On("List", mock.Anything, mock.AnythingOfType("vitalinstrument.VitalInstrumentFilter")).
			Return(nil, 0, errors.New("database error"))
		
		// Execute
		handler.List(w, req)
		
		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Error fetching vital instruments")
		
		mockSearchService.AssertExpectations(t)
	})
}

func TestHandler_ListByCharacter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/vital-instruments/character/1", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		characterID := uint(1)
		instruments := []*VitalInstrument{
			{ID: 1, Name: "3R", CharacterID: &characterID},
			{ID: 3, Name: "Another Instrument", CharacterID: &characterID},
		}
		mockSearchService.On("ListByCharacter", mock.Anything, characterID, 10).Return(instruments, nil)
		
		// Execute
		handler.ListByCharacter(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Parse response
		var response []*VitalInstrument
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 2)
		
		mockSearchService.AssertExpectations(t)
	})
	
	t.Run("invalid_id", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request with invalid ID
		req := httptest.NewRequest(http.MethodGet, "/vital-instruments/character/invalid", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Execute
		handler.ListByCharacter(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid ID format")
	})
	
	t.Run("invalid_input", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/vital-instruments/character/1", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockSearchService.On("ListByCharacter", mock.Anything, uint(1), 10).Return(nil, ErrInvalidInput)
		
		// Execute
		handler.ListByCharacter(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		mockSearchService.AssertExpectations(t)
	})
	
	t.Run("internal_error", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/vital-instruments/character/1", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockSearchService.On("ListByCharacter", mock.Anything, uint(1), 10).Return(nil, errors.New("database error"))
		
		// Execute
		handler.ListByCharacter(w, req)
		
		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Error fetching vital instruments")
		
		mockSearchService.AssertExpectations(t)
	})
}

func TestHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodDelete, "/admin/vital-instruments/1", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Delete", mock.Anything, uint(1)).Return(nil)
		
		// Execute
		handler.Delete(w, req)
		
		// Assert
		assert.Equal(t, http.StatusNoContent, w.Code)
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("invalid_id", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request with invalid ID
		req := httptest.NewRequest(http.MethodDelete, "/admin/vital-instruments/invalid", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Execute
		handler.Delete(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid ID format")
	})
	
	t.Run("instrument_not_found", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodDelete, "/admin/vital-instruments/999", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Delete", mock.Anything, uint(999)).Return(ErrVitalInstrumentNotFound)
		
		// Execute
		handler.Delete(w, req)
		
		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Vital instrument not found")
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("internal_error", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodDelete, "/admin/vital-instruments/1", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Delete", mock.Anything, uint(1)).Return(errors.New("database error"))
		
		// Execute
		handler.Delete(w, req)
		
		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
		
		mockCrudService.AssertExpectations(t)
	})
}

func TestHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test input
		newName := "Updated 3R"
		input := UpdateVitalInstrumentInput{
			Name: &newName,
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPut, "/admin/vital-instruments/1", bytes.NewBuffer(inputJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		characterID := uint(2)
		updatedInstrument := &VitalInstrument{
			ID:          1,
			Name:        newName,
			Description: "A vital instrument",
			CharacterID: &characterID,
		}
		mockCrudService.On("Update", mock.Anything, uint(1), input).Return(updatedInstrument, nil)
		
		// Execute
		handler.Update(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Parse response
		var response VitalInstrument
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, updatedInstrument.ID, response.ID)
		assert.Equal(t, updatedInstrument.Name, response.Name)
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("invalid_id", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test input
		newName := "Updated 3R"
		input := UpdateVitalInstrumentInput{
			Name: &newName,
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request with invalid ID
		req := httptest.NewRequest(http.MethodPut, "/admin/vital-instruments/invalid", bytes.NewBuffer(inputJSON))
    req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Execute
		handler.Update(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid ID format")
	})
	
	t.Run("invalid_request_body", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request with invalid JSON
		req := httptest.NewRequest(http.MethodPut, "/admin/vital-instruments/1", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Execute
		handler.Update(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request payload")
	})
	
	t.Run("instrument_not_found", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test input
		newName := "Updated 3R"
		input := UpdateVitalInstrumentInput{
			Name: &newName,
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPut, "/admin/vital-instruments/999", bytes.NewBuffer(inputJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Update", mock.Anything, uint(999), input).Return(nil, ErrVitalInstrumentNotFound)
		
		// Execute
		handler.Update(w, req)
		
		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Vital instrument not found")
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("invalid_input", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test input with invalid data
		invalidFirstAppearance := 0
		input := UpdateVitalInstrumentInput{
			FirstAppearance: &invalidFirstAppearance,
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPut, "/admin/vital-instruments/1", bytes.NewBuffer(inputJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Update", mock.Anything, uint(1), input).Return(nil, ErrInvalidInput)
		
		// Execute
		handler.Update(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("internal_error", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test input
		newName := "Updated 3R"
		input := UpdateVitalInstrumentInput{
			Name: &newName,
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPut, "/admin/vital-instruments/1", bytes.NewBuffer(inputJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Update", mock.Anything, uint(1), input).Return(nil, errors.New("database error"))
		
		// Execute
		handler.Update(w, req)
		
		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
		
		mockCrudService.AssertExpectations(t)
	})
}

func TestHandler_Restore(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPatch, "/admin/vital-instruments/1/restore", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Restore", mock.Anything, uint(1)).Return(nil)
		
		// Execute
		handler.Restore(w, req)
		
		// Assert
		assert.Equal(t, http.StatusNoContent, w.Code)
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("invalid_id", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request with invalid ID
		req := httptest.NewRequest(http.MethodPatch, "/admin/vital-instruments/invalid/restore", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Execute
		handler.Restore(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid ID format")
	})
	
	t.Run("instrument_not_found", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPatch, "/admin/vital-instruments/999/restore", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Restore", mock.Anything, uint(999)).Return(ErrVitalInstrumentNotFound)
		
		// Execute
		handler.Restore(w, req)
		
		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Vital instrument not found")
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("internal_error", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPatch, "/admin/vital-instruments/1/restore", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Restore", mock.Anything, uint(1)).Return(errors.New("database error"))
		
		// Execute
		handler.Restore(w, req)
		
		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
		
		mockCrudService.AssertExpectations(t)
	})
}

func TestHandler_DeletePermanently(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodDelete, "/admin/vital-instruments/1/permanent", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("DeletePermanently", mock.Anything, uint(1)).Return(nil)
		
		// Execute
		handler.DeletePermanently(w, req)
		
		// Assert
		assert.Equal(t, http.StatusNoContent, w.Code)
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("invalid_id", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request with invalid ID
		req := httptest.NewRequest(http.MethodDelete, "/admin/vital-instruments/invalid/permanent", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Execute
		handler.DeletePermanently(w, req)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid ID format")
	})
	
	t.Run("instrument_not_found", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodDelete, "/admin/vital-instruments/999/permanent", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("DeletePermanently", mock.Anything, uint(999)).Return(ErrVitalInstrumentNotFound)
		
		// Execute
		handler.DeletePermanently(w, req)
		
		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Vital instrument not found")
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("internal_error", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodDelete, "/admin/vital-instruments/1/permanent", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("DeletePermanently", mock.Anything, uint(1)).Return(errors.New("database error"))
		
		// Execute
		handler.DeletePermanently(w, req)
		
		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
		
		mockCrudService.AssertExpectations(t)
	})
}
