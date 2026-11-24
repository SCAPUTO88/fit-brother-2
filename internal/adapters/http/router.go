package http

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/handler"
	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	"github.com/SCAPUTO88/fit-brother/internal/adapters/repository/postgres"
	"github.com/SCAPUTO88/fit-brother/internal/adapters/storage"
	"github.com/SCAPUTO88/fit-brother/internal/application/auth"
	"github.com/SCAPUTO88/fit-brother/internal/application/student"
	appvideo "github.com/SCAPUTO88/fit-brother/internal/application/video"
	appworkout "github.com/SCAPUTO88/fit-brother/internal/application/workout"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

// RouterDeps permite substituir adapters externos (ex.: storage) em testes de integração.
type RouterDeps struct {
	VideoStorage appvideo.StorageClient
}

func SetupRouter(cfg *config.Config, db *sqlx.DB) (*gin.Engine, error) {
	return SetupRouterWithDeps(cfg, db, RouterDeps{})
}

func SetupRouterWithDeps(cfg *config.Config, db *sqlx.DB, deps RouterDeps) (*gin.Engine, error) {
	if cfg.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.SecurityHeaders())
	router.Use(cors.New(newCORSConfig(cfg.CORS.AllowedOrigins)))

	passwordSvc, err := crypto.NewPasswordService(cfg.Crypto.PasswordPepper)
	if err != nil {
		return nil, fmt.Errorf("falha configurando auth crypto: %w", err)
	}

	jwtSvc, err := newJWTService(cfg.JWT)
	if err != nil {
		return nil, fmt.Errorf("falha configurando jwt service: %w", err)
	}

	aesEncryptor, err := crypto.NewAESEncryptor(cfg.Crypto.AESKey)
	if err != nil {
		return nil, fmt.Errorf("falha configurando AES: %w", err)
	}

	hashGen, err := crypto.NewHash(cfg.Crypto.CPFHMACSecret)
	if err != nil {
		return nil, fmt.Errorf("falha configurando hash gen: %w", err)
	}

	txManager := postgres.NewTxManager(db)
	userRepo := postgres.NewUserRepository(db)
	studentRepo := postgres.NewStudentRepository(db)
	workoutRepo := postgres.NewWorkoutRepository(db)
	videoRepo := postgres.NewVideoRepository(db)

	videoStorage := deps.VideoStorage
	if videoStorage == nil {
		r2Storage, err := storage.NewR2Storage(&cfg.R2)
		if err != nil {
			return nil, fmt.Errorf("falha configurando R2 storage: %w", err)
		}
		videoStorage = r2Storage
	}

	loginUC := auth.NewLoginUseCase(userRepo, passwordSvc, jwtSvc, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	refreshUC := auth.NewRefreshUseCase(userRepo, txManager, jwtSvc, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	logoutUC := auth.NewLogoutUseCase(userRepo)

	createStudentUC := student.NewCreateStudentUseCase(userRepo, studentRepo, txManager, passwordSvc, aesEncryptor, hashGen)
	getStudentUC := student.NewGetStudentUseCase(userRepo, studentRepo, aesEncryptor)
	listStudentsUC := student.NewListStudentsUseCase(studentRepo)
	anonymizeStudentUC := student.NewAnonymizeStudentUseCase(userRepo, studentRepo, txManager)

	createPlanUC := appworkout.NewCreatePlanUseCase(workoutRepo, studentRepo)
	createSessionUC := appworkout.NewCreateSessionUseCase(workoutRepo)
	createExerciseUC := appworkout.NewCreateExerciseUseCase(workoutRepo)
	listPlansUC := appworkout.NewListPlansUseCase(workoutRepo, studentRepo)
	getPlanDetailsUC := appworkout.NewGetPlanDetailsUseCase(workoutRepo, studentRepo)

	requestUploadURLUC := appvideo.NewRequestUploadURLUseCase(videoRepo, workoutRepo, videoStorage, &cfg.R2)
	confirmUploadUC := appvideo.NewConfirmUploadUseCase(videoRepo, workoutRepo, videoStorage)
	getSignedURLUC := appvideo.NewGetSignedURLUseCase(videoRepo, workoutRepo, studentRepo, videoStorage, &cfg.R2)

	authHandler := handler.NewAuthHandler(loginUC, refreshUC, logoutUC)
	studentHandler := handler.NewStudentHandler(createStudentUC, getStudentUC, listStudentsUC, anonymizeStudentUC)
	workoutHandler := handler.NewWorkoutHandler(createPlanUC, createSessionUC, createExerciseUC, listPlansUC, getPlanDetailsUC)
	videoHandler := handler.NewVideoHandler(requestUploadURLUC, confirmUploadUC, getSignedURLUC)

	authMiddleware := middleware.RequireAuth(jwtSvc, userRepo)

	v1 := router.Group("/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"version": "1.0",
				"env":     cfg.App.Env,
			})
		})

		authHandler.RegisterRoutes(v1)
		studentHandler.RegisterRoutes(v1, authMiddleware)
		workoutHandler.RegisterRoutes(v1, authMiddleware)
		videoHandler.RegisterRoutes(v1, authMiddleware)
	}

	return router, nil
}

func newJWTService(cfg config.JWTConfig) (*jwt.Service, error) {
	if cfg.PrivateKeyPEM != "" || cfg.PublicKeyPEM != "" {
		if cfg.PrivateKeyPEM == "" || cfg.PublicKeyPEM == "" {
			return nil, fmt.Errorf("JWT_PRIVATE_KEY_PEM e JWT_PUBLIC_KEY_PEM devem ser definidos juntos")
		}
		return jwt.NewServiceFromPEM(cfg.PrivateKeyPEM, cfg.PublicKeyPEM, "fit-brother-api")
	}

	return jwt.NewService(cfg.PrivateKeyPath, cfg.PublicKeyPath, "fit-brother-api")
}

func newCORSConfig(allowedOrigins string) cors.Config {
	origins := make([]string, 0)
	for _, origin := range strings.Split(allowedOrigins, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}

	return cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}
