package postgres

import (
    "context"
    "fmt"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/TommySanDev/gachiakuta-hispano/internal/character"
    "github.com/TommySanDev/gachiakuta-hispano/internal/vitalinstrument"
)

func TestVitalInstrumentReader_GetByID(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    viReader := NewVitalInstrumentReader(db)
    viWriter := NewVitalInstrumentWriter(db)
    charWriter := NewCharacterWriter(db)
    ctx := context.Background()

    t.Run("should return vital instrument when found", func(t *testing.T) {
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
            Description: "A powerful vital instrument",
        })
        
        err = viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)
        require.NotZero(t, testInstrument.ID)

        // Get vital instrument by ID
        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        require.NoError(t, err)
        assert.Equal(t, testInstrument.Name, foundInstrument.Name)
        assert.Equal(t, testInstrument.Description, foundInstrument.Description)
        assert.Equal(t, testInstrument.ID, foundInstrument.ID)
        assert.Equal(t, testCharacter.ID, *foundInstrument.CharacterID)
    })

    t.Run("should return vital instrument without character", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create vital instrument without character
        testInstrument := CreateTestVitalInstrument(t, nil, &vitalinstrument.VitalInstrument{
            Name: "Unowned Instrument",
        })
        
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)

        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        require.NoError(t, err)
        assert.Equal(t, testInstrument.Name, foundInstrument.Name)
        assert.Nil(t, foundInstrument.CharacterID)
    })

    t.Run("should return error when vital instrument not found", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        foundInstrument, err := viReader.GetByID(ctx, 999)
        assert.Nil(t, foundInstrument)
        assert.ErrorIs(t, err, vitalinstrument.ErrVitalInstrumentNotFound)
    })

    t.Run("should not return soft deleted vital instrument", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testInstrument := CreateTestVitalInstrument(t, nil)
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)
        
        err = viWriter.Delete(ctx, testInstrument.ID)
        require.NoError(t, err)

        foundInstrument, err := viReader.GetByID(ctx, testInstrument.ID)
        assert.Nil(t, foundInstrument)
        assert.ErrorIs(t, err, vitalinstrument.ErrVitalInstrumentNotFound)
    })
}

func TestVitalInstrumentReader_List(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    viReader := NewVitalInstrumentReader(db)
    viWriter := NewVitalInstrumentWriter(db)
    charWriter := NewCharacterWriter(db)
    ctx := context.Background()

    t.Run("should return paginated vital instruments", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create character
        testCharacter := CreateTestCharacter(t)
        err := charWriter.Create(ctx, testCharacter)
        require.NoError(t, err)

        // Create vital instruments
        instruments := []*vitalinstrument.VitalInstrument{
            CreateTestVitalInstrument(t, &testCharacter.ID, &vitalinstrument.VitalInstrument{Name: "3R"}),
            CreateTestVitalInstrument(t, nil, &vitalinstrument.VitalInstrument{Name: "Another Instrument"}),
        }
        
        for _, instrument := range instruments {
            err := viWriter.Create(ctx, instrument)
            require.NoError(t, err)
        }

        filter := vitalinstrument.VitalInstrumentFilter{
            Page:     1,
            PageSize: 10,
            SortBy:   "name",
            SortDir:  "asc",
        }

        foundInstruments, total, err := viReader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 2, total)
        assert.Len(t, foundInstruments, 2)
        assert.Equal(t, "3R", foundInstruments[0].Name)
        assert.Equal(t, "Another Instrument", foundInstruments[1].Name)
    })

    t.Run("should filter by search term", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        instruments := []*vitalinstrument.VitalInstrument{
            CreateTestVitalInstrument(t, nil, &vitalinstrument.VitalInstrument{
                Name:        "3R Instrument",
                Description: "A powerful weapon",
            }),
            CreateTestVitalInstrument(t, nil, &vitalinstrument.VitalInstrument{
                Name:        "Other Tool",
                Description: "Different tool",
            }),
        }
        
        for _, instrument := range instruments {
            err := viWriter.Create(ctx, instrument)
            require.NoError(t, err)
        }

        filter := vitalinstrument.VitalInstrumentFilter{
            Search:   "3R",
            Page:     1,
            PageSize: 10,
        }

        foundInstruments, total, err := viReader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 1, total)
        assert.Len(t, foundInstruments, 1)
        assert.Equal(t, "3R Instrument", foundInstruments[0].Name)
    })

    t.Run("should filter by character ID", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create characters
        char1 := CreateTestCharacter(t, &character.Character{Name: "Character 1"})
        char2 := CreateTestCharacter(t, &character.Character{Name: "Character 2"})
        err := charWriter.Create(ctx, char1)
        require.NoError(t, err)
        err = charWriter.Create(ctx, char2)
        require.NoError(t, err)

        // Create instruments for different characters
        instruments := []*vitalinstrument.VitalInstrument{
            CreateTestVitalInstrument(t, &char1.ID, &vitalinstrument.VitalInstrument{Name: "Char1 Instrument"}),
            CreateTestVitalInstrument(t, &char2.ID, &vitalinstrument.VitalInstrument{Name: "Char2 Instrument"}),
            CreateTestVitalInstrument(t, nil, &vitalinstrument.VitalInstrument{Name: "No Owner"}),
        }
        
        for _, instrument := range instruments {
            err := viWriter.Create(ctx, instrument)
            require.NoError(t, err)
        }

        filter := vitalinstrument.VitalInstrumentFilter{
            CharacterID: &char1.ID,
            Page:        1,
            PageSize:    10,
        }

        foundInstruments, total, err := viReader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 1, total)
        assert.Len(t, foundInstruments, 1)
        assert.Equal(t, "Char1 Instrument", foundInstruments[0].Name)
        assert.Equal(t, char1.ID, *foundInstruments[0].CharacterID)
    })

    t.Run("should not include deleted instruments by default", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testInstrument := CreateTestVitalInstrument(t, nil)
        err := viWriter.Create(ctx, testInstrument)
        require.NoError(t, err)
        
        err = viWriter.Delete(ctx, testInstrument.ID)
        require.NoError(t, err)

        filter := vitalinstrument.VitalInstrumentFilter{
            Page:     1,
            PageSize: 10,
        }

        foundInstruments, total, err := viReader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 0, total)
        assert.Len(t, foundInstruments, 0)
    })
}

func TestVitalInstrumentReader_ListByCharacter(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    viReader := NewVitalInstrumentReader(db)
    viWriter := NewVitalInstrumentWriter(db)
    charWriter := NewCharacterWriter(db)
    ctx := context.Background()

    t.Run("should return instruments for specified character", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create characters
        char1 := CreateTestCharacter(t, &character.Character{Name: "Character 1"})
        char2 := CreateTestCharacter(t, &character.Character{Name: "Character 2"})
        err := charWriter.Create(ctx, char1)
        require.NoError(t, err)
        err = charWriter.Create(ctx, char2)
        require.NoError(t, err)

        // Create instruments
        instruments := []*vitalinstrument.VitalInstrument{
            CreateTestVitalInstrument(t, &char1.ID, &vitalinstrument.VitalInstrument{Name: "Instrument 1"}),
            CreateTestVitalInstrument(t, &char1.ID, &vitalinstrument.VitalInstrument{Name: "Instrument 2"}),
            CreateTestVitalInstrument(t, &char2.ID, &vitalinstrument.VitalInstrument{Name: "Other Instrument"}),
        }
        
        for _, instrument := range instruments {
            err := viWriter.Create(ctx, instrument)
            require.NoError(t, err)
        }

        foundInstruments, err := viReader.ListByCharacter(ctx, char1.ID, 10)
        require.NoError(t, err)
        assert.Len(t, foundInstruments, 2)
        
        for _, instrument := range foundInstruments {
            assert.Equal(t, char1.ID, *instrument.CharacterID)
        }
    })

    t.Run("should respect limit parameter", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create character
        testCharacter := CreateTestCharacter(t)
        err := charWriter.Create(ctx, testCharacter)
        require.NoError(t, err)

        // Create multiple instruments for the character
        for i := 0; i < 5; i++ {
            instrument := CreateTestVitalInstrument(t, &testCharacter.ID, &vitalinstrument.VitalInstrument{
                Name: fmt.Sprintf("Instrument %d", i),
            })
            err := viWriter.Create(ctx, instrument)
            require.NoError(t, err)
        }

        foundInstruments, err := viReader.ListByCharacter(ctx, testCharacter.ID, 3)
        require.NoError(t, err)
        assert.Len(t, foundInstruments, 3)
    })

    t.Run("should return empty list for character with no instruments", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        // Create character without instruments
        testCharacter := CreateTestCharacter(t)
        err := charWriter.Create(ctx, testCharacter)
        require.NoError(t, err)

        foundInstruments, err := viReader.ListByCharacter(ctx, testCharacter.ID, 10)
        require.NoError(t, err)
        assert.Len(t, foundInstruments, 0)
    })

    t.Run("should not return instruments for non-existent character", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        foundInstruments, err := viReader.ListByCharacter(ctx, 999, 10)
        require.NoError(t, err)
        assert.Len(t, foundInstruments, 0)
    })
}
