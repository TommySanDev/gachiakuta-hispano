package postgres

import (
    "context"
    "fmt"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/TommySanDev/gachiakuta-hispano/internal/character"
)

func TestCharacterReader_GetByID(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    reader := NewCharacterReader(db)
    writer := NewCharacterWriter(db)
    ctx := context.Background()

    t.Run("should return character when found", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testCharacter := CreateTestCharacter(t, &character.Character{
            Name:   "Rudo Surebrec",
            Status: "Alive",
        })
        
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)
        require.NotZero(t, testCharacter.ID)

        foundCharacter, err := reader.GetByID(ctx, testCharacter.ID)
        require.NoError(t, err)
        assert.Equal(t, testCharacter.Name, foundCharacter.Name)
        assert.Equal(t, testCharacter.Status, foundCharacter.Status)
        assert.Equal(t, testCharacter.ID, foundCharacter.ID)
    })

    t.Run("should return error when character not found", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        foundCharacter, err := reader.GetByID(ctx, 999)
        assert.Nil(t, foundCharacter)
        assert.ErrorIs(t, err, character.ErrCharacterNotFound)
    })

    t.Run("should not return soft deleted character", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testCharacter := CreateTestCharacter(t)
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)
        
        err = writer.Delete(ctx, testCharacter.ID)
        require.NoError(t, err)

        foundCharacter, err := reader.GetByID(ctx, testCharacter.ID)
        assert.Nil(t, foundCharacter)
        assert.ErrorIs(t, err, character.ErrCharacterNotFound)
    })
}

func TestCharacterReader_List(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    reader := NewCharacterReader(db)
    writer := NewCharacterWriter(db)
    ctx := context.Background()

    t.Run("should return paginated characters", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        characters := []*character.Character{
            CreateTestCharacter(t, &character.Character{Name: "Rudo", Status: "Alive"}),
            CreateTestCharacter(t, &character.Character{Name: "Engine", Status: "Unknown"}),
        }
        
        for _, char := range characters {
            err := writer.Create(ctx, char)
            require.NoError(t, err)
        }

        filter := character.CharacterFilter{
            Page:     1,
            PageSize: 10,
            SortBy:   "name",
            SortDir:  "asc",
        }

        foundCharacters, total, err := reader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 2, total)
        assert.Len(t, foundCharacters, 2)
        assert.Equal(t, "Engine", foundCharacters[0].Name)
        assert.Equal(t, "Rudo", foundCharacters[1].Name)
    })

    t.Run("should filter by search term", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        characters := []*character.Character{
            CreateTestCharacter(t, &character.Character{Name: "Rudo Surebrec", Description: "Main character"}),
            CreateTestCharacter(t, &character.Character{Name: "Engine", Description: "Side character"}),
        }
        
        for _, char := range characters {
            err := writer.Create(ctx, char)
            require.NoError(t, err)
        }

        filter := character.CharacterFilter{
            Search:   "Rudo",
            Page:     1,
            PageSize: 10,
        }

        foundCharacters, total, err := reader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 1, total)
        assert.Len(t, foundCharacters, 1)
        assert.Equal(t, "Rudo Surebrec", foundCharacters[0].Name)
    })

    t.Run("should filter by status", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        characters := []*character.Character{
            CreateTestCharacter(t, &character.Character{Name: "Alive Character", Status: "Alive"}),
            CreateTestCharacter(t, &character.Character{Name: "Dead Character", Status: "Dead"}),
        }
        
        for _, char := range characters {
            err := writer.Create(ctx, char)
            require.NoError(t, err)
        }

        filter := character.CharacterFilter{
            Status:   "Alive",
            Page:     1,
            PageSize: 10,
        }

        foundCharacters, total, err := reader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 1, total)
        assert.Len(t, foundCharacters, 1)
        assert.Equal(t, "Alive", foundCharacters[0].Status)
    })

    t.Run("should filter by affiliation", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        characters := []*character.Character{
            CreateTestCharacter(t, &character.Character{Name: "Cleaner 1", Affiliation: "Limpiadores"}),
            CreateTestCharacter(t, &character.Character{Name: "Cleaner 2", Affiliation: "Limpiadores"}),
            CreateTestCharacter(t, &character.Character{Name: "Other", Affiliation: "Other Group"}),
        }
        
        for _, char := range characters {
            err := writer.Create(ctx, char)
            require.NoError(t, err)
        }

        filter := character.CharacterFilter{
            Affiliation: "Limpiadores",
            Page:        1,
            PageSize:    10,
        }

        foundCharacters, total, err := reader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 2, total)
        assert.Len(t, foundCharacters, 2)
    })

    t.Run("should not include deleted characters by default", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        testCharacter := CreateTestCharacter(t)
        err := writer.Create(ctx, testCharacter)
        require.NoError(t, err)
        
        err = writer.Delete(ctx, testCharacter.ID)
        require.NoError(t, err)

        filter := character.CharacterFilter{
            Page:     1,
            PageSize: 10,
        }

        foundCharacters, total, err := reader.List(ctx, filter)
        require.NoError(t, err)
        assert.Equal(t, 0, total)
        assert.Len(t, foundCharacters, 0)
    })
}

func TestCharacterReader_ListByAffiliation(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    reader := NewCharacterReader(db)
    writer := NewCharacterWriter(db)
    ctx := context.Background()

    t.Run("should return characters with specified affiliation", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        characters := []*character.Character{
            CreateTestCharacter(t, &character.Character{Name: "Cleaner 1", Affiliation: "Limpiadores"}),
            CreateTestCharacter(t, &character.Character{Name: "Cleaner 2", Affiliation: "Limpiadores"}),
            CreateTestCharacter(t, &character.Character{Name: "Other", Affiliation: "Other Group"}),
        }
        
        for _, char := range characters {
            err := writer.Create(ctx, char)
            require.NoError(t, err)
        }

        foundCharacters, err := reader.ListByAffiliation(ctx, "Limpiadores", 10)
        require.NoError(t, err)
        assert.Len(t, foundCharacters, 2)
        
        for _, char := range foundCharacters {
            assert.Contains(t, char.Affiliation, "Limpiadores")
        }
    })

    t.Run("should respect limit parameter", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        for i := 0; i < 5; i++ {
            char := CreateTestCharacter(t, &character.Character{
                Name:        fmt.Sprintf("Cleaner %d", i),
                Affiliation: "Limpiadores",
            })
            err := writer.Create(ctx, char)
            require.NoError(t, err)
        }

        foundCharacters, err := reader.ListByAffiliation(ctx, "Limpiadores", 3)
        require.NoError(t, err)
        assert.Len(t, foundCharacters, 3)
    })
}

func TestCharacterReader_ListByStatus(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    reader := NewCharacterReader(db)
    writer := NewCharacterWriter(db)
    ctx := context.Background()

    t.Run("should return characters with specified status", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        characters := []*character.Character{
            CreateTestCharacter(t, &character.Character{Name: "Alive 1", Status: "Alive"}),
            CreateTestCharacter(t, &character.Character{Name: "Alive 2", Status: "Alive"}),
            CreateTestCharacter(t, &character.Character{Name: "Dead", Status: "Dead"}),
        }
        
        for _, char := range characters {
            err := writer.Create(ctx, char)
            require.NoError(t, err)
        }

        foundCharacters, err := reader.ListByStatus(ctx, "Alive", 10)
        require.NoError(t, err)
        assert.Len(t, foundCharacters, 2)
        
        for _, char := range foundCharacters {
            assert.Equal(t, "Alive", char.Status)
        }
    })
}

func TestCharacterReader_ListBySpecies(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)

    reader := NewCharacterReader(db)
    writer := NewCharacterWriter(db)
    ctx := context.Background()

    t.Run("should return characters with specified species", func(t *testing.T) {
        CleanupTestDB(t, db)
        
        characters := []*character.Character{
            CreateTestCharacter(t, &character.Character{Name: "Human 1", Species: "Human"}),
            CreateTestCharacter(t, &character.Character{Name: "Human 2", Species: "Human"}),
            CreateTestCharacter(t, &character.Character{Name: "Other", Species: "Other Species"}),
        }
        
        for _, char := range characters {
            err := writer.Create(ctx, char)
            require.NoError(t, err)
        }

        foundCharacters, err := reader.ListBySpecies(ctx, "Human", 10)
        require.NoError(t, err)
        assert.Len(t, foundCharacters, 2)
        
        for _, char := range foundCharacters {
            assert.Equal(t, "Human", char.Species)
        }
    })
}
