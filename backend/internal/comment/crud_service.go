package comment

import (
    "context"
    "fmt"
    "time"

    "go.uber.org/zap"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements CRUD operations for comments
type CrudService struct {
    reader Reader
    writer Writer
}

func NewCrudService(reader Reader, writer Writer) *CrudService {
    return &CrudService{reader: reader, writer: writer}
}

func (s *CrudService) Get(ctx context.Context, id uint) (*Comment, error) {
    if id == 0 {
        return nil, fmt.Errorf("%w: id is required", ErrInvalidInput)
    }

    comment, err := s.reader.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get comment: %w", err)
    }

    return comment, nil
}

func (s *CrudService) ListByChapter(ctx context.Context, chapterID uint, includeDeleted bool) ([]*Comment, error) {
    if chapterID == 0 {
        return nil, fmt.Errorf("%w: chapter_id is required", ErrInvalidInput)
    }

    filter := ChapterCommentFilter{
        ChapterID:      chapterID,
        IncludeDeleted: includeDeleted,
    }

    return s.reader.ListByChapter(ctx, filter)
}

func (s *CrudService) Create(ctx context.Context, userID uint, input CreateCommentInput) (*Comment, error) {
    if userID == 0 || input.ChapterID == 0 || input.Content == "" {
        return nil, fmt.Errorf("%w: all fields are required", ErrInvalidInput)
    }

    now := time.Now()
    comment := &Comment{
        UserID:    userID,
        ChapterID: input.ChapterID,
        Content:   input.Content,
        CreatedAt: now,
        UpdatedAt: now,
    }

    if err := s.writer.Create(ctx, comment); err != nil {
        return nil, fmt.Errorf("create comment: %w", err)
    }

    return comment, nil
}

func (s *CrudService) Update(ctx context.Context, userID, id uint, input UpdateCommentInput) (*Comment, error) {
    if id == 0 || userID == 0 {
        return nil, fmt.Errorf("%w: id and user_id are required", ErrInvalidInput)
    }

    comment, err := s.reader.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("get comment: %w", err)
    }

    // Only owner or elevated roles should be allowed (roles checked at handler/router level)
    if comment.UserID != userID {
        return nil, ErrForbidden
    }

    updated := false
    if input.Content != nil {
        comment.Content = *input.Content
        updated = true
    }

    if !updated {
        return comment, nil
    }

    comment.UpdatedAt = time.Now()

    if err := s.writer.Update(ctx, comment); err != nil {
        return nil, fmt.Errorf("update comment: %w", err)
    }

    return comment, nil
}

func (s *CrudService) Delete(ctx context.Context, userID, id uint) error {
    if id == 0 || userID == 0 {
        return fmt.Errorf("%w: id and user_id are required", ErrInvalidInput)
    }

    comment, err := s.reader.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("get comment: %w", err)
    }

    // Only owner or elevated roles should be allowed (roles checked externally)
    if comment.UserID != userID {
        return ErrForbidden
    }

    return s.writer.Delete(ctx, id)
}

func (s *CrudService) DeletePermanently(ctx context.Context, id uint) error {
    if id == 0 {
        return fmt.Errorf("%w: id is required", ErrInvalidInput)
    }

    return s.writer.DeletePermanently(ctx, id)
}

