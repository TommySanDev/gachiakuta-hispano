package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/TommySanDev/gachiakuta-hispano/internal/vitalinstrument"
)

func TestVitalInstrumentReader_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewVitalInstrumentReader(db)
		
		// Test data
		id := uint(1)
		characterID := uint(2)
		expectedInstrument := &vitalinstrument.VitalInstrument{
			ID:              id,
			Name:            "3R",
			MainImage:       "/images/3r.jpg",
			Description:     "A vital instrument",
			Powers:          "Create vital instruments",
			CharacterID:     &characterID,
			FirstAppearance: 1,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		
		// Setup expectations
		rows := sqlmock.NewRows([]string{
			"id", "name", "main_image", "description", "powers", "character_id", 
			"first_appearance", "created_at", "updated_at", "deleted_at",
		}).
			AddRow(
				expectedInstrument.ID, expectedInstrument.Name, expectedInstrument.MainImage,
				expectedInstrument.Description, expectedInstrument.Powers, expectedInstrument.CharacterID,
				expectedInstrument.FirstAppearance, expectedInstrument.CreatedAt, expectedInstrument.UpdatedAt, nil,
			)
		
		mock.ExpectQuery("^SELECT (.+) FROM vital_instruments WHERE").
			WithArgs(id).
			WillReturnRows(rows)
		
		// Execute
		ctx := context.Background()
		result, err := reader.GetByID(ctx, id)
		
		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedInstrument.ID, result.ID)
		assert.Equal(t, expectedInstrument.Name, result.Name)
		assert.Equal(t, *expectedInstrument.CharacterID, *result.CharacterID)
		
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
		reader := NewVitalInstrumentReader(db)
		
		// Test data
		id := uint(999)
		
		// Setup expectations - empty result set
		mock.ExpectQuery("^SELECT (.+) FROM vital_instruments WHERE").
			WithArgs(id).
			WillReturnError(errors.New("sql: no rows in result set"))
		
		// Execute
		ctx := context.Background()
		result, err := reader.GetByID(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, vitalinstrument.ErrVitalInstrumentNotFound))
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
		reader := NewVitalInstrumentReader(db)
		
		// Test data
		id := uint(1)
		
		// Setup expectations - generic database error
		mock.ExpectQuery("^SELECT (.+) FROM vital_instruments WHERE").
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

func TestVitalInstrumentReader_List(t *testing.T) {
	t.Run("success_with_no_filters", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewVitalInstrumentReader(db)
		
		// Test data
		filter := vitalinstrument.VitalInstrumentFilter{
			Page:     1,
			PageSize: 10,
		}
		
		now := time.Now()
		characterID1 := uint(1)
		characterID2 := uint(2)
		instruments := []*vitalinstrument.VitalInstrument{
			{ID: 1, Name: "3R", CharacterID: &characterID1, CreatedAt: now, UpdatedAt: now},
			{ID: 2, Name: "Another Instrument", CharacterID: &characterID2, CreatedAt: now, UpdatedAt: now},
		}
		
		// Setup expectations - count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery("^SELECT COUNT\\(\\*\\) FROM vital_instruments WHERE").
			WillReturnRows(countRows)
		
		// Setup expectations - main query
		rows := sqlmock.NewRows([]string{
			"id", "name", "main_image", "description", "powers", "character_id", 
			"first_appearance", "created_at", "updated_at", "deleted_at",
		})
		
		for _, vi := range instruments {
			rows.AddRow(
				vi.ID, vi.Name, vi.MainImage, vi.Description, vi.Powers, vi.CharacterID,
				vi.FirstAppearance, vi.CreatedAt, vi.UpdatedAt, vi.DeletedAt,
			)
		}
		
		mock.ExpectQuery("^SELECT (.+) FROM vital_instruments WHERE").
			WillReturnRows(rows)
		
		// Execute
		ctx := context.Background()
		results, total, err := reader.List(ctx, filter)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, results, 2)
		assert.Equal(t, uint(1), results[0].ID)
		assert.Equal(t, "3R", results[0].Name)
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
	
	t.Run("success_with_character_id_filter", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewVitalInstrumentReader(db)
		
		// Test data
		characterID := uint(1)
		filter := vitalinstrument.VitalInstrumentFilter{
			CharacterID: &characterID,
			Page:        1,
			PageSize:    10,
		}
		
		now := time.Now()
		instruments := []*vitalinstrument.VitalInstrument{
			{ID: 1, Name: "3R", CharacterID: &characterID, CreatedAt: now, UpdatedAt: now},
		}
		
		// Setup expectations - count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("^SELECT COUNT\\(\\*\\) FROM vital_instruments WHERE").
			WillReturnRows(countRows)
		
		// Setup expectations - main query
		rows := sqlmock.NewRows([]string{
			"id", "name", "main_image", "description", "powers", "character_id", 
			"first_appearance", "created_at", "updated_at", "deleted_at",
		})
		
		for _, vi := range instruments {
			rows.AddRow(
				vi.ID, vi.Name, vi.MainImage, vi.Description, vi.Powers, vi.CharacterID,
				vi.FirstAppearance, vi.CreatedAt, vi.UpdatedAt, vi.DeletedAt,
			)
		}
		
		mock.ExpectQuery("^SELECT (.+) FROM vital_instruments WHERE").
			WillReturnRows(rows)
		
		// Execute
		ctx := context.Background()
		results, total, err := reader.List(ctx, filter)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, results, 1)
		assert.Equal(t, "3R", results[0].Name)
		assert.Equal(t, characterID, *results[0].CharacterID)
		
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
		reader := NewVitalInstrumentReader(db)
		
		// Test data
		filter := vitalinstrument.VitalInstrumentFilter{
			Page:     1,
			PageSize: 10,
		}
		
		// Setup expectations - error on count query
		mock.ExpectQuery("^SELECT COUNT\\(\\*\\) FROM vital_instruments WHERE").
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		results, total, err := reader.List(ctx, filter)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, 0, total)
		assert.Contains(t, err.Error(), "count vital instruments")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestVitalInstrumentReader_ListByCharacter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		reader := NewVitalInstrumentReader(db)
		
		// Test data
		characterID := uint(1)
		limit := 10
		
		now := time.Now()
		instruments := []*vitalinstrument.VitalInstrument{
			{ID: 1, Name: "3R", CharacterID: &characterID, CreatedAt: now, UpdatedAt: now},
			{ID: 3, Name: "Instrument 3", CharacterID: &characterID, CreatedAt: now, UpdatedAt: now},
		}
		
		// Setup expectations
		rows := sqlmock.NewRows([]string{
			"id", "name", "main_image", "description", "powers", "character_id", 
			"first_appearance", "created_at", "updated_at", "deleted_at",
		})
		
		for _, vi := range instruments {
			rows.AddRow(
				vi.ID, vi.Name, vi.MainImage, vi.Description, vi.Powers, vi.CharacterID,
				vi.FirstAppearance, vi.CreatedAt, vi.UpdatedAt, vi.DeletedAt,
			)
		}
		
		mock.ExpectQuery("^SELECT (.+) FROM vital_instruments WHERE character_id =").
			WithArgs(characterID, limit).
			WillReturnRows(rows)
		
		// Execute
		ctx := context.Background()
		results, err := reader.ListByCharacter(ctx, characterID, limit)
		
		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, characterID, *results[0].CharacterID)
		
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
		reader := NewVitalInstrumentReader(db)
		
		// Test data
		characterID := uint(1)
		limit := 10
		
		// Setup expectations
		mock.ExpectQuery("^SELECT (.+) FROM vital_instruments WHERE character_id =").
			WithArgs(characterID, limit).
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		results, err := reader.ListByCharacter(ctx, characterID, limit)
		
		// Assert
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "query vital instruments by character")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}
