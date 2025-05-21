package character

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSearchService_List(t *testing.T) {
	t.Run("with_default_parameters", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		filter := CharacterFilter{
			// Empty filter, should use defaults
		}
		
		expectedCharacters := []*Character{
			{ID: 1, Name: "Rudo"},
			{ID: 2, Name: "Another Character"},
		}
		totalCount := 2
		
		// Set expectations on the reader mock
		mockReader.On("List", ctx, mock.AnythingOfType("character.CharacterFilter")).
			Return(expectedCharacters, totalCount, nil).
			Run(func(args mock.Arguments) {
				// Verify default values are applied
				appliedFilter := args.Get(1).(CharacterFilter)
				assert.Equal(t, 1, appliedFilter.Page)
				assert.Equal(t, 20, appliedFilter.PageSize)
				assert.Equal(t, "created_at", appliedFilter.SortBy)
				assert.Equal(t, "desc", appliedFilter.SortDir)
			})
		
		// Execute
		characters, total, err := service.List(ctx, filter)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCharacters, characters)
		assert.Equal(t, totalCount, total)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("with_custom_parameters", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		filter := CharacterFilter{
			Search:  "Rudo",
			Status:  "Alive",
			Page:    2,
			PageSize: 15,
			SortBy:  "name",
			SortDir: "asc",
		}
		
		expectedCharacters := []*Character{
			{ID: 1, Name: "Rudo"},
		}
		totalCount := 1
		
		// Set expectations on the reader mock
		mockReader.On("List", ctx, mock.AnythingOfType("character.CharacterFilter")).
			Return(expectedCharacters, totalCount, nil).
			Run(func(args mock.Arguments) {
				// Verify filter values are passed correctly
				appliedFilter := args.Get(1).(CharacterFilter)
				assert.Equal(t, filter.Search, appliedFilter.Search)
				assert.Equal(t, filter.Status, appliedFilter.Status)
				assert.Equal(t, filter.Page, appliedFilter.Page)
				assert.Equal(t, filter.PageSize, appliedFilter.PageSize)
				assert.Equal(t, filter.SortBy, appliedFilter.SortBy)
				assert.Equal(t, filter.SortDir, appliedFilter.SortDir)
			})
		
		// Execute
		characters, total, err := service.List(ctx, filter)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCharacters, characters)
		assert.Equal(t, totalCount, total)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("with_invalid_parameters", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		filter := CharacterFilter{
			Page:     -1,          // Invalid page
			PageSize: 150,         // Too large
			SortBy:   "invalid",   // Invalid sort field
			SortDir:  "invalid",   // Invalid sort direction
		}
		
		expectedCharacters := []*Character{
			{ID: 1, Name: "Rudo"},
		}
		totalCount := 1
		
		// Set expectations on the reader mock
		mockReader.On("List", ctx, mock.AnythingOfType("character.CharacterFilter")).
			Return(expectedCharacters, totalCount, nil).
			Run(func(args mock.Arguments) {
				// Verify invalid values are normalized
				appliedFilter := args.Get(1).(CharacterFilter)
				assert.Equal(t, 1, appliedFilter.Page)
				assert.Equal(t, 100, appliedFilter.PageSize)
				assert.Equal(t, "created_at", appliedFilter.SortBy)
				assert.Equal(t, "desc", appliedFilter.SortDir)
			})
		
		// Execute
		characters, total, err := service.List(ctx, filter)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCharacters, characters)
		assert.Equal(t, totalCount, total)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("reader_error", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		filter := CharacterFilter{
			Page:     1,
			PageSize: 20,
		}
		
		expectedError := errors.New("database error")
		
		// Set expectations on the reader mock
		mockReader.On("List", ctx, mock.AnythingOfType("character.CharacterFilter")).
			Return(nil, 0, expectedError)
		
		// Execute
		characters, total, err := service.List(ctx, filter)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, characters)
		assert.Zero(t, total)
		mockReader.AssertExpectations(t)
	})
}

func TestSearchService_ListByAffiliation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		affiliation := "Limpiadores"
		limit := 10
		
		expectedCharacters := []*Character{
			{ID: 1, Name: "Rudo", Affiliation: "Limpiadores"},
			{ID: 3, Name: "Character 3", Affiliation: "Limpiadores"},
		}
		
		// Set expectations
		mockReader.On("ListByAffiliation", ctx, affiliation, limit).
			Return(expectedCharacters, nil)
		
		// Execute
		characters, err := service.ListByAffiliation(ctx, affiliation, limit)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCharacters, characters)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("with_empty_affiliation", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		// Execute
		characters, err := service.ListByAffiliation(ctx, "", 10)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidInput))
		assert.Nil(t, characters)
	})
	
	t.Run("with_adjusted_limit", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		affiliation := "Limpiadores"
		
		expectedCharacters := []*Character{
			{ID: 1, Name: "Rudo", Affiliation: "Limpiadores"},
		}
		
		// Test cases with different limits
		testCases := []struct {
			inputLimit  int
			adjustedLimit int
		}{
			{inputLimit: -5, adjustedLimit: 10}, // Negative should become default
			{inputLimit: 0, adjustedLimit: 10},  // Zero should become default
			{inputLimit: 100, adjustedLimit: 50}, // Too large should be capped
		}
		
		for _, tc := range testCases {
			// Set expectations for this case
			mockReader.On("ListByAffiliation", ctx, affiliation, tc.adjustedLimit).
				Return(expectedCharacters, nil).Once()
			
			// Execute
			characters, err := service.ListByAffiliation(ctx, affiliation, tc.inputLimit)
			
			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedCharacters, characters)
		}
		
		mockReader.AssertExpectations(t)
	})
	
	t.Run("reader_error", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		affiliation := "Limpiadores"
		limit := 10
		
		expectedError := errors.New("database error")
		
		// Set expectations
		mockReader.On("ListByAffiliation", ctx, affiliation, limit).
			Return(nil, expectedError)
		
		// Execute
		characters, err := service.ListByAffiliation(ctx, affiliation, limit)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, characters)
		mockReader.AssertExpectations(t)
	})
}

func TestSearchService_ListByStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		status := "Alive"
		limit := 10
		
		expectedCharacters := []*Character{
			{ID: 1, Name: "Rudo", Status: "Alive"},
			{ID: 3, Name: "Character 3", Status: "Alive"},
		}
		
		// Set expectations
		mockReader.On("ListByStatus", ctx, status, limit).
			Return(expectedCharacters, nil)
		
		// Execute
		characters, err := service.ListByStatus(ctx, status, limit)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCharacters, characters)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("with_empty_status", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		// Execute
		characters, err := service.ListByStatus(ctx, "", 10)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidInput))
		assert.Nil(t, characters)
	})
	
	t.Run("reader_error", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		status := "Alive"
		limit := 10
		
		expectedError := errors.New("database error")
		
		// Set expectations
		mockReader.On("ListByStatus", ctx, status, limit).
			Return(nil, expectedError)
		
		// Execute
		characters, err := service.ListByStatus(ctx, status, limit)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, characters)
		mockReader.AssertExpectations(t)
	})
}

func TestSearchService_ListBySpecies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		species := "Human"
		limit := 10
		
		expectedCharacters := []*Character{
			{ID: 1, Name: "Rudo", Species: "Human"},
			{ID: 3, Name: "Character 3", Species: "Human"},
		}
		
		// Set expectations
		mockReader.On("ListBySpecies", ctx, species, limit).
			Return(expectedCharacters, nil)
		
		// Execute
		characters, err := service.ListBySpecies(ctx, species, limit)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedCharacters, characters)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("with_empty_species", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		// Execute
		characters, err := service.ListBySpecies(ctx, "", 10)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidInput))
		assert.Nil(t, characters)
	})
	
	t.Run("reader_error", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		species := "Human"
		limit := 10
		
		expectedError := errors.New("database error")
		
		// Set expectations
		mockReader.On("ListBySpecies", ctx, species, limit).
			Return(nil, expectedError)
		
		// Execute
		characters, err := service.ListBySpecies(ctx, species, limit)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, characters)
		mockReader.AssertExpectations(t)
	})
}
