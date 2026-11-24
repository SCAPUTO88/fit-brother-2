//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apphttp "github.com/SCAPUTO88/fit-brother/internal/adapters/http"
	"github.com/SCAPUTO88/fit-brother/internal/adapters/storage"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

type integrationVideoStorage struct {
	mu      sync.Mutex
	objects map[string]storage.ObjectMetadata
}

func newIntegrationVideoStorage() *integrationVideoStorage {
	return &integrationVideoStorage{
		objects: make(map[string]storage.ObjectMetadata),
	}
}

func (s *integrationVideoStorage) GeneratePresignedUploadURL(_ context.Context, key, _ string, _ time.Duration) (string, error) {
	return "https://r2.test.local/upload/" + key, nil
}

func (s *integrationVideoStorage) HeadObject(_ context.Context, key string) (*storage.ObjectMetadata, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	meta, ok := s.objects[key]
	if !ok {
		return nil, storage.ErrObjectNotFound
	}
	return &meta, nil
}

func (s *integrationVideoStorage) GenerateSignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://r2.test.local/signed/" + key, nil
}

func (s *integrationVideoStorage) DeleteObject(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, key)
	return nil
}

func (s *integrationVideoStorage) putFakeObject(key string, size int64, mimeType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = storage.ObjectMetadata{
		SizeBytes: size,
		MimeType:  mimeType,
	}
}

func generateTestRSAKeys(t *testing.T) (string, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	})
	return string(privPEM), string(pubPEM)
}

func applyMigrations(t *testing.T, db *sqlx.DB) {
	t.Helper()
	_, err := db.Exec(`
		DROP TABLE IF EXISTS videos, exercises, workout_sessions, workout_plans,
		audit_logs, biometric_data, students, refresh_tokens, users CASCADE;
	`)
	require.NoError(t, err)

	files, err := filepath.Glob("../../migrations/*.up.sql")
	require.NoError(t, err)
	require.NotEmpty(t, files)
	sort.Strings(files)

	for _, f := range files {
		sqlBytes, err := os.ReadFile(f)
		require.NoError(t, err, "lendo migration %s", f)
		_, err = db.Exec(string(sqlBytes))
		require.NoError(t, err, "executando migration %s", f)
	}
}

func seedTrainerUser(t *testing.T, db *sqlx.DB, passSvc *crypto.PasswordService, email, name, password string) uuid.UUID {
	t.Helper()
	hash, err := passSvc.Hash(password)
	require.NoError(t, err)
	id := uuid.New()
	_, err = db.Exec(`
		INSERT INTO users (id, email, full_name, password_hash, role, is_active)
		VALUES ($1, $2, $3, $4, 'TRAINER', true)
	`, id, email, name, hash)
	require.NoError(t, err)
	return id
}

func doJSONRequest(t *testing.T, router http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func loginAndGetTokens(t *testing.T, router http.Handler, email, password string) (string, string) {
	t.Helper()
	w := doJSONRequest(t, router, http.MethodPost, "/v1/auth/login", "", map[string]string{
		"email":    email,
		"password": password,
	})
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var out struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	return out.AccessToken, out.RefreshToken
}

func TestSecurityBaselinePostgresIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL não configurada; pulando teste de integração PostgreSQL")
	}
	if !strings.Contains(dbURL, "test") {
		t.Skip("DATABASE_URL não aponta para banco de teste; abortando reset de schema por segurança")
	}

	db, err := sqlx.Open("pgx", dbURL)
	require.NoError(t, err)
	defer db.Close()

	pingCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	require.NoError(t, db.PingContext(pingCtx))
	t.Log("Postgres conectado; aplicando migrations")

	applyMigrations(t, db)
	t.Log("Migrations aplicadas com sucesso")

	privPEM, pubPEM := generateTestRSAKeys(t)
	pepper := "integration-test-pepper-32-chars!!"
	aesKey := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
	hmacSecret := base64.StdEncoding.EncodeToString([]byte("integration-hmac-secret-32-bytes"))

	cfg := &config.Config{
		App: config.AppConfig{Env: "test", Port: "8080"},
		Crypto: config.CryptoConfig{
			AESKey:         aesKey,
			CPFHMACSecret:  hmacSecret,
			PasswordPepper: pepper,
		},
		JWT: config.JWTConfig{
			PrivateKeyPEM:   privPEM,
			PublicKeyPEM:    pubPEM,
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 24 * time.Hour,
		},
		R2: config.R2Config{
			SignedURLTTL: 5 * time.Minute,
		},
		CORS: config.CORSConfig{AllowedOrigins: "http://localhost:3000"},
	}

	fakeStorage := newIntegrationVideoStorage()
	router, err := apphttp.SetupRouterWithDeps(cfg, db, apphttp.RouterDeps{
		VideoStorage: fakeStorage,
	})
	require.NoError(t, err)

	passSvc, err := crypto.NewPasswordService(pepper)
	require.NoError(t, err)

	trainerAEmail := "trainer.a@fit.test"
	trainerBEmail := "trainer.b@fit.test"
	trainerAPass := "TrainerA!2026"
	trainerBPass := "TrainerB!2026"
	trainerAID := seedTrainerUser(t, db, passSvc, trainerAEmail, "Treinador A", trainerAPass)
	_ = seedTrainerUser(t, db, passSvc, trainerBEmail, "Treinador B", trainerBPass)

	trainerAToken, _ := loginAndGetTokens(t, router, trainerAEmail, trainerAPass)
	trainerBToken, _ := loginAndGetTokens(t, router, trainerBEmail, trainerBPass)

	// 1. POST /v1/students exige autenticacao de treinador
	wAnon := doJSONRequest(t, router, http.MethodPost, "/v1/students", "", map[string]string{
		"email": "anon@fit.test", "full_name": "Anonimo", "password": "Password123!", "cpf": "11122233344",
	})
	assert.Equal(t, http.StatusUnauthorized, wAnon.Code)

	// 2. Treinador A cadastra Aluno 1
	wCreateStud := doJSONRequest(t, router, http.MethodPost, "/v1/students", trainerAToken, map[string]string{
		"email":     "aluno1@fit.test",
		"full_name": "Aluno Um",
		"password":  "AlunoPass!2026",
		"cpf":       "123.456.789-01",
	})
	require.Equal(t, http.StatusCreated, wCreateStud.Code, wCreateStud.Body.String())
	var createdStud struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(wCreateStud.Body.Bytes(), &createdStud))

	var storedTrainerID uuid.UUID
	err = db.Get(&storedTrainerID, `SELECT trainer_id FROM students WHERE id = $1`, createdStud.ID)
	require.NoError(t, err)
	assert.Equal(t, trainerAID, storedTrainerID)

	// 3. Treinador B nao enxerga Aluno 1 na listagem nem nos detalhes
	wListB := doJSONRequest(t, router, http.MethodGet, "/v1/students", trainerBToken, nil)
	require.Equal(t, http.StatusOK, wListB.Code)
	assert.NotContains(t, wListB.Body.String(), createdStud.ID)

	wGetB := doJSONRequest(t, router, http.MethodGet, "/v1/students/"+createdStud.ID, trainerBToken, nil)
	assert.Equal(t, http.StatusForbidden, wGetB.Code)

	// 4. Aluno 1 loga e nao pode listar alunos ou criar novos alunos
	student1Token, student1Refresh := loginAndGetTokens(t, router, "aluno1@fit.test", "AlunoPass!2026")
	wStudList := doJSONRequest(t, router, http.MethodGet, "/v1/students", student1Token, nil)
	assert.Equal(t, http.StatusForbidden, wStudList.Code)

	wStudCreate := doJSONRequest(t, router, http.MethodPost, "/v1/students", student1Token, map[string]string{
		"email": "outro@fit.test", "full_name": "Outro", "password": "Password123!", "cpf": "99988877766",
	})
	assert.Equal(t, http.StatusForbidden, wStudCreate.Code)

	// 5. Treinador B nao cria plano para aluno do Treinador A
	wPlanB := doJSONRequest(t, router, http.MethodPost, "/v1/workouts", trainerBToken, map[string]string{
		"student_id": createdStud.ID,
		"title":      "Plano Indevido",
	})
	assert.Equal(t, http.StatusForbidden, wPlanB.Code)

	// 6. Treinador A cria plano, sessao e exercicio -> Aluno 1 visualiza usando users.id resolvido para students.id
	wPlanA := doJSONRequest(t, router, http.MethodPost, "/v1/workouts", trainerAToken, map[string]string{
		"student_id":  createdStud.ID,
		"title":       "Plano Hipertrofia",
		"description": "Fase 1",
	})
	require.Equal(t, http.StatusCreated, wPlanA.Code, wPlanA.Body.String())
	var planOut struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(wPlanA.Body.Bytes(), &planOut))

	wSessA := doJSONRequest(t, router, http.MethodPost, "/v1/workouts/"+planOut.ID+"/sessions", trainerAToken, map[string]any{
		"workout_plan_id": planOut.ID,
		"title":           "Treino A",
		"day_of_week":     1,
	})
	require.Equal(t, http.StatusCreated, wSessA.Code, wSessA.Body.String())
	var sessOut struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(wSessA.Body.Bytes(), &sessOut))

	wExA := doJSONRequest(t, router, http.MethodPost, "/v1/workouts/sessions/"+sessOut.ID+"/exercises", trainerAToken, map[string]any{
		"workout_session_id": sessOut.ID,
		"name":               "Supino Reto",
		"sets":               4,
		"reps":               "10",
		"rest_seconds":       60,
		"position":           1,
	})
	require.Equal(t, http.StatusCreated, wExA.Code, wExA.Body.String())
	var exOut struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(wExA.Body.Bytes(), &exOut))

	wStudPlans := doJSONRequest(t, router, http.MethodGet, "/v1/workouts", student1Token, nil)
	require.Equal(t, http.StatusOK, wStudPlans.Code)
	assert.Contains(t, wStudPlans.Body.String(), planOut.ID)

	wStudDetail := doJSONRequest(t, router, http.MethodGet, "/v1/workouts/"+planOut.ID, student1Token, nil)
	require.Equal(t, http.StatusOK, wStudDetail.Code)
	assert.Contains(t, wStudDetail.Body.String(), "Supino Reto")

	// 7. Fluxo de video com intencao PENDING -> HeadObject -> CONFIRMED -> Signed URL com ownership
	wUploadURL := doJSONRequest(t, router, http.MethodPost, "/v1/videos/upload-url", trainerAToken, map[string]string{
		"exercise_id": exOut.ID,
		"filename":    "supino.mp4",
		"mime_type":   "video/mp4",
	})
	require.Equal(t, http.StatusOK, wUploadURL.Code, wUploadURL.Body.String())
	var uploadOut struct {
		R2Key string `json:"r2_key"`
	}
	require.NoError(t, json.Unmarshal(wUploadURL.Body.Bytes(), &uploadOut))

	// Confirmacao sem objeto no R2 deve falhar
	wConfirmBeforePut := doJSONRequest(t, router, http.MethodPost, "/v1/videos", trainerAToken, map[string]string{
		"exercise_id": exOut.ID,
		"r2_key":      uploadOut.R2Key,
	})
	assert.Equal(t, http.StatusBadRequest, wConfirmBeforePut.Code)

	// Simula PUT concluido no R2 e confirma
	fakeStorage.putFakeObject(uploadOut.R2Key, 1024*512, "video/mp4")
	wConfirmAfterPut := doJSONRequest(t, router, http.MethodPost, "/v1/videos", trainerAToken, map[string]string{
		"exercise_id": exOut.ID,
		"r2_key":      uploadOut.R2Key,
	})
	require.Equal(t, http.StatusCreated, wConfirmAfterPut.Code, wConfirmAfterPut.Body.String())

	wSignedStud := doJSONRequest(t, router, http.MethodGet, "/v1/videos/exercises/"+exOut.ID+"/signed-url", student1Token, nil)
	require.Equal(t, http.StatusOK, wSignedStud.Code)
	assert.Contains(t, wSignedStud.Body.String(), uploadOut.R2Key)

	wSignedB := doJSONRequest(t, router, http.MethodGet, "/v1/videos/exercises/"+exOut.ID+"/signed-url", trainerBToken, nil)
	assert.Equal(t, http.StatusForbidden, wSignedB.Code)

	// 8. LGPD Anonymize remove PII, revoga tokens e invalida JWT ativo em tempo real
	wDel := doJSONRequest(t, router, http.MethodDelete, "/v1/students/"+createdStud.ID, trainerAToken, nil)
	require.Equal(t, http.StatusOK, wDel.Code, wDel.Body.String())

	wAfterDel := doJSONRequest(t, router, http.MethodGet, "/v1/workouts", student1Token, nil)
	assert.Equal(t, http.StatusUnauthorized, wAfterDel.Code)

	wRefreshAfterDel := doJSONRequest(t, router, http.MethodPost, "/v1/auth/refresh", "", map[string]string{
		"refresh_token": student1Refresh,
	})
	assert.Equal(t, http.StatusUnauthorized, wRefreshAfterDel.Code)

	var anonUser struct {
		Email    string  `db:"email"`
		FullName string  `db:"full_name"`
		IsActive bool    `db:"is_active"`
		Deleted  *string `db:"deleted_at"`
	}
	err = db.Get(&anonUser, `SELECT email, full_name, is_active, deleted_at::text FROM users WHERE full_name = 'Usuário Anonimizado'`)
	require.NoError(t, err)
	assert.False(t, anonUser.IsActive)
	assert.NotNil(t, anonUser.Deleted)
	assert.True(t, strings.HasPrefix(anonUser.Email, "deleted-"))
}
