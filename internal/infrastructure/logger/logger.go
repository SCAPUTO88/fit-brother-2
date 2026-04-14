package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup inicializa o logger global zerolog com o formato adequado ao ambiente.
// - development: saída human-readable com cores no console
// - production: saída JSON estruturada (compatível com Grafana/Loki)
func Setup(env string) {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		log.Logger = zerolog.New(os.Stdout).With().
			Timestamp().
			Str("service", "fit-brother-api").
			Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// FromRequest retorna um logger enriquecido com dados da requisição.
// Use nos handlers para logging contextualizado.
func FromRequest(requestID, method, path, clientIP string) zerolog.Logger {
	return log.With().
		Str("request_id", requestID).
		Str("method", method).
		Str("path", path).
		Str("client_ip", clientIP).
		Logger()
}
