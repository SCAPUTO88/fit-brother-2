package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/handler"
	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	"github.com/SCAPUTO88/fit-brother/internal/adapters/repository/postgres"
	"github.com/SCAPUTO88/fit-brother/internal/application/auth"
	"github.com/SCAPUTO88/fit-brother/internal/application/student"
	appworkout "github.com/SCAPUTO88/fit-brother/internal/application/workout"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

func SetupRouter(cfg *config.Config, db *sqlx.DB) (*gin.Engine, error) {
	if cfg.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	passwordSvc, err := crypto.NewPasswordService(cfg.Crypto.PasswordPepper)
	if err != nil {
		return nil, fmt.Errorf("falha configurando auth crypto: %w", err)
	}

	jwtSvc, err := jwt.NewService(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath, "fit-brother-api")
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

	userRepo := postgres.NewUserRepository(db)
	studentRepo := postgres.NewStudentRepository(db)
	workoutRepo := postgres.NewWorkoutRepository(db)

	loginUC := auth.NewLoginUseCase(userRepo, passwordSvc, jwtSvc, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	refreshUC := auth.NewRefreshUseCase(userRepo, jwtSvc, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	logoutUC := auth.NewLogoutUseCase(userRepo)

	createStudentUC := student.NewCreateStudentUseCase(userRepo, studentRepo, passwordSvc, aesEncryptor, hashGen)
	getStudentUC := student.NewGetStudentUseCase(userRepo, studentRepo, aesEncryptor)
	anonymizeStudentUC := student.NewAnonymizeStudentUseCase(userRepo, studentRepo)

	createPlanUC := appworkout.NewCreatePlanUseCase(workoutRepo)
	createSessionUC := appworkout.NewCreateSessionUseCase(workoutRepo)
	createExerciseUC := appworkout.NewCreateExerciseUseCase(workoutRepo)
	listPlansUC := appworkout.NewListPlansUseCase(workoutRepo)

	authHandler := handler.NewAuthHandler(loginUC, refreshUC, logoutUC)
	studentHandler := handler.NewStudentHandler(createStudentUC, getStudentUC, anonymizeStudentUC)
	workoutHandler := handler.NewWorkoutHandler(createPlanUC, createSessionUC, createExerciseUC, listPlansUC)

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
		studentHandler.RegisterRoutes(v1, middleware.RequireAuth(jwtSvc))
		workoutHandler.RegisterRoutes(v1, middleware.RequireAuth(jwtSvc))
	}

	return router, nil
}
