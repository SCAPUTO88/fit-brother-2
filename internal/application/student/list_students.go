package student

import (
	"context"

	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
)

type ListStudentsUseCase struct {
	studRepo domainS.Repository
}

func NewListStudentsUseCase(sr domainS.Repository) *ListStudentsUseCase {
	return &ListStudentsUseCase{
		studRepo: sr,
	}
}

func (c *ListStudentsUseCase) Execute(ctx context.Context) ([]*domainS.StudentBaseInfo, error) {
	// Futuramente podemos injetar o `TrainerID` aqui se quisermos que treinadores
	// vejam apenas seus próprios alunos em um ambiente SaaS multi-tenant.
	return c.studRepo.ListAll(ctx)
}
