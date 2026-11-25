package config

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setMinimalRequiredEnv(t *testing.T) {
	t.Helper()
	testAESKey := base64.StdEncoding.EncodeToString([]byte("test-aes-key-32-bytes-unit-only!"))
	t.Setenv("DATABASE_URL", "postgres://fitbrother:secret@localhost:5432/fitbrother_test?sslmode=disable")
	t.Setenv("AES_KEY", testAESKey)
	t.Setenv("CPF_HMAC_SECRET", "test-hmac-secret-for-ci-only-32chars!")
	t.Setenv("PASSWORD_PEPPER", "test-pepper-for-ci-only-32chars!!")
	t.Setenv("REDIS_URL", "")
	t.Setenv("STORAGE_DRIVER", "")
	t.Setenv("R2_ACCOUNT_ID", "")
	t.Setenv("R2_ACCESS_KEY_ID", "")
	t.Setenv("R2_SECRET_ACCESS_KEY", "")
	t.Setenv("R2_BUCKET", "")
	t.Setenv("R2_PUBLIC_URL", "")
}

func TestResolveAppPortFallsBackToPortEnvVar(t *testing.T) {
	t.Setenv("APP_PORT", "")
	t.Setenv("PORT", "49152")

	if got := resolveAppPort(); got != "49152" {
		t.Fatalf("resolveAppPort() = %q, want %q", got, "49152")
	}
}

func TestLoadSucceedsWithoutRedisAndWithoutR2InDevelopment(t *testing.T) {
	setMinimalRequiredEnv(t)
	t.Setenv("APP_ENV", "development")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "", cfg.Redis.URL)
	assert.Equal(t, StorageDriverLocal, cfg.Storage.Driver)
	assert.Equal(t, "./tmp/storage", cfg.Storage.LocalDir)
	assert.Equal(t, 5*time.Minute, cfg.Storage.SignedURLTTL)
	assert.Equal(t, 20, cfg.RateLimit.AuthMaxRequests)
	assert.Equal(t, time.Minute, cfg.RateLimit.AuthWindow)
}

func TestLoadDefaultsToDisabledStorageInProductionWhenR2Unset(t *testing.T) {
	setMinimalRequiredEnv(t)
	t.Setenv("APP_ENV", "production")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, StorageDriverDisabled, cfg.Storage.Driver)
}

func TestLoadAutoSelectsR2WhenCredentialsPresent(t *testing.T) {
	setMinimalRequiredEnv(t)
	t.Setenv("R2_ACCOUNT_ID", "acc-123")
	t.Setenv("R2_ACCESS_KEY_ID", "key-123")
	t.Setenv("R2_SECRET_ACCESS_KEY", "sec-123")
	t.Setenv("R2_BUCKET", "fit-bucket")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, StorageDriverR2, cfg.Storage.Driver)
}

func TestLoadRejectsR2DriverWithoutCredentials(t *testing.T) {
	setMinimalRequiredEnv(t)
	t.Setenv("STORAGE_DRIVER", "r2")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "R2_ACCOUNT_ID")
}

func TestLoadRejectsUnknownStorageDriver(t *testing.T) {
	setMinimalRequiredEnv(t)
	t.Setenv("STORAGE_DRIVER", "gcs")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "STORAGE_DRIVER")
}
