package user

import (
    "context"
    "fmt"
    "time"

    "golang.org/x/crypto/bcrypt"
    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Implements basic CRUD operations for users
type CrudService struct {
    reader UserReader
    writer UserWriter
}

func NewCrudService(reader UserReader, writer UserWriter) *CrudService {
    return &CrudService{
        reader: reader,
        writer: writer,
    }
}

func (s *CrudService) Get(ctx context.Context, id uint) (*User, error) {
    log := logger.GetLogger(zap.String("service", "CrudService"), zap.String("method", "Get"))

    if id == 0 {
        log.Error("Invalid ID provided")
        return nil, fmt.Errorf("%w: id is required", ErrInvalidInput)
    }

    user, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get user", zap.Error(err), zap.Uint("id", id))
        return nil, fmt.Errorf("get user: %w", err)
    }

    return user, nil
}

func (s *CrudService) Create(ctx context.Context, input CreateUserInput) (*User, error) {
    log := logger.GetLogger(zap.String("service", "CrudService"), zap.String("method", "Create"))

    // Validate input
    if input.Email == "" {
        return nil, fmt.Errorf("%w: email is required", ErrInvalidInput)
    }
    if input.Username == "" {
        return nil, fmt.Errorf("%w: username is required", ErrInvalidInput)
    }
    if input.Password == "" {
        return nil, fmt.Errorf("%w: password is required", ErrInvalidInput)
    }
    if input.FirstName == "" {
        return nil, fmt.Errorf("%w: first_name is required", ErrInvalidInput)
    }
    if input.LastName == "" {
        return nil, fmt.Errorf("%w: last_name is required", ErrInvalidInput)
    }
    if input.Role == "" {
        return nil, fmt.Errorf("%w: role is required", ErrInvalidInput)
    }

    // Validate role
    if input.Role != RoleAdmin && input.Role != RoleEditor && input.Role != RoleUser {
        return nil, fmt.Errorf("%w: invalid role", ErrInvalidInput)
    }

    // Check if user already exists
    existingUser, err := s.reader.GetByEmail(ctx, input.Email)
    if err != nil && err != ErrUserNotFound {
        log.Error("Error checking existing user", zap.Error(err))
        return nil, fmt.Errorf("check existing user: %w", err)
    }
    
    if existingUser != nil {
        return nil, ErrUserAlreadyExists
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        log.Error("Error hashing password", zap.Error(err))
        return nil, fmt.Errorf("hash password: %w", err)
    }

    // Create entity
    now := time.Now()
    user := &User{
        Email:            input.Email,
        Username:         input.Username,
        PasswordHash:     string(hashedPassword),
        FirstName:        input.FirstName,
        LastName:         input.LastName,
        Role:             input.Role,
        IsActive:         true,
        EmailVerified:    false,
        MagicLinkEnabled: input.MagicLinkEnabled,
        TOTPEnabled:      false,
        CreatedAt:        now,
        UpdatedAt:        now,
    }

    // Save to database
    if err := s.writer.Create(ctx, user); err != nil {
        log.Error("Failed to create user", 
            zap.Error(err), 
            zap.String("email", input.Email))
        return nil, fmt.Errorf("create user: %w", err)
    }

    log.Info("User created successfully", zap.Uint("id", user.ID), zap.String("email", input.Email))
    return user, nil
}

func (s *CrudService) Update(ctx context.Context, id uint, input UpdateUserInput) (*User, error) {
    log := logger.GetLogger(zap.String("service", "CrudService"), zap.String("method", "Update"))
    
    // Get existing user
    user, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get user for update", zap.Error(err), zap.Uint("id", id))
        return nil, fmt.Errorf("get user: %w", err)
    }

    // Update fields if provided
    updated := false

    if input.Username != nil {
        user.Username = *input.Username
        updated = true
    }

    if input.FirstName != nil {
        user.FirstName = *input.FirstName
        updated = true
    }

    if input.LastName != nil {
        user.LastName = *input.LastName
        updated = true
    }

    if input.MagicLinkEnabled != nil {
        user.MagicLinkEnabled = *input.MagicLinkEnabled
        updated = true
    }

    // If no changes, return user without update
    if !updated {
        log.Info("No changes to update", zap.Uint("id", id))
        return user, nil
    }

    // Update timestamp
    user.UpdatedAt = time.Now()

    // Save to database
    if err := s.writer.Update(ctx, user); err != nil {
        log.Error("Failed to update user", zap.Error(err), zap.Uint("id", id))
        return nil, fmt.Errorf("update user: %w", err)
    }

    log.Info("User updated successfully", zap.Uint("id", id))
    return user, nil
}

func (s *CrudService) AdminUpdate(ctx context.Context, id uint, input AdminUpdateUserInput) (*User, error) {
    log := logger.GetLogger(zap.String("service", "CrudService"), zap.String("method", "AdminUpdate"))
    
    // Get existing user
    user, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get user for admin update", zap.Error(err), zap.Uint("id", id))
        return nil, fmt.Errorf("get user: %w", err)
    }

    // Update fields if provided
    updated := false

    if input.Username != nil {
        user.Username = *input.Username
        updated = true
    }

    if input.FirstName != nil {
        user.FirstName = *input.FirstName
        updated = true
    }

    if input.LastName != nil {
        user.LastName = *input.LastName
        updated = true
    }

    if input.Role != nil {
        // Validate role
        if *input.Role != RoleAdmin && *input.Role != RoleEditor && *input.Role != RoleUser {
            return nil, fmt.Errorf("%w: invalid role", ErrInvalidInput)
        }
        user.Role = *input.Role
        updated = true
    }

    if input.IsActive != nil {
        user.IsActive = *input.IsActive
        updated = true
    }

    if input.EmailVerified != nil {
        user.EmailVerified = *input.EmailVerified
        updated = true
    }

    if input.MagicLinkEnabled != nil {
        user.MagicLinkEnabled = *input.MagicLinkEnabled
        updated = true
    }

    // If no changes, return user without update
    if !updated {
        log.Info("No changes to update", zap.Uint("id", id))
        return user, nil
    }

    // Update timestamp
    user.UpdatedAt = time.Now()

    // Save to database
    if err := s.writer.Update(ctx, user); err != nil {
        log.Error("Failed to admin update user", zap.Error(err), zap.Uint("id", id))
        return nil, fmt.Errorf("admin update user: %w", err)
    }

    log.Info("User admin updated successfully", zap.Uint("id", id))
    return user, nil
}

func (s *CrudService) Delete(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("service", "CrudService"), zap.String("method", "Delete"))

    // Verify user exists
    _, err := s.reader.GetByID(ctx, id)
    if err != nil {
        log.Error("Failed to get user for deletion", zap.Error(err), zap.Uint("id", id))
        return fmt.Errorf("get user: %w", err)
    }

    if err := s.writer.Delete(ctx, id); err != nil {
        log.Error("Failed to delete user", zap.Error(err), zap.Uint("id", id))
        return fmt.Errorf("delete user: %w", err)
    }

    log.Info("User deleted successfully", zap.Uint("id", id))
    return nil
}

func (s *CrudService) DeletePermanently(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("service", "CrudService"), zap.String("method", "DeletePermanently"))

    // Verify user exists
    _, err := s.reader.GetByID(ctx, id)
    if err != nil {
        // If soft-deleted, delete anyway
        if err != ErrUserNotFound {
            log.Error("Failed to check user for permanent deletion", zap.Error(err), zap.Uint("id", id))
            return fmt.Errorf("check user: %w", err)
        }
    }

    if err := s.writer.DeletePermanently(ctx, id); err != nil {
        log.Error("Failed to delete user permanently", zap.Error(err), zap.Uint("id", id))
        return fmt.Errorf("delete user permanently: %w", err)
    }

    log.Info("User permanently deleted", zap.Uint("id", id))
    return nil
}

func (s *CrudService) Restore(ctx context.Context, id uint) error {
    log := logger.GetLogger(zap.String("service", "CrudService"), zap.String("method", "Restore"))

    if err := s.writer.Restore(ctx, id); err != nil {
        log.Error("Failed to restore user", zap.Error(err), zap.Uint("id", id))
        return fmt.Errorf("restore user: %w", err)
    }

    log.Info("User restored successfully", zap.Uint("id", id))
    return nil
}
