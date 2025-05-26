package postgres

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/TommySanDev/gachiakuta-hispano/internal/character"
    "github.com/TommySanDev/gachiakuta-hispano/internal/vitalinstrument"
)

func TestVitalInstrumentWriter_Create(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    viWriter := NewVitalInstrumentWriter(db)
    viReader := NewVitalInstrumentReader(db)
    charWriter := NewCharacterWriter(db)
    ctx := context.Background()

    t.Run("should create vital instrument with character successfully", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create character first
        testCharacter := CreateTestCharacter(t, &character.Character{
            Name: "Rudo Surebrec",
        })
        err := charWriter.Create(ctx, testCharacter)
        require.NoError(t, err)

        // Create vital instrument
        testInstrument := CreateTestVitalInstrument(t, &testCharacter.ID, &vitalinstrument.VitalInstrument{
            Name:        "3R",
            Description: "A powerful vital instrument from Gachiakuta",
            Powers:      "Creation of vital instruments",
        })

        err = viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)
        assert.NotZero(t, testInstrument.ID)

        // Verify instrument was created
        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        require.NoError(t, err)
        assert.Equal(t, testInstrument.Name, foundInstrument.Name)
        assert.Equal(t, testInstrument.Description, foundInstrument.Description)
        assert.Equal(t, testInstrument.Powers, foundInstrument.Powers)
        assert.Equal(t, testCharacter.ID, *foundInstrument.CharacterID)
    })

    t.Run("should create vital instrument without character", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testInstrument := CreateTestVitalInstrument(t, nil, &vitalinstrument.VitalInstrument{
            Name: "Unowned Instrument",
        })

        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)
        assert.NotZero(t, testInstrument.ID)

        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        require.NoError(t, err)
        assert.Equal(t, testInstrument.Name, foundInstrument.Name)
        assert.Nil(t, foundInstrument.CharacterID)
    })

    t.Run("should set timestamps correctly", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testInstrument := CreateTestVitalInstrument(t, nil)
        
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)

        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        require.NoError(t, err)
        
        assert.False(t, foundInstrument.CreatedAt.IsZero(), "CreatedAt should be set")
        assert.False(t, foundInstrument.UpdatedAt.IsZero(), "UpdatedAt should be set")
        assert.Nil(t, foundInstrument.DeletedAt, "DeletedAt should be nil for new instrument")
        
        // That's it! No more timezone/precision headaches
    })

    t.Run("should handle all vital instrument fields", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testCharacter := CreateTestCharacter(t)
        err := charWriter.Create(ctx, testCharacter)
        require.NoError(t, err)

        testInstrument := CreateTestVitalInstrument(t, &testCharacter.ID, &vitalinstrument.VitalInstrument{
            Name:            "Complete Instrument",
            MainImage:       "/images/complete-instrument.jpg",
            Description:     "A vital instrument with all fields populated",
            Powers:          "Multiple powerful abilities",
            FirstAppearance: 5,
        })

        err = viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)

        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        require.NoError(t, err)
        assert.Equal(t, testInstrument.Name, foundInstrument.Name)
        assert.Equal(t, testInstrument.MainImage, foundInstrument.MainImage)
        assert.Equal(t, testInstrument.Description, foundInstrument.Description)
        assert.Equal(t, testInstrument.Powers, foundInstrument.Powers)
        assert.Equal(t, testInstrument.FirstAppearance, foundInstrument.FirstAppearance)
        assert.Equal(t, testCharacter.ID, *foundInstrument.CharacterID)
    })
}

func TestVitalInstrumentWriter_Update(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    viWriter := NewVitalInstrumentWriter(db)
    viReader := NewVitalInstrumentReader(db)
    charWriter := NewCharacterWriter(db)
    ctx := context.Background()

    t.Run("should update vital instrument successfully", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create instrument
        testInstrument := CreateTestVitalInstrument(t, nil)
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)

        // Update instrument fields
        testInstrument.Name = "Updated Name"
        testInstrument.Description = "Updated description"
        testInstrument.Powers = "Updated powers"
        testInstrument.UpdatedAt = time.Now()

        err = viWriter.Update(ctx, testInstrument)
        require.NoError(t, err)

        // Verify updates
        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        require.NoError(t, err)
        assert.Equal(t, "Updated Name", foundInstrument.Name)
        assert.Equal(t, "Updated description", foundInstrument.Description)
        assert.Equal(t, "Updated powers", foundInstrument.Powers)
    })

    t.Run("should update character ownership", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create characters
        char1 := CreateTestCharacter(t, &character.Character{Name: "Character 1"})
        char2 := CreateTestCharacter(t, &character.Character{Name: "Character 2"})
        err := charWriter.Create(ctx, char1)
        require.NoError(t, err)
        err = charWriter.Create(ctx, char2)
        require.NoError(t, err)

        // Create instrument owned by char1
        testInstrument := CreateTestVitalInstrument(t, &char1.ID)
        err = viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)

        // Transfer ownership to char2
        testInstrument.CharacterID = &char2.ID
        testInstrument.UpdatedAt = time.Now()

        err = viWriter.Update(ctx, testInstrument)
        require.NoError(t, err)

        // Verify ownership transfer
        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        require.NoError(t, err)
        assert.Equal(t, char2.ID, *foundInstrument.CharacterID)
    })

    t.Run("should update timestamps correctly", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create instrument
        testInstrument := CreateTestVitalInstrument(t, nil)
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)

        originalUpdatedAt := testInstrument.UpdatedAt.UTC().Truncate(time.Second)
        time.Sleep(1100 * time.Millisecond) // Ensure time difference of at least 1 second

        // Update instrument
        testInstrument.Name = "Updated Name"
        testInstrument.UpdatedAt = time.Now()
        
        err = viWriter.Update(ctx, testInstrument)
        require.NoError(t, err)

        // Verify timestamp was updated
        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        require.NoError(t, err)
        
        newUpdatedAt := foundInstrument.UpdatedAt.UTC().Truncate(time.Second)
        assert.True(t, newUpdatedAt.After(originalUpdatedAt), 
            "UpdatedAt %v should be after original %v", newUpdatedAt, originalUpdatedAt)
    })

    t.Run("should return error when instrument not found", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testInstrument := CreateTestVitalInstrument(t, nil)
        testInstrument.ID = 999 // Non-existent ID

        err := viWriter.Update(ctx, testInstrument)
        assert.ErrorIs(t, err, vitalinstrument.ErrVitalInstrumentNotFound)
    })

    t.Run("should not update soft deleted instrument", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create and delete instrument
        testInstrument := CreateTestVitalInstrument(t, nil)
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)
        
        err = viWriter.Delete(ctx, testInstrument.ID)
        require.NoError(t, err)

        // Try to update deleted instrument
        testInstrument.Name = "Should Not Update"
        err = viWriter.Update(ctx, testInstrument)
        assert.ErrorIs(t, err, vitalinstrument.ErrVitalInstrumentNotFound)
    })
}

func TestVitalInstrumentWriter_Delete(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    viWriter := NewVitalInstrumentWriter(db)
    viReader := NewVitalInstrumentReader(db)
    ctx := context.Background()

    t.Run("should soft delete vital instrument successfully", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create instrument
        testInstrument := CreateTestVitalInstrument(t, nil)
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)

        // Delete instrument
        err = viWriter.Delete(ctx, testInstrument.ID)
        require.NoError(t, err)

        // Verify instrument is not found (soft deleted)
        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        assert.Nil(t, foundInstrument)
        assert.ErrorIs(t, err, vitalinstrument.ErrVitalInstrumentNotFound)
    })

    t.Run("should return error when instrument not found", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        err := viWriter.Delete(ctx, 999)
        assert.ErrorIs(t, err, vitalinstrument.ErrVitalInstrumentNotFound)
    })

    t.Run("should not delete already deleted instrument", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create and delete instrument
        testInstrument := CreateTestVitalInstrument(t, nil)
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)
        
        err = viWriter.Delete(ctx, testInstrument.ID)
        require.NoError(t, err)

        // Try to delete again
        err = viWriter.Delete(ctx, testInstrument.ID)
        assert.ErrorIs(t, err, vitalinstrument.ErrVitalInstrumentNotFound)
    })
}

func TestVitalInstrumentWriter_Restore(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    viWriter := NewVitalInstrumentWriter(db)
    viReader := NewVitalInstrumentReader(db)
    ctx := context.Background()

    t.Run("should restore soft deleted vital instrument", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create and delete instrument
        testInstrument := CreateTestVitalInstrument(t, nil)
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)
        
        err = viWriter.Delete(ctx, testInstrument.ID)
        require.NoError(t, err)

        // Restore instrument
        err = viWriter.Restore(ctx, testInstrument.ID)
        require.NoError(t, err)

        // Verify instrument is accessible again
        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        require.NoError(t, err)
        assert.Equal(t, testInstrument.Name, foundInstrument.Name)
        assert.Nil(t, foundInstrument.DeletedAt)
    })

    t.Run("should return error when instrument not found", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        err := viWriter.Restore(ctx, 999)
        assert.ErrorIs(t, err, vitalinstrument.ErrVitalInstrumentNotFound)
    })

    t.Run("should return error when instrument not deleted", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create active instrument
        testInstrument := CreateTestVitalInstrument(t, nil)
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)

        // Try to restore non-deleted instrument
        err = viWriter.Restore(ctx, testInstrument.ID)
        assert.ErrorIs(t, err, vitalinstrument.ErrVitalInstrumentNotFound)
    })
}

func TestVitalInstrumentWriter_DeletePermanently(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    viWriter := NewVitalInstrumentWriter(db)
    viReader := NewVitalInstrumentReader(db)
    ctx := context.Background()

    t.Run("should permanently delete vital instrument", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create instrument
        testInstrument := CreateTestVitalInstrument(t, nil)
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)

        // Permanently delete instrument
        err = viWriter.DeletePermanently(ctx, testInstrument.ID)
        require.NoError(t, err)

        // Verify instrument is completely gone
        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        assert.Nil(t, foundInstrument)
        assert.ErrorIs(t, err, vitalinstrument.ErrVitalInstrumentNotFound)

        // Verify with include deleted filter
        filter := vitalinstrument.VitalInstrumentFilter{
            Page:           1,
            PageSize:       10,
            IncludeDeleted: true,
        }
        instruments, total, err := viReader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 0, total)
        assert.Len(t, instruments, 0)
    })

    t.Run("should permanently delete soft deleted instrument", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create and soft delete instrument
        testInstrument := CreateTestVitalInstrument(t, nil)
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)
        
        err = viWriter.Delete(ctx, testInstrument.ID)
        require.NoError(t, err)

        // Permanently delete
        err = viWriter.DeletePermanently(ctx, testInstrument.ID)
        require.NoError(t, err)

        // Verify instrument is completely gone even with include deleted
        filter := vitalinstrument.VitalInstrumentFilter{
            Page:           1,
            PageSize:       10,
            IncludeDeleted: true,
        }
        instruments, total, err := viReader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 0, total)
        assert.Len(t, instruments, 0)
    })

    t.Run("should return error when instrument not found", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        err := viWriter.DeletePermanently(ctx, 999)
        assert.ErrorIs(t, err, vitalinstrument.ErrVitalInstrumentNotFound)
    })
}
