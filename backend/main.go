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
    "github.com/TommySanDev/gachiakuta-hispano/internal/vitalinstrument"
)

func main() {
    // Initialize logger
    logger.Init("development")
    log := logger.GetLogger(zap.String("component", "main"))
    log.Info("Starting Gachiakuta Hispano server")

    // Database connection
    db := config.ConnectDB()
    defer db.Close()

    // Initialize character repositories
    characterReader := postgres.NewCharacterReader(db)
    characterWriter := postgres.NewCharacterWriter(db)

    // Initialize vital instrument repositories
    vitalInstrumentReader := postgres.NewVitalInstrumentReader(db)
    vitalInstrumentWriter := postgres.NewVitalInstrumentWriter(db)

    // Initialize services
    characterCrudService := character.NewCrudService(characterReader, characterWriter)
    characterSearchService := character.NewSearchService(characterReader)

    vitalInstrumentCrudService := vitalinstrument.NewCrudService(vitalInstrumentReader, vitalInstrumentWriter)
    vitalInstrumentSearchService := vitalinstrument.NewSearchService(vitalInstrumentReader)

    // Initialize handlers
    characterHandler := character.NewHandler(characterCrudService, characterSearchService)
    vitalInstrumentHandler := vitalinstrument.NewHandler(vitalInstrumentCrudService, vitalInstrumentSearchService)

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
        w.Write([]byte(`{"status":"ok"}`))
    })

    // API routes
    router.Route("/api/v1", func(r chi.Router) {
        // Register character routes
        characterHandler.RegisterRoutes(r)
        
        // Register vital instrument routes
        vitalInstrumentHandler.RegisterRoutes(r)
    })

    // Start server
    log.Info("Server running at http://localhost:8080")
    if err := http.ListenAndServe(":8080", router); err != nil {
        log.Fatal("Server startup error", zap.Error(err))
    }
}
