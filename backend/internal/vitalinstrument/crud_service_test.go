package vitalinstrument

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Implements the Reader interface for testing purposes
type MockReader struct {
	mock.Mock
}

func (m *MockReader) GetByID(ctx context.Context, id uint) (*VitalInstrument, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*VitalInstrument), args.Error(1)
}

func (m *MockReader) List(ctx context.Context, filter VitalInstrumentFilter) ([]*VitalInstrument, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*VitalInstrument), args.Int(1), args.Error(2)
}

func (m *MockReader) ListByCharacter(ctx context.Context, characterID uint, limit int) ([]*VitalInstrument, error) {
	args := m.Called(ctx, characterID, limit)
	return args.Get(0).([]*VitalInstrument), args.Error(1)
}

// Implements the Writer interface for testing purposes
type MockWriter struct {
	mock.Mock
}

func (m *MockWriter) Create(ctx context.Context, instrument *VitalInstrument) error {
	args := m.Called(ctx, instrument)
	return args.Error(0)
}

func (m *MockWriter) Update(ctx context.Context, instrument *VitalInstrument) error {
	args := m.Called(ctx, instrument)
	return args.Error(0)
}

func (m *MockWriter) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWriter) DeletePermanently(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWriter) Restore(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCrudService_Get(t *testing.T) {
	t.Run("with_valid_id", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		character_id := uint(2)
		instrument := &VitalInstrument{
			ID:             id,
			Name:           "3R",
			CharacterID:    &character_id,
			Description:    "A vital instrument",
		}
		
		mockReader.On("GetByID", ctx, id).Return(instrument, nil)
		
		result, err := service.Get(ctx, id)
		
		assert.NoError(t, err)
		assert.Equal(t, instrument, result)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("with_invalid_id", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		result, err := service.Get(ctx, 0)
		
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidInput))
		assert.Nil(t, result)
	})
	
	t.Run("when_instrument_not_found", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		id := uint(999)
		
		mockReader.On("GetByID", ctx, id).Return(nil, ErrVitalInstrumentNotFound)
		
		result, err := service.Get(ctx, id)
		
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrVitalInstrumentNotFound))
		assert.Nil(t, result)
		mockReader.AssertExpectations(t)
	})
}

func TestCrudService_Create(t *testing.T) {
	t.Run("with_valid_input", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		characterID := uint(1)
		input := CreateVitalInstrumentInput{
			Name:            "3R",
			MainImage:       "/images/3r.jpg",
			Description:     "A vital instrument from Gachiakuta",
			Powers:          "Creation of vital instruments",
			CharacterID:     &characterID,
			FirstAppearance: 1,
		}
		
		mockWriter.On("Create", ctx, mock.AnythingOfType("*vitalinstrument.VitalInstrument")).Return(nil).Run(func(args mock.Arguments) {
			instrument := args.Get(1).(*VitalInstrument)
			assert.Equal(t, input.Name, instrument.Name)
			assert.Equal(t, input.MainImage, instrument.MainImage)
			assert.Equal(t, input.Description, instrument.Description)
			assert.Equal(t, input.Powers, instrument.Powers)
			assert.Equal(t, input.CharacterID, instrument.CharacterID)
			assert.Equal(t, input.FirstAppearance, instrument.FirstAppearance)
			assert.NotZero(t, instrument.CreatedAt)
			assert.NotZero(t, instrument.UpdatedAt)
			
			instrument.ID = 1
		})
		
		result, err := service.Create(ctx, input)
		
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, input.Name, result.Name)
		mockWriter.AssertExpectations(t)
	})
	
	t.Run("with_invalid_input", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		testCases := []struct {
			name  string
			input CreateVitalInstrumentInput
		}{
			{
				name: "missing_name",
				input: CreateVitalInstrumentInput{
					MainImage:       "/images/3r.jpg",
					Description:     "A vital instrument",
					FirstAppearance: 1,
				},
			},
			{
				name: "missing_main_image",
				input: CreateVitalInstrumentInput{
					Name:            "3R",
					Description:     "A vital instrument",
					FirstAppearance: 1,
				},
			},
			{
				name: "missing_description",
				input: CreateVitalInstrumentInput{
					Name:            "3R",
					MainImage:       "/images/3r.jpg",
					FirstAppearance: 1,
				},
			},
			{
				name: "invalid_first_appearance",
				input: CreateVitalInstrumentInput{
					Name:            "3R",
					MainImage:       "/images/3r.jpg",
					Description:     "A vital instrument",
					FirstAppearance: 0,
				},
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := service.Create(ctx, tc.input)
				
				assert.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidInput))
				assert.Nil(t, result)
			})
		}
	})
	
	t.Run("when_writer_returns_error", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		characterID := uint(1)
		input := CreateVitalInstrumentInput{
			Name:            "3R",
			MainImage:       "/images/3r.jpg",
			Description:     "A vital instrument from Gachiakuta",
			Powers:          "Creation of vital instruments",
			CharacterID:     &characterID,
			FirstAppearance: 1,
		}
		
		expectedError := errors.New("database error")
		mockWriter.On("Create", ctx, mock.AnythingOfType("*vitalinstrument.VitalInstrument")).Return(expectedError)
		
		result, err := service.Create(ctx, input)
		
		assert.Error(t, err)
		assert.Nil(t, result)
		mockWriter.AssertExpectations(t)
	})
}

func TestCrudService_Update(t *testing.T) {
	t.Run("successful_update", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		characterID := uint(2)
		existingInstrument := &VitalInstrument{
			ID:          id,
			Name:        "3R",
			Description: "Original description",
			CharacterID: &characterID,
			UpdatedAt:   time.Now().Add(-24 * time.Hour), // 1 day ago
		}
		
		newDescription := "Updated description"
		updateInput := UpdateVitalInstrumentInput{
			Description: &newDescription,
		}
		
		mockReader.On("GetByID", ctx, id).Return(existingInstrument, nil)
		mockWriter.On("Update", ctx, mock.AnythingOfType("*vitalinstrument.VitalInstrument")).Return(nil).Run(func(args mock.Arguments) {
			updatedInstrument := args.Get(1).(*VitalInstrument)
			assert.Equal(t, newDescription, updatedInstrument.Description)
			assert.Equal(t, existingInstrument.Name, updatedInstrument.Name) // Name shouldn't change
			assert.True(t, updatedInstrument.UpdatedAt.After(existingInstrument.UpdatedAt)) // Should have newer timestamp
		})
		
		result, err := service.Update(ctx, id, updateInput)
		
		assert.NoError(t, err)
		assert.Equal(t, newDescription, result.Description)
		mockReader.AssertExpectations(t)
		mockWriter.AssertExpectations(t)
	})
	
	t.Run("instrument_not_found", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(999)
		updateInput := UpdateVitalInstrumentInput{
			Name: stringPtr("New Name"),
		}
		
		mockReader.On("GetByID", ctx, id).Return(nil, ErrVitalInstrumentNotFound)
		
		result, err := service.Update(ctx, id, updateInput)
		
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrVitalInstrumentNotFound))
		assert.Nil(t, result)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("no_changes", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		existingInstrument := &VitalInstrument{
			ID:     id,
			Name:   "3R",
			Powers: "Creation of vital instruments",
		}
		
		updateInput := UpdateVitalInstrumentInput{} // Empty update input
		
		mockReader.On("GetByID", ctx, id).Return(existingInstrument, nil)
		
		result, err := service.Update(ctx, id, updateInput)
		
		assert.NoError(t, err)
		assert.Equal(t, existingInstrument, result)
		mockReader.AssertExpectations(t)
		// Writer shouldn't be called since there are no changes
	})
	
	t.Run("writer_error", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		existingInstrument := &VitalInstrument{
			ID:     id,
			Name:   "3R",
			Powers: "Creation of vital instruments",
		}
		
		newName := "Updated Name"
		updateInput := UpdateVitalInstrumentInput{
			Name: &newName,
		}
		
		expectedError := errors.New("database error")
		mockReader.On("GetByID", ctx, id).Return(existingInstrument, nil)
		mockWriter.On("Update", ctx, mock.AnythingOfType("*vitalinstrument.VitalInstrument")).Return(expectedError)
		
		result, err := service.Update(ctx, id, updateInput)
		
		assert.Error(t, err)
		assert.Nil(t, result)
		mockReader.AssertExpectations(t)
		mockWriter.AssertExpectations(t)
	})
	
	t.Run("invalid_first_appearance", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		existingInstrument := &VitalInstrument{
			ID:              id,
			Name:            "3R",
			FirstAppearance: 1,
		}
		
		invalidFirstAppearance := 0
		updateInput := UpdateVitalInstrumentInput{
			FirstAppearance: &invalidFirstAppearance,
		}
		
		mockReader.On("GetByID", ctx, id).Return(existingInstrument, nil)
		
		result, err := service.Update(ctx, id, updateInput)
		
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidInput))
		assert.Nil(t, result)
		mockReader.AssertExpectations(t)
	})
}

func TestCrudService_Delete(t *testing.T) {
	t.Run("successful_deletion", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		existingInstrument := &VitalInstrument{
			ID:   id,
			Name: "3R",
		}
		
		mockReader.On("GetByID", ctx, id).Return(existingInstrument, nil)
		mockWriter.On("Delete", ctx, id).Return(nil)
		
		err := service.Delete(ctx, id)
		
		assert.NoError(t, err)
		mockReader.AssertExpectations(t)
		mockWriter.AssertExpectations(t)
	})
	
	t.Run("instrument_not_found", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(999)
		
		mockReader.On("GetByID", ctx, id).Return(nil, ErrVitalInstrumentNotFound)
		
		err := service.Delete(ctx, id)
		
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrVitalInstrumentNotFound))
		mockReader.AssertExpectations(t)
	})
	
	t.Run("deletion_error", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		existingInstrument := &VitalInstrument{
			ID:   id,
			Name: "3R",
		}
		
		expectedError := errors.New("database error")
		mockReader.On("GetByID", ctx, id).Return(existingInstrument, nil)
		mockWriter.On("Delete", ctx, id).Return(expectedError)
		
		err := service.Delete(ctx, id)
		
		assert.Error(t, err)
		mockReader.AssertExpectations(t)
		mockWriter.AssertExpectations(t)
	})
}

func TestCrudService_Restore(t *testing.T) {
	t.Run("successful_restoration", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		
		mockWriter.On("Restore", ctx, id).Return(nil)
		
		err := service.Restore(ctx, id)
		
		assert.NoError(t, err)
		mockWriter.AssertExpectations(t)
	})
	
	t.Run("restoration_error", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		expectedError := errors.New("database error")
		
		mockWriter.On("Restore", ctx, id).Return(expectedError)
		
		err := service.Restore(ctx, id)
		
		assert.Error(t, err)
		mockWriter.AssertExpectations(t)
	})
}

func TestCrudService_DeletePermanently(t *testing.T) {
	t.Run("successful_permanent_deletion", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		existingInstrument := &VitalInstrument{
			ID:   id,
			Name: "3R",
		}
		
		mockReader.On("GetByID", ctx, id).Return(existingInstrument, nil)
		mockWriter.On("DeletePermanently", ctx, id).Return(nil)
		
		err := service.DeletePermanently(ctx, id)
		
		assert.NoError(t, err)
		mockReader.AssertExpectations(t)
		mockWriter.AssertExpectations(t)
	})
	
	t.Run("instrument_not_found_but_proceed", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		
		mockReader.On("GetByID", ctx, id).Return(nil, ErrVitalInstrumentNotFound)
		mockWriter.On("DeletePermanently", ctx, id).Return(nil)
		
		err := service.DeletePermanently(ctx, id)
		
		assert.NoError(t, err)
		mockReader.AssertExpectations(t)
		mockWriter.AssertExpectations(t)
	})
	
	t.Run("permanent_deletion_error", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		existingInstrument := &VitalInstrument{
			ID:   id,
			Name: "3R",
		}
		
		expectedError := errors.New("database error")
		mockReader.On("GetByID", ctx, id).Return(existingInstrument, nil)
		mockWriter.On("DeletePermanently", ctx, id).Return(expectedError)
		
		err := service.DeletePermanently(ctx, id)
		
		assert.Error(t, err)
		mockReader.AssertExpectations(t)
		mockWriter.AssertExpectations(t)
	})
}

// Helper function for string pointers in tests
func stringPtr(s string) *string {
	return &s
}
