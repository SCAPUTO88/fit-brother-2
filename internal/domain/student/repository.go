package student

import (
	"context"

	"github.com/google/uuid"
)

// Repository isola a camada de Domínio de Alunos das implementações de BD.
type Repository interface {
	Create(ctx context.Context, u *Student) error
	GetByID(ctx context.Context, id uuid.UUID) (*Student, error)
	GetByCPFHash(ctx context.Context, cpfHash string) (*Student, error)
	ListAll(ctx context.Context) ([]*StudentBaseInfo, error)
	Delete(ctx context.Context, id uuid.UUID) error // Usado no Direito ao Esquecimento (LGPD)

	SaveBiometric(ctx context.Context, b *BiometricData) error
}
