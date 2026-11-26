//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base32"
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
	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	"github.com/SCAPUTO88/fit-brother/internal/adapters/repository/postgres"
	appCatalog "github.com/SCAPUTO88/fit-brother/internal/application/catalog"
	domainMailer "github.com/SCAPUTO88/fit-brother/internal/domain/mailer"
	domainStorage "github.com/SCAPUTO88/fit-brother/internal/domain/storage"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/totp"
)

type integrationMailer struct {
	mu           sync.Mutex
	inviteTokens []string
	resetTokens  []string
}

func (m *integrationMailer) SendInvitation(_ context.Context, msg domainMailer.InvitationEmail) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inviteTokens = append(m.inviteTokens, msg.RawToken)
	return nil
}

func (m *integrationMailer) SendPasswordReset(_ context.Context, msg domainMailer.PasswordResetEmail) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resetTokens = append(m.resetTokens, msg.RawToken)
	return nil
}

type integrationVideoStorage struct {
	mu      sync.Mutex
	objects map[string]domainStorage.ObjectMetadata
}

func newIntegrationVideoStorage() *integrationVideoStorage {
	return &integrationVideoStorage{
		objects: make(map[string]domainStorage.ObjectMetadata),
	}
}

func (s *integrationVideoStorage) GeneratePresignedUploadURL(_ context.Context, key, _ string, _ time.Duration) (string, error) {
	return "https://r2.test.local/upload/" + key, nil
}

func (s *integrationVideoStorage) HeadObject(_ context.Context, key string) (*domainStorage.ObjectMetadata, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	meta, ok := s.objects[key]
	if !ok {
		return nil, domainStorage.ErrObjectNotFound
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
	s.objects[key] = domainStorage.ObjectMetadata{
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
		DROP TABLE IF EXISTS
			exercise_media, exercise_catalog, equipments, muscle_groups,
			audit_events, privacy_requests, consent_records, user_mfa,
			password_reset_tokens, invitations, trainer_student_relationships,
			memberships, organizations,
			videos, exercises, workout_sessions, workout_plans,
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

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
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
		App: config.AppConfig{Env: "test", Port: "8080", BaseURL: "http://localhost:8080"},
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
		Storage: config.StorageConfig{
			Driver:       config.StorageDriverLocal,
			LocalDir:     t.TempDir(),
			SignedURLTTL: 5 * time.Minute,
		},
		CORS: config.CORSConfig{AllowedOrigins: "http://localhost:3000"},
	}

	fakeStorage := newIntegrationVideoStorage()
	fakeMailer := &integrationMailer{}
	router, err := apphttp.SetupRouterWithDeps(cfg, db, apphttp.RouterDeps{
		VideoStorage: fakeStorage,
		Mailer:       fakeMailer,
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

	// Simula PUT concluido e confirma
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

	// 8. Sessao Web com cookies HttpOnly + protecao CSRF (ADR-008) e LocalStorageAdapter ponta-a-ponta (ADR-005)
	t.Run("Fluxo Web Cookie + CSRF e LocalStorage sem Redis/R2", func(t *testing.T) {
		localRouter, err := apphttp.SetupRouter(cfg, db)
		require.NoError(t, err)

		var loginBody bytes.Buffer
		require.NoError(t, json.NewEncoder(&loginBody).Encode(map[string]string{
			"email":    trainerAEmail,
			"password": trainerAPass,
		}))
		reqCookieLogin := httptest.NewRequest(http.MethodPost, "/v1/auth/login", &loginBody)
		reqCookieLogin.Header.Set("Content-Type", "application/json")
		reqCookieLogin.Header.Set(middleware.SessionModeHeaderName, "cookie")
		wCookieLogin := httptest.NewRecorder()
		localRouter.ServeHTTP(wCookieLogin, reqCookieLogin)
		require.Equal(t, http.StatusOK, wCookieLogin.Code, wCookieLogin.Body.String())

		// Nao vaza raw access_token/refresh_token no corpo JSON
		assert.NotContains(t, wCookieLogin.Body.String(), "access_token")
		assert.NotContains(t, wCookieLogin.Body.String(), "refresh_token")

		cookies := wCookieLogin.Result().Cookies()
		accessCookie := findCookie(cookies, middleware.AccessCookieName)
		refreshCookie := findCookie(cookies, middleware.RefreshCookieName)
		csrfCookie := findCookie(cookies, middleware.CSRFCookieName)
		require.NotNil(t, accessCookie)
		require.NotNil(t, refreshCookie)
		require.NotNil(t, csrfCookie)
		assert.True(t, accessCookie.HttpOnly)
		assert.True(t, refreshCookie.HttpOnly)
		assert.False(t, csrfCookie.HttpOnly)

		// GET autenticado via cookie
		reqStudents := httptest.NewRequest(http.MethodGet, "/v1/students", nil)
		reqStudents.AddCookie(accessCookie)
		wStudents := httptest.NewRecorder()
		localRouter.ServeHTTP(wStudents, reqStudents)
		require.Equal(t, http.StatusOK, wStudents.Code)

		// POST via cookie sem X-CSRF-Token deve falhar com 403
		reqNoCSRF := httptest.NewRequest(http.MethodPost, "/v1/videos/upload-url", strings.NewReader(`{"exercise_id":"`+exOut.ID+`","filename":"local.mp4","mime_type":"video/mp4"}`))
		reqNoCSRF.Header.Set("Content-Type", "application/json")
		reqNoCSRF.Header.Set("Origin", "http://localhost:3000")
		reqNoCSRF.AddCookie(accessCookie)
		reqNoCSRF.AddCookie(csrfCookie)
		wNoCSRF := httptest.NewRecorder()
		localRouter.ServeHTTP(wNoCSRF, reqNoCSRF)
		assert.Equal(t, http.StatusForbidden, wNoCSRF.Code)

		// POST via cookie com X-CSRF-Token -> gera URL local assinada -> PUT local -> Confirm -> GET playback
		reqWithCSRF := httptest.NewRequest(http.MethodPost, "/v1/videos/upload-url", strings.NewReader(`{"exercise_id":"`+exOut.ID+`","filename":"local.mp4","mime_type":"video/mp4"}`))
		reqWithCSRF.Header.Set("Content-Type", "application/json")
		reqWithCSRF.Header.Set("Origin", "http://localhost:3000")
		reqWithCSRF.Header.Set(middleware.CSRFHeaderName, csrfCookie.Value)
		reqWithCSRF.AddCookie(accessCookie)
		reqWithCSRF.AddCookie(csrfCookie)
		wWithCSRF := httptest.NewRecorder()
		localRouter.ServeHTTP(wWithCSRF, reqWithCSRF)
		require.Equal(t, http.StatusOK, wWithCSRF.Code, wWithCSRF.Body.String())

		var localUpload struct {
			PresignedURL string `json:"presigned_url"`
			R2Key        string `json:"r2_key"`
		}
		require.NoError(t, json.Unmarshal(wWithCSRF.Body.Bytes(), &localUpload))
		assert.Contains(t, localUpload.PresignedURL, "/v1/storage/local")

		// PUT binario direto para URL local assinada
		reqPut := httptest.NewRequest(http.MethodPut, localUpload.PresignedURL, strings.NewReader("video-bytes-payload"))
		reqPut.Header.Set("Content-Type", "video/mp4")
		wPut := httptest.NewRecorder()
		localRouter.ServeHTTP(wPut, reqPut)
		require.Equal(t, http.StatusOK, wPut.Code, wPut.Body.String())

		// Refresh via cookie rotaciona cookies sem enviar body JSON
		reqCookieRefresh := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
		reqCookieRefresh.Header.Set("Origin", "http://localhost:3000")
		reqCookieRefresh.Header.Set(middleware.CSRFHeaderName, csrfCookie.Value)
		reqCookieRefresh.AddCookie(refreshCookie)
		reqCookieRefresh.AddCookie(csrfCookie)
		wCookieRefresh := httptest.NewRecorder()
		localRouter.ServeHTTP(wCookieRefresh, reqCookieRefresh)
		require.Equal(t, http.StatusOK, wCookieRefresh.Code, wCookieRefresh.Body.String())
		rotatedCookies := wCookieRefresh.Result().Cookies()
		require.NotNil(t, findCookie(rotatedCookies, middleware.AccessCookieName))
		rotatedCSRF := findCookie(rotatedCookies, middleware.CSRFCookieName)
		require.NotNil(t, rotatedCSRF)

		// Logout via cookie expira cookies e retorna 204
		reqCookieLogout := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
		reqCookieLogout.Header.Set("Origin", "http://localhost:3000")
		reqCookieLogout.Header.Set(middleware.CSRFHeaderName, rotatedCSRF.Value)
		reqCookieLogout.AddCookie(findCookie(rotatedCookies, middleware.RefreshCookieName))
		reqCookieLogout.AddCookie(rotatedCSRF)
		wCookieLogout := httptest.NewRecorder()
		localRouter.ServeHTTP(wCookieLogout, reqCookieLogout)
		require.Equal(t, http.StatusNoContent, wCookieLogout.Code)
	})

	// 9. LGPD Anonymize remove PII, revoga tokens e invalida JWT ativo em tempo real
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

	// 10. Fase 3 — Multi-tenancy, Convites, CPF opcional, Password Reset, MFA TOTP e Central LGPD
	t.Run("Fase 3 - Identidade Multi-Tenancy Convites MFA Reset e LGPD", func(t *testing.T) {
		// 10.1 Cadastro direto de aluno sem CPF (ADR-006)
		wNoCPFStud := doJSONRequest(t, router, http.MethodPost, "/v1/students", trainerAToken, map[string]string{
			"email":     "nocpf@fit.test",
			"full_name": "Aluno Sem CPF",
			"password":  "NoCPFPass!2026",
		})
		require.Equal(t, http.StatusCreated, wNoCPFStud.Code, wNoCPFStud.Body.String())
		var noCPFStud struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(wNoCPFStud.Body.Bytes(), &noCPFStud))

		var cpfEnc *string
		err := db.Get(&cpfEnc, `SELECT cpf_encrypted FROM students WHERE id = $1`, noCPFStud.ID)
		require.NoError(t, err)
		assert.Nil(t, cpfEnc)

		// 10.2 Organizações e Membros multi-tenant
		wMyOrgs := doJSONRequest(t, router, http.MethodGet, "/v1/organizations", trainerAToken, nil)
		require.Equal(t, http.StatusOK, wMyOrgs.Code, wMyOrgs.Body.String())
		var myOrgs []struct {
			OrganizationID string `json:"organization_id"`
		}
		require.NoError(t, json.Unmarshal(wMyOrgs.Body.Bytes(), &myOrgs))
		require.NotEmpty(t, myOrgs)
		defaultOrgID := myOrgs[0].OrganizationID

		wStudio := doJSONRequest(t, router, http.MethodPost, "/v1/organizations", trainerAToken, map[string]string{
			"name": "Fit Studio Integration",
			"type": "STUDIO",
		})
		require.Equal(t, http.StatusCreated, wStudio.Code, wStudio.Body.String())
		var studioOut struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(wStudio.Body.Bytes(), &studioOut))

		wAddMem := doJSONRequest(t, router, http.MethodPost, "/v1/organizations/"+studioOut.ID+"/members", trainerAToken, map[string]string{
			"email": trainerBEmail,
			"role":  "TRAINER",
		})
		require.Equal(t, http.StatusCreated, wAddMem.Code, wAddMem.Body.String())

		wListMem := doJSONRequest(t, router, http.MethodGet, "/v1/organizations/"+studioOut.ID+"/members", trainerBToken, nil)
		require.Equal(t, http.StatusOK, wListMem.Code)
		assert.Contains(t, wListMem.Body.String(), trainerAEmail)
		assert.Contains(t, wListMem.Body.String(), trainerBEmail)

		// Encerrar vínculo com Aluno Sem CPF bloqueia acesso do Treinador A
		wEndRel := doJSONRequest(t, router, http.MethodPatch, "/v1/organizations/"+defaultOrgID+"/relationships/"+noCPFStud.ID, trainerAToken, map[string]string{
			"status": "ENDED",
		})
		require.Equal(t, http.StatusOK, wEndRel.Code, wEndRel.Body.String())

		wGetEnded := doJSONRequest(t, router, http.MethodGet, "/v1/students/"+noCPFStud.ID, trainerAToken, nil)
		assert.Equal(t, http.StatusForbidden, wGetEnded.Code)

		wWorkoutEnded := doJSONRequest(t, router, http.MethodPost, "/v1/workouts", trainerAToken, map[string]string{
			"student_id": noCPFStud.ID,
			"title":      "Treino Bloqueado",
		})
		assert.Equal(t, http.StatusForbidden, wWorkoutEnded.Code)

		// 10.3 Convites (criar, revogar, aceitar sem CPF com consentimentos LGPD)
		wRevokeCreate := doJSONRequest(t, router, http.MethodPost, "/v1/invitations", trainerAToken, map[string]string{
			"email":     "revoked.invite@fit.test",
			"full_name": "Convidado Revogado",
		})
		require.Equal(t, http.StatusCreated, wRevokeCreate.Code, wRevokeCreate.Body.String())
		var revokedInv struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(wRevokeCreate.Body.Bytes(), &revokedInv))
		revokedToken := fakeMailer.inviteTokens[len(fakeMailer.inviteTokens)-1]

		wDelInv := doJSONRequest(t, router, http.MethodDelete, "/v1/invitations/"+revokedInv.ID, trainerAToken, nil)
		require.Equal(t, http.StatusOK, wDelInv.Code)

		wAcceptRevoked := doJSONRequest(t, router, http.MethodPost, "/v1/auth/invitations/accept", "", map[string]any{
			"token":               revokedToken,
			"password":            "InvitePass!2026",
			"consent_terms":       true,
			"consent_health_data": true,
		})
		assert.Equal(t, http.StatusUnauthorized, wAcceptRevoked.Code)

		wInvite := doJSONRequest(t, router, http.MethodPost, "/v1/invitations", trainerAToken, map[string]string{
			"organization_id": studioOut.ID,
			"email":           "invited.student@fit.test",
			"full_name":       "Aluno Convidado",
		})
		require.Equal(t, http.StatusCreated, wInvite.Code, wInvite.Body.String())
		validInviteToken := fakeMailer.inviteTokens[len(fakeMailer.inviteTokens)-1]

		wAcceptNoConsent := doJSONRequest(t, router, http.MethodPost, "/v1/auth/invitations/accept", "", map[string]any{
			"token":               validInviteToken,
			"password":            "InvitePass!2026",
			"consent_terms":       true,
			"consent_health_data": false,
		})
		assert.Equal(t, http.StatusBadRequest, wAcceptNoConsent.Code)

		wAcceptOK := doJSONRequest(t, router, http.MethodPost, "/v1/auth/invitations/accept", "", map[string]any{
			"token":               validInviteToken,
			"password":            "InvitePass!2026",
			"consent_terms":       true,
			"consent_health_data": true,
		})
		require.Equal(t, http.StatusCreated, wAcceptOK.Code, wAcceptOK.Body.String())
		var acceptOut struct {
			StudentID      string `json:"student_id"`
			OrganizationID string `json:"organization_id"`
			Session        struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			} `json:"session"`
		}
		require.NoError(t, json.Unmarshal(wAcceptOK.Body.Bytes(), &acceptOut))
		assert.Equal(t, studioOut.ID, acceptOut.OrganizationID)
		require.NotEmpty(t, acceptOut.Session.AccessToken)

		// Reuso do convite é bloqueado
		wReuseInv := doJSONRequest(t, router, http.MethodPost, "/v1/auth/invitations/accept", "", map[string]any{
			"token":               validInviteToken,
			"password":            "InvitePass!2026",
			"consent_terms":       true,
			"consent_health_data": true,
		})
		assert.Equal(t, http.StatusUnauthorized, wReuseInv.Code)

		// 10.4 Redefinição de senha revoga refresh token anterior e bloqueia reuso do token
		wResetUnknown := doJSONRequest(t, router, http.MethodPost, "/v1/auth/password-reset/request", "", map[string]string{
			"email": "nao.existe@fit.test",
		})
		assert.Equal(t, http.StatusOK, wResetUnknown.Code)

		beforeResetCount := len(fakeMailer.resetTokens)
		wResetReq := doJSONRequest(t, router, http.MethodPost, "/v1/auth/password-reset/request", "", map[string]string{
			"email": "invited.student@fit.test",
		})
		require.Equal(t, http.StatusOK, wResetReq.Code)
		require.Equal(t, beforeResetCount+1, len(fakeMailer.resetTokens))
		resetToken := fakeMailer.resetTokens[len(fakeMailer.resetTokens)-1]

		wResetConfirm := doJSONRequest(t, router, http.MethodPost, "/v1/auth/password-reset/confirm", "", map[string]string{
			"token":        resetToken,
			"new_password": "NewInvitePass!2026",
		})
		require.Equal(t, http.StatusOK, wResetConfirm.Code, wResetConfirm.Body.String())

		wOldRefresh := doJSONRequest(t, router, http.MethodPost, "/v1/auth/refresh", "", map[string]string{
			"refresh_token": acceptOut.Session.RefreshToken,
		})
		assert.Equal(t, http.StatusUnauthorized, wOldRefresh.Code)

		wResetReuse := doJSONRequest(t, router, http.MethodPost, "/v1/auth/password-reset/confirm", "", map[string]string{
			"token":        resetToken,
			"new_password": "AnotherPass!2026",
		})
		assert.Equal(t, http.StatusUnauthorized, wResetReuse.Code)

		invitedAccess, _ := loginAndGetTokens(t, router, "invited.student@fit.test", "NewInvitePass!2026")

		// 10.5 Central LGPD (consentimentos, solicitação, exportação segura, audit events e auto-anonimização)
		wConsent := doJSONRequest(t, router, http.MethodPost, "/v1/privacy/consents", invitedAccess, map[string]any{
			"purpose":        "MEDIA_UPLOAD",
			"policy_version": "2026-11",
			"granted":        true,
		})
		require.Equal(t, http.StatusCreated, wConsent.Code, wConsent.Body.String())

		wPrivReq := doJSONRequest(t, router, http.MethodPost, "/v1/privacy/requests", invitedAccess, map[string]string{
			"request_type": "RECTIFY",
			"details":      "Atualizar nome de exibição",
		})
		require.Equal(t, http.StatusCreated, wPrivReq.Code, wPrivReq.Body.String())

		wExport := doJSONRequest(t, router, http.MethodGet, "/v1/privacy/export", invitedAccess, nil)
		require.Equal(t, http.StatusOK, wExport.Code, wExport.Body.String())
		assert.Contains(t, wExport.Body.String(), "invited.student@fit.test")
		assert.NotContains(t, wExport.Body.String(), "password_hash")
		assert.NotContains(t, wExport.Body.String(), "cpf_encrypted")

		wAuditStud := doJSONRequest(t, router, http.MethodGet, "/v1/audit-events", invitedAccess, nil)
		assert.Equal(t, http.StatusForbidden, wAuditStud.Code)

		wAuditTrainer := doJSONRequest(t, router, http.MethodGet, "/v1/audit-events", trainerAToken, nil)
		require.Equal(t, http.StatusOK, wAuditTrainer.Code)
		assert.Contains(t, wAuditTrainer.Body.String(), "INVITATION_CREATED")

		wSelfAnon := doJSONRequest(t, router, http.MethodPost, "/v1/privacy/anonymize", invitedAccess, nil)
		require.Equal(t, http.StatusOK, wSelfAnon.Code, wSelfAnon.Body.String())

		wAfterSelfAnon := doJSONRequest(t, router, http.MethodGet, "/v1/workouts", invitedAccess, nil)
		assert.Equal(t, http.StatusUnauthorized, wAfterSelfAnon.Code)

		// 10.6 MFA TOTP para Treinador (Setup -> Enable step-1 -> Login sem código 401 mfa_required -> Login step -> Replay bloqueado -> Disable step+1)
		trainerCEmail := "trainer.c@fit.test"
		trainerCPass := "TrainerC!2026"
		_ = seedTrainerUser(t, db, passSvc, trainerCEmail, "Treinador C", trainerCPass)
		trainerCToken, _ := loginAndGetTokens(t, router, trainerCEmail, trainerCPass)

		wMFASetup := doJSONRequest(t, router, http.MethodPost, "/v1/auth/mfa/setup", trainerCToken, nil)
		require.Equal(t, http.StatusOK, wMFASetup.Code, wMFASetup.Body.String())
		var setupOut struct {
			Secret     string `json:"secret"`
			OTPAuthURI string `json:"otpauth_uri"`
		}
		require.NoError(t, json.Unmarshal(wMFASetup.Body.Bytes(), &setupOut))
		require.NotEmpty(t, setupOut.Secret)

		totpKey, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(setupOut.Secret)
		require.NoError(t, err)
		curStep := time.Now().UTC().Unix() / totp.PeriodSeconds
		codePrev := totp.GenerateCodeForStep(totpKey, curStep-1)
		codeCur := totp.GenerateCodeForStep(totpKey, curStep)
		codeNext := totp.GenerateCodeForStep(totpKey, curStep+1)

		wMFAEnable := doJSONRequest(t, router, http.MethodPost, "/v1/auth/mfa/enable", trainerCToken, map[string]string{
			"code": codePrev,
		})
		require.Equal(t, http.StatusOK, wMFAEnable.Code, wMFAEnable.Body.String())

		wLoginNoCode := doJSONRequest(t, router, http.MethodPost, "/v1/auth/login", "", map[string]string{
			"email":    trainerCEmail,
			"password": trainerCPass,
		})
		assert.Equal(t, http.StatusUnauthorized, wLoginNoCode.Code)
		assert.Contains(t, wLoginNoCode.Body.String(), `"mfa_required":true`)

		wLoginWithCode := doJSONRequest(t, router, http.MethodPost, "/v1/auth/login", "", map[string]string{
			"email":     trainerCEmail,
			"password":  trainerCPass,
			"totp_code": codeCur,
		})
		require.Equal(t, http.StatusOK, wLoginWithCode.Code, wLoginWithCode.Body.String())

		wLoginReplay := doJSONRequest(t, router, http.MethodPost, "/v1/auth/login", "", map[string]string{
			"email":     trainerCEmail,
			"password":  trainerCPass,
			"totp_code": codeCur,
		})
		assert.Equal(t, http.StatusUnauthorized, wLoginReplay.Code)

		wMFADisable := doJSONRequest(t, router, http.MethodPost, "/v1/auth/mfa/disable", trainerCToken, map[string]string{
			"password": trainerCPass,
			"code":     codeNext,
		})
		require.Equal(t, http.StatusOK, wMFADisable.Code, wMFADisable.Body.String())

		wMFAStatus := doJSONRequest(t, router, http.MethodGet, "/v1/auth/mfa/status", trainerCToken, nil)
		require.Equal(t, http.StatusOK, wMFAStatus.Code)
		assert.Contains(t, wMFAStatus.Body.String(), `"enabled":false`)
	})

	// 11. Fase 4 — Biblioteca de Exercícios (Catálogo PT-BR, Importador Idempotente, Custom CRUD e Mídia Segura)
	t.Run("Fase 4 - Biblioteca de Exercicios Catalogo PT-BR Importador Idempotente e Midia Segura", func(t *testing.T) {
		// 11.1 Filtros iniciais e exercícios globais PT-BR vindos da migration 008
		wFilters := doJSONRequest(t, router, http.MethodGet, "/v1/catalog/filters", trainerAToken, nil)
		require.Equal(t, http.StatusOK, wFilters.Code, wFilters.Body.String())
		assert.Contains(t, wFilters.Body.String(), "Peitoral")
		assert.Contains(t, wFilters.Body.String(), "Halteres")

		wSearchGlobal := doJSONRequest(t, router, http.MethodGet, "/v1/catalog/exercises?q=Supino+Reto&muscle=chest&equipment=barbell", trainerAToken, nil)
		require.Equal(t, http.StatusOK, wSearchGlobal.Code, wSearchGlobal.Body.String())
		var searchGlobalOut struct {
			Items []struct {
				ID          string `json:"id"`
				NamePT      string `json:"name_pt"`
				LicenseInfo string `json:"license_info"`
				IsCustom    bool   `json:"is_custom"`
			} `json:"items"`
			Total int `json:"total"`
		}
		require.NoError(t, json.Unmarshal(wSearchGlobal.Body.Bytes(), &searchGlobalOut))
		require.GreaterOrEqual(t, searchGlobalOut.Total, 1)
		assert.Equal(t, "Supino Reto com Barra", searchGlobalOut.Items[0].NamePT)
		assert.Contains(t, searchGlobalOut.Items[0].LicenseInfo, "The Unlicense")
		assert.False(t, searchGlobalOut.Items[0].IsCustom)
		globalExerciseID := searchGlobalOut.Items[0].ID

		// Treinador não pode editar ou excluir exercício do catálogo global
		wPatchGlobal := doJSONRequest(t, router, http.MethodPatch, "/v1/catalog/exercises/"+globalExerciseID, trainerAToken, map[string]any{
			"name_pt":        "Tentativa",
			"primary_muscle": "chest",
			"equipment":      "barbell",
			"category":       "strength",
			"level":          "intermediate",
		})
		assert.Equal(t, http.StatusForbidden, wPatchGlobal.Code)

		wDeleteGlobal := doJSONRequest(t, router, http.MethodDelete, "/v1/catalog/exercises/"+globalExerciseID, trainerAToken, nil)
		assert.Equal(t, http.StatusForbidden, wDeleteGlobal.Code)

		// 11.2 Importador idempotente free-exercise-db contra PostgreSQL real
		importer := appCatalog.NewImporter(postgres.NewCatalogRepository(db))
		fixtureJSON := `[
			{
				"id": "Barbell_Bench_Press_-_Medium_Grip",
				"name": "Barbell Bench Press - Medium Grip",
				"force": "push",
				"level": "beginner",
				"mechanic": "compound",
				"equipment": "barbell",
				"primaryMuscles": ["chest"],
				"secondaryMuscles": ["shoulders", "triceps"],
				"instructions": ["Lie back on a flat bench."],
				"category": "strength"
			},
			{
				"id": "Integration_Test_Incline_Dumbbell_Press",
				"name": "Incline Dumbbell Press",
				"force": "push",
				"level": "intermediate",
				"mechanic": "compound",
				"equipment": "dumbbell",
				"primaryMuscles": ["chest"],
				"secondaryMuscles": ["shoulders"],
				"instructions": ["Press the dumbbells overhead with control."],
				"category": "strength"
			}
		]`
		impRun1, err := importer.ImportFreeExerciseDB(context.Background(), strings.NewReader(fixtureJSON))
		require.NoError(t, err)
		assert.Equal(t, 2, impRun1.Processed)
		assert.Equal(t, 1, impRun1.Inserted)
		assert.Equal(t, 1, impRun1.Updated)

		impRun2, err := importer.ImportFreeExerciseDB(context.Background(), strings.NewReader(fixtureJSON))
		require.NoError(t, err)
		assert.Equal(t, 2, impRun2.Processed)
		assert.Equal(t, 0, impRun2.Inserted)
		assert.Equal(t, 2, impRun2.Updated)

		// 11.3 Treinador A cria exercício customizado na sua org pessoal e envia mídia (Imagem PNG)
		wCreateCustom := doJSONRequest(t, router, http.MethodPost, "/v1/catalog/exercises", trainerAToken, map[string]any{
			"name_pt":           "Crucifixo Customizado do Studio",
			"primary_muscle":    "chest",
			"secondary_muscles": []string{"shoulders"},
			"equipment":         "cable",
			"category":          "strength",
			"level":             "intermediate",
			"instructions_pt":   []string{"Mantenha escápulas estabilizadas durante o arco."},
		})
		require.Equal(t, http.StatusCreated, wCreateCustom.Code, wCreateCustom.Body.String())
		var customOut struct {
			ID       string `json:"id"`
			IsCustom bool   `json:"is_custom"`
		}
		require.NoError(t, json.Unmarshal(wCreateCustom.Body.Bytes(), &customOut))
		assert.True(t, customOut.IsCustom)

		wMediaReq := doJSONRequest(t, router, http.MethodPost, "/v1/catalog/exercises/"+customOut.ID+"/media/upload-url", trainerAToken, map[string]string{
			"filename":   "execucao.png",
			"mime_type":  "image/png",
			"media_type": "IMAGE",
		})
		require.Equal(t, http.StatusOK, wMediaReq.Code, wMediaReq.Body.String())
		var mediaReqOut struct {
			MediaID    string `json:"media_id"`
			StorageKey string `json:"storage_key"`
		}
		require.NoError(t, json.Unmarshal(wMediaReq.Body.Bytes(), &mediaReqOut))
		assert.True(t, strings.HasPrefix(mediaReqOut.StorageKey, "catalog/"))

		fakeStorage.putFakeObject(mediaReqOut.StorageKey, 204800, "image/png")
		wMediaConfirm := doJSONRequest(t, router, http.MethodPost, "/v1/catalog/exercises/"+customOut.ID+"/media/confirm", trainerAToken, map[string]any{
			"storage_key": mediaReqOut.StorageKey,
			"filename":    "execucao.png",
			"mime_type":   "image/png",
			"size_bytes":  204800,
		})
		require.Equal(t, http.StatusCreated, wMediaConfirm.Code, wMediaConfirm.Body.String())

		// 11.4 Aluno com vínculo ENDED não acessa customizados; aluno com vínculo ACTIVE vê exercício e obtém URL assinada, mas não pode criar
		noCPFToken, _ := loginAndGetTokens(t, router, "nocpf@fit.test", "NoCPFPass!2026")
		wEndedStudDetail := doJSONRequest(t, router, http.MethodGet, "/v1/catalog/exercises/"+customOut.ID, noCPFToken, nil)
		assert.Equal(t, http.StatusForbidden, wEndedStudDetail.Code)

		wActiveStudCreate := doJSONRequest(t, router, http.MethodPost, "/v1/students", trainerAToken, map[string]string{
			"email":     "aluno.orga@fit.test",
			"full_name": "Aluno Ativo Org A",
			"password":  "OrgAPass!2026",
		})
		require.Equal(t, http.StatusCreated, wActiveStudCreate.Code, wActiveStudCreate.Body.String())
		activeStudToken, _ := loginAndGetTokens(t, router, "aluno.orga@fit.test", "OrgAPass!2026")

		wStudCatalog := doJSONRequest(t, router, http.MethodGet, "/v1/catalog/exercises?scope=custom", activeStudToken, nil)
		require.Equal(t, http.StatusOK, wStudCatalog.Code, wStudCatalog.Body.String())
		assert.Contains(t, wStudCatalog.Body.String(), customOut.ID)

		wStudMedia := doJSONRequest(t, router, http.MethodGet, "/v1/catalog/exercises/"+customOut.ID+"/media/"+mediaReqOut.MediaID+"/signed-url", activeStudToken, nil)
		require.Equal(t, http.StatusOK, wStudMedia.Code, wStudMedia.Body.String())
		assert.Contains(t, wStudMedia.Body.String(), mediaReqOut.StorageKey)

		wStudCreate := doJSONRequest(t, router, http.MethodPost, "/v1/catalog/exercises", activeStudToken, map[string]any{
			"name_pt":        "Proibido para Aluno",
			"primary_muscle": "chest",
			"equipment":      "barbell",
			"category":       "strength",
			"level":          "beginner",
		})
		assert.Equal(t, http.StatusForbidden, wStudCreate.Code)

		// 11.5 Treinador B e aluno de outra org não enxergam nem acessam o exercício customizado do Treinador A
		wTrainerBSearch := doJSONRequest(t, router, http.MethodGet, "/v1/catalog/exercises?scope=custom", trainerBToken, nil)
		require.Equal(t, http.StatusOK, wTrainerBSearch.Code)
		assert.NotContains(t, wTrainerBSearch.Body.String(), customOut.ID)

		wTrainerBDetail := doJSONRequest(t, router, http.MethodGet, "/v1/catalog/exercises/"+customOut.ID, trainerBToken, nil)
		assert.Equal(t, http.StatusForbidden, wTrainerBDetail.Code)

		wTrainerBMedia := doJSONRequest(t, router, http.MethodGet, "/v1/catalog/exercises/"+customOut.ID+"/media/"+mediaReqOut.MediaID+"/signed-url", trainerBToken, nil)
		assert.Equal(t, http.StatusForbidden, wTrainerBMedia.Code)

		wOtherStudCreate := doJSONRequest(t, router, http.MethodPost, "/v1/students", trainerBToken, map[string]string{
			"email":     "aluno.orgb@fit.test",
			"full_name": "Aluno Org B",
			"password":  "OrgBPass!2026",
		})
		require.Equal(t, http.StatusCreated, wOtherStudCreate.Code, wOtherStudCreate.Body.String())
		otherStudToken, _ := loginAndGetTokens(t, router, "aluno.orgb@fit.test", "OrgBPass!2026")

		wOtherStudDetail := doJSONRequest(t, router, http.MethodGet, "/v1/catalog/exercises/"+customOut.ID, otherStudToken, nil)
		assert.Equal(t, http.StatusForbidden, wOtherStudDetail.Code)
	})
}
