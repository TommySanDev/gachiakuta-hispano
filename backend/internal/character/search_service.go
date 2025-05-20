package character

import (
    "context"
    "fmt"
    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements advanced search and filtering operations
type SearchService struct {
    reader Reader
}

func NewSearchService(reader Reader) *SearchService {
    return &SearchService{
        reader: reader,
    }
}

func (s *SearchService) List(ctx context.Context, filter CharacterFilter) ([]*Character, int, error) {
    log := logger.GetLogger(zap.String("method", "List"))
    
    // Set default values
    if filter.Page <= 0 {
        filter.Page = 1
    }
    if filter.PageSize <= 0 {
        filter.PageSize = 20
    }
    if filter.PageSize > 100 {
        filter.PageSize = 100
    }

    // Validate sort fields
    validSortFields := map[string]bool{
        "name":             true,
        "first_appearance": true,
        "created_at":       true,
        "status":           true,
        "affiliation":      true,
    }

    if !validSortFields[filter.SortBy] {
        filter.SortBy = "created_at"
    }

    // Normalize sort direction
    if filter.SortDir != "asc" && filter.SortDir != "desc" {
        filter.SortDir = "desc"
    }

    characters, total, err := s.reader.List(ctx, filter)
    if err != nil {
        log.Error("Failed to list characters", zap.Error(err))
        return nil, 0, fmt.Errorf("list characters: %w", err)
    }

    return characters, total, nil
}

func (s *SearchService) ListByAffiliation(ctx context.Context, affiliation string, limit int) ([]*Character, error) {
    log := logger.GetLogger(
        zap.String("method", "ListByAffiliation"), 
        zap.String("affiliation", affiliation),
    )
    
    if affiliation == "" {
        return nil, fmt.Errorf("%w: affiliation is required", ErrInvalidInput)
    }

    if limit <= 0 {
        limit = 10
    }
    if limit > 50 {
        limit = 50
    }

    characters, err := s.reader.ListByAffiliation(ctx, affiliation, limit)
    if err != nil {
        log.Error("Failed to list characters by affiliation", zap.Error(err))
        return nil, fmt.Errorf("list characters by affiliation: %w", err)
    }

    return characters, nil
}

func (s *SearchService) ListByStatus(ctx context.Context, status string, limit int) ([]*Character, error) {
    log := logger.GetLogger(
        zap.String("method", "ListByStatus"), 
        zap.String("status", status),
    )
    
    if status == "" {
        return nil, fmt.Errorf("%w: status is required", ErrInvalidInput)
    }

    if limit <= 0 {
        limit = 10
    }
    if limit > 50 {
        limit = 50
    }

    characters, err := s.reader.ListByStatus(ctx, status, limit)
    if err != nil {
        log.Error("Failed to list characters by status", zap.Error(err))
        return nil, fmt.Errorf("list characters by status: %w", err)
    }

    return characters, nil
}

func (s *SearchService) ListBySpecies(ctx context.Context, species string, limit int) ([]*Character, error) {
    log := logger.GetLogger(
        zap.String("method", "ListBySpecies"), 
        zap.String("species", species),
    )
    
    if species == "" {
        return nil, fmt.Errorf("%w: species is required", ErrInvalidInput)
    }

    if limit <= 0 {
        limit = 10
    }
    if limit > 50 {
        limit = 50
    }

    characters, err := s.reader.ListBySpecies(ctx, species, limit)
    if err != nil {
        log.Error("Failed to list characters by species", zap.Error(err))
        return nil, fmt.Errorf("list characters by species: %w", err)
    }

    return characters, nil
}
