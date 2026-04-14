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

// Execute cadastra um aluno novo no sistema.
// É uma operação multi-repositório que protege ativamente o dado pessoal CPF.
func (c *CreateStudentUseCase) Execute(ctx context.Context, input CreateStudentInput) (*StudentOutput, error) {

	// 1. Verificar duplicações (Email)
	existingUsr, _ := c.userRepo.GetByEmail(ctx, input.Email)
	if existingUsr != nil {
		return nil, errors.New("email ou CPF já cadastrados")
	}

	// 2. Cria Entidade Domínio de Student temporária pra ter o CleanCPF
	// Note que precisamos de um UserID fictício temporario pois o User só ganha ID ao persistir.
	// Porem, como somos UUID based, nós mesmos geramos o ID do User de antemão!
	newUserID := uuid.New()

	studentDomain, err := domainS.NewStudent(newUserID, input.CPF)
	if err != nil {
		return nil, err // Ex: Erro de 11 dígitos, etc.
	}

	// 3. Verificar duplicações de CPF (via Hash, pois nós não conseguimos buscar por CPF cifrado)
	cpfHash := c.hashGen.Generate(studentDomain.GetCleanCPF())
	existingStud, _ := c.studRepo.GetByCPFHash(ctx, cpfHash)
	if existingStud != nil {
		return nil, errors.New("email ou CPF já cadastrados")
	}

	// 4. Se tudo está certo, processamos o Hash da Senha
	hashedPass, err := c.passSvc.Hash(input.Password)
	if err != nil {
		return nil, errors.New("falha interna ao processar senha")
	}

	// 5. Configurar Domínio do Usuário
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

	// 6. Criptografar o CPF usando o ID DA ENTIDADE ESTUDANTE como contexto (AAD)
	// Isso garante autenticidade vinculada do AES-GCM
	encryptedCpf, err := c.encryptor.Encrypt(studentDomain.GetCleanCPF(), []byte(studentDomain.ID.String()))
	if err != nil {
		return nil, errors.New("falha interna encriptando CPF")
	}

	// 7. Atualizar entidade de Domínio com os cifrados e LIMPAR A MEMORIA do puro
	studentDomain.CPFEncrypted = encryptedCpf
	studentDomain.CPFHash = cpfHash
	studentDomain.ClearSensitiveData() // <= CRÍTICO! Evita vazamento acidental em logs/panics

	// 8. Persistência (Num futuro, ideal usar Transaction transversal aqui)
	if err := c.userRepo.Create(ctx, userDomain); err != nil {
		return nil, errors.New("falha ao criar credenciais base")
	}

	if err := c.studRepo.Create(ctx, studentDomain); err != nil {
		// Compensacao rollback (pseudo-transacao)
		// Isso mostra a gravidade de desacoplar, SQLX tem metodos melhores mas mantemos simples.
		return nil, errors.New("falha persistindo aluno")
	}

	return &StudentOutput{
		ID:        studentDomain.ID.String(),
		Email:     userDomain.Email,
		FullName:  userDomain.FullName,
		CreatedAt: studentDomain.CreatedAt.Format(time.RFC3339),
	}, nil
}
