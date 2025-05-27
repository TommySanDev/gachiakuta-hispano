package user

import (
    "context"
    "crypto/rand"
    "encoding/base32"
    "fmt"
    "strings"
    "time"

    "github.com/pquerna/otp"
    "github.com/pquerna/otp/totp"
    "github.com/skip2/go-qrcode"
    "go.uber.org/zap"
    
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
)

// Handles TOTP 2FA operations
type TOTPService struct {
    userReader   UserGetter
    userWriter   UserWriter
    totpReader   TOTPGetter
    totpWriter   TOTPWriter
}

func NewTOTPService(
    userReader UserGetter,
    userWriter UserWriter,
    totpReader TOTPGetter,
    totpWriter TOTPWriter,
) *TOTPService {
    return &TOTPService{
        userReader: userReader,
        userWriter: userWriter,
        totpReader: totpReader,
        totpWriter: totpWriter,
    }
}

func (s *TOTPService) SetupTOTP(ctx context.Context, userID uint) (*TOTPSetupResponse, error) {
    log := logger.GetLogger(zap.String("service", "TOTPService"), zap.String("method", "SetupTOTP"))
    
    // Get user
    user, err := s.userReader.GetByID(ctx, userID)
    if err != nil {
        log.Error("Error getting user", zap.Error(err))
        return nil, fmt.Errorf("get user: %w", err)
    }
    
    // Check if TOTP already exists
    existingTOTP, err := s.totpReader.GetByUserID(ctx, userID)
    if err != nil && err != ErrTOTPNotFound {
        log.Error("Error checking existing TOTP", zap.Error(err))
        return nil, fmt.Errorf("check existing TOTP: %w", err)
    }
    
    if existingTOTP != nil && existingTOTP.Verified {
        return nil, ErrTOTPAlreadyEnabled
    }
    
    // Generate secret
    secret, err := s.generateSecret()
    if err != nil {
        log.Error("Error generating TOTP secret", zap.Error(err))
        return nil, fmt.Errorf("generate secret: %w", err)
    }
    
    // Create TOTP key
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "Gachiakuta Hispano",
        AccountName: user.Email,
        Secret:      []byte(secret),
    })
    if err != nil {
        log.Error("Error generating TOTP key", zap.Error(err))
        return nil, fmt.Errorf("generate TOTP key: %w", err)
    }
    
    // Generate QR code
    qrCode, err := qrcode.Encode(key.String(), qrcode.Medium, 256)
    if err != nil {
        log.Error("Error generating QR code", zap.Error(err))
        return nil, fmt.Errorf("generate QR code: %w", err)
    }
    
    // Save TOTP secret (unverified)
    now := time.Now()
    totpSecret := &TOTPSecret{
        UserID:    userID,
        Secret:    secret,
        Verified:  false,
        CreatedAt: now,
        UpdatedAt: now,
    }
    
    if existingTOTP != nil {
        // Delete existing unverified TOTP
        err = s.totpWriter.DeleteTOTPSecret(ctx, userID)
        if err != nil {
            log.Warn("Error deleting existing unverified TOTP", zap.Error(err))
        }
    }
    
    err = s.totpWriter.CreateTOTPSecret(ctx, totpSecret)
    if err != nil {
        log.Error("Error saving TOTP secret", zap.Error(err))
        return nil, fmt.Errorf("save TOTP secret: %w", err)
    }
    
    response := &TOTPSetupResponse{
        Secret:    secret,
        QRCode:    qrCode,
        BackupURL: key.String(),
    }
    
    log.Info("TOTP setup initiated", zap.Uint("user_id", userID))
    return response, nil
}

func (s *TOTPService) VerifyTOTP(ctx context.Context, userID uint, code string) error {
    log := logger.GetLogger(zap.String("service", "TOTPService"), zap.String("method", "VerifyTOTP"))
    
    // Get TOTP secret
    totpSecret, err := s.totpReader.GetByUserID(ctx, userID)
    if err != nil {
        if err == ErrTOTPNotFound {
            return ErrTOTPNotEnabled
        }
        log.Error("Error getting TOTP secret", zap.Error(err))
        return fmt.Errorf("get TOTP secret: %w", err)
    }
    
    // Validate code
    valid := totp.Validate(code, totpSecret.Secret, time.Now())
    if !valid {
        return ErrInvalidTOTPCode
    }
    
    // If not verified yet, verify and enable 2FA
    if !totpSecret.Verified {
        err = s.totpWriter.VerifyTOTPSecret(ctx, userID)
        if err != nil {
            log.Error("Error verifying TOTP secret", zap.Error(err))
            return fmt.Errorf("verify TOTP secret: %w", err)
        }
        
        // Enable TOTP for user
        user, err := s.userReader.GetByID(ctx, userID)
        if err != nil {
            log.Error("Error getting user for TOTP enable", zap.Error(err))
            return fmt.Errorf("get user: %w", err)
        }
        
        user.TOTPEnabled = true
        user.UpdatedAt = time.Now()
        
        err = s.userWriter.Update(ctx, user)
        if err != nil {
            log.Error("Error enabling TOTP for user", zap.Error(err))
            return fmt.Errorf("enable TOTP: %w", err)
        }
        
        // Generate recovery codes
        err = s.generateRecoveryCodes(ctx, userID)
        if err != nil {
            log.Error("Error generating recovery codes", zap.Error(err))
            // Don't fail TOTP verification for this
        }
        
        log.Info("TOTP enabled successfully", zap.Uint("user_id", userID))
    }
    
    return nil
}

func (s *TOTPService) ValidateTOTP(ctx context.Context, userID uint, code string) error {
    log := logger.GetLogger(zap.String("service", "TOTPService"), zap.String("method", "ValidateTOTP"))
    
    // Get TOTP secret
    totpSecret, err := s.totpReader.GetByUserID(ctx, userID)
    if err != nil {
        if err == ErrTOTPNotFound {
            return ErrTOTPNotEnabled
        }
        log.Error("Error getting TOTP secret", zap.Error(err))
        return fmt.Errorf("get TOTP secret: %w", err)
    }
    
    if !totpSecret.Verified {
        return ErrTOTPNotEnabled
    }
    
    // Validate code
    valid := totp.Validate(code, totpSecret.Secret, time.Now())
    if !valid {
        return ErrInvalidTOTPCode
    }
    
    return nil
}

func (s *TOTPService) DisableTOTP(ctx context.Context, userID uint) error {
    log := logger.GetLogger(zap.String("service", "TOTPService"), zap.String("method", "DisableTOTP"))
    
    // Delete TOTP secret
    err := s.totpWriter.DeleteTOTPSecret(ctx, userID)
    if err != nil && err != ErrTOTPNotFound {
        log.Error("Error deleting TOTP secret", zap.Error(err))
        return fmt.Errorf("delete TOTP secret: %w", err)
    }
    
    // Delete recovery codes
    err = s.totpWriter.DeleteRecoveryCodes(ctx, userID)
    if err != nil {
        log.Warn("Error deleting recovery codes", zap.Error(err))
    }
    
    // Disable TOTP for user
    user, err := s.userReader.GetByID(ctx, userID)
    if err != nil {
        log.Error("Error getting user for TOTP disable", zap.Error(err))
        return fmt.Errorf("get user: %w", err)
    }
    
    user.TOTPEnabled = false
    user.UpdatedAt = time.Now()
    
    err = s.userWriter.Update(ctx, user)
    if err != nil {
        log.Error("Error disabling TOTP for user", zap.Error(err))
        return fmt.Errorf("disable TOTP: %w", err)
    }
    
    log.Info("TOTP disabled successfully", zap.Uint("user_id", userID))
    return nil
}

func (s *TOTPService) GenerateRecoveryCodes(ctx context.Context, userID uint) ([]string, error) {
    log := logger.GetLogger(zap.String("service", "TOTPService"), zap.String("method", "GenerateRecoveryCodes"))
    
    // Delete existing codes
    err := s.totpWriter.DeleteRecoveryCodes(ctx, userID)
    if err != nil {
        log.Warn("Error deleting existing recovery codes", zap.Error(err))
    }
    
    // Generate new codes
    err = s.generateRecoveryCodes(ctx, userID)
    if err != nil {
        log.Error("Error generating recovery codes", zap.Error(err))
        return nil, fmt.Errorf("generate recovery codes: %w", err)
    }
    
    // Get codes for return
    codes, err := s.totpReader.GetUnusedRecoveryCodesByUserID(ctx, userID)
    if err != nil {
        log.Error("Error getting recovery codes", zap.Error(err))
        return nil, fmt.Errorf("get recovery codes: %w", err)
    }
    
    result := make([]string, len(codes))
    for i, code := range codes {
        result[i] = code.Code
    }
    
    log.Info("Recovery codes generated", zap.Uint("user_id", userID), zap.Int("count", len(result)))
    return result, nil
}

func (s *TOTPService) UseRecoveryCode(ctx context.Context, userID uint, code string) error {
    log := logger.GetLogger(zap.String("service", "TOTPService"), zap.String("method", "UseRecoveryCode"))
    
    err := s.totpWriter.UseRecoveryCode(ctx, userID, code)
    if err != nil {
        if err == ErrInvalidRecoveryCode {
            return err
        }
        log.Error("Error using recovery code", zap.Error(err))
        return fmt.Errorf("use recovery code: %w", err)
    }
    
    log.Info("Recovery code used successfully", zap.Uint("user_id", userID))
    return nil
}

// Helper methods
func (s *TOTPService) generateSecret() (string, error) {
    secret := make([]byte, 20)
    _, err := rand.Read(secret)
    if err != nil {
        return "", err
    }
    return base32.StdEncoding.EncodeToString(secret), nil
}

func (s *TOTPService) generateRecoveryCodes(ctx context.Context, userID uint) error {
    codes := make([]*RecoveryCode, 8) // Generate 8 recovery codes
    now := time.Now()
    
    for i := 0; i < 8; i++ {
        code, err := s.generateRecoveryCode()
        if err != nil {
            return fmt.Errorf("generate recovery code: %w", err)
        }
        
        codes[i] = &RecoveryCode{
            UserID:    userID,
            Code:      code,
            Used:      false,
            CreatedAt: now,
            UpdatedAt: now,
        }
    }
    
    return s.totpWriter.CreateRecoveryCodes(ctx, codes)
}

func (s *TOTPService) generateRecoveryCode() (string, error) {
    bytes := make([]byte, 6)
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }
    
    code := fmt.Sprintf("%02x%02x-%02x%02x-%02x%02x", 
        bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5])
    return strings.ToUpper(code), nil
}
