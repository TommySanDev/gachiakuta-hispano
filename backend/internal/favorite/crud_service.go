package favorite

import (
    "context"
    "fmt"
    "strings"
    "time"

    "go.uber.org/zap"
)

// Implements basic operations for favorites
type CrudService struct {
    reader Reader
    writer Writer
}

func NewCrudService(reader Reader, writer Writer) *CrudService {
    return &CrudService{
        reader: reader,
        writer: writer,
    }
}

func (s *CrudService) Add(ctx context.Context, userID uint, input CreateFavoriteInput) (*Favorite, error) {
    log := logger.GetLogger(zap.String("method", "AddFavorite"))

    if userID == 0 || input.EntityType == "" || input.EntityID == 0 {
        return nil, fmt.Errorf("%w: userID, entity_type and entity_id are required", ErrInvalidInput)
    }

    normalizedType := strings.ToLower(input.EntityType)
    if normalizedType != "character" && normalizedType != "chapter" && normalizedType != "vital_instrument" {
        return nil, fmt.Errorf("%w: unsupported entity_type", ErrInvalidInput)
    }

    fav := &Favorite{
        UserID:     userID,
        EntityType: normalizedType,
        EntityID:   input.EntityID,
        CreatedAt:  time.Now(),
    }

    if err := s.writer.Create(ctx, fav); err != nil {
        return nil, fmt.Errorf("create favorite: %w", err)
    }

    return fav, nil
}

func (s *CrudService) Remove(ctx context.Context, userID uint, input DeleteFavoriteInput) error {
    log := logger.GetLogger(zap.String("method", "RemoveFavorite"))

    if userID == 0 || input.EntityType == "" || input.EntityID == 0 {
        return fmt.Errorf("%w: userID, entity_type and entity_id are required", ErrInvalidInput)
    }

    return s.writer.Delete(ctx, userID, strings.ToLower(input.EntityType), input.EntityID)
}

func (s *CrudService) List(ctx context.Context, userID uint) ([]*Favorite, error) {
    log := logger.GetLogger(zap.String("method", "ListFavorites"), zap.Uint("user_id", userID))

    if userID == 0 {
        return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
    }

    return s.reader.ListByUser(ctx, userID)
}

