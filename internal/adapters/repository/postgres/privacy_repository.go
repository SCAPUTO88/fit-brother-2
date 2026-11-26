package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
)

type PrivacyRepository struct {
	db *sqlx.DB
}

func NewPrivacyRepository(db *sqlx.DB) *PrivacyRepository {
	return &PrivacyRepository{db: db}
}

func (r *PrivacyRepository) CreateConsent(ctx context.Context, record *privacy.ConsentRecord) error {
	query := `
		INSERT INTO consent_records (id, user_id, purpose, policy_version, granted, ip_address, user_agent, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query,
		record.ID, record.UserID, record.Purpose, record.PolicyVersion,
		record.Granted, record.IPAddress, record.UserAgent, record.RecordedAt,
	)
	return err
}

func (r *PrivacyRepository) ListConsentsByUser(ctx context.Context, userID uuid.UUID) ([]*privacy.ConsentRecord, error) {
	query := `
		SELECT id, user_id, purpose, policy_version, granted, host(ip_address) AS ip_address, user_agent, recorded_at
		FROM consent_records
		WHERE user_id = $1
		ORDER BY recorded_at DESC
	`
	var items []*privacy.ConsentRecord
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &items, query, userID); err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*privacy.ConsentRecord, 0)
	}
	return items, nil
}

func (r *PrivacyRepository) CreatePrivacyRequest(ctx context.Context, req *privacy.PrivacyRequest) error {
	query := `
		INSERT INTO privacy_requests (id, user_id, request_type, status, details, requested_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query,
		req.ID, req.UserID, req.RequestType, req.Status, req.Details, req.RequestedAt, req.CompletedAt,
	)
	return err
}

func (r *PrivacyRepository) ListPrivacyRequestsByUser(ctx context.Context, userID uuid.UUID) ([]*privacy.PrivacyRequest, error) {
	query := `
		SELECT id, user_id, request_type, status, details, requested_at, completed_at
		FROM privacy_requests
		WHERE user_id = $1
		ORDER BY requested_at DESC
	`
	var items []*privacy.PrivacyRequest
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &items, query, userID); err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*privacy.PrivacyRequest, 0)
	}
	return items, nil
}

func (r *PrivacyRepository) CompletePrivacyRequestsByUserAndType(ctx context.Context, userID uuid.UUID, reqType string) error {
	query := `
		UPDATE privacy_requests
		SET status = 'COMPLETED', completed_at = NOW()
		WHERE user_id = $1 AND request_type = $2 AND status = 'OPEN'
	`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query, userID, reqType)
	return err
}

func (r *PrivacyRepository) RecordAuditEvent(ctx context.Context, evt *privacy.AuditEvent) error {
	metadata := evt.MetadataJSON
	if metadata == "" {
		metadata = "{}"
	}
	query := `
		INSERT INTO audit_events (id, actor_user_id, organization_id, action, target_type, target_id, metadata, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9)
	`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query,
		evt.ID, evt.ActorUserID, evt.OrganizationID, evt.Action,
		evt.TargetType, evt.TargetID, metadata, evt.IPAddress, evt.CreatedAt,
	)
	return err
}

func (r *PrivacyRepository) ListAuditEventsForUser(ctx context.Context, userID uuid.UUID, limit int) ([]*privacy.AuditEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := `
		SELECT
			id,
			actor_user_id,
			organization_id,
			action,
			target_type,
			target_id,
			metadata::text AS metadata,
			host(ip_address) AS ip_address,
			created_at
		FROM audit_events
		WHERE actor_user_id = $1
		   OR organization_id IN (
				SELECT organization_id FROM memberships WHERE user_id = $1 AND status = 'ACTIVE'
		   )
		ORDER BY created_at DESC
		LIMIT $2
	`
	var items []*privacy.AuditEvent
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &items, query, userID, limit); err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*privacy.AuditEvent, 0)
	}
	return items, nil
}
