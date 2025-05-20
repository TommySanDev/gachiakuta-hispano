package postgres

import (
    "context"
    "fmt"
    "time"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/character"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements character.Writer using PostgreSQL
type CharacterWriter struct {
    db *sqlx.DB
}

func NewCharacterWriter(db *sqlx.DB) *CharacterWriter {
    return &CharacterWriter{
        db: db,
    }
}

func (w *CharacterWriter) Create(ctx context.Context, c *character.Character) error {
    log := logger.GetLogger(zap.String("repository", "CharacterWriter"), zap.String("method", "Create"))
    
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

    // Using NamedQuery for cleaner parameter binding
    rows, err := w.db.NamedQueryContext(ctx, query, c)
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

func (w *CharacterWriter) Update(ctx context.Context, c *character.Character) error {
    log := logger.GetLogger(zap.String("repository", "CharacterWriter"), zap.String("method", "Update"))
    
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

    result, err := w.db.NamedExecContext(ctx, query, c)
    if err != nil {
        log.Error("Error updating character", zap.Error(err), zap.Uint("id", c.ID))
        return fmt.Errorf("update character: %w", err)
    }

    // Check if any row was affected
    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No character found to update", zap.Uint("id", c.ID))
        return character.ErrCharacterNotFound
    }

    return nil
}

func (w *CharacterWriter) Delete(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "CharacterWriter"), zap.String("method", "Delete"))
    
    query := `
        UPDATE characters
        SET deleted_at = :deleted_at
        WHERE id = :id AND deleted_at IS NULL
    `

    params := map[string]interface{}{
        "id":         id,
        "deleted_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
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
        return character.ErrCharacterNotFound
    }

    return nil
}

func (w *CharacterWriter) DeletePermanently(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "CharacterWriter"), zap.String("method", "DeletePermanently"))
    
    query := `DELETE FROM characters WHERE id = $1`

    result, err := w.db.ExecContext(ctx, query, id)
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
        return character.ErrCharacterNotFound
    }

    return nil
}

func (w *CharacterWriter) Restore(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "CharacterWriter"), zap.String("method", "Restore"))
    
    query := `
        UPDATE characters
        SET deleted_at = NULL, updated_at = :updated_at
        WHERE id = :id AND deleted_at IS NOT NULL
    `

    params := map[string]interface{}{
        "id":         id,
        "updated_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
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
        return character.ErrCharacterNotFound
    }

    return nil
}
