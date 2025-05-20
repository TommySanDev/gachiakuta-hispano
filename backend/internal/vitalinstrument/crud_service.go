package vitalinstrument

import (
    "context"
    "fmt"
    "time"

    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements basic CRUD operations for vital instruments
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

func (s *CrudService) Get(ctx context.Context, id uint) (*VitalInstrument, error) {
    log := logger.GetLogger(zap.String("method", "Get"), zap.Uint("id", id))
    
    if id == 0 {
        log.Error("Invalid ID provided")
        return nil, fmt.Errorf("%w: id is required", ErrInvalidInput)
    }

    instrument, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get vital instrument", zap.Error(err))
        return nil, fmt.Errorf("get vital instrument: %w", err)
    }

    return instrument, nil
}

func (s *CrudService) Create(ctx context.Context, input CreateVitalInstrumentInput) (*VitalInstrument, error) {
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
    if input.FirstAppearance <= 0 {
        return nil, fmt.Errorf("%w: first_appearance must be positive", ErrInvalidInput)
    }

    // Create entity
    now := time.Now()
    instrument := &VitalInstrument{
        Name:            input.Name,
        MainImage:       input.MainImage,
        Description:     input.Description,
        Powers:          input.Powers,
        CharacterID:     input.CharacterID,
        FirstAppearance: input.FirstAppearance,
        CreatedAt:       now,
        UpdatedAt:       now,
    }

    // Save to database
    if err := s.writer.Create(ctx, instrument); err != nil {
        log.Error("Failed to create vital instrument", 
            zap.Error(err), 
            zap.String("name", input.Name))
        return nil, fmt.Errorf("create vital instrument: %w", err)
    }

    log.Info("Vital instrument created successfully", zap.Uint("id", instrument.ID))
    return instrument, nil
}

func (s *CrudService) Update(ctx context.Context, id uint, input UpdateVitalInstrumentInput) (*VitalInstrument, error) {
    log := logger.GetLogger(zap.String("method", "Update"), zap.Uint("id", id))
    
    // Get existing vital instrument
    instrument, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get vital instrument for update", zap.Error(err))
        return nil, fmt.Errorf("get vital instrument: %w", err)
    }

    // Update fields if provided
    updated := false

    if input.Name != nil {
        instrument.Name = *input.Name
        updated = true
    }

    if input.MainImage != nil {
        instrument.MainImage = *input.MainImage
        updated = true
    }

    if input.Description != nil {
        instrument.Description = *input.Description
        updated = true
    }

    if input.Powers != nil {
        instrument.Powers = *input.Powers
        updated = true
    }

    if input.CharacterID != nil {
        instrument.CharacterID = input.CharacterID
        updated = true
    }

    if input.FirstAppearance != nil {
        if *input.FirstAppearance <= 0 {
            return nil, fmt.Errorf("%w: first_appearance must be positive", ErrInvalidInput)
        }
        instrument.FirstAppearance = *input.FirstAppearance
        updated = true
    }

    // If no changes, return instrument without update
    if !updated {
        log.Info("No changes to update")
        return instrument, nil
    }

    // Update timestamp
    instrument.UpdatedAt = time.Now()

    // Save to database
    if err := s.writer.Update(ctx, instrument); err != nil {
        log.Error("Failed to update vital instrument", zap.Error(err))
        return nil, fmt.Errorf("update vital instrument: %w", err)
    }

    log.Info("Vital instrument updated successfully")
    return instrument, nil
}

func (s *CrudService) Delete(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("method", "Delete"), zap.Uint("id", id))
    
    // Verify vital instrument exists
    _, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get vital instrument for deletion", zap.Error(err))
        return fmt.Errorf("get vital instrument: %w", err)
    }

    if err := s.writer.Delete(ctx, id); err != nil {
        log.Error("Failed to delete vital instrument", zap.Error(err))
        return fmt.Errorf("delete vital instrument: %w", err)
    }

    log.Info("Vital instrument deleted successfully")
    return nil
}

func (s *CrudService) DeletePermanently(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("method", "DeletePermanently"), zap.Uint("id", id))
    
    // Verify vital instrument exists
    _, err := s.reader.GetByID(ctx, id)
    if err != nil {
        // If soft-deleted, delete anyway
        if err != ErrVitalInstrumentNotFound {
            log.Error("Failed to check vital instrument for permanent deletion", zap.Error(err))
            return fmt.Errorf("check vital instrument: %w", err)
        }
    }

    if err := s.writer.DeletePermanently(ctx, id); err != nil {
        log.Error("Failed to delete vital instrument permanently", zap.Error(err))
        return fmt.Errorf("delete vital instrument permanently: %w", err)
    }

    log.Info("Vital instrument permanently deleted")
    return nil
}

func (s *CrudService) Restore(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("method", "Restore"), zap.Uint("id", id))
    
    if err := s.writer.Restore(ctx, id); err != nil {
        log.Error("Failed to restore vital instrument", zap.Error(err))
        return fmt.Errorf("restore vital instrument: %w", err)
    }

    log.Info("Vital instrument restored successfully")
    return nil
}
