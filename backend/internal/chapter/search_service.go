package chapter

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

func (s *SearchService) List(ctx context.Context, filter ChapterFilter) ([]*Chapter, int, error) {
    log := logger.GetLogger(zap.String("method", "ListChapters"))

    // Set default pagination values
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
        "title":      true,
        "number":     true,
        "created_at": true,
    }

    if !validSortFields[filter.SortBy] {
        filter.SortBy = "created_at"
    }

    // Normalize sort direction
    if filter.SortDir != "asc" && filter.SortDir != "desc" {
        filter.SortDir = "desc"
    }

    chapters, total, err := s.reader.List(ctx, filter)
    if err != nil {
        log.Error("Failed to list chapters", zap.Error(err))
        return nil, 0, fmt.Errorf("list chapters: %w", err)
    }

    return chapters, total, nil
}

