package main

import (
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "go.uber.org/zap"

    "github.com/TommySanDev/gachiakuta-hispano/config"
    "github.com/TommySanDev/gachiakuta-hispano/internal/character"
    "github.com/TommySanDev/gachiakuta-hispano/internal/logger"
    customMiddleware "github.com/TommySanDev/gachiakuta-hispano/internal/middleware"
    "github.com/TommySanDev/gachiakuta-hispano/internal/postgres"
    "github.com/TommySanDev/gachiakuta-hispano/internal/user"
    "github.com/TommySanDev/gachiakuta-hispano/internal/vitalinstrument"
)

func main() {
    // Initialize logger
    logger.Init("development")
    log := logger.GetLogger(zap.String("component", "main"))
    log.Info("Starting Gachiakuta Hispano server")

    // Load configurations
    authConfig := config.LoadAuthConfig()
    smtpConfig := config.LoadSMTPConfig()

    // Database connection
    db := config.ConnectDB()
    defer db.Close()

    // Initialize character repositories
    characterReader := postgres.NewCharacterReader(db)
    characterWriter := postgres.NewCharacterWriter(db)

    // Initialize vital instrument repositories
    vitalInstrumentReader := postgres.NewVitalInstrumentReader(db)
    vitalInstrumentWriter := postgres.NewVitalInstrumentWriter(db)

    // Initialize user repositories
    userReader := postgres.NewUserReader(db)
    userWriter := postgres.NewUserWriter(db)
    sessionReader := postgres.NewSessionReader(db)
    sessionWriter := postgres.NewSessionWriter(db)
    magicLinkReader := postgres.NewMagicLinkReader(db)
    magicLinkWriter := postgres.NewMagicLinkWriter(db)
    resetTokenReader := postgres.NewResetTokenReader(db)
    resetTokenWriter := postgres.NewResetTokenWriter(db)
    totpReader := postgres.NewTOTPReader(db)
    totpWriter := postgres.NewTOTPWriter(db)

    // Initialize character services
    characterCrudService := character.NewCrudService(characterReader, characterWriter)
    characterSearchService := character.NewSearchService(characterReader)

    // Initialize vital instrument services
    vitalInstrumentCrudService := vitalinstrument.NewCrudService(vitalInstrumentReader, vitalInstrumentWriter)
    vitalInstrumentSearchService := vitalinstrument.NewSearchService(vitalInstrumentReader)

    // Initialize user services
    var emailService *user.EmailService
    if smtpConfig != nil {
        emailConfig := config.GetEmailConfig()
        emailService = user.NewEmailService(emailConfig)
        log.Info("Email service initialized")
    } else {
        log.Warn("Email service disabled - SMTP configuration not found")
    }

    userAuthService := user.NewAuthService(
        userReader,
        userWriter,
        sessionWriter,
        magicLinkReader,
        magicLinkWriter,
        emailService,
    )
    userCrudService := user.NewCrudService(userReader, userWriter)
    userSearchService := user.NewSearchService(userReader)
    userPasswordService := user.NewPasswordService(
        userReader,
        userWriter,
        resetTokenReader,
        resetTokenWriter,
        emailService,
    )
    userTOTPService := user.NewTOTPService(
        userReader,
        userWriter,
        totpReader,
        totpWriter,
    )

    // Initialize handlers
    characterHandler := character.NewHandler(characterCrudService, characterSearchService)
    vitalInstrumentHandler := vitalinstrument.NewHandler(vitalInstrumentCrudService, vitalInstrumentSearchService)
    userHandler := user.NewHandler(userAuthService, userCrudService, userSearchService, userPasswordService)

    // Initialize authentication middleware
    authMiddleware := customMiddleware.NewAuthMiddleware(sessionReader, userReader)

    // Create router
    router := chi.NewRouter()

    // Common middleware
    router.Use(middleware.RequestID)
    router.Use(middleware.RealIP)
    router.Use(customMiddleware.RequestLogger)
    router.Use(middleware.Recoverer)
    router.Use(middleware.Timeout(30 * time.Second))

    // Health check endpoint
    router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"ok","auth_enabled":true}`))
    })

    // API routes
    router.Route("/api/v1", func(r chi.Router) {
        // Register character routes
        characterHandler.RegisterRoutes(r)
        
        // Register vital instrument routes
        vitalInstrumentHandler.RegisterRoutes(r)
        
        // Register user routes with authentication
        userHandler.RegisterRoutes(r, authMiddleware)
    })

    // Start server
    log.Info("Server running at http://localhost:8080",
        zap.Bool("auth_enabled", true),
        zap.Bool("email_enabled", smtpConfig != nil),
        zap.Bool("magic_link_enabled", authConfig.MagicLinkEnabled),
        zap.Bool("password_reset_enabled", smtpConfig != nil),
    )
    
    if err := http.ListenAndServe(":8080", router); err != nil {
        log.Fatal("Server startup error", zap.Error(err))
    }
}
