package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Crypto   CryptoConfig
	JWT      JWTConfig
	R2       R2Config
	CORS     CORSConfig
}

type AppConfig struct {
	Env     string
	Port    string
	BaseURL string
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type CryptoConfig struct {
	AESKey         string
	CPFHMACSecret  string
	PasswordPepper string
}

type JWTConfig struct {
	PrivateKeyPath  string
	PublicKeyPath   string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	PublicURL       string
	SignedURLTTL    time.Duration
}

type CORSConfig struct {
	AllowedOrigins string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:     getEnv("APP_ENV", "development"),
			Port:    getEnv("APP_PORT", "8080"),
			BaseURL: getEnv("APP_BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			URL: requireEnv("DATABASE_URL"),
		},
		Redis: RedisConfig{
			URL: requireEnv("REDIS_URL"),
		},
		Crypto: CryptoConfig{
			AESKey:         requireEnv("AES_KEY"),
			CPFHMACSecret:  requireEnv("CPF_HMAC_SECRET"),
			PasswordPepper: requireEnv("PASSWORD_PEPPER"),
		},
		JWT: JWTConfig{
			PrivateKeyPath:  requireEnv("JWT_PRIVATE_KEY_PATH"),
			PublicKeyPath:   requireEnv("JWT_PUBLIC_KEY_PATH"),
			AccessTokenTTL:  parseDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: parseDuration("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour),
		},
		R2: R2Config{
			AccountID:       requireEnv("R2_ACCOUNT_ID"),
			AccessKeyID:     requireEnv("R2_ACCESS_KEY_ID"),
			SecretAccessKey: requireEnv("R2_SECRET_ACCESS_KEY"),
			Bucket:          requireEnv("R2_BUCKET"),
			PublicURL:       requireEnv("R2_PUBLIC_URL"),
			SignedURLTTL:    parseDuration("VIDEO_SIGNED_URL_TTL_MINUTES", 5*time.Minute),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		},
	}

	return cfg, nil
}

func (c *AppConfig) IsDevelopment() bool {
	return c.Env == "development"
}

func (c *AppConfig) IsProduction() bool {
	return c.Env == "production"
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("variavel de ambiente obrigatoria nao definida: %s", key))
	}
	return v
}

func parseDuration(key string, defaultVal time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if mins, err := strconv.Atoi(v); err == nil {
		return time.Duration(mins) * time.Minute
	}
	return defaultVal
}
