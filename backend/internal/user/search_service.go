package user

import (
    "context"
    "fmt"

    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements advanced search and filtering operations for users
type SearchService struct {
    reader UserReader
}

func NewSearchService(reader UserReader) *SearchService {
    return &SearchService{
        reader: reader,
    }
}

func (s *SearchService) List(ctx context.Context, filter UserFilter) ([]*User, int, error) {
    log := logger.GetLogger(zap.String("service", "SearchService"), zap.String("method", "List"))
    
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
        "email":         true,
        "username":      true,
        "first_name":    true,
        "last_name":     true,
        "role":          true,
        "is_active":     true,
        "email_verified": true,
        "created_at":    true,
        "last_login_at": true,
    }

    if !validSortFields[filter.SortBy] {
        filter.SortBy = "created_at"
    }

    // Normalize sort direction
    if filter.SortDir != "asc" && filter.SortDir != "desc" {
        filter.SortDir = "desc"
    }

    users, total, err := s.reader.List(ctx, filter)
    if err != nil {
        log.Error("Failed to list users", zap.Error(err))
        return nil, 0, fmt.Errorf("list users: %w", err)
    }

    return users, total, nil
}

func (s *SearchService) ListByRole(ctx context.Context, role string, limit int) ([]*User, error) {
    log := logger.GetLogger(
        zap.String("service", "SearchService"), 
        zap.String("method", "ListByRole"),
        zap.String("role", role),
    )
    
    if role == "" {
        return nil, fmt.Errorf("%w: role is required", ErrInvalidInput)
    }

    // Validate role
    if role != RoleAdmin && role != RoleEditor && role != RoleUser {
        return nil, fmt.Errorf("%w: invalid role", ErrInvalidInput)
    }

    if limit <= 0 {
        limit = 10
    }
    if limit > 50 {
        limit = 50
    }

    users, err := s.reader.ListByRole(ctx, role, limit)
    if err != nil {
        log.Error("Failed to list users by role", zap.Error(err))
        return nil, fmt.Errorf("list users by role: %w", err)
    }

    return users, nil
}

func (s *SearchService) ListActive(ctx context.Context, limit int) ([]*User, error) {
    log := logger.GetLogger(
        zap.String("service", "SearchService"), 
        zap.String("method", "ListActive"),
    )

    if limit <= 0 {
        limit = 10
    }
    if limit > 50 {
        limit = 50
    }

    users, err := s.reader.ListActive(ctx, limit)
    if err != nil {
        log.Error("Failed to list active users", zap.Error(err))
        return nil, fmt.Errorf("list active users: %w", err)
    }

    return users, nil
}

func (s *SearchService) ListRecentlyCreated(ctx context.Context, limit int) ([]*User, error) {
    log := logger.GetLogger(
        zap.String("service", "SearchService"), 
        zap.String("method", "ListRecentlyCreated"),
    )

    if limit <= 0 {
        limit = 10
    }
    if limit > 50 {
        limit = 50
    }

    users, err := s.reader.ListRecentlyCreated(ctx, limit)
    if err != nil {
        log.Error("Failed to list recently created users", zap.Error(err))
        return nil, fmt.Errorf("list recently created users: %w", err)
    }

    return users, nil
}

func (s *SearchService) SearchByEmail(ctx context.Context, emailPattern string, limit int) ([]*User, error) {
    log := logger.GetLogger(
        zap.String("service", "SearchService"), 
        zap.String("method", "SearchByEmail"),
    )

    if emailPattern == "" {
        return nil, fmt.Errorf("%w: email pattern is required", ErrInvalidInput)
    }

    if limit <= 0 {
        limit = 10
    }
    if limit > 50 {
        limit = 50
    }

    users, err := s.reader.SearchByEmail(ctx, emailPattern, limit)
    if err != nil {
        log.Error("Failed to search users by email", zap.Error(err))
        return nil, fmt.Errorf("search users by email: %w", err)
    }

    return users, nil
}
