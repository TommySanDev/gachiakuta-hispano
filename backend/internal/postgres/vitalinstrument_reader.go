package postgres

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
    "github.com/TommySanDev/gachiakuta-hispano/internal/vitalinstrument"
)

// Implements vitalinstrument.Reader using PostgreSQL
type VitalInstrumentReader struct {
    db *sqlx.DB
}

func NewVitalInstrumentReader(db *sqlx.DB) *VitalInstrumentReader {
    return &VitalInstrumentReader{
        db: db,
    }
}

func (r *VitalInstrumentReader) GetByID(ctx context.Context, id uint) (*vitalinstrument.VitalInstrument, error) {
    log := logger.GetLogger(zap.String("repository", "VitalInstrumentReader"), zap.String("method", "GetByID"))
    
    query := `
        SELECT id, name, main_image, description, powers, character_id, 
               first_appearance, created_at, updated_at, deleted_at
        FROM vital_instruments
        WHERE id = $1 AND deleted_at IS NULL
    `

    var vi vitalinstrument.VitalInstrument
    err := r.db.GetContext(ctx, &vi, query, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("Vital instrument not found", zap.Uint("id", id))
            return nil, vitalinstrument.ErrVitalInstrumentNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &vi, nil
}

func (r *VitalInstrumentReader) List(ctx context.Context, filter vitalinstrument.VitalInstrumentFilter) ([]*vitalinstrument.VitalInstrument, int, error) {
    log := logger.GetLogger(zap.String("repository", "VitalInstrumentReader"), zap.String("method", "List"))
    
    // Build query parts
    whereClauses := []string{"1=1"}
    args := []interface{}{}
    argPos := 1

    // Base where clause
    if !filter.IncludeDeleted {
        whereClauses = append(whereClauses, "deleted_at IS NULL")
    }
    
    // Apply search filter
    if filter.Search != "" {
        whereClauses = append(whereClauses, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d OR powers ILIKE $%d)", argPos, argPos+1, argPos+2))
        searchTerm := "%" + filter.Search + "%"
        args = append(args, searchTerm, searchTerm, searchTerm)
        argPos += 3
    }

    // Filter by character
    if filter.CharacterID != nil {
        whereClauses = append(whereClauses, fmt.Sprintf("character_id = $%d", argPos))
        args = append(args, *filter.CharacterID)
        argPos++
    }

    whereClause := "WHERE " + strings.Join(whereClauses, " AND ")

    // Count total matching records
    countQuery := "SELECT COUNT(*) FROM vital_instruments " + whereClause
    var total int
    err := r.db.GetContext(ctx, &total, countQuery, args...)
    if err != nil {
        log.Error("Error counting vital instruments", zap.Error(err))
        return nil, 0, fmt.Errorf("count vital instruments: %w", err)
   }

   // Main query with sorting and pagination
   query := `
       SELECT id, name, main_image, description, powers, character_id, 
              first_appearance, created_at, updated_at, deleted_at
       FROM vital_instruments
       ` + whereClause

   // Add sorting
   validSortFields := map[string]bool{
       "name":             true,
       "first_appearance": true,
       "created_at":       true,
   }
   
   if validSortFields[filter.SortBy] {
       query += fmt.Sprintf(" ORDER BY %s", filter.SortBy)
       
       if strings.ToLower(filter.SortDir) == "desc" {
           query += " DESC"
       } else {
           query += " ASC"
       }
   } else {
       query += " ORDER BY created_at DESC"
   }

   // Add pagination
   offset := (filter.Page - 1) * filter.PageSize
   query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
   args = append(args, filter.PageSize, offset)

   // Execute query
   var instruments []*vitalinstrument.VitalInstrument
   err = r.db.SelectContext(ctx, &instruments, query, args...)
   if err != nil {
       log.Error("Error querying vital instruments", zap.Error(err))
       return nil, 0, fmt.Errorf("query vital instruments: %w", err)
   }

   return instruments, total, nil
}

func (r *VitalInstrumentReader) ListByCharacter(ctx context.Context, characterID uint, limit int) ([]*vitalinstrument.VitalInstrument, error) {
   log := logger.GetLogger(
       zap.String("repository", "VitalInstrumentReader"), 
       zap.String("method", "ListByCharacter"),
   )
   
   query := `
       SELECT id, name, main_image, description, powers, character_id, 
              first_appearance, created_at, updated_at, deleted_at
       FROM vital_instruments
       WHERE character_id = $1 AND deleted_at IS NULL
       ORDER BY name ASC
       LIMIT $2
   `

   var instruments []*vitalinstrument.VitalInstrument
   err := r.db.SelectContext(ctx, &instruments, query, characterID, limit)
   if err != nil {
       log.Error("Error querying vital instruments by character", zap.Error(err))
       return nil, fmt.Errorf("query vital instruments by character: %w", err)
   }

   return instruments, nil
}
