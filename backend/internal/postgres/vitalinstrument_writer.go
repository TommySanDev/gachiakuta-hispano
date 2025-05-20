package postgres

import (
    "context"
    "fmt"
    "time"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
    "github.com/TommySanDev/gachiakuta-hispano/internal/vitalinstrument"
)

// Implements vitalinstrument.Writer using PostgreSQL
type VitalInstrumentWriter struct {
    db *sqlx.DB
}

func NewVitalInstrumentWriter(db *sqlx.DB) *VitalInstrumentWriter {
    return &VitalInstrumentWriter{
        db: db,
    }
}

func (w *VitalInstrumentWriter) Create(ctx context.Context, vi *vitalinstrument.VitalInstrument) error {
    log := logger.GetLogger(zap.String("repository", "VitalInstrumentWriter"), zap.String("method", "Create"))
    
    query := `
        INSERT INTO vital_instruments (
            name, main_image, description, powers, character_id,
            first_appearance, created_at, updated_at
        ) VALUES (
            :name, :main_image, :description, :powers, :character_id,
            :first_appearance, :created_at, :updated_at
        )
        RETURNING id
    `

    // Using NamedQuery for cleaner parameter binding
    rows, err := w.db.NamedQueryContext(ctx, query, vi)
    if err != nil {
        log.Error("Error creating vital instrument", zap.Error(err), zap.String("name", vi.Name))
        return fmt.Errorf("insert vital instrument: %w", err)
    }
    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&vi.ID)
        if err != nil {
            log.Error("Error scanning ID", zap.Error(err))
            return fmt.Errorf("scan ID: %w", err)
        }
    }
    
    return nil
}

func (w *VitalInstrumentWriter) Update(ctx context.Context, vi *vitalinstrument.VitalInstrument) error {
    log := logger.GetLogger(zap.String("repository", "VitalInstrumentWriter"), zap.String("method", "Update"))
    
    query := `
        UPDATE vital_instruments
        SET name = :name, 
            main_image = :main_image,
            description = :description,
            powers = :powers,
            character_id = :character_id,
            first_appearance = :first_appearance,
            updated_at = :updated_at
        WHERE id = :id AND deleted_at IS NULL
    `

    result, err := w.db.NamedExecContext(ctx, query, vi)
    if err != nil {
        log.Error("Error updating vital instrument", zap.Error(err), zap.Uint("id", vi.ID))
        return fmt.Errorf("update vital instrument: %w", err)
    }

    // Check if any row was affected
    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No vital instrument found to update", zap.Uint("id", vi.ID))
        return vitalinstrument.ErrVitalInstrumentNotFound
    }

    return nil
}

func (w *VitalInstrumentWriter) Delete(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "VitalInstrumentWriter"), zap.String("method", "Delete"))
    
    query := `
        UPDATE vital_instruments
        SET deleted_at = :deleted_at
        WHERE id = :id AND deleted_at IS NULL
    `

    params := map[string]interface{}{
        "id":         id,
        "deleted_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
    if err != nil {
        log.Error("Error deleting vital instrument", zap.Error(err), zap.Uint("id", id))
        return fmt.Errorf("delete vital instrument: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No vital instrument found to delete", zap.Uint("id", id))
        return vitalinstrument.ErrVitalInstrumentNotFound
    }

    return nil
}

func (w *VitalInstrumentWriter) DeletePermanently(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "VitalInstrumentWriter"), zap.String("method", "DeletePermanently"))
    
    query := `DELETE FROM vital_instruments WHERE id = $1`

    result, err := w.db.ExecContext(ctx, query, id)
    if err != nil {
        log.Error("Error permanently deleting vital instrument", zap.Error(err), zap.Uint("id", id))
        return fmt.Errorf("delete vital instrument permanently: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No vital instrument found to delete permanently", zap.Uint("id", id))
        return vitalinstrument.ErrVitalInstrumentNotFound
    }

    return nil
}

func (w *VitalInstrumentWriter) Restore(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("repository", "VitalInstrumentWriter"), zap.String("method", "Restore"))
    
    query := `
        UPDATE vital_instruments
        SET deleted_at = NULL, updated_at = :updated_at
        WHERE id = :id AND deleted_at IS NOT NULL
    `

    params := map[string]interface{}{
        "id":         id,
        "updated_at": time.Now(),
    }

    result, err := w.db.NamedExecContext(ctx, query, params)
    if err != nil {
        log.Error("Error restoring vital instrument", zap.Error(err), zap.Uint("id", id))
        return fmt.Errorf("restore vital instrument: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        log.Error("Error getting rows affected", zap.Error(err))
        return fmt.Errorf("get rows affected: %w", err)
    }

    if rows == 0 {
        log.Warn("No vital instrument found to restore", zap.Uint("id", id))
        return vitalinstrument.ErrVitalInstrumentNotFound
    }

    return nil
}
