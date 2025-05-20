package postgres

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"

    "github.com/jmoiron/sqlx"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/character"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements character.Reader using PostgreSQL
type CharacterReader struct {
    db *sqlx.DB
}

func NewCharacterReader(db *sqlx.DB) *CharacterReader {
    return &CharacterReader{
        db: db,
    }
}

func (r *CharacterReader) GetByID(ctx context.Context, id uint) (*character.Character, error) {
    log := logger.GetLogger(zap.String("repository", "CharacterReader"), zap.String("method", "GetByID"))
    
    query := `
        SELECT id, name, name_japanese, main_image, description, species, gender, age,
               height, status, affiliation, occupation, birth_date, birth_place,
               relatives, first_appearance, created_at, updated_at, deleted_at
        FROM characters
        WHERE id = $1 AND deleted_at IS NULL
    `

    var c character.Character
    err := r.db.GetContext(ctx, &c, query, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            log.Debug("Character not found", zap.Uint("id", id))
            return nil, character.ErrCharacterNotFound
        }
        log.Error("Database error", zap.Error(err))
        return nil, fmt.Errorf("database error: %w", err)
    }

    return &c, nil
}

func (r *CharacterReader) List(ctx context.Context, filter character.CharacterFilter) ([]*character.Character, int, error) {
    log := logger.GetLogger(zap.String("repository", "CharacterReader"), zap.String("method", "List"))
    
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
        whereClauses = append(whereClauses, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d OR name_japanese ILIKE $%d)", argPos, argPos+1, argPos+2))
        searchTerm := "%" + filter.Search + "%"
        args = append(args, searchTerm, searchTerm, searchTerm)
        argPos += 3
    }

    // Fields filters
    if filter.Status != "" {
        whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argPos))
        args = append(args, filter.Status)
        argPos++
    }

    if filter.Affiliation != "" {
        whereClauses = append(whereClauses, fmt.Sprintf("affiliation ILIKE $%d", argPos))
        args = append(args, "%"+filter.Affiliation+"%")
        argPos++
    }

    if filter.Species != "" {
        whereClauses = append(whereClauses, fmt.Sprintf("species = $%d", argPos))
        args = append(args, filter.Species)
        argPos++
    }

    whereClause := "WHERE " + strings.Join(whereClauses, " AND ")

    // Count total matching records
    countQuery := "SELECT COUNT(*) FROM characters " + whereClause
    var total int
    err := r.db.GetContext(ctx, &total, countQuery, args...)
    if err != nil {
        log.Error("Error counting characters", zap.Error(err))
        return nil, 0, fmt.Errorf("count characters: %w", err)
    }

    // Main query with sorting and pagination
    query := `
        SELECT id, name, name_japanese, main_image, description, species, gender, age,
               height, status, affiliation, occupation, birth_date, birth_place,
               relatives, first_appearance, created_at, updated_at, deleted_at
        FROM characters
        ` + whereClause

    // Add sorting
    validSortFields := map[string]bool{
        "name":             true,
        "status":           true,
        "affiliation":      true,
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
    var characters []*character.Character
    err = r.db.SelectContext(ctx, &characters, query, args...)
    if err != nil {
        log.Error("Error querying characters", zap.Error(err))
        return nil, 0, fmt.Errorf("query characters: %w", err)
    }

    return characters, total, nil
}

func (r *CharacterReader) ListByAffiliation(ctx context.Context, affiliation string, limit int) ([]*character.Character, error) {
    log := logger.GetLogger(
        zap.String("repository", "CharacterReader"), 
        zap.String("method", "ListByAffiliation"),
    )
    
    query := `
        SELECT id, name, name_japanese, main_image, description, species, gender, age,
               height, status, affiliation, occupation, birth_date, birth_place,
               relatives, first_appearance, created_at, updated_at, deleted_at
        FROM characters
        WHERE affiliation ILIKE $1 AND deleted_at IS NULL
        ORDER BY name ASC
        LIMIT $2
    `

    var characters []*character.Character
    err := r.db.SelectContext(ctx, &characters, query, "%"+affiliation+"%", limit)
    if err != nil {
        log.Error("Error querying characters by affiliation", zap.Error(err))
        return nil, fmt.Errorf("query characters by affiliation: %w", err)
    }

    return characters, nil
}

func (r *CharacterReader) ListByStatus(ctx context.Context, status string, limit int) ([]*character.Character, error) {
    log := logger.GetLogger(
        zap.String("repository", "CharacterReader"), 
        zap.String("method", "ListByStatus"),
    )
    
    query := `
        SELECT id, name, name_japanese, main_image, description, species, gender, age,
               height, status, affiliation, occupation, birth_date, birth_place,
               relatives, first_appearance, created_at, updated_at, deleted_at
        FROM characters
        WHERE status = $1 AND deleted_at IS NULL
        ORDER BY name ASC
        LIMIT $2
    `

    var characters []*character.Character
    err := r.db.SelectContext(ctx, &characters, query, status, limit)
    if err != nil {
        log.Error("Error querying characters by status", zap.Error(err))
        return nil, fmt.Errorf("query characters by status: %w", err)
    }

    return characters, nil
}

func (r *CharacterReader) ListBySpecies(ctx context.Context, species string, limit int) ([]*character.Character, error) {
    log := logger.GetLogger(
        zap.String("repository", "CharacterReader"), 
        zap.String("method", "ListBySpecies"),
    )
    
    query := `
        SELECT id, name, name_japanese, main_image, description, species, gender, age,
               height, status, affiliation, occupation, birth_date, birth_place,
               relatives, first_appearance, created_at, updated_at, deleted_at
        FROM characters
        WHERE species = $1 AND deleted_at IS NULL
        ORDER BY name ASC
        LIMIT $2
    `

    var characters []*character.Character
    err := r.db.SelectContext(ctx, &characters, query, species, limit)
    if err != nil {
        log.Error("Error querying characters by species", zap.Error(err))
        return nil, fmt.Errorf("query characters by species: %w", err)
    }

    return characters, nil
}
