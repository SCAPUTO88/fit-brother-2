package http

import (
	"crypto/sha256"
	"errors"
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
	domainStorage "github.com/SCAPUTO88/fit-brother/internal/domain/storage"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

// RouterDeps permite substituir adapters externos (ex.: storage) em testes de integração.
type RouterDeps struct {
	VideoStorage domainStorage.Port
}

func SetupRouter(cfg *config.Config, db *sqlx.DB) (*gin.Engine, error) {
	return SetupRouterWithDeps(cfg, db, RouterDeps{})
}

func SetupRouterWithDeps(cfg *config.Config, db *sqlx.DB, deps RouterDeps) (*gin.Engine, error) {
	if cfg.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	corsCfg, err := newCORSConfig(cfg.CORS.AllowedOrigins)
	if err != nil {
		return nil, fmt.Errorf("falha configurando CORS: %w", err)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.SecurityHeaders())
	router.Use(cors.New(corsCfg))
	router.Use(middleware.CSRFProtection(corsCfg.AllowOrigins))

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

	videoStorage, localStore, err := resolveVideoStorage(cfg, deps)
	if err != nil {
		return nil, fmt.Errorf("falha configurando storage: %w", err)
	}

	signedURLTTL := cfg.Storage.SignedURLTTL
	if signedURLTTL <= 0 {
		signedURLTTL = cfg.R2.SignedURLTTL
	}
	if signedURLTTL <= 0 {
		signedURLTTL = 5 * time.Minute
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

	requestUploadURLUC := appvideo.NewRequestUploadURLUseCase(videoRepo, workoutRepo, videoStorage, signedURLTTL)
	confirmUploadUC := appvideo.NewConfirmUploadUseCase(videoRepo, workoutRepo, videoStorage)
	getSignedURLUC := appvideo.NewGetSignedURLUseCase(videoRepo, workoutRepo, studentRepo, videoStorage, signedURLTTL)

	authHandler := handler.NewAuthHandler(loginUC, refreshUC, logoutUC, cfg.App.IsProduction(), cfg.JWT.RefreshTokenTTL)
	studentHandler := handler.NewStudentHandler(createStudentUC, getStudentUC, listStudentsUC, anonymizeStudentUC)
	workoutHandler := handler.NewWorkoutHandler(createPlanUC, createSessionUC, createExerciseUC, listPlansUC, getPlanDetailsUC)
	videoHandler := handler.NewVideoHandler(requestUploadURLUC, confirmUploadUC, getSignedURLUC)

	authLimiter := middleware.NewRateLimiter(cfg.RateLimit.AuthMaxRequests, cfg.RateLimit.AuthWindow, 5*time.Minute)
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

		authHandler.RegisterRoutes(v1, authLimiter.Middleware("auth"))
		studentHandler.RegisterRoutes(v1, authMiddleware)
		workoutHandler.RegisterRoutes(v1, authMiddleware)
		videoHandler.RegisterRoutes(v1, authMiddleware)

		if localStore != nil {
			localStorageHandler := handler.NewLocalStorageHandler(localStore)
			localStorageHandler.RegisterRoutes(v1)
		}
	}

	return router, nil
}

func resolveVideoStorage(cfg *config.Config, deps RouterDeps) (domainStorage.Port, *storage.LocalStorage, error) {
	if deps.VideoStorage != nil {
		if ls, ok := deps.VideoStorage.(*storage.LocalStorage); ok {
			return deps.VideoStorage, ls, nil
		}
		return deps.VideoStorage, nil, nil
	}

	driver := strings.ToLower(strings.TrimSpace(cfg.Storage.Driver))
	if driver == "" {
		if cfg.R2.HasCredentials() {
			driver = config.StorageDriverR2
		} else if cfg.App.IsProduction() {
			driver = config.StorageDriverDisabled
		} else {
			driver = config.StorageDriverLocal
		}
	}

	switch driver {
	case config.StorageDriverDisabled:
		return storage.NewDisabledStorage(), nil, nil
	case config.StorageDriverR2:
		r2Storage, err := storage.NewR2Storage(&cfg.R2)
		if err != nil {
			return nil, nil, err
		}
		return r2Storage, nil, nil
	case config.StorageDriverLocal:
		localDir := cfg.Storage.LocalDir
		if strings.TrimSpace(localDir) == "" {
			localDir = "./tmp/storage"
		}
		baseURL := cfg.App.BaseURL
		if strings.TrimSpace(baseURL) == "" {
			baseURL = "http://localhost:8080"
		}
		signingSecret := deriveLocalSigningSecret(cfg)
		localStore, err := storage.NewLocalStorage(localDir, baseURL, signingSecret)
		if err != nil {
			return nil, nil, err
		}
		return localStore, localStore, nil
	default:
		return nil, nil, fmt.Errorf("driver de storage desconhecido: %q", driver)
	}
}

func deriveLocalSigningSecret(cfg *config.Config) []byte {
	if strings.TrimSpace(cfg.Storage.SigningSecret) != "" {
		return []byte(cfg.Storage.SigningSecret)
	}
	sum := sha256.Sum256([]byte("fit-brother-local-storage:" + cfg.Crypto.AESKey))
	return sum[:]
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

func newCORSConfig(allowedOrigins string) (cors.Config, error) {
	origins := make([]string, 0)
	for _, origin := range strings.Split(allowedOrigins, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			if origin == "*" {
				return cors.Config{}, errors.New("wildcard '*' em CORS_ALLOWED_ORIGINS é proibido quando AllowCredentials=true")
			}
			origins = append(origins, origin)
		}
	}

	return cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", middleware.CSRFHeaderName, middleware.SessionModeHeaderName},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}, nil
}
