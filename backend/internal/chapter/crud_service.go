package chapter

import (
    "context"
    "fmt"
    "time"

    "go.uber.org/zap"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements basic CRUD operations for chapters
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

func (s *CrudService) Get(ctx context.Context, id uint) (*Chapter, error) {
    log := logger.GetLogger(zap.String("method", "GetChapter"), zap.Uint("id", id))

    if id == 0 {
        log.Error("Invalid ID provided")
        return nil, fmt.Errorf("%w: id is required", ErrInvalidInput)
    }

    chapter, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get chapter", zap.Error(err))
        return nil, fmt.Errorf("get chapter: %w", err)
    }

    return chapter, nil
}

func (s *CrudService) Create(ctx context.Context, input CreateChapterInput) (*Chapter, error) {
    log := logger.GetLogger(zap.String("method", "CreateChapter"))

    if input.Title == "" {
        return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
    }
    if input.Number <= 0 {
        return nil, fmt.Errorf("%w: number must be positive", ErrInvalidInput)
    }
    if input.Image == "" {
        return nil, fmt.Errorf("%w: image is required", ErrInvalidInput)
    }

    now := time.Now()
    chapter := &Chapter{
        Title:     input.Title,
        Number:    input.Number,
        Image:     input.Image,
        CreatedAt: now,
        UpdatedAt: now,
    }

    if err := s.writer.Create(ctx, chapter); err != nil {
        log.Error("Failed to create chapter", zap.Error(err), zap.String("title", input.Title))
        return nil, fmt.Errorf("create chapter: %w", err)
    }

    log.Info("Chapter created successfully", zap.Uint("id", chapter.ID))
    return chapter, nil
}

func (s *CrudService) Update(ctx context.Context, id uint, input UpdateChapterInput) (*Chapter, error) {
    log := logger.GetLogger(zap.String("method", "UpdateChapter"), zap.Uint("id", id))

    chapter, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get chapter for update", zap.Error(err))
        return nil, fmt.Errorf("get chapter: %w", err)
    }

    updated := false

    if input.Title != nil {
        chapter.Title = *input.Title
        updated = true
    }
    if input.Number != nil {
        chapter.Number = *input.Number
        updated = true
    }
    if input.Image != nil {
        chapter.Image = *input.Image
        updated = true
    }

    if !updated {
        log.Info("No changes to update")
        return chapter, nil
    }

    chapter.UpdatedAt = time.Now()

    if err := s.writer.Update(ctx, chapter); err != nil {
        log.Error("Failed to update chapter", zap.Error(err))
        return nil, fmt.Errorf("update chapter: %w", err)
    }

    log.Info("Chapter updated successfully")
    return chapter, nil
}

func (s *CrudService) Delete(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("method", "DeleteChapter"), zap.Uint("id", id))

    _, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get chapter for deletion", zap.Error(err))
        return fmt.Errorf("get chapter: %w", err)
    }

    if err := s.writer.Delete(ctx, id); err != nil {
        log.Error("Failed to delete chapter", zap.Error(err))
        return fmt.Errorf("delete chapter: %w", err)
    }

    log.Info("Chapter deleted successfully")
    return nil
}

func (s *CrudService) DeletePermanently(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("method", "DeleteChapterPermanently"), zap.Uint("id", id))

    _, err := s.reader.GetByID(ctx, id)
    if err != nil && err != ErrChapterNotFound {
        log.Error("Failed to check chapter for permanent deletion", zap.Error(err))
        return fmt.Errorf("check chapter: %w", err)
    }

    if err := s.writer.DeletePermanently(ctx, id); err != nil {
        log.Error("Failed to permanently delete chapter", zap.Error(err))
        return fmt.Errorf("delete chapter permanently: %w", err)
    }

    log.Info("Chapter permanently deleted")
    return nil
}

func (s *CrudService) Restore(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("method", "RestoreChapter"), zap.Uint("id", id))

    if err := s.writer.Restore(ctx, id); err != nil {
        log.Error("Failed to restore chapter", zap.Error(err))
        return fmt.Errorf("restore chapter: %w", err)
    }

    log.Info("Chapter restored successfully")
    return nil
}

