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

func TestVitalInstrumentWriter_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		writer := NewVitalInstrumentWriter(db)
		
		// Test data
		characterID := uint(1)
		now := time.Now()
		vi := &vitalinstrument.VitalInstrument{
			Name:            "3R",
			MainImage:       "/images/3r.jpg",
			Description:     "A vital instrument",
			Powers:          "Create vital instruments",
			CharacterID:     &characterID,
			FirstAppearance: 1,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		
		// Setup expectations
		mock.ExpectQuery("^INSERT INTO vital_instruments").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		
		// Execute
		ctx := context.Background()
		err = writer.Create(ctx, vi)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, uint(1), vi.ID) // ID should be updated
		
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
		writer := NewVitalInstrumentWriter(db)
		
		// Test data
		characterID := uint(1)
		now := time.Now()
		vi := &vitalinstrument.VitalInstrument{
			Name:            "3R",
			MainImage:       "/images/3r.jpg",
			Description:     "A vital instrument",
			Powers:          "Create vital instruments",
			CharacterID:     &characterID,
			FirstAppearance: 1,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		
		// Setup expectations
		mock.ExpectQuery("^INSERT INTO vital_instruments").
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		err = writer.Create(ctx, vi)
		
		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insert vital instrument")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestVitalInstrumentWriter_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		writer := NewVitalInstrumentWriter(db)
		
		// Test data
		characterID := uint(1)
		now := time.Now()
		vi := &vitalinstrument.VitalInstrument{
			ID:              1,
			Name:            "Updated Name",
			Description:     "Updated description",
			CharacterID:     &characterID,
			FirstAppearance: 2,
			UpdatedAt:       now,
		}
		
		// Setup expectations
		mock.ExpectExec("^UPDATE vital_instruments SET").
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
		
		// Execute
		ctx := context.Background()
		err = writer.Update(ctx, vi)
		
		// Assert
		assert.NoError(t, err)
		
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
		writer := NewVitalInstrumentWriter(db)
		
		// Test data
		characterID := uint(1)
		now := time.Now()
		vi := &vitalinstrument.VitalInstrument{
			ID:              999, // Non-existent ID
			Name:            "Updated Name",
			CharacterID:     &characterID,
			UpdatedAt:       now,
		}
		
		// Setup expectations
		mock.ExpectExec("^UPDATE vital_instruments SET").
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
		
		// Execute
		ctx := context.Background()
		err = writer.Update(ctx, vi)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, vitalinstrument.ErrVitalInstrumentNotFound))
		
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
		writer := NewVitalInstrumentWriter(db)
		
		// Test data
		characterID := uint(1)
		now := time.Now()
		vi := &vitalinstrument.VitalInstrument{
			ID:              1,
			Name:            "Updated Name",
			CharacterID:     &characterID,
			UpdatedAt:       now,
		}
		
		// Setup expectations
		mock.ExpectExec("^UPDATE vital_instruments SET").
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		err = writer.Update(ctx, vi)
		
		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update vital instrument")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestVitalInstrumentWriter_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		writer := NewVitalInstrumentWriter(db)
		
		// Test data
		id := uint(1)
		
		// Setup expectations
		mock.ExpectExec("^UPDATE vital_instruments SET deleted_at").
			WithArgs(sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
		
		// Execute
		ctx := context.Background()
		err = writer.Delete(ctx, id)
		
		// Assert
		assert.NoError(t, err)
		
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
		writer := NewVitalInstrumentWriter(db)
		
		// Test data
		id := uint(999) // Non-existent ID
		
		// Setup expectations
		mock.ExpectExec("^UPDATE vital_instruments SET deleted_at").
			WithArgs(sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
		
		// Execute
		ctx := context.Background()
		err = writer.Delete(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, vitalinstrument.ErrVitalInstrumentNotFound))
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestVitalInstrumentWriter_Restore(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		writer := NewVitalInstrumentWriter(db)
		
		// Test data
		id := uint(1)
		
		// Setup expectations
		mock.ExpectExec("^UPDATE vital_instruments SET deleted_at = NULL").
			WithArgs(sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
		
		// Execute
		ctx := context.Background()
		err = writer.Restore(ctx, id)
		
		// Assert
		assert.NoError(t, err)
		
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
		writer := NewVitalInstrumentWriter(db)
		
		// Test data
		id := uint(999) // Non-existent ID
		
		// Setup expectations
		mock.ExpectExec("^UPDATE vital_instruments SET deleted_at = NULL").
			WithArgs(sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
		
		// Execute
		ctx := context.Background()
		err = writer.Restore(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, vitalinstrument.ErrVitalInstrumentNotFound))
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestVitalInstrumentWriter_DeletePermanently(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		writer := NewVitalInstrumentWriter(db)
		
		// Test data
		id := uint(1)
		
		// Setup expectations
		mock.ExpectExec("^DELETE FROM vital_instruments WHERE id =").
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
		
		// Execute
		ctx := context.Background()
		err = writer.DeletePermanently(ctx, id)
		
		// Assert
		assert.NoError(t, err)
		
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
		writer := NewVitalInstrumentWriter(db)
		
		// Test data
		id := uint(999) // Non-existent ID
		
		// Setup expectations
		mock.ExpectExec("^DELETE FROM vital_instruments WHERE id =").
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
		
		// Execute
		ctx := context.Background()
		err = writer.DeletePermanently(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, vitalinstrument.ErrVitalInstrumentNotFound))
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}
