package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	StorageDriverLocal    = "local"
	StorageDriverR2       = "r2"
	StorageDriverDisabled = "disabled"
)

type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Crypto    CryptoConfig
	JWT       JWTConfig
	Storage   StorageConfig
	R2        R2Config
	CORS      CORSConfig
	RateLimit RateLimitConfig
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
	PrivateKeyPEM   string
	PublicKeyPEM    string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type StorageConfig struct {
	Driver        string
	LocalDir      string
	SigningSecret string
	SignedURLTTL  time.Duration
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

type RateLimitConfig struct {
	AuthMaxRequests int
	AuthWindow      time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	databaseURL, err := requireEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}
	aesKey, err := requireEnv("AES_KEY")
	if err != nil {
		return nil, err
	}
	cpfHMACSecret, err := requireEnv("CPF_HMAC_SECRET")
	if err != nil {
		return nil, err
	}
	passwordPepper, err := requireEnv("PASSWORD_PEPPER")
	if err != nil {
		return nil, err
	}

	appEnv := getEnv("APP_ENV", "development")
	signedURLTTL := parseDuration("VIDEO_SIGNED_URL_TTL_MINUTES", 5*time.Minute)

	r2Cfg := R2Config{
		AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		Bucket:          getEnv("R2_BUCKET", ""),
		PublicURL:       getEnv("R2_PUBLIC_URL", ""),
		SignedURLTTL:    signedURLTTL,
	}

	driver, err := resolveStorageDriver(getEnv("STORAGE_DRIVER", ""), appEnv, r2Cfg)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		App: AppConfig{
			Env:     appEnv,
			Port:    resolveAppPort(),
			BaseURL: getEnv("APP_BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			URL: databaseURL,
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", ""),
		},
		Crypto: CryptoConfig{
			AESKey:         aesKey,
			CPFHMACSecret:  cpfHMACSecret,
			PasswordPepper: passwordPepper,
		},
		JWT: JWTConfig{
			PrivateKeyPath:  getEnv("JWT_PRIVATE_KEY_PATH", ""),
			PublicKeyPath:   getEnv("JWT_PUBLIC_KEY_PATH", ""),
			PrivateKeyPEM:   getEnv("JWT_PRIVATE_KEY_PEM", ""),
			PublicKeyPEM:    getEnv("JWT_PUBLIC_KEY_PEM", ""),
			AccessTokenTTL:  parseDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: parseDuration("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour),
		},
		Storage: StorageConfig{
			Driver:        driver,
			LocalDir:      getEnv("STORAGE_LOCAL_DIR", "./tmp/storage"),
			SigningSecret: getEnv("LOCAL_STORAGE_SIGNING_SECRET", ""),
			SignedURLTTL:  signedURLTTL,
		},
		R2: r2Cfg,
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		},
		RateLimit: RateLimitConfig{
			AuthMaxRequests: parsePositiveInt("AUTH_RATE_LIMIT_MAX", 20),
			AuthWindow:      parseDuration("AUTH_RATE_LIMIT_WINDOW", time.Minute),
		},
	}

	return cfg, nil
}

func resolveStorageDriver(rawDriver string, appEnv string, r2Cfg R2Config) (string, error) {
	driver := strings.ToLower(strings.TrimSpace(rawDriver))
	if driver == "" {
		if r2Cfg.HasCredentials() {
			return StorageDriverR2, nil
		}
		if appEnv == "production" {
			return StorageDriverDisabled, nil
		}
		return StorageDriverLocal, nil
	}

	switch driver {
	case StorageDriverLocal, StorageDriverDisabled:
		return driver, nil
	case StorageDriverR2:
		if !r2Cfg.HasCredentials() {
			return "", fmt.Errorf("STORAGE_DRIVER=r2 exige R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY e R2_BUCKET")
		}
		return StorageDriverR2, nil
	default:
		return "", fmt.Errorf("STORAGE_DRIVER inválido: %q (valores aceitos: local, r2, disabled)", rawDriver)
	}
}

func (c R2Config) HasCredentials() bool {
	return c.AccountID != "" &&
		c.AccessKeyID != "" &&
		c.SecretAccessKey != "" &&
		c.Bucket != ""
}

func resolveAppPort() string {
	return getEnv("APP_PORT", getEnv("PORT", "8080"))
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

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("variavel de ambiente obrigatoria nao definida: %s", key)
	}
	return v, nil
}

func parsePositiveInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return n
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
