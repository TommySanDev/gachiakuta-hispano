package character

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

func (m *MockCrudService) Get(ctx context.Context, id uint) (*Character, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Character), args.Error(1)
}

func (m *MockCrudService) Create(ctx context.Context, input CreateCharacterInput) (*Character, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Character), args.Error(1)
}

func (m *MockCrudService) Update(ctx context.Context, id uint, input UpdateCharacterInput) (*Character, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Character), args.Error(1)
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

func (m *MockSearchService) List(ctx context.Context, filter CharacterFilter) ([]*Character, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*Character), args.Int(1), args.Error(2)
}

func (m *MockSearchService) ListByAffiliation(ctx context.Context, affiliation string, limit int) ([]*Character, error) {
	args := m.Called(ctx, affiliation, limit)
	return args.Get(0).([]*Character), args.Error(1)
}

func (m *MockSearchService) ListByStatus(ctx context.Context, status string, limit int) ([]*Character, error) {
	args := m.Called(ctx, status, limit)
	return args.Get(0).([]*Character), args.Error(1)
}

func (m *MockSearchService) ListBySpecies(ctx context.Context, species string, limit int) ([]*Character, error) {
	args := m.Called(ctx, species, limit)
	return args.Get(0).([]*Character), args.Error(1)
}

func TestHandler_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/characters/1", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		character := &Character{
			ID:     1,
			Name:   "Rudo Surebrec",
			Status: "Alive",
		}
		mockCrudService.On("Get", mock.Anything, uint(1)).Return(character, nil)
		
		// Execute
		handler.Get(w, req)
		
		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		
		// Parse response
		var response Character
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, character.ID, response.ID)
		assert.Equal(t, character.Name, response.Name)
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("invalid_id", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request with invalid ID
		req := httptest.NewRequest(http.MethodGet, "/characters/invalid", nil)
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
	
	t.Run("character_not_found", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/characters/999", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Get", mock.Anything, uint(999)).Return(nil, ErrCharacterNotFound)
		
		// Execute
		handler.Get(w, req)
		
		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Character not found")
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("internal_error", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodGet, "/characters/1", nil)
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
		input := CreateCharacterInput{
			Name:            "Rudo Surebrec",
			MainImage:       "/images/rudo.jpg",
			Description:     "A character from Gachiakuta",
			Status:          "Alive",
			FirstAppearance: 1,
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPost, "/admin/characters", bytes.NewBuffer(inputJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Setup mock expectations
		character := &Character{
			ID:              1,
			Name:            input.Name,
			MainImage:       input.MainImage,
			Description:     input.Description,
			Status:          input.Status,
			FirstAppearance: input.FirstAppearance,
		}
		mockCrudService.On("Create", mock.Anything, input).Return(character, nil)
		
		// Execute
		handler.Create(w, req)
		
		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)
		
		// Parse response
		var response Character
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, character.ID, response.ID)
		assert.Equal(t, character.Name, response.Name)
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("invalid_request_body", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request with invalid JSON
		req := httptest.NewRequest(http.MethodPost, "/admin/characters", bytes.NewBuffer([]byte("invalid json")))
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
		input := CreateCharacterInput{
			Name: "Rudo Surebrec",
			// Missing other required fields
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPost, "/admin/characters", bytes.NewBuffer(inputJSON))
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
	
	t.Run("character_already_exists", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test input
		input := CreateCharacterInput{
			Name:            "Rudo Surebrec",
			MainImage:       "/images/rudo.jpg",
			Description:     "A character from Gachiakuta",
			Status:          "Alive",
			FirstAppearance: 1,
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPost, "/admin/characters", bytes.NewBuffer(inputJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		// Setup mock expectations
		mockCrudService.On("Create", mock.Anything, input).Return(nil, ErrCharacterAlreadyExists)
		
		// Execute
		handler.Create(w, req)
		
		// Assert
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "Character already exists")
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("internal_error", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test input
		input := CreateCharacterInput{
			Name:            "Rudo Surebrec",
			MainImage:       "/images/rudo.jpg",
			Description:     "A character from Gachiakuta",
			Status:          "Alive",
			FirstAppearance: 1,
		}
		
		// Convert to JSON for request body
		inputJSON, _ := json.Marshal(input)
		
		// Create test request
		req := httptest.NewRequest(http.MethodPost, "/admin/characters", bytes.NewBuffer(inputJSON))
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
		req := httptest.NewRequest(http.MethodGet, "/characters?page=1&page_size=10", nil)
		w := httptest.NewRecorder()
		
		// Setup mock expectations
		characters := []*Character{
			{ID: 1, Name: "Rudo Surebrec", Status: "Alive"},
			{ID: 2, Name: "Another Character", Status: "Unknown"},
		}
		mockSearchService.On("List", mock.Anything, mock.AnythingOfType("character.CharacterFilter")).
			Return(characters, len(characters), nil).
			Run(func(args mock.Arguments) {
				// Verify filter parameters
				filter := args.Get(1).(CharacterFilter)
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
		req := httptest.NewRequest(http.MethodGet, "/characters?search=Rudo&status=Alive&sort_by=name&sort_dir=asc", nil)
		w := httptest.NewRecorder()
		
		// Setup mock expectations
		characters := []*Character{
			{ID: 1, Name: "Rudo Surebrec", Status: "Alive"},
		}
		mockSearchService.On("List", mock.Anything, mock.AnythingOfType("character.CharacterFilter")).
			Return(characters, len(characters), nil).
			Run(func(args mock.Arguments) {
				// Verify filter parameters
				filter := args.Get(1).(CharacterFilter)
				assert.Equal(t, "Rudo", filter.Search)
				assert.Equal(t, "Alive", filter.Status)
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
		req := httptest.NewRequest(http.MethodGet, "/characters", nil)
		w := httptest.NewRecorder()
		
		// Setup mock expectations
		mockSearchService.On("List", mock.Anything, mock.AnythingOfType("character.CharacterFilter")).
			Return(nil, 0, errors.New("database error"))
		
		// Execute
		handler.List(w, req)
		
		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Error fetching characters")
		
		mockSearchService.AssertExpectations(t)
	})
}

func TestHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodDelete, "/admin/characters/1", nil)
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
		req := httptest.NewRequest(http.MethodDelete, "/admin/characters/invalid", nil)
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
	
	t.Run("character_not_found", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodDelete, "/admin/characters/999", nil)
		w := httptest.NewRecorder()
		
		// Setup chi router context with URL params
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "999")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		
		// Setup mock expectations
		mockCrudService.On("Delete", mock.Anything, uint(999)).Return(ErrCharacterNotFound)
		
		// Execute
		handler.Delete(w, req)
		
		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Character not found")
		
		mockCrudService.AssertExpectations(t)
	})
	
	t.Run("internal_error", func(t *testing.T) {
		mockCrudService := new(MockCrudService)
		mockSearchService := new(MockSearchService)
		handler := NewHandler(mockCrudService, mockSearchService)
		
		// Create test request
		req := httptest.NewRequest(http.MethodDelete, "/admin/characters/1", nil)
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

// Helper function for test requests with URL parameters
func setupRequestWithURLParam(method, urlPath, paramName, paramValue string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, urlPath, nil)
	w := httptest.NewRecorder()
	
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add(paramName, paramValue)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	
	return req, w
}
