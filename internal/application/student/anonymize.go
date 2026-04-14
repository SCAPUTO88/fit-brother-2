package student

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

type AnonymizeStudentUseCase struct {
	userRepo domainU.Repository
	studRepo domainS.Repository
}

func NewAnonymizeStudentUseCase(ur domainU.Repository, sr domainS.Repository) *AnonymizeStudentUseCase {
	return &AnonymizeStudentUseCase{
		userRepo: ur,
		studRepo: sr,
	}
}

// Execute cumpre os requisitos de Direito ao Esquecimento da LGPD.
// Ação Irreversível. O Perfil core do User entra em SoftDelete e a Ficha Student/Biometria é Purgada do BD permanentemente.
func (c *AnonymizeStudentUseCase) Execute(ctx context.Context, studentID uuid.UUID) error {
	stud, err := c.studRepo.GetByID(ctx, studentID)
	if err != nil || stud == nil {
		return errors.New("aluno não encontrado")
	}

	usr, err := c.userRepo.GetByID(ctx, stud.UserID)
	if err != nil || usr == nil {
		return errors.New("perfil base corrompido ou inativo")
	}

	usr.IsActive = false
	// Podemos "Zerar" o email do cara aqui no profile tbm para que o email dele seja reaproveitavel num signup novo
	// usr.Email = fmt.Sprintf("erased_%d", time.Now().UnixNano())

	err = c.userRepo.Update(ctx, usr)
	if err != nil {
		return err
	}

	// Como fizemos a tabela biometric_data ter ON DELETE CASCADE (foreign key constraint) do StudentID,
	// HardDeleatar o Student removerá para sempre o PDF do CPF_Encrypted, o Hash, o peso e a altura do banco
	// limpando a trilha Pessoal Identificável de fato.
	return c.studRepo.Delete(ctx, stud.ID)
}
