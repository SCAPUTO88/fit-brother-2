package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

// UserRepository implementa a interface do Domínio usando a lib sqlx para abstração de DB.
type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (id, email, full_name, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		u.ID, u.Email, u.FullName, u.PasswordHash, u.Role, u.IsActive, u.CreatedAt, u.UpdatedAt,
	)

	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	query := `
		SELECT id, email, full_name, password_hash, role, is_active, created_at, updated_at, deleted_at 
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL
	`
	var usr user.User
	err := r.db.GetContext(ctx, &usr, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Não encontrado
		}
		return nil, err
	}
	return &usr, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id, email, full_name, password_hash, role, is_active, created_at, updated_at, deleted_at 
		FROM users 
		WHERE email = $1 AND deleted_at IS NULL
	`
	var usr user.User
	err := r.db.GetContext(ctx, &usr, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &usr, nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE users 
		SET email = $1, full_name = $2, password_hash = $3, role = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query,
		u.Email, u.FullName, u.PasswordHash, u.Role, u.IsActive, u.ID,
	)
	return err
}

func (r *UserRepository) CreateRefreshToken(ctx context.Context, rt *user.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, family_id, is_revoked, expires_at, created_ip)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, query,
		rt.UserID, rt.TokenHash, rt.FamilyID, rt.IsRevoked, rt.ExpiresAt, rt.CreatedIP,
	).Scan(&rt.ID, &rt.CreatedAt)

	return err
}

// GetRefreshTokenByHash pega o token direto por Hash. Isso faz da busca extremamente rápida indexada.
func (r *UserRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*user.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, family_id, is_revoked, expires_at, created_at, created_ip
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var rt user.RefreshToken
	err := r.db.GetContext(ctx, &rt, query, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rt, nil
}

// RevokeRefreshToken marca UM RT como revogado em um Update
func (r *UserRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET is_revoked = TRUE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// RevokeTokenFamily é chamado via "Token Family Rotation". Revoga tudo que tiver o mesmo Family ID!
func (r *UserRepository) RevokeTokenFamily(ctx context.Context, familyID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET is_revoked = TRUE WHERE family_id = $1`
	_, err := r.db.ExecContext(ctx, query, familyID)
	return err
}
