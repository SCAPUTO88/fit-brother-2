package student_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/application/student"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

// --- Mocks ---

type MockTxManager struct {
	calls int
}

func (m *MockTxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	m.calls++
	return fn(ctx)
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, u *domainU.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}
func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainU.User, error) {
	args := m.Called(ctx, id)
	if usr := args.Get(0); usr != nil {
		return usr.(*domainU.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*domainU.User, error) {
	args := m.Called(ctx, email)
	if usr := args.Get(0); usr != nil {
		return usr.(*domainU.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepo) Update(ctx context.Context, u *domainU.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}
func (m *MockUserRepo) CreateRefreshToken(ctx context.Context, rt *domainU.RefreshToken) error {
	return nil
}
func (m *MockUserRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*domainU.RefreshToken, error) {
	return nil, nil
}
func (m *MockUserRepo) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error      { return nil }
func (m *MockUserRepo) RevokeTokenFamily(ctx context.Context, familyID uuid.UUID) error { return nil }
func (m *MockUserRepo) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockStudentRepo struct {
	mock.Mock
}

func (m *MockStudentRepo) Create(ctx context.Context, u *domainS.Student) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}
func (m *MockStudentRepo) GetByID(ctx context.Context, id uuid.UUID) (*domainS.Student, error) {
	args := m.Called(ctx, id)
	if std := args.Get(0); std != nil {
		return std.(*domainS.Student), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockStudentRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domainS.Student, error) {
	args := m.Called(ctx, userID)
	if std := args.Get(0); std != nil {
		return std.(*domainS.Student), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockStudentRepo) GetByCPFHash(ctx context.Context, cpfHash string) (*domainS.Student, error) {
	args := m.Called(ctx, cpfHash)
	if std := args.Get(0); std != nil {
		return std.(*domainS.Student), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockStudentRepo) ListByTrainer(ctx context.Context, trainerID uuid.UUID) ([]*domainS.StudentBaseInfo, error) {
	args := m.Called(ctx, trainerID)
	if students := args.Get(0); students != nil {
		return students.([]*domainS.StudentBaseInfo), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockStudentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockStudentRepo) SaveBiometric(ctx context.Context, b *domainS.BiometricData) error {
	return nil
}

// --- Test Suite ---

func TestCreateStudentUseCase(t *testing.T) {
	ctx := context.Background()

	aesKey := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
	encryptor, _ := crypto.NewAESEncryptor(aesKey)

	hmacSecret := base64.StdEncoding.EncodeToString([]byte("HMAC_SECRET_LITERAL_010101010101"))
	hashGen, _ := crypto.NewHash(hmacSecret)

	passSvc, _ := crypto.NewPasswordService("pepper")

	t.Run("Deve criar Aluno em transacao vinculado ao TrainerID, criptografar CPF e limpar string crua", func(t *testing.T) {
		mockURepo := new(MockUserRepo)
		mockSRepo := new(MockStudentRepo)
		txMgr := &MockTxManager{}

		sut := student.NewCreateStudentUseCase(mockURepo, mockSRepo, txMgr, passSvc, encryptor, hashGen)
		trainerID := uuid.New()

		input := student.CreateStudentInput{
			TrainerID: trainerID,
			Email:     "test@fit.com",
			FullName:  "Joao Teste",
			Password:  "senha_segura_123",
			CPF:       "123.456.789-01",
		}

		mockURepo.On("GetByEmail", ctx, "test@fit.com").Return((*domainU.User)(nil), nil)
		mockSRepo.On("GetByCPFHash", ctx, mock.Anything).Return((*domainS.Student)(nil), nil)

		mockURepo.On("Create", ctx, mock.MatchedBy(func(u *domainU.User) bool {
			return u.Email == input.Email && u.Role == domainU.RoleStudent && u.IsActive
		})).Return(nil)

		mockSRepo.On("Create", ctx, mock.MatchedBy(func(s *domainS.Student) bool {
			if s.GetCleanCPF() != "" {
				return false
			}
			return s.TrainerID != nil && *s.TrainerID == trainerID && s.CPFHash != "" && s.CPFEncrypted != "" && s.CPFEncrypted != "12345678901"
		})).Return(nil)

		out, err := sut.Execute(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, "test@fit.com", out.Email)
		assert.Equal(t, 1, txMgr.calls)

		mockURepo.AssertExpectations(t)
		mockSRepo.AssertExpectations(t)
	})

	t.Run("Deve falhar com ErrConflict se Email ja existe", func(t *testing.T) {
		mockURepo := new(MockUserRepo)
		mockSRepo := new(MockStudentRepo)
		txMgr := &MockTxManager{}

		mockURepo.On("GetByEmail", ctx, "duplicado@fit.com").Return(&domainU.User{}, nil)

		sut := student.NewCreateStudentUseCase(mockURepo, mockSRepo, txMgr, passSvc, encryptor, hashGen)
		_, err := sut.Execute(ctx, student.CreateStudentInput{
			TrainerID: uuid.New(),
			Email:     "duplicado@fit.com",
			FullName:  "Fulano",
			Password:  "123456",
			CPF:       "00000000000",
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrConflict))
		assert.Contains(t, err.Error(), "email ou CPF já cadastrados")
	})

	t.Run("Deve propagar erro de infraestrutura sem ignorar GetByEmail", func(t *testing.T) {
		mockURepo := new(MockUserRepo)
		mockSRepo := new(MockStudentRepo)
		txMgr := &MockTxManager{}

		mockURepo.On("GetByEmail", ctx, "db@fit.com").Return((*domainU.User)(nil), errors.New("connection refused"))

		sut := student.NewCreateStudentUseCase(mockURepo, mockSRepo, txMgr, passSvc, encryptor, hashGen)
		_, err := sut.Execute(ctx, student.CreateStudentInput{
			TrainerID: uuid.New(),
			Email:     "db@fit.com",
			FullName:  "Fulano",
			Password:  "123456",
			CPF:       "00000000000",
		})
		require.Error(t, err)
		assert.False(t, errors.Is(err, errs.ErrConflict))
	})
}
