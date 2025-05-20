package vitalinstrument

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

func (s *SearchService) List(ctx context.Context, filter VitalInstrumentFilter) ([]*VitalInstrument, int, error) {
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
    }

    if !validSortFields[filter.SortBy] {
        filter.SortBy = "created_at"
    }

    // Normalize sort direction
    if filter.SortDir != "asc" && filter.SortDir != "desc" {
        filter.SortDir = "desc"
    }

    instruments, total, err := s.reader.List(ctx, filter)
    if err != nil {
        log.Error("Failed to list vital instruments", zap.Error(err))
        return nil, 0, fmt.Errorf("list vital instruments: %w", err)
    }

    return instruments, total, nil
}

func (s *SearchService) ListByCharacter(ctx context.Context, characterID uint, limit int) ([]*VitalInstrument, error) {
    log := logger.GetLogger(
        zap.String("method", "ListByCharacter"), 
        zap.Uint("characterID", characterID),
    )
    
    if characterID == 0 {
        return nil, fmt.Errorf("%w: character_id is required", ErrInvalidInput)
    }

    if limit <= 0 {
        limit = 10
    }
    if limit > 50 {
        limit = 50
    }

    instruments, err := s.reader.ListByCharacter(ctx, characterID, limit)
    if err != nil {
        log.Error("Failed to list vital instruments by character", zap.Error(err))
        return nil, fmt.Errorf("list vital instruments by character: %w", err)
    }

    return instruments, nil
}
