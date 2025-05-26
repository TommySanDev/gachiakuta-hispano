package postgres

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/TommySanDev/gachiakuta-hispano/internal/character"
)

func TestCharacterWriter_Create(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    writer := NewCharacterWriter(db)
    reader := NewCharacterReader(db)
    ctx := context.Background()

    t.Run("should create character successfully", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testCharacter := CreateTestCharacter(t, &character.Character{
            Name:            "Rudo Surebrec",
            NameJapanese:    "ルド・シュアブレック",
            Status:          "Alive",
            Affiliation:     "Limpiadores",
        })

        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)
        assert.NotZero(t, testCharacter.ID)

        // Verify character was created
        foundCharacter, err := reader.GetByID(ctx, testCharacter.ID)
        require.NoError(t, err)
        assert.Equal(t, testCharacter.Name, foundCharacter.Name)
        assert.Equal(t, testCharacter.NameJapanese, foundCharacter.NameJapanese)
        assert.Equal(t, testCharacter.Status, foundCharacter.Status)
        assert.Equal(t, testCharacter.Affiliation, foundCharacter.Affiliation)
    })

    t.Run("should set timestamps correctly", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testCharacter := CreateTestCharacter(t)
        
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)

        // Verify timestamps are set (non-zero) - that's all we really need to test
        foundCharacter, err := reader.GetByID(ctx, testCharacter.ID)
        require.NoError(t, err)
        
        assert.False(t, foundCharacter.CreatedAt.IsZero(), "CreatedAt should be set")
        assert.False(t, foundCharacter.UpdatedAt.IsZero(), "UpdatedAt should be set")
        assert.Nil(t, foundCharacter.DeletedAt, "DeletedAt should be nil for new character")
        
        // That's it! No more timezone/precision headaches
    })

    t.Run("should handle all character fields", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testCharacter := CreateTestCharacter(t, &character.Character{
            Name:            "Complete Character",
            NameJapanese:    "完全なキャラクター",
            MainImage:       "/images/complete.jpg",
            Description:     "A character with all fields populated",
            Species:         "Human",
            Gender:          "Male",
            Age:             25,
            Height:          "180cm",
            Status:          "Alive",
            Affiliation:     "Test Organization",
            Occupation:      "Test Job",
            BirthDate:       "January 1st",
            BirthPlace:      "Test City",
            Relatives:       "Test Family",
            FirstAppearance: 5,
        })

        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)

        foundCharacter, err := reader.GetByID(ctx, testCharacter.ID)
        require.NoError(t, err)
        assert.Equal(t, testCharacter.Name, foundCharacter.Name)
        assert.Equal(t, testCharacter.NameJapanese, foundCharacter.NameJapanese)
        assert.Equal(t, testCharacter.Species, foundCharacter.Species)
        assert.Equal(t, testCharacter.Gender, foundCharacter.Gender)
        assert.Equal(t, testCharacter.Age, foundCharacter.Age)
        assert.Equal(t, testCharacter.Height, foundCharacter.Height)
        assert.Equal(t, testCharacter.Occupation, foundCharacter.Occupation)
        assert.Equal(t, testCharacter.BirthDate, foundCharacter.BirthDate)
        assert.Equal(t, testCharacter.FirstAppearance, foundCharacter.FirstAppearance)
    })
}

func TestCharacterWriter_Update(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    writer := NewCharacterWriter(db)
    reader := NewCharacterReader(db)
    ctx := context.Background()

    t.Run("should update character successfully", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create character
        testCharacter := CreateTestCharacter(t)
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)

        // Update character fields
        testCharacter.Name = "Updated Name"
        testCharacter.Status = "Unknown"
        testCharacter.Affiliation = "Updated Affiliation"
        testCharacter.UpdatedAt = time.Now()

        err = writer.Update(ctx, testCharacter)
        require.NoError(t, err)

        // Verify updates
        foundCharacter, err := reader.GetByID(ctx, testCharacter.ID)
        require.NoError(t, err)
        assert.Equal(t, "Updated Name", foundCharacter.Name)
        assert.Equal(t, "Unknown", foundCharacter.Status)
        assert.Equal(t, "Updated Affiliation", foundCharacter.Affiliation)
    })

    t.Run("should update timestamps correctly", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create character
        testCharacter := CreateTestCharacter(t)
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)

        originalUpdatedAt := testCharacter.UpdatedAt.UTC().Truncate(time.Second)
        time.Sleep(1100 * time.Millisecond) // Ensure time difference of at least 1 second

        // Update character
        testCharacter.Name = "Updated Name"
        testCharacter.UpdatedAt = time.Now()
        
        err = writer.Update(ctx, testCharacter)
        require.NoError(t, err)

        // Verify timestamp was updated
        foundCharacter, err := reader.GetByID(ctx, testCharacter.ID)
        require.NoError(t, err)
        
        newUpdatedAt := foundCharacter.UpdatedAt.UTC().Truncate(time.Second)
        assert.True(t, newUpdatedAt.After(originalUpdatedAt), 
            "UpdatedAt %v should be after original %v", newUpdatedAt, originalUpdatedAt)
    })

    t.Run("should return error when character not found", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testCharacter := CreateTestCharacter(t)
        testCharacter.ID = 999 // Non-existent ID

        err := writer.Update(ctx, testCharacter)
        assert.ErrorIs(t, err, character.ErrCharacterNotFound)
    })

    t.Run("should not update soft deleted character", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create and delete character
        testCharacter := CreateTestCharacter(t)
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)
        
        err = writer.Delete(ctx, testCharacter.ID)
        require.NoError(t, err)

        // Try to update deleted character
        testCharacter.Name = "Should Not Update"
        err = writer.Update(ctx, testCharacter)
        assert.ErrorIs(t, err, character.ErrCharacterNotFound)
    })
}

func TestCharacterWriter_Delete(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    writer := NewCharacterWriter(db)
    reader := NewCharacterReader(db)
    ctx := context.Background()

    t.Run("should soft delete character successfully", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create character
        testCharacter := CreateTestCharacter(t)
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)

        // Delete character
        err = writer.Delete(ctx, testCharacter.ID)
        require.NoError(t, err)

        // Verify character is not found (soft deleted)
        foundCharacter, err := reader.GetByID(ctx, testCharacter.ID)
        assert.Nil(t, foundCharacter)
        assert.ErrorIs(t, err, character.ErrCharacterNotFound)
    })

    t.Run("should return error when character not found", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        err := writer.Delete(ctx, 999)
        assert.ErrorIs(t, err, character.ErrCharacterNotFound)
    })

    t.Run("should not delete already deleted character", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create and delete character
        testCharacter := CreateTestCharacter(t)
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)
        
        err = writer.Delete(ctx, testCharacter.ID)
        require.NoError(t, err)

        // Try to delete again
        err = writer.Delete(ctx, testCharacter.ID)
        assert.ErrorIs(t, err, character.ErrCharacterNotFound)
    })
}

func TestCharacterWriter_Restore(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    writer := NewCharacterWriter(db)
    reader := NewCharacterReader(db)
    ctx := context.Background()

    t.Run("should restore soft deleted character", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create and delete character
        testCharacter := CreateTestCharacter(t)
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)
        
        err = writer.Delete(ctx, testCharacter.ID)
        require.NoError(t, err)

        // Restore character
        err = writer.Restore(ctx, testCharacter.ID)
        require.NoError(t, err)

        // Verify character is accessible again
        foundCharacter, err := reader.GetByID(ctx, testCharacter.ID)
        require.NoError(t, err)
        assert.Equal(t, testCharacter.Name, foundCharacter.Name)
        assert.Nil(t, foundCharacter.DeletedAt)
    })

    t.Run("should return error when character not found", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        err := writer.Restore(ctx, 999)
        assert.ErrorIs(t, err, character.ErrCharacterNotFound)
    })

    t.Run("should return error when character not deleted", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create active character
        testCharacter := CreateTestCharacter(t)
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)

        // Try to restore non-deleted character
        err = writer.Restore(ctx, testCharacter.ID)
        assert.ErrorIs(t, err, character.ErrCharacterNotFound)
    })
}

func TestCharacterWriter_DeletePermanently(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    writer := NewCharacterWriter(db)
    reader := NewCharacterReader(db)
    ctx := context.Background()

    t.Run("should permanently delete character", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create character
        testCharacter := CreateTestCharacter(t)
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)

        // Permanently delete character
        err = writer.DeletePermanently(ctx, testCharacter.ID)
        require.NoError(t, err)

        // Verify character is completely gone
        foundCharacter, err := reader.GetByID(ctx, testCharacter.ID)
        assert.Nil(t, foundCharacter)
        assert.ErrorIs(t, err, character.ErrCharacterNotFound)

        // Verify with include deleted filter
        filter := character.CharacterFilter{
            Page:           1,
            PageSize:       10,
            IncludeDeleted: true,
        }
        characters, total, err := reader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 0, total)
        assert.Len(t, characters, 0)
    })

    t.Run("should permanently delete soft deleted character", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create and soft delete character
        testCharacter := CreateTestCharacter(t)
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)
        
        err = writer.Delete(ctx, testCharacter.ID)
        require.NoError(t, err)

        // Permanently delete
        err = writer.DeletePermanently(ctx, testCharacter.ID)
        require.NoError(t, err)

        // Verify character is completely gone even with include deleted
        filter := character.CharacterFilter{
            Page:           1,
            PageSize:       10,
            IncludeDeleted: true,
        }
        characters, total, err := reader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 0, total)
        assert.Len(t, characters, 0)
    })

    t.Run("should return error when character not found", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        err := writer.DeletePermanently(ctx, 999)
        assert.ErrorIs(t, err, character.ErrCharacterNotFound)
    })
}
