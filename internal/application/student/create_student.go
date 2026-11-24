package student

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	"github.com/SCAPUTO88/fit-brother/internal/domain/tx"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

type CreateStudentUseCase struct {
	userRepo  domainU.Repository
	studRepo  domainS.Repository
	txManager tx.Manager
	passSvc   *crypto.PasswordService
	encryptor *crypto.AESEncryptor
	hashGen   *crypto.Hash
}

func NewCreateStudentUseCase(
	ur domainU.Repository,
	sr domainS.Repository,
	txManager tx.Manager,
	ps *crypto.PasswordService,
	enc *crypto.AESEncryptor,
	hg *crypto.Hash,
) *CreateStudentUseCase {
	return &CreateStudentUseCase{
		userRepo:  ur,
		studRepo:  sr,
		txManager: txManager,
		passSvc:   ps,
		encryptor: enc,
		hashGen:   hg,
	}
}

func (c *CreateStudentUseCase) Execute(ctx context.Context, input CreateStudentInput) (*StudentOutput, error) {
	if input.TrainerID == uuid.Nil {
		return nil, fmt.Errorf("%w: treinador responsável obrigatório", errs.ErrForbidden)
	}

	email := strings.TrimSpace(input.Email)
	newUserID := uuid.New()

	studentDomain, err := domainS.NewStudent(newUserID, input.TrainerID, input.CPF)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrInvalidInput, err)
	}

	cpfHash := c.hashGen.Generate(studentDomain.GetCleanCPF())

	hashedPass, err := c.passSvc.Hash(input.Password)
	if err != nil {
		studentDomain.ClearSensitiveData()
		return nil, errors.New("falha interna ao processar senha")
	}

	userDomain := &domainU.User{
		ID:           newUserID,
		Email:        email,
		FullName:     strings.TrimSpace(input.FullName),
		PasswordHash: hashedPass,
		Role:         domainU.RoleStudent,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := userDomain.Validate(); err != nil {
		studentDomain.ClearSensitiveData()
		return nil, fmt.Errorf("%w: %v", errs.ErrInvalidInput, err)
	}

	encryptedCpf, err := c.encryptor.Encrypt(studentDomain.GetCleanCPF(), []byte(studentDomain.ID.String()))
	studentDomain.ClearSensitiveData()
	if err != nil {
		return nil, errors.New("falha interna encriptando CPF")
	}

	studentDomain.CPFEncrypted = encryptedCpf
	studentDomain.CPFHash = cpfHash

	run := func(execCtx context.Context) error {
		existingUsr, err := c.userRepo.GetByEmail(execCtx, email)
		if err != nil {
			return errors.New("falha ao verificar email cadastrado")
		}
		if existingUsr != nil {
			return fmt.Errorf("%w: email ou CPF já cadastrados", errs.ErrConflict)
		}

		existingStud, err := c.studRepo.GetByCPFHash(execCtx, cpfHash)
		if err != nil {
			return errors.New("falha ao verificar CPF cadastrado")
		}
		if existingStud != nil {
			return fmt.Errorf("%w: email ou CPF já cadastrados", errs.ErrConflict)
		}

		if err := c.userRepo.Create(execCtx, userDomain); err != nil {
			return errors.New("falha ao criar credenciais base")
		}

		if err := c.studRepo.Create(execCtx, studentDomain); err != nil {
			return errors.New("falha persistindo aluno")
		}

		return nil
	}

	if c.txManager != nil {
		if err := c.txManager.RunInTx(ctx, run); err != nil {
			return nil, err
		}
	} else if err := run(ctx); err != nil {
		return nil, err
	}

	return &StudentOutput{
		ID:        studentDomain.ID.String(),
		Email:     userDomain.Email,
		FullName:  userDomain.FullName,
		CreatedAt: studentDomain.CreatedAt.Format(time.RFC3339),
	}, nil
}
