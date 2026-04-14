package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	router "github.com/SCAPUTO88/fit-brother/internal/adapters/http"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/logger"
)

// @title           Fit Brother API
// @version         1.0
// @description     API de gestão de treinos — Backend em Go
// @BasePath        /v1
func main() {
	// 1. Carregar configuração
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro ao carregar configuracao: %v\n", err)
		os.Exit(1)
	}

	// 2. Inicializar logger
	logger.Setup(cfg.App.Env)
	log.Info().
		Str("env", cfg.App.Env).
		Str("port", cfg.App.Port).
		Msg("Iniciando Fit Brother API")

	// 3. Inicializar banco de dados
	db, err := sqlx.Connect("pgx", cfg.Database.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("Erro ao conectar no banco de dados Postgres")
	}
	defer db.Close()

	// Configurar pool de conexões (melhoria de resiliência e performance)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 4. Montar router HTTP
	mux, err := router.SetupRouter(cfg, db)
	if err != nil {
		log.Fatal().Err(err).Msg("Erro configurando o Router HTTP")
	}

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 6. Graceful shutdown
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("Servidor HTTP iniciado")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Erro fatal no servidor HTTP")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Sinal de encerramento recebido — iniciando graceful shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Erro durante o graceful shutdown")
	}

	log.Info().Msg("Servidor encerrado com sucesso")
}
