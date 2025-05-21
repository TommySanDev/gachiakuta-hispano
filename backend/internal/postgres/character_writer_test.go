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

func TestCharacterWriter_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		writer := NewCharacterWriter(db)
		
		// Test data
		now := time.Now()
		char := &character.Character{
			Name:            "Rudo Surebrec",
			NameJapanese:    "ルド・シュアブレック",
			MainImage:       "/images/rudo.jpg",
			Description:     "A character from Gachiakuta",
			Status:          "Alive",
			FirstAppearance: 1,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		
		// Setup expectations
		mock.ExpectQuery("^INSERT INTO characters").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		
		// Execute
		ctx := context.Background()
		err = writer.Create(ctx, char)
		
		// Assert
		assert.NoError(t, err)
		assert.Equal(t, uint(1), char.ID) // ID should be updated
		
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
		writer := NewCharacterWriter(db)
		
		// Test data
		now := time.Now()
		char := &character.Character{
			Name:            "Rudo Surebrec",
			NameJapanese:    "ルド・シュアブレック",
			MainImage:       "/images/rudo.jpg",
			Description:     "A character from Gachiakuta",
			Status:          "Alive",
			FirstAppearance: 1,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		
		// Setup expectations
		mock.ExpectQuery("^INSERT INTO characters").
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		err = writer.Create(ctx, char)
		
		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insert character")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestCharacterWriter_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		writer := NewCharacterWriter(db)
		
		// Test data
		now := time.Now()
		char := &character.Character{
			ID:              1,
			Name:            "Updated Name",
			Description:     "Updated description",
			FirstAppearance: 2,
			UpdatedAt:       now,
		}
		
		// Setup expectations
		mock.ExpectExec("^UPDATE characters SET").
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected
		
		// Execute
		ctx := context.Background()
		err = writer.Update(ctx, char)
		
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
		writer := NewCharacterWriter(db)
		
		// Test data
		now := time.Now()
		char := &character.Character{
			ID:              999, // Non-existent ID
			Name:            "Updated Name",
			UpdatedAt:       now,
		}
		
		// Setup expectations
		mock.ExpectExec("^UPDATE characters SET").
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
		
		// Execute
		ctx := context.Background()
		err = writer.Update(ctx, char)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, character.ErrCharacterNotFound))
		
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
		writer := NewCharacterWriter(db)
		
		// Test data
		now := time.Now()
		char := &character.Character{
			ID:              1,
			Name:            "Updated Name",
			UpdatedAt:       now,
		}
		
		// Setup expectations
		mock.ExpectExec("^UPDATE characters SET").
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		err = writer.Update(ctx, char)
		
		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update character")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestCharacterWriter_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		writer := NewCharacterWriter(db)
		
		// Test data
		id := uint(1)
		
		// Setup expectations
		mock.ExpectExec("^UPDATE characters SET deleted_at").
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
		writer := NewCharacterWriter(db)
		
		// Test data
		id := uint(999) // Non-existent ID
		
		// Setup expectations
		mock.ExpectExec("^UPDATE characters SET deleted_at").
			WithArgs(sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
		
		// Execute
		ctx := context.Background()
		err = writer.Delete(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, character.ErrCharacterNotFound))
		
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
		writer := NewCharacterWriter(db)
		
		// Test data
		id := uint(1)
		
		// Setup expectations
		mock.ExpectExec("^UPDATE characters SET deleted_at").
			WithArgs(sqlmock.AnyArg(), id).
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		err = writer.Delete(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete character")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestCharacterWriter_Restore(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		writer := NewCharacterWriter(db)
		
		// Test data
		id := uint(1)
		
		// Setup expectations
		mock.ExpectExec("^UPDATE characters SET deleted_at = NULL").
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
		writer := NewCharacterWriter(db)
		
		// Test data
		id := uint(999) // Non-existent ID
		
		// Setup expectations
		mock.ExpectExec("^UPDATE characters SET deleted_at = NULL").
			WithArgs(sqlmock.AnyArg(), id).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
		
		// Execute
		ctx := context.Background()
		err = writer.Restore(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, character.ErrCharacterNotFound))
		
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
		writer := NewCharacterWriter(db)
		
		// Test data
		id := uint(1)
		
		// Setup expectations
		mock.ExpectExec("^UPDATE characters SET deleted_at = NULL").
			WithArgs(sqlmock.AnyArg(), id).
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		err = writer.Restore(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "restore character")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestCharacterWriter_DeletePermanently(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Setup mock DB
		mockDB, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("Error creating mock database: %v", err)
		}
		defer mockDB.Close()
		
		db := sqlx.NewDb(mockDB, "sqlmock")
		writer := NewCharacterWriter(db)
		
		// Test data
		id := uint(1)
		
		// Setup expectations
		mock.ExpectExec("^DELETE FROM characters WHERE id =").
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
		writer := NewCharacterWriter(db)
		
		// Test data
		id := uint(999) // Non-existent ID
		
		// Setup expectations
		mock.ExpectExec("^DELETE FROM characters WHERE id =").
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
		
		// Execute
		ctx := context.Background()
		err = writer.DeletePermanently(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.True(t, errors.Is(err, character.ErrCharacterNotFound))
		
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
		writer := NewCharacterWriter(db)
		
		// Test data
		id := uint(1)
		
		// Setup expectations
		mock.ExpectExec("^DELETE FROM characters WHERE id =").
			WithArgs(id).
			WillReturnError(errors.New("database error"))
		
		// Execute
		ctx := context.Background()
		err = writer.DeletePermanently(ctx, id)
		
		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete character permanently")
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}
