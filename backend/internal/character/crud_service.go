package character

import (
    "context"
    "fmt"
    "time"

    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements basic CRUD operations for characters
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

func (s *CrudService) Get(ctx context.Context, id uint) (*Character, error) {
    log := logger.GetLogger(zap.String("method", "Get"), zap.Uint("id", id))

    if id == 0 {
        log.Error("Invalid ID provided")
        return nil, fmt.Errorf("%w: id is required", ErrInvalidInput)
    }

    character, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get character", zap.Error(err))
        return nil, fmt.Errorf("get character: %w", err)
    }

    return character, nil
}

func (s *CrudService) Create(ctx context.Context, input CreateCharacterInput) (*Character, error) {
    log := logger.GetLogger(zap.String("method", "Create"))

    // Validate input
    if input.Name == "" {
        return nil, fmt.Errorf("%w: name is required", ErrInvalidInput)
    }
    if input.MainImage == "" {
        return nil, fmt.Errorf("%w: main_image is required", ErrInvalidInput)
    }
    if input.Description == "" {
        return nil, fmt.Errorf("%w: description is required", ErrInvalidInput)
    }
    if input.Status == "" {
        return nil, fmt.Errorf("%w: status is required", ErrInvalidInput)
    }
    if input.FirstAppearance <= 0 {
        return nil, fmt.Errorf("%w: first_appearance must be positive", ErrInvalidInput)
    }

    // Create entity
    now := time.Now()
    character := &Character{
        Name:            input.Name,
        NameJapanese:    input.NameJapanese,
        MainImage:       input.MainImage,
        Description:     input.Description,
        Species:         input.Species,
        Gender:          input.Gender,
        Age:             input.Age,
        Height:          input.Height,
        Status:          input.Status,
        Affiliation:     input.Affiliation,
        Occupation:      input.Occupation,
        BirthDate:       input.BirthDate,
        BirthPlace:      input.BirthPlace,
        Relatives:       input.Relatives,
        FirstAppearance: input.FirstAppearance,
        CreatedAt:       now,
        UpdatedAt:       now,
    }

    // Save to database
    if err := s.writer.Create(ctx, character); err != nil {
        log.Error("Failed to create character", 
            zap.Error(err), 
            zap.String("name", input.Name))
        return nil, fmt.Errorf("create character: %w", err)
    }

    log.Info("Character created successfully", zap.Uint("id", character.ID))
    return character, nil
}

func (s *CrudService) Update(ctx context.Context, id uint, input UpdateCharacterInput) (*Character, error) {
    log := logger.GetLogger(zap.String("method", "Update"), zap.Uint("id", id))
    
    // Get existing character
    character, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get character for update", zap.Error(err))
        return nil, fmt.Errorf("get character: %w", err)
    }

    // Update fields if provided
    updated := false

    if input.Name != nil {
        character.Name = *input.Name
        updated = true
    }

    if input.NameJapanese != nil {
        character.NameJapanese = *input.NameJapanese
        updated = true
    }

    if input.MainImage != nil {
        character.MainImage = *input.MainImage
        updated = true
    }

    if input.Description != nil {
        character.Description = *input.Description
        updated = true
    }

    if input.Species != nil {
        character.Species = *input.Species
        updated = true
    }

    if input.Gender != nil {
        character.Gender = *input.Gender
        updated = true
    }

    if input.Age != nil {
        character.Age = *input.Age
        updated = true
    }

    if input.Height != nil {
        character.Height = *input.Height
        updated = true
    }

    if input.Status != nil {
        character.Status = *input.Status
        updated = true
    }

    if input.Affiliation != nil {
        character.Affiliation = *input.Affiliation
        updated = true
    }

    if input.Occupation != nil {
        character.Occupation = *input.Occupation
        updated = true
    }

    if input.BirthDate != nil {
        character.BirthDate = *input.BirthDate
        updated = true
    }

    if input.BirthPlace != nil {
        character.BirthPlace = *input.BirthPlace
        updated = true
    }

    if input.Relatives != nil {
        character.Relatives = *input.Relatives
        updated = true
    }

    if input.FirstAppearance != nil {
        character.FirstAppearance = *input.FirstAppearance
        updated = true
    }

    // If no changes, return character without update
    if !updated {
        log.Info("No changes to update")
        return character, nil
    }

    // Update timestamp
    character.UpdatedAt = time.Now()

    // Save to database
    if err := s.writer.Update(ctx, character); err != nil {
        log.Error("Failed to update character", zap.Error(err))
        return nil, fmt.Errorf("update character: %w", err)
    }

    log.Info("Character updated successfully")
    return character, nil
}

func (s *CrudService) Delete(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("method", "Delete"), zap.Uint("id", id))

    // Verify character exists
    _, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get character for deletion", zap.Error(err))
        return fmt.Errorf("get character: %w", err)
    }

    if err := s.writer.Delete(ctx, id); err != nil {
        log.Error("Failed to delete character", zap.Error(err))
        return fmt.Errorf("delete character: %w", err)
    }

    log.Info("Character deleted successfully")
    return nil
}

func (s *CrudService) DeletePermanently(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("method", "DeletePermanently"), zap.Uint("id", id))

    // Verify character exists
    _, err := s.reader.GetByID(ctx, id)
    if err != nil {
        // If soft-deleted, delete anyway
        if err != ErrCharacterNotFound {
            log.Error("Failed to check character for permanent deletion", zap.Error(err))
            return fmt.Errorf("check character: %w", err)
        }
    }

    if err := s.writer.DeletePermanently(ctx, id); err != nil {
        log.Error("Failed to delete character permanently", zap.Error(err))
        return fmt.Errorf("delete character permanently: %w", err)
    }

    log.Info("Character permanently deleted")
    return nil
}

func (s *CrudService) Restore(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("method", "Restore"), zap.Uint("id", id))

    if err := s.writer.Restore(ctx, id); err != nil {
        log.Error("Failed to restore character", zap.Error(err))
        return fmt.Errorf("restore character: %w", err)
    }

    log.Info("Character restored successfully")
    return nil
}
