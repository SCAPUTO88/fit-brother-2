package student

import (
	"context"

	"github.com/google/uuid"

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

func (c *ListStudentsUseCase) Execute(ctx context.Context, trainerID uuid.UUID) ([]*domainS.StudentBaseInfo, error) {
	return c.studRepo.ListByTrainer(ctx, trainerID)
}
