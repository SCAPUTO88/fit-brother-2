package handler_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/http/handler"
	"github.com/SCAPUTO88/fit-brother/internal/application/student"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

// O teste HTTP injeta Repositorios Falsos pra testar se o Gin converte bem o Request pra Response.

// --- Mocks simplificados ---
type HTTPMockUserRepo struct{ mock.Mock }
func (m *HTTPMockUserRepo) Create(ctx context.Context, u *domainU.User) error                                { return nil }
func (m *HTTPMockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainU.User, error)                 { return nil, nil }
func (m *HTTPMockUserRepo) GetByEmail(ctx context.Context, email string) (*domainU.User, error)              { return nil, nil }
func (m *HTTPMockUserRepo) Update(ctx context.Context, u *domainU.User) error                                { return nil }
func (m *HTTPMockUserRepo) CreateRefreshToken(ctx context.Context, rt *domainU.RefreshToken) error           { return nil }
func (m *HTTPMockUserRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*domainU.RefreshToken, error) { return nil, nil }
func (m *HTTPMockUserRepo) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error                       { return nil }
func (m *HTTPMockUserRepo) RevokeTokenFamily(ctx context.Context, familyID uuid.UUID) error                  { return nil }

type HTTPMockStudentRepo struct{ mock.Mock }
func (m *HTTPMockStudentRepo) Create(ctx context.Context, u *domainS.Student) error                          { return nil }
func (m *HTTPMockStudentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainS.Student, error)           { return nil, nil }
func (m *HTTPMockStudentRepo) GetByCPFHash(ctx context.Context, cpf string) (*domainS.Student, error)        { return nil, nil }
func (m *HTTPMockStudentRepo) Delete(ctx context.Context, id uuid.UUID) error                                { return nil }
func (m *HTTPMockStudentRepo) SaveBiometric(ctx context.Context, b *domainS.BiometricData) error             { return nil }

func TestStudentHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Instanciando UseCase c/ dependências dummy
	aesKey := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
	encryptor, _ := crypto.NewAESEncryptor(aesKey)
	hashGen, _ := crypto.NewHash(base64.StdEncoding.EncodeToString([]byte("pepper123-------")))
	passSvc, _ := crypto.NewPasswordService("pepper")

	ur := new(HTTPMockUserRepo)
	sr := new(HTTPMockStudentRepo)

	uc := student.NewCreateStudentUseCase(ur, sr, passSvc, encryptor, hashGen)
	h := handler.NewStudentHandler(uc, nil, nil) // Mock get/anonymize as nil for now

	router := gin.New()
	group := router.Group("")
	h.RegisterRoutes(group, nil) // Middlewares mocked to pass-through

	t.Run("Deve retornar 200 OK criando via rota /students", func(t *testing.T) {
		body := map[string]string{
			"email":     "lucas@fit.com",
			"full_name": "Lucas",
			"password":  "senha123",
			"cpf":       "123.456.789-01",
		}
		jsonValue, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/students", bytes.NewBuffer(jsonValue))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Deve falhar HTTP 400 por erro de binding (Falta de atributos obrigatórios)", func(t *testing.T) {
		body := map[string]string{
			"email": "invalid_email_no_at_sign",
		}
		jsonValue, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/students", bytes.NewBuffer(jsonValue))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "dados inválidos")
	})
}
