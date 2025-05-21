package character

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

func (m *MockReader) GetByID(ctx context.Context, id uint) (*Character, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Character), args.Error(1)
}

func (m *MockReader) List(ctx context.Context, filter CharacterFilter) ([]*Character, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*Character), args.Int(1), args.Error(2)
}

func (m *MockReader) ListByAffiliation(ctx context.Context, affiliation string, limit int) ([]*Character, error) {
	args := m.Called(ctx, affiliation, limit)
	return args.Get(0).([]*Character), args.Error(1)
}

func (m *MockReader) ListByStatus(ctx context.Context, status string, limit int) ([]*Character, error) {
	args := m.Called(ctx, status, limit)
	return args.Get(0).([]*Character), args.Error(1)
}

func (m *MockReader) ListBySpecies(ctx context.Context, species string, limit int) ([]*Character, error) {
	args := m.Called(ctx, species, limit)
	return args.Get(0).([]*Character), args.Error(1)
}

// Implements the Writer interface for testing purposes
type MockWriter struct {
	mock.Mock
}

func (m *MockWriter) Create(ctx context.Context, character *Character) error {
	args := m.Called(ctx, character)
	return args.Error(0)
}

func (m *MockWriter) Update(ctx context.Context, character *Character) error {
	args := m.Called(ctx, character)
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
		// Setup
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		// Define test data
		id := uint(1)
		character := &Character{
			ID:   id,
			Name: "Rudo Surebrec",
		}
		
		// Set expectations
		mockReader.On("GetByID", ctx, id).Return(character, nil)
		
		// Execute
		result, err := service.Get(ctx, id)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, character, result)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("with_invalid_id", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		// Execute
		result, err := service.Get(ctx, 0)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidInput))
		assert.Nil(t, result)
	})
	
	t.Run("when_character_not_found", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		id := uint(999)
		
		// Set expectations
		mockReader.On("GetByID", ctx, id).Return(nil, ErrCharacterNotFound)
		
		// Execute
		result, err := service.Get(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrCharacterNotFound))
		assert.Nil(t, result)
		mockReader.AssertExpectations(t)
	})
}

func TestCrudService_Create(t *testing.T) {
	t.Run("with_valid_input", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		// Define test data
		input := CreateCharacterInput{
			Name:            "Rudo Surebrec",
			MainImage:       "/images/rudo.jpg",
			Description:     "A character from Gachiakuta",
			Status:          "Alive",
			FirstAppearance: 1,
		}
		
		// Set expectations
		mockWriter.On("Create", ctx, mock.AnythingOfType("*character.Character")).Return(nil).Run(func(args mock.Arguments) {
			// Verify that the character was populated correctly
			character := args.Get(1).(*Character)
			assert.Equal(t, input.Name, character.Name)
			assert.Equal(t, input.MainImage, character.MainImage)
			assert.Equal(t, input.Description, character.Description)
			assert.Equal(t, input.Status, character.Status)
			assert.Equal(t, input.FirstAppearance, character.FirstAppearance)
			assert.NotZero(t, character.CreatedAt)
			assert.NotZero(t, character.UpdatedAt)
			
			// Set ID to simulate database insert
			character.ID = 1
		})
		
		// Execute
		result, err := service.Create(ctx, input)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, input.Name, result.Name)
		mockWriter.AssertExpectations(t)
	})
	
	t.Run("with_invalid_input", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		// Define invalid inputs to test
		testCases := []struct {
			name  string
			input CreateCharacterInput
		}{
			{
				name: "missing_name",
				input: CreateCharacterInput{
					MainImage:       "/images/rudo.jpg",
					Description:     "A character",
					Status:          "Alive",
					FirstAppearance: 1,
				},
			},
			{
				name: "missing_main_image",
				input: CreateCharacterInput{
					Name:            "Rudo",
					Description:     "A character",
					Status:          "Alive",
					FirstAppearance: 1,
				},
			},
			{
				name: "missing_description",
				input: CreateCharacterInput{
					Name:            "Rudo",
					MainImage:       "/images/rudo.jpg",
					Status:          "Alive",
					FirstAppearance: 1,
				},
			},
			{
				name: "missing_status",
				input: CreateCharacterInput{
					Name:            "Rudo",
					MainImage:       "/images/rudo.jpg",
					Description:     "A character",
					FirstAppearance: 1,
				},
			},
			{
				name: "invalid_first_appearance",
				input: CreateCharacterInput{
					Name:            "Rudo",
					MainImage:       "/images/rudo.jpg",
					Description:     "A character",
					Status:          "Alive",
					FirstAppearance: 0,
				},
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Execute
				result, err := service.Create(ctx, tc.input)
				
				// Assert
				assert.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidInput))
				assert.Nil(t, result)
			})
		}
	})
	
	t.Run("when_writer_returns_error", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		// Define test data
		input := CreateCharacterInput{
			Name:            "Rudo Surebrec",
			MainImage:       "/images/rudo.jpg",
			Description:     "A character from Gachiakuta",
			Status:          "Alive",
			FirstAppearance: 1,
		}
		
		// Set expectations
		expectedError := errors.New("database error")
		mockWriter.On("Create", ctx, mock.AnythingOfType("*character.Character")).Return(expectedError)
		
		// Execute
		result, err := service.Create(ctx, input)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		mockWriter.AssertExpectations(t)
	})
}

func TestCrudService_Update(t *testing.T) {
	t.Run("successful_update", func(t *testing.T) {
		// Setup
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		existingChar := &Character{
			ID:          id,
			Name:        "Rudo",
			Description: "Original description",
			Status:      "Alive",
			UpdatedAt:   time.Now().Add(-24 * time.Hour), // 1 day ago
		}
		
		newDescription := "Updated description"
		updateInput := UpdateCharacterInput{
			Description: &newDescription,
		}
		
		// Configure mocks
		mockReader.On("GetByID", ctx, id).Return(existingChar, nil)
		mockWriter.On("Update", ctx, mock.AnythingOfType("*character.Character")).Return(nil).Run(func(args mock.Arguments) {
			// Verify the character was updated correctly
			updatedChar := args.Get(1).(*Character)
			assert.Equal(t, newDescription, updatedChar.Description)
			assert.Equal(t, existingChar.Name, updatedChar.Name) // Name shouldn't change
			assert.True(t, updatedChar.UpdatedAt.After(existingChar.UpdatedAt)) // Should have newer timestamp
		})
		
		// Execute
		result, err := service.Update(ctx, id, updateInput)
		
		// Verify
		assert.NoError(t, err)
		assert.Equal(t, newDescription, result.Description)
		mockReader.AssertExpectations(t)
		mockWriter.AssertExpectations(t)
	})
	
	t.Run("character_not_found", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(999)
		updateInput := UpdateCharacterInput{
			Name: stringPtr("New Name"),
		}
		
		mockReader.On("GetByID", ctx, id).Return(nil, ErrCharacterNotFound)
		
		result, err := service.Update(ctx, id, updateInput)
		
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrCharacterNotFound))
		assert.Nil(t, result)
		mockReader.AssertExpectations(t)
	})
	
	t.Run("no_changes", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		existingChar := &Character{
			ID:     id,
			Name:   "Rudo",
			Status: "Alive",
		}
		
		// Empty update input
		updateInput := UpdateCharacterInput{}
		
		mockReader.On("GetByID", ctx, id).Return(existingChar, nil)
		// Writer shouldn't be called since there are no changes
		
		result, err := service.Update(ctx, id, updateInput)
		
		assert.NoError(t, err)
		assert.Equal(t, existingChar, result)
		mockReader.AssertExpectations(t)
		// Writer expectations not checked because it shouldn't be called
	})
	
	t.Run("writer_error", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		existingChar := &Character{
			ID:     id,
			Name:   "Rudo",
			Status: "Alive",
		}
		
		newName := "Updated Name"
		updateInput := UpdateCharacterInput{
			Name: &newName,
		}
		
		expectedError := errors.New("database error")
		mockReader.On("GetByID", ctx, id).Return(existingChar, nil)
		mockWriter.On("Update", ctx, mock.AnythingOfType("*character.Character")).Return(expectedError)
		
		result, err := service.Update(ctx, id, updateInput)
		
		assert.Error(t, err)
		assert.Nil(t, result)
		mockReader.AssertExpectations(t)
		mockWriter.AssertExpectations(t)
	})
}

func TestCrudService_Delete(t *testing.T) {
	t.Run("successful_deletion", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		existingChar := &Character{
			ID:   id,
			Name: "Rudo",
		}
		
		mockReader.On("GetByID", ctx, id).Return(existingChar, nil)
		mockWriter.On("Delete", ctx, id).Return(nil)
		
		err := service.Delete(ctx, id)
		
		assert.NoError(t, err)
		mockReader.AssertExpectations(t)
		mockWriter.AssertExpectations(t)
	})
	
	t.Run("character_not_found", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(999)
		
		mockReader.On("GetByID", ctx, id).Return(nil, ErrCharacterNotFound)
		
		err := service.Delete(ctx, id)
		
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrCharacterNotFound))
		mockReader.AssertExpectations(t)
	})
	
	t.Run("deletion_error", func(t *testing.T) {
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		existingChar := &Character{
			ID:   id,
			Name: "Rudo",
		}
		
		expectedError := errors.New("database error")
		mockReader.On("GetByID", ctx, id).Return(existingChar, nil)
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
		existingChar := &Character{
			ID:   id,
			Name: "Rudo",
		}
		
		mockReader.On("GetByID", ctx, id).Return(existingChar, nil)
		mockWriter.On("DeletePermanently", ctx, id).Return(nil)
		
		err := service.DeletePermanently(ctx, id)
		
		assert.NoError(t, err)
		mockReader.AssertExpectations(t)
		mockWriter.AssertExpectations(t)
	})
	
	t.Run("character_not_found_but_proceed", func(t *testing.T) {
		// Even if character not found, we proceed with deletion
		// because it might be soft-deleted
		mockReader := new(MockReader)
		mockWriter := new(MockWriter)
		service := NewCrudService(mockReader, mockWriter)
		ctx := context.Background()
		
		id := uint(1)
		
		mockReader.On("GetByID", ctx, id).Return(nil, ErrCharacterNotFound)
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
		existingChar := &Character{
			ID:   id,
			Name: "Rudo",
		}
		
		expectedError := errors.New("database error")
		mockReader.On("GetByID", ctx, id).Return(existingChar, nil)
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
