package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/SCAPUTO88/fit-brother/internal/domain/invitation"
)

type InvitationRepository struct {
	db *sqlx.DB
}

func NewInvitationRepository(db *sqlx.DB) *InvitationRepository {
	return &InvitationRepository{db: db}
}

func (r *InvitationRepository) Create(ctx context.Context, inv *invitation.Invitation) error {
	query := `
		INSERT INTO invitations (id, organization_id, trainer_user_id, email, full_name, token_hash, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query,
		inv.ID, inv.OrganizationID, inv.TrainerUserID, inv.Email, inv.FullName,
		inv.TokenHash, inv.Status, inv.ExpiresAt, inv.CreatedAt,
	)
	return err
}

func (r *InvitationRepository) GetByID(ctx context.Context, id uuid.UUID) (*invitation.Invitation, error) {
	query := `
		SELECT id, organization_id, trainer_user_id, email, full_name, token_hash, status, expires_at, accepted_at, created_at
		FROM invitations
		WHERE id = $1
	`
	var inv invitation.Invitation
	err := getExecutor(ctx, r.db).GetContext(ctx, &inv, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

func (r *InvitationRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*invitation.Invitation, error) {
	query := `
		SELECT id, organization_id, trainer_user_id, email, full_name, token_hash, status, expires_at, accepted_at, created_at
		FROM invitations
		WHERE token_hash = $1
	`
	if txFromContext(ctx) != nil {
		query += ` FOR UPDATE`
	}
	var inv invitation.Invitation
	err := getExecutor(ctx, r.db).GetContext(ctx, &inv, query, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

func (r *InvitationRepository) ListByTrainer(ctx context.Context, trainerUserID uuid.UUID) ([]*invitation.Invitation, error) {
	query := `
		SELECT id, organization_id, trainer_user_id, email, full_name, token_hash, status, expires_at, accepted_at, created_at
		FROM invitations
		WHERE trainer_user_id = $1
		ORDER BY created_at DESC
	`
	var items []*invitation.Invitation
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &items, query, trainerUserID); err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*invitation.Invitation, 0)
	}
	return items, nil
}

func (r *InvitationRepository) RevokePendingByEmailAndOrg(ctx context.Context, orgID uuid.UUID, email string) error {
	query := `
		UPDATE invitations
		SET status = 'REVOKED'
		WHERE organization_id = $1 AND email = $2 AND status = 'PENDING'
	`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query, orgID, email)
	return err
}

func (r *InvitationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE invitations SET status = $1 WHERE id = $2`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query, status, id)
	return err
}

func (r *InvitationRepository) MarkAccepted(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE invitations SET status = 'ACCEPTED', accepted_at = NOW() WHERE id = $1`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query, id)
	return err
}
