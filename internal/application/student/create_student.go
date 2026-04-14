package student

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

type CreateStudentUseCase struct {
	userRepo  domainU.Repository
	studRepo  domainS.Repository
	passSvc   *crypto.PasswordService
	encryptor *crypto.AESEncryptor
	hashGen   *crypto.Hash
}

func NewCreateStudentUseCase(
	ur domainU.Repository,
	sr domainS.Repository,
	ps *crypto.PasswordService,
	enc *crypto.AESEncryptor,
	hg *crypto.Hash,
) *CreateStudentUseCase {
	return &CreateStudentUseCase{
		userRepo:  ur,
		studRepo:  sr,
		passSvc:   ps,
		encryptor: enc,
		hashGen:   hg,
	}
}

func (c *CreateStudentUseCase) Execute(ctx context.Context, input CreateStudentInput) (*StudentOutput, error) {
	existingUsr, _ := c.userRepo.GetByEmail(ctx, input.Email)
	if existingUsr != nil {
		return nil, errors.New("email ou CPF já cadastrados")
	}

	newUserID := uuid.New()

	studentDomain, err := domainS.NewStudent(newUserID, input.CPF)
	if err != nil {
		return nil, err
	}

	cpfHash := c.hashGen.Generate(studentDomain.GetCleanCPF())
	existingStud, _ := c.studRepo.GetByCPFHash(ctx, cpfHash)
	if existingStud != nil {
		return nil, errors.New("email ou CPF já cadastrados")
	}

	hashedPass, err := c.passSvc.Hash(input.Password)
	if err != nil {
		return nil, errors.New("falha interna ao processar senha")
	}

	userDomain := &domainU.User{
		ID:           newUserID,
		Email:        input.Email,
		FullName:     input.FullName,
		PasswordHash: hashedPass,
		Role:         domainU.RoleStudent,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	encryptedCpf, err := c.encryptor.Encrypt(studentDomain.GetCleanCPF(), []byte(studentDomain.ID.String()))
	if err != nil {
		return nil, errors.New("falha interna encriptando CPF")
	}

	studentDomain.CPFEncrypted = encryptedCpf
	studentDomain.CPFHash = cpfHash
	studentDomain.ClearSensitiveData()

	if err := c.userRepo.Create(ctx, userDomain); err != nil {
		return nil, errors.New("falha ao criar credenciais base")
	}

	if err := c.studRepo.Create(ctx, studentDomain); err != nil {
		return nil, errors.New("falha persistindo aluno")
	}

	return &StudentOutput{
		ID:        studentDomain.ID.String(),
		Email:     userDomain.Email,
		FullName:  userDomain.FullName,
		CreatedAt: studentDomain.CreatedAt.Format(time.RFC3339),
	}, nil
}
