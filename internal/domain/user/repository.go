package user

import (
	"context"

	"github.com/google/uuid"
)

// Repository define a Port (Hexagonal Architecture) que deve ser
// implementada pelo Adapter de infraestrutura (PostgreSQL).
type Repository interface {
	// User
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error

	// Refresh Tokens
	CreateRefreshToken(ctx context.Context, rt *RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
	RevokeTokenFamily(ctx context.Context, familyID uuid.UUID) error
}
