package invitation

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, inv *Invitation) error
	GetByID(ctx context.Context, id uuid.UUID) (*Invitation, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*Invitation, error)
	ListByTrainer(ctx context.Context, trainerUserID uuid.UUID) ([]*Invitation, error)
	RevokePendingByEmailAndOrg(ctx context.Context, orgID uuid.UUID, email string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	MarkAccepted(ctx context.Context, id uuid.UUID) error
}
