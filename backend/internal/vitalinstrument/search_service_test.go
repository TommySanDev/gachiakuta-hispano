package vitalinstrument

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
		
		filter := VitalInstrumentFilter{
			// Empty filter, should use defaults
		}
		
		expectedInstruments := []*VitalInstrument{
			{ID: 1, Name: "3R"},
			{ID: 2, Name: "Another Instrument"},
		}
		totalCount := 2
		
		// Set expectations on the reader mock
		mockReader.On("List", ctx, mock.AnythingOfType("vitalinstrument.VitalInstrumentFilter")).
			Return(expectedInstruments, totalCount, nil).
			Run(func(args mock.Arguments) {
				// Verify default values are applied
				appliedFilter := args.Get(1).(VitalInstrumentFilter)
				assert.Equal(t, 1, appliedFilter.Page)
				assert.Equal(t, 20, appliedFilter.PageSize)
				assert.Equal(t, "created_at", appliedFilter.SortBy)
				assert.Equal(t, "desc", appliedFilter.SortDir)
			})
		
		// Execute
		instruments, total, err := service.List(ctx, filter)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedInstruments, instruments)
		assert.Equal(t, totalCount, total)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("with_custom_parameters", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		characterID := uint(1)
		filter := VitalInstrumentFilter{
			Search:     "3R",
			CharacterID: &characterID,
			Page:       2,
			PageSize:   15,
			SortBy:     "name",
			SortDir:    "asc",
		}
		
		expectedInstruments := []*VitalInstrument{
			{ID: 1, Name: "3R"},
		}
		totalCount := 1
		
		// Set expectations on the reader mock
		mockReader.On("List", ctx, mock.AnythingOfType("vitalinstrument.VitalInstrumentFilter")).
			Return(expectedInstruments, totalCount, nil).
			Run(func(args mock.Arguments) {
				// Verify filter values are passed correctly
				appliedFilter := args.Get(1).(VitalInstrumentFilter)
				assert.Equal(t, filter.Search, appliedFilter.Search)
				assert.Equal(t, filter.CharacterID, appliedFilter.CharacterID)
				assert.Equal(t, filter.Page, appliedFilter.Page)
				assert.Equal(t, filter.PageSize, appliedFilter.PageSize)
				assert.Equal(t, filter.SortBy, appliedFilter.SortBy)
				assert.Equal(t, filter.SortDir, appliedFilter.SortDir)
			})
		
		// Execute
		instruments, total, err := service.List(ctx, filter)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedInstruments, instruments)
		assert.Equal(t, totalCount, total)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("with_invalid_parameters", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		filter := VitalInstrumentFilter{
			Page:     -1,          // Invalid page
			PageSize: 150,         // Too large
			SortBy:   "invalid",   // Invalid sort field
			SortDir:  "invalid",   // Invalid sort direction
		}
		
		expectedInstruments := []*VitalInstrument{
			{ID: 1, Name: "3R"},
		}
		totalCount := 1
		
		// Set expectations on the reader mock
		mockReader.On("List", ctx, mock.AnythingOfType("vitalinstrument.VitalInstrumentFilter")).
			Return(expectedInstruments, totalCount, nil).
			Run(func(args mock.Arguments) {
				// Verify invalid values are normalized
				appliedFilter := args.Get(1).(VitalInstrumentFilter)
				assert.Equal(t, 1, appliedFilter.Page)
				assert.Equal(t, 100, appliedFilter.PageSize)
				assert.Equal(t, "created_at", appliedFilter.SortBy)
				assert.Equal(t, "desc", appliedFilter.SortDir)
			})
		
		// Execute
		instruments, total, err := service.List(ctx, filter)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedInstruments, instruments)
		assert.Equal(t, totalCount, total)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("reader_error", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		filter := VitalInstrumentFilter{
			Page:     1,
			PageSize: 20,
		}
		
		expectedError := errors.New("database error")
		
		// Set expectations on the reader mock
		mockReader.On("List", ctx, mock.AnythingOfType("vitalinstrument.VitalInstrumentFilter")).
			Return(nil, 0, expectedError)
		
		// Execute
		instruments, total, err := service.List(ctx, filter)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, instruments)
		assert.Zero(t, total)
		mockReader.AssertExpectations(t)
	})
}

func TestSearchService_ListByCharacter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		characterID := uint(1)
		limit := 10
		
		expectedInstruments := []*VitalInstrument{
			{ID: 1, Name: "3R", CharacterID: &characterID},
			{ID: 3, Name: "Instrument 3", CharacterID: &characterID},
		}
		
		// Set expectations
		mockReader.On("ListByCharacter", ctx, characterID, limit).
			Return(expectedInstruments, nil)
		
		// Execute
		instruments, err := service.ListByCharacter(ctx, characterID, limit)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedInstruments, instruments)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("with_zero_character_id", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		// Execute
		instruments, err := service.ListByCharacter(ctx, 0, 10)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidInput))
		assert.Nil(t, instruments)
	})
	
	t.Run("with_adjusted_limit", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		characterID := uint(1)
		
		expectedInstruments := []*VitalInstrument{
			{ID: 1, Name: "3R", CharacterID: &characterID},
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
			mockReader.On("ListByCharacter", ctx, characterID, tc.adjustedLimit).
				Return(expectedInstruments, nil).Once()
			
			// Execute
			instruments, err := service.ListByCharacter(ctx, characterID, tc.inputLimit)
			
			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedInstruments, instruments)
		}
		
		mockReader.AssertExpectations(t)
	})
	
	t.Run("reader_error", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		service := NewSearchService(mockReader)
		ctx := context.Background()
		
		characterID := uint(1)
		limit := 10
		
		expectedError := errors.New("database error")
		
		// Set expectations
		mockReader.On("ListByCharacter", ctx, characterID, limit).
			Return(nil, expectedError)
		
		// Execute
		instruments, err := service.ListByCharacter(ctx, characterID, limit)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, instruments)
		mockReader.AssertExpectations(t)
	})
}
