package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/handler"
	"github.com/SCAPUTO88/fit-brother/internal/adapters/repository/postgres"
	"github.com/SCAPUTO88/fit-brother/internal/application/auth"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

// SetupRouter injeta todas as dependências do projeto e retorna a Engine do Gin configurada.
func SetupRouter(cfg *config.Config, db *sqlx.DB) (*gin.Engine, error) {
	if cfg.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// ==========================================
	// INFRAESTRUTURA COMPARTILHADA
	// ==========================================
	passwordSvc, err := crypto.NewPasswordService(cfg.Crypto.PasswordPepper)
	if err != nil {
		return nil, fmt.Errorf("falha configurando auth crypto: %w", err)
	}

	jwtSvc, err := jwt.NewService(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath, "fit-brother-api")
	if err != nil {
		return nil, fmt.Errorf("falha configurando jwt service: %w", err)
	}

	// ==========================================
	// DOMÍNIO USER / AUTH
	// ==========================================
	userRepo := postgres.NewUserRepository(db)

	loginUC := auth.NewLoginUseCase(userRepo, passwordSvc, jwtSvc, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	refreshUC := auth.NewRefreshUseCase(userRepo, jwtSvc, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	logoutUC := auth.NewLogoutUseCase(userRepo)

	authHandler := handler.NewAuthHandler(loginUC, refreshUC, logoutUC)

	// ==========================================
	// REGISTRO DE ROTAS
	// ==========================================
	v1 := router.Group("/v1")
	{
		// Health Check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"version": "1.0",
				"env":     cfg.App.Env,
			})
		})

		// Rotas de Auth (Publicas)
		authHandler.RegisterRoutes(v1)
	}

	return router, nil
}
