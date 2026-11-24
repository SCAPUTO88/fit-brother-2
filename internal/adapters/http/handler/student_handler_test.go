package handler_test

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
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/handler"
	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/middleware"
	"github.com/SCAPUTO88/fit-brother/internal/application/student"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

type HTTPMockUserRepo struct {
	mock.Mock
	activeUsers map[uuid.UUID]bool
}

func (m *HTTPMockUserRepo) Create(ctx context.Context, u *domainU.User) error { return nil }
func (m *HTTPMockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainU.User, error) {
	return nil, nil
}
func (m *HTTPMockUserRepo) GetByEmail(ctx context.Context, email string) (*domainU.User, error) {
	return nil, nil
}
func (m *HTTPMockUserRepo) Update(ctx context.Context, u *domainU.User) error { return nil }
func (m *HTTPMockUserRepo) CreateRefreshToken(ctx context.Context, rt *domainU.RefreshToken) error {
	return nil
}
func (m *HTTPMockUserRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*domainU.RefreshToken, error) {
	return nil, nil
}
func (m *HTTPMockUserRepo) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error { return nil }
func (m *HTTPMockUserRepo) RevokeTokenFamily(ctx context.Context, familyID uuid.UUID) error {
	return nil
}
func (m *HTTPMockUserRepo) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *HTTPMockUserRepo) IsUserActive(_ context.Context, userID uuid.UUID) (bool, error) {
	if m.activeUsers == nil {
		return true, nil
	}
	return m.activeUsers[userID], nil
}

type HTTPMockStudentRepo struct {
	mock.Mock
	createdStudent *domainS.Student
	studentByID    map[uuid.UUID]*domainS.Student
	listResult     []*domainS.StudentBaseInfo
	lastTrainerID  uuid.UUID
}

func (m *HTTPMockStudentRepo) Create(_ context.Context, s *domainS.Student) error {
	cp := *s
	m.createdStudent = &cp
	return nil
}
func (m *HTTPMockStudentRepo) GetByID(_ context.Context, id uuid.UUID) (*domainS.Student, error) {
	if m.studentByID != nil {
		return m.studentByID[id], nil
	}
	return nil, nil
}
func (m *HTTPMockStudentRepo) GetByUserID(_ context.Context, _ uuid.UUID) (*domainS.Student, error) {
	return nil, nil
}
func (m *HTTPMockStudentRepo) GetByCPFHash(_ context.Context, _ string) (*domainS.Student, error) {
	return nil, nil
}
func (m *HTTPMockStudentRepo) ListByTrainer(_ context.Context, trainerID uuid.UUID) ([]*domainS.StudentBaseInfo, error) {
	m.lastTrainerID = trainerID
	if m.listResult == nil {
		return []*domainS.StudentBaseInfo{}, nil
	}
	return m.listResult, nil
}
func (m *HTTPMockStudentRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (m *HTTPMockStudentRepo) SaveBiometric(_ context.Context, _ *domainS.BiometricData) error {
	return nil
}

func newHandlerTestJWT(t *testing.T) *jwt.Service {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	svc, err := jwt.NewServiceFromPEM(string(privPEM), string(pubPEM), "fit-brother-test")
	require.NoError(t, err)
	return svc
}

func TestStudentHandler_SecurityBaseline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	aesKey := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
	encryptor, _ := crypto.NewAESEncryptor(aesKey)
	hashGen, _ := crypto.NewHash(base64.StdEncoding.EncodeToString([]byte("pepper123-------")))
	passSvc, _ := crypto.NewPasswordService("pepper")

	jwtSvc := newHandlerTestJWT(t)
	trainerA := uuid.New()
	trainerB := uuid.New()
	studentUser := uuid.New()

	trainerAToken, err := jwtSvc.GenerateAccessToken(trainerA, domainU.RoleTrainer, uuid.NewString(), 15*time.Minute)
	require.NoError(t, err)
	trainerBToken, err := jwtSvc.GenerateAccessToken(trainerB, domainU.RoleTrainer, uuid.NewString(), 15*time.Minute)
	require.NoError(t, err)
	studentToken, err := jwtSvc.GenerateAccessToken(studentUser, domainU.RoleStudent, uuid.NewString(), 15*time.Minute)
	require.NoError(t, err)

	ur := &HTTPMockUserRepo{
		activeUsers: map[uuid.UUID]bool{
			trainerA:    true,
			trainerB:    true,
			studentUser: true,
		},
	}
	stud1ID := uuid.New()
	sr := &HTTPMockStudentRepo{
		studentByID: map[uuid.UUID]*domainS.Student{
			stud1ID: {ID: stud1ID, UserID: studentUser, TrainerID: &trainerA},
		},
		listResult: []*domainS.StudentBaseInfo{
			{ID: stud1ID.String(), FullName: "Lucas", Email: "lucas@fit.com"},
		},
	}

	createUC := student.NewCreateStudentUseCase(ur, sr, nil, passSvc, encryptor, hashGen)
	getUC := student.NewGetStudentUseCase(ur, sr, encryptor)
	listUC := student.NewListStudentsUseCase(sr)
	anonymizeUC := student.NewAnonymizeStudentUseCase(ur, sr, nil)

	h := handler.NewStudentHandler(createUC, getUC, listUC, anonymizeUC)

	router := gin.New()
	group := router.Group("")
	h.RegisterRoutes(group, middleware.RequireAuth(jwtSvc, ur))

	t.Run("POST /students sem token deve retornar 401", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"email":     "lucas@fit.com",
			"full_name": "Lucas",
			"password":  "senha123",
			"cpf":       "12345678901",
		})
		req := httptest.NewRequest(http.MethodPost, "/students", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("POST /students com token de STUDENT deve retornar 403", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"email":     "lucas@fit.com",
			"full_name": "Lucas",
			"password":  "senha123",
			"cpf":       "12345678901",
		})
		req := httptest.NewRequest(http.MethodPost, "/students", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+studentToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("POST /students com token de TRAINER deve criar aluno 201 vinculado ao trainer", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"email":     "lucas@fit.com",
			"full_name": "Lucas",
			"password":  "senha123",
			"cpf":       "123.456.789-01",
		})
		req := httptest.NewRequest(http.MethodPost, "/students", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+trainerAToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		require.NotNil(t, sr.createdStudent)
		require.NotNil(t, sr.createdStudent.TrainerID)
		assert.Equal(t, trainerA, *sr.createdStudent.TrainerID)
	})

	t.Run("POST /students com payload invalido deve retornar 400", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"email": "invalido"})
		req := httptest.NewRequest(http.MethodPost, "/students", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+trainerAToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /students deve bloquear STUDENT e filtrar por Trainer para TRAINER", func(t *testing.T) {
		reqStudent := httptest.NewRequest(http.MethodGet, "/students", nil)
		reqStudent.Header.Set("Authorization", "Bearer "+studentToken)
		wStudent := httptest.NewRecorder()
		router.ServeHTTP(wStudent, reqStudent)
		assert.Equal(t, http.StatusForbidden, wStudent.Code)

		reqTrainer := httptest.NewRequest(http.MethodGet, "/students", nil)
		reqTrainer.Header.Set("Authorization", "Bearer "+trainerAToken)
		wTrainer := httptest.NewRecorder()
		router.ServeHTTP(wTrainer, reqTrainer)
		assert.Equal(t, http.StatusOK, wTrainer.Code)
		assert.Equal(t, trainerA, sr.lastTrainerID)
	})

	t.Run("GET e DELETE /students/:id devem bloquear acesso cruzado entre treinadores", func(t *testing.T) {
		reqGet := httptest.NewRequest(http.MethodGet, "/students/"+stud1ID.String(), nil)
		reqGet.Header.Set("Authorization", "Bearer "+trainerBToken)
		wGet := httptest.NewRecorder()
		router.ServeHTTP(wGet, reqGet)
		assert.Equal(t, http.StatusForbidden, wGet.Code)

		reqDel := httptest.NewRequest(http.MethodDelete, "/students/"+stud1ID.String(), nil)
		reqDel.Header.Set("Authorization", "Bearer "+trainerBToken)
		wDel := httptest.NewRecorder()
		router.ServeHTTP(wDel, reqDel)
		assert.Equal(t, http.StatusForbidden, wDel.Code)
	})
}
