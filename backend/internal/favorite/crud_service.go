package favorite

import (
    "context"
    "fmt"
    "strings"
    "time"

    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
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
        log.Warn("Invalid input for AddFavorite", zap.Uint("user_id", userID), zap.Any("input", input))
        return nil, fmt.Errorf("%w: userID, entity_type and entity_id are required", ErrInvalidInput)
    }

    normalizedType := strings.ToLower(input.EntityType)
    if normalizedType != "character" && normalizedType != "chapter" && normalizedType != "vital_instrument" {
        log.Warn("Unsupported entity_type", zap.String("entity_type", input.EntityType))
        return nil, fmt.Errorf("%w: unsupported entity_type", ErrInvalidInput)
    }

    fav := &Favorite{
        UserID:     userID,
        EntityType: normalizedType,
        EntityID:   input.EntityID,
        CreatedAt:  time.Now(),
    }

    if err := s.writer.Create(ctx, fav); err != nil {
        log.Error("Failed to create favorite", zap.Error(err))
        return nil, fmt.Errorf("create favorite: %w", err)
    }

    log.Info("Favorite added successfully", zap.Uint("user_id", userID), zap.String("type", normalizedType), zap.Uint("entity_id", input.EntityID))
    return fav, nil
}

func (s *CrudService) Remove(ctx context.Context, userID uint, input DeleteFavoriteInput) error {
    log := logger.GetLogger(zap.String("method", "RemoveFavorite"))

    if userID == 0 || input.EntityType == "" || input.EntityID == 0 {
        log.Warn("Invalid input for RemoveFavorite", zap.Uint("user_id", userID), zap.Any("input", input))
        return fmt.Errorf("%w: userID, entity_type and entity_id are required", ErrInvalidInput)
    }

    err := s.writer.Delete(ctx, userID, strings.ToLower(input.EntityType), input.EntityID)
    if err != nil {
        log.Error("Failed to remove favorite", zap.Error(err))
        return err
    }

    log.Info("Favorite removed successfully", zap.Uint("user_id", userID), zap.String("type", input.EntityType), zap.Uint("entity_id", input.EntityID))
    return nil
}

func (s *CrudService) List(ctx context.Context, userID uint) ([]*Favorite, error) {
    log := logger.GetLogger(zap.String("method", "ListFavorites"), zap.Uint("user_id", userID))

    if userID == 0 {
        log.Warn("Missing user_id in ListFavorites")
        return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
    }

    favorites, err := s.reader.ListByUser(ctx, userID)
    if err != nil {
        log.Error("Failed to list favorites", zap.Error(err))
        return nil, err
    }

    log.Info("Favorites listed successfully", zap.Int("count", len(favorites)))
    return favorites, nil
}

