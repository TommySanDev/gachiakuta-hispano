package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/TommySanDev/gachiakuta-hispano/internal/character"
)

func TestCharacterReader_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		id := uint(1)
		expectedCharacter := &character.Character{
			ID:              id,
			Name:            "Rudo Surebrec",
			NameJapanese:    "ルド・シュアブレック",
			MainImage:       "/images/rudo.jpg",
			Description:     "A character from Gachiakuta",
			Status:          "Alive",
			FirstAppearance: 1,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		
		// Setup expectations
		rows := sqlmock.NewRows([]string{
			"id", "name", "name_japanese", "main_image", "description", "species", 
			"gender", "age", "height", "status", "affiliation", "occupation", 
			"birth_date", "birth_place", "relatives", "first_appearance", 
			"created_at", "updated_at", "deleted_at",
		}).
			AddRow(
				expectedCharacter.ID, expectedCharacter.Name, expectedCharacter.NameJapanese,
				expectedCharacter.MainImage, expectedCharacter.Description, expectedCharacter.Species,
				expectedCharacter.Gender, expectedCharacter.Age, expectedCharacter.Height,
				expectedCharacter.Status, expectedCharacter.Affiliation, expectedCharacter.Occupation,
				expectedCharacter.BirthDate, expectedCharacter.BirthPlace, expectedCharacter.Relatives,
				expectedCharacter.FirstAppearance, expectedCharacter.CreatedAt, expectedCharacter.UpdatedAt, nil,
			)
		
		mock.ExpectQuery("^SELECT (.+) FROM characters WHERE").
			WithArgs(id).
			WillReturnRows(rows)
		
		// Execute
		ctx := context.Background()
		result, err := reader.GetByID(ctx, id)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedCharacter.ID, result.ID)
		assert.Equal(t, expectedCharacter.Name, result.Name)
		assert.Equal(t, expectedCharacter.Status, result.Status)
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
	
	t.Run("not_found", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		id := uint(999)
		
		// Setup expectations - empty result set
		mock.ExpectQuery("^SELECT (.+) FROM characters WHERE").
			WithArgs(id).
			WillReturnError(errors.New("sql: no rows in result set"))
		
		// Execute
		ctx := context.Background()
		result, err := reader.GetByID(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, character.ErrCharacterNotFound))
		assert.Nil(t, result)
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
	
	t.Run("database_error", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		id := uint(1)
		
		// Setup expectations - generic database error
		mock.ExpectQuery("^SELECT (.+) FROM characters WHERE").
			WithArgs(id).
			WillReturnError(errors.New("database connection failed"))
		
		// Execute
		ctx := context.Background()
		result, err := reader.GetByID(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "database error")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestCharacterReader_List(t *testing.T) {
	t.Run("success_with_no_filters", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		filter := character.CharacterFilter{
			Page:     1,
			PageSize: 10,
		}
		
		now := time.Now()
		characters := []*character.Character{
			{ID: 1, Name: "Rudo", Status: "Alive", CreatedAt: now, UpdatedAt: now},
			{ID: 2, Name: "Character 2", Status: "Unknown", CreatedAt: now, UpdatedAt: now},
		}
		
		// Setup expectations - count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery("^SELECT COUNT\\(\\*\\) FROM characters WHERE").
			WillReturnRows(countRows)
		
		// Setup expectations - main query
		rows := sqlmock.NewRows([]string{
			"id", "name", "name_japanese", "main_image", "description", "species", 
			"gender", "age", "height", "status", "affiliation", "occupation", 
			"birth_date", "birth_place", "relatives", "first_appearance", 
			"created_at", "updated_at", "deleted_at",
		})
		
		for _, c := range characters {
			rows.AddRow(
				c.ID, c.Name, c.NameJapanese, c.MainImage, c.Description, c.Species,
				c.Gender, c.Age, c.Height, c.Status, c.Affiliation, c.Occupation,
				c.BirthDate, c.BirthPlace, c.Relatives, c.FirstAppearance,
				c.CreatedAt, c.UpdatedAt, c.DeletedAt,
			)
		}
		
		mock.ExpectQuery("^SELECT (.+) FROM characters WHERE").
			WillReturnRows(rows)
		
		// Execute
		ctx := context.Background()
		results, total, err := reader.List(ctx, filter)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, results, 2)
		assert.Equal(t, uint(1), results[0].ID)
		assert.Equal(t, "Rudo", results[0].Name)
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
	
	t.Run("success_with_search_filter", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		filter := character.CharacterFilter{
			Search:   "Rudo",
			Page:     1,
			PageSize: 10,
		}
		
		now := time.Now()
		characters := []*character.Character{
			{ID: 1, Name: "Rudo", Status: "Alive", CreatedAt: now, UpdatedAt: now},
		}
		
		// Setup expectations - count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("^SELECT COUNT\\(\\*\\) FROM characters WHERE").
			WithArgs("%Rudo%", "%Rudo%", "%Rudo%").
			WillReturnRows(countRows)
		
		// Setup expectations - main query
		rows := sqlmock.NewRows([]string{
			"id", "name", "name_japanese", "main_image", "description", "species", 
			"gender", "age", "height", "status", "affiliation", "occupation", 
			"birth_date", "birth_place", "relatives", "first_appearance", 
			"created_at", "updated_at", "deleted_at",
		})
		
		for _, c := range characters {
			rows.AddRow(
				c.ID, c.Name, c.NameJapanese, c.MainImage, c.Description, c.Species,
				c.Gender, c.Age, c.Height, c.Status, c.Affiliation, c.Occupation,
				c.BirthDate, c.BirthPlace, c.Relatives, c.FirstAppearance,
				c.CreatedAt, c.UpdatedAt, c.DeletedAt,
			)
		}
		
		mock.ExpectQuery("^SELECT (.+) FROM characters WHERE").
			WithArgs("%Rudo%", "%Rudo%", "%Rudo%", 10, 0).
			WillReturnRows(rows)
		
		// Execute
		ctx := context.Background()
		results, total, err := reader.List(ctx, filter)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, results, 1)
		assert.Equal(t, "Rudo", results[0].Name)
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
	
	t.Run("count_query_error", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		filter := character.CharacterFilter{
			Page:     1,
			PageSize: 10,
		}
		
		// Setup expectations - error on count query
		mock.ExpectQuery("^SELECT COUNT\\(\\*\\) FROM characters WHERE").
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		results, total, err := reader.List(ctx, filter)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "count characters")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
	
	t.Run("main_query_error", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		filter := character.CharacterFilter{
			Page:     1,
			PageSize: 10,
		}
		
		// Setup expectations - count query succeeds
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery("^SELECT COUNT\\(\\*\\) FROM characters WHERE").
			WillReturnRows(countRows)
		
		// Setup expectations - main query fails
		mock.ExpectQuery("^SELECT (.+) FROM characters WHERE").
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		results, total, err := reader.List(ctx, filter)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "query characters")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestCharacterReader_ListByAffiliation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		affiliation := "Limpiadores"
		limit := 10
		
		now := time.Now()
		characters := []*character.Character{
			{ID: 1, Name: "Rudo", Affiliation: "Limpiadores", CreatedAt: now, UpdatedAt: now},
			{ID: 3, Name: "Character 3", Affiliation: "Limpiadores", CreatedAt: now, UpdatedAt: now},
		}
		
		// Setup expectations
		rows := sqlmock.NewRows([]string{
			"id", "name", "name_japanese", "main_image", "description", "species", 
			"gender", "age", "height", "status", "affiliation", "occupation", 
			"birth_date", "birth_place", "relatives", "first_appearance", 
			"created_at", "updated_at", "deleted_at",
		})
		
		for _, c := range characters {
			rows.AddRow(
				c.ID, c.Name, c.NameJapanese, c.MainImage, c.Description, c.Species,
				c.Gender, c.Age, c.Height, c.Status, c.Affiliation, c.Occupation,
				c.BirthDate, c.BirthPlace, c.Relatives, c.FirstAppearance,
				c.CreatedAt, c.UpdatedAt, c.DeletedAt,
			)
		}
		
		mock.ExpectQuery("^SELECT (.+) FROM characters WHERE affiliation ILIKE").
			WithArgs("%"+affiliation+"%", limit).
			WillReturnRows(rows)
		
		// Execute
		ctx := context.Background()
		results, err := reader.ListByAffiliation(ctx, affiliation, limit)
		
		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "Limpiadores", results[0].Affiliation)
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
	
	t.Run("database_error", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		affiliation := "Limpiadores"
		limit := 10
		
		// Setup expectations
		mock.ExpectQuery("^SELECT (.+) FROM characters WHERE affiliation ILIKE").
			WithArgs("%"+affiliation+"%", limit).
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		results, err := reader.ListByAffiliation(ctx, affiliation, limit)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "query characters by affiliation")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestCharacterReader_ListByStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		status := "Alive"
		limit := 10
		
		now := time.Now()
		characters := []*character.Character{
			{ID: 1, Name: "Rudo", Status: "Alive", CreatedAt: now, UpdatedAt: now},
			{ID: 3, Name: "Character 3", Status: "Alive", CreatedAt: now, UpdatedAt: now},
		}
		
		// Setup expectations
		rows := sqlmock.NewRows([]string{
			"id", "name", "name_japanese", "main_image", "description", "species", 
			"gender", "age", "height", "status", "affiliation", "occupation", 
			"birth_date", "birth_place", "relatives", "first_appearance", 
			"created_at", "updated_at", "deleted_at",
		})
		
		for _, c := range characters {
			rows.AddRow(
				c.ID, c.Name, c.NameJapanese, c.MainImage, c.Description, c.Species,
				c.Gender, c.Age, c.Height, c.Status, c.Affiliation, c.Occupation,
				c.BirthDate, c.BirthPlace, c.Relatives, c.FirstAppearance,
				c.CreatedAt, c.UpdatedAt, c.DeletedAt,
			)
		}
		
		mock.ExpectQuery("^SELECT (.+) FROM characters WHERE status =").
			WithArgs(status, limit).
			WillReturnRows(rows)
		
		// Execute
		ctx := context.Background()
		results, err := reader.ListByStatus(ctx, status, limit)
		
		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "Alive", results[0].Status)
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
	
	t.Run("database_error", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		status := "Alive"
		limit := 10
		
		// Setup expectations
		mock.ExpectQuery("^SELECT (.+) FROM characters WHERE status =").
			WithArgs(status, limit).
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		results, err := reader.ListByStatus(ctx, status, limit)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "query characters by status")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestCharacterReader_ListBySpecies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		species := "Human"
		limit := 10
		
		now := time.Now()
		characters := []*character.Character{
			{ID: 1, Name: "Rudo", Species: "Human", CreatedAt: now, UpdatedAt: now},
			{ID: 3, Name: "Character 3", Species: "Human", CreatedAt: now, UpdatedAt: now},
		}
		
		// Setup expectations
		rows := sqlmock.NewRows([]string{
			"id", "name", "name_japanese", "main_image", "description", "species", 
			"gender", "age", "height", "status", "affiliation", "occupation", 
			"birth_date", "birth_place", "relatives", "first_appearance", 
			"created_at", "updated_at", "deleted_at",
		})
		
		for _, c := range characters {
			rows.AddRow(
				c.ID, c.Name, c.NameJapanese, c.MainImage, c.Description, c.Species,
				c.Gender, c.Age, c.Height, c.Status, c.Affiliation, c.Occupation,
				c.BirthDate, c.BirthPlace, c.Relatives, c.FirstAppearance,
				c.CreatedAt, c.UpdatedAt, c.DeletedAt,
			)
		}
		
		mock.ExpectQuery("^SELECT (.+) FROM characters WHERE species =").
			WithArgs(species, limit).
			WillReturnRows(rows)
		
		// Execute
		ctx := context.Background()
		results, err := reader.ListBySpecies(ctx, species, limit)
		
		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, "Human", results[0].Species)
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
	
	t.Run("database_error", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewCharacterReader(db)
		
		// Test data
		species := "Human"
		limit := 10
		
		// Setup expectations
		mock.ExpectQuery("^SELECT (.+) FROM characters WHERE species =").
			WithArgs(species, limit).
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		results, err := reader.ListBySpecies(ctx, species, limit)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "query characters by species")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}
