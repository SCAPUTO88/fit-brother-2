package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
}

func (t *PasswordResetToken) IsValid(now time.Time) bool {
	return t.UsedAt == nil && now.Before(t.ExpiresAt)
}

type UserMFA struct {
	UserID          uuid.UUID  `db:"user_id"`
	SecretEncrypted string     `db:"secret_encrypted"`
	Enabled         bool       `db:"enabled"`
	LastUsedStep    int64      `db:"last_used_step"`
	VerifiedAt      *time.Time `db:"verified_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

// PasswordResetRepository isola operações de token de redefinição de senha.
type PasswordResetRepository interface {
	CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error
	GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error
	InvalidatePasswordResetTokensByUser(ctx context.Context, userID uuid.UUID) error
}

// MFARepository isola o armazenamento do segredo TOTP e controle contra replay.
type MFARepository interface {
	GetMFA(ctx context.Context, userID uuid.UUID) (*UserMFA, error)
	UpsertMFA(ctx context.Context, mfa *UserMFA) error
	UpdateMFALastUsedStep(ctx context.Context, userID uuid.UUID, step int64) error
	DisableMFA(ctx context.Context, userID uuid.UUID) error
}
