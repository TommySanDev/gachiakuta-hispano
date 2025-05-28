package character

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// PostgresRepository implements Repository interface for PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewRepository creates a new character repository with PostgreSQL
func NewRepository(db *sqlx.DB) Repository {
	return &PostgresRepository{db: db}
}

// GetByID retrieves a character by its ID
func (r *PostgresRepository) GetByID(ctx context.Context, id uint) (*Character, error) {
	log := logger.GetLogger(zap.String("repository", "Character"), zap.String("method", "GetByID"))
	
	query := `
		SELECT id, name, name_japanese, main_image, description, species, gender, age,
		       height, status, affiliation, occupation, birth_date, birth_place,
		       relatives, first_appearance, created_at, updated_at, deleted_at
		FROM characters
		WHERE id = $1 AND deleted_at IS NULL
	`

	var c Character
	err := r.db.GetContext(ctx, &c, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Debug("Character not found", zap.Uint("id", id))
			return nil, ErrCharacterNotFound
		}
		log.Error("Database error", zap.Error(err))
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &c, nil
}

// Create inserts a new character
func (r *PostgresRepository) Create(ctx context.Context, c *Character) error {
	log := logger.GetLogger(zap.String("repository", "Character"), zap.String("method", "Create"))
	
	query := `
		INSERT INTO characters (
			name, name_japanese, main_image, description, species, gender, age,
			height, status, affiliation, occupation, birth_date, birth_place,
			relatives, first_appearance, created_at, updated_at
		) VALUES (
			:name, :name_japanese, :main_image, :description, :species, :gender, :age,
			:height, :status, :affiliation, :occupation, :birth_date, :birth_place,
			:relatives, :first_appearance, :created_at, :updated_at
		)
		RETURNING id
	`

	rows, err := r.db.NamedQueryContext(ctx, query, c)
	if err != nil {
		log.Error("Error creating character", zap.Error(err), zap.String("name", c.Name))
		return fmt.Errorf("insert character: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&c.ID)
		if err != nil {
			log.Error("Error scanning ID", zap.Error(err))
			return fmt.Errorf("scan ID: %w", err)
		}
	}
	
	return nil
}

// Update modifies an existing character
func (r *PostgresRepository) Update(ctx context.Context, c *Character) error {
	log := logger.GetLogger(zap.String("repository", "Character"), zap.String("method", "Update"))
	
	query := `
		UPDATE characters
		SET name = :name, 
			name_japanese = :name_japanese,
			main_image = :main_image,
			description = :description,
			species = :species,
			gender = :gender,
			age = :age,
			height = :height,
			status = :status,
			affiliation = :affiliation,
			occupation = :occupation,
			birth_date = :birth_date,
			birth_place = :birth_place,
			relatives = :relatives,
			first_appearance = :first_appearance,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	result, err := r.db.NamedExecContext(ctx, query, c)
	if err != nil {
		log.Error("Error updating character", zap.Error(err), zap.Uint("id", c.ID))
		return fmt.Errorf("update character: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Error("Error getting rows affected", zap.Error(err))
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		log.Warn("No character found to update", zap.Uint("id", c.ID))
		return ErrCharacterNotFound
	}

	return nil
}

// Delete performs a soft delete on a character
func (r *PostgresRepository) Delete(ctx context.Context, id uint) error {
	log := logger.GetLogger(zap.String("repository", "Character"), zap.String("method", "Delete"))
	
	query := `
		UPDATE characters
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		log.Error("Error deleting character", zap.Error(err), zap.Uint("id", id))
		return fmt.Errorf("delete character: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Error("Error getting rows affected", zap.Error(err))
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		log.Warn("No character found to delete", zap.Uint("id", id))
		return ErrCharacterNotFound
	}

	return nil
}

// Restore undoes a soft delete
func (r *PostgresRepository) Restore(ctx context.Context, id uint) error {
	log := logger.GetLogger(zap.String("repository", "Character"), zap.String("method", "Restore"))
	
	query := `
		UPDATE characters
		SET deleted_at = NULL, updated_at = $1
		WHERE id = $2 AND deleted_at IS NOT NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		log.Error("Error restoring character", zap.Error(err), zap.Uint("id", id))
		return fmt.Errorf("restore character: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Error("Error getting rows affected", zap.Error(err))
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		log.Warn("No character found to restore", zap.Uint("id", id))
		return ErrCharacterNotFound
	}

	return nil
}

// DeletePermanently completely removes a character from the database
func (r *PostgresRepository) DeletePermanently(ctx context.Context, id uint) error {
	log := logger.GetLogger(zap.String("repository", "Character"), zap.String("method", "DeletePermanently"))
	
	query := `DELETE FROM characters WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Error("Error permanently deleting character", zap.Error(err), zap.Uint("id", id))
		return fmt.Errorf("delete character permanently: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Error("Error getting rows affected", zap.Error(err))
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		log.Warn("No character found to delete permanently", zap.Uint("id", id))
		return ErrCharacterNotFound
	}

	return nil
}
