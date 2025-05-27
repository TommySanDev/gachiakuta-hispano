package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/TommySanDev/gachiakuta-hispano/config"
	"github.com/TommySanDev/gachiakuta-hispano/internal/chapter"
	"github.com/TommySanDev/gachiakuta-hispano/internal/character"
	"github.com/TommySanDev/gachiakuta-hispano/internal/comment"
	"github.com/TommySanDev/gachiakuta-hispano/internal/favorite"
	"github.com/TommySanDev/gachiakuta-hispano/internal/logger"
	customMiddleware "github.com/TommySanDev/gachiakuta-hispano/internal/middleware"
	"github.com/TommySanDev/gachiakuta-hispano/internal/postgres"
	"github.com/TommySanDev/gachiakuta-hispano/internal/user"
	"github.com/TommySanDev/gachiakuta-hispano/internal/vitalinstrument"
)

func main() {
	logger.Init("development")
	log := logger.GetLogger(zap.String("component", "main"))
	log.Info("Starting Gachiakuta Hispano server")

	authConfig := config.LoadAuthConfig()
	smtpConfig := config.LoadSMTPConfig()

	db := config.ConnectDB()
	defer db.Close()

	characterReader := postgres.NewCharacterReader(db)
	characterWriter := postgres.NewCharacterWriter(db)

	vitalInstrumentReader := postgres.NewVitalInstrumentReader(db)
	vitalInstrumentWriter := postgres.NewVitalInstrumentWriter(db)

	chapterReader := postgres.NewChapterReader(db)
	chapterWriter := postgres.NewChapterWriter(db)

	commentReader := postgres.NewCommentReader(db)
	commentWriter := postgres.NewCommentWriter(db)

	favoriteReader := postgres.NewFavoriteReader(db)
	favoriteWriter := postgres.NewFavoriteWriter(db)

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

	characterCrudService := character.NewCrudService(characterReader, characterWriter)
	characterSearchService := character.NewSearchService(characterReader)

	vitalInstrumentCrudService := vitalinstrument.NewCrudService(vitalInstrumentReader, vitalInstrumentWriter)
	vitalInstrumentSearchService := vitalinstrument.NewSearchService(vitalInstrumentReader)

	chapterCrudService := chapter.NewCrudService(chapterReader, chapterWriter)
	chapterSearchService := chapter.NewSearchService(chapterReader)

	commentCrudService := comment.NewCrudService(commentReader, commentWriter)
	favoriteCrudService := favorite.NewCrudService(favoriteReader, favoriteWriter)

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

	characterHandler := character.NewHandler(characterCrudService, characterSearchService)
	vitalInstrumentHandler := vitalinstrument.NewHandler(vitalInstrumentCrudService, vitalInstrumentSearchService)
	chapterHandler := chapter.NewHandler(chapterCrudService, chapterSearchService)
	commentHandler := comment.NewHandler(commentCrudService)
	favoriteHandler := favorite.NewHandler(favoriteCrudService)
	userHandler := user.NewHandler(userAuthService, userCrudService, userSearchService, userPasswordService, userTOTPService)

  userAdapter := user.NewUserAdapter(userReader)
  sessionAdapter := user.NewSessionAdapter(sessionReader)

  authMiddleware := middleware.NewAuthMiddleware(sessionAdapter, userAdapter)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(customMiddleware.RequestLogger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","auth_enabled":true}`))
	})

	router.Route("/api/v1", func(r chi.Router) {
		// Public content
		characterHandler.RegisterRoutes(r)
		chapterHandler.RegisterRoutes(r)
		vitalInstrumentHandler.RegisterRoutes(r)
		commentHandler.RegisterRoutes(r)

		// Authenticated features
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			favoriteHandler.RegisterRoutes(r)
		})

		// Content and comment moderation
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Use(authMiddleware.RequireAnyRole("editor", "admin"))
			characterHandler.RegisterRoutes(r)
			chapterHandler.RegisterRoutes(r)
			vitalInstrumentHandler.RegisterRoutes(r)
			commentHandler.RegisterRoutes(r)
		})

		// Admin-only access
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Use(authMiddleware.RequireRole("admin"))
			userHandler.RegisterRoutes(r, authMiddleware)
		})
	})

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

