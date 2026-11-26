package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/SCAPUTO88/fit-brother/internal/domain/organization"
)

type OrganizationRepository struct {
	db *sqlx.DB
}

func NewOrganizationRepository(db *sqlx.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) CreateOrganizationWithOwner(ctx context.Context, org *organization.Organization, ownerMembership *organization.Membership) error {
	orgQuery := `
		INSERT INTO organizations (id, name, slug, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if _, err := getExecutor(ctx, r.db).ExecContext(ctx, orgQuery,
		org.ID, org.Name, org.Slug, org.Type, org.CreatedAt, org.UpdatedAt,
	); err != nil {
		return err
	}

	memQuery := `
		INSERT INTO memberships (id, organization_id, user_id, role, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, memQuery,
		ownerMembership.ID, ownerMembership.OrganizationID, ownerMembership.UserID,
		ownerMembership.Role, ownerMembership.Status, ownerMembership.CreatedAt,
	)
	return err
}

func (r *OrganizationRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error) {
	query := `SELECT id, name, slug, type, created_at, updated_at FROM organizations WHERE id = $1`
	var org organization.Organization
	err := getExecutor(ctx, r.db).GetContext(ctx, &org, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

func (r *OrganizationRepository) ListOrganizationsByUser(ctx context.Context, userID uuid.UUID) ([]*organization.OrganizationSummary, error) {
	query := `
		SELECT
			o.id AS organization_id,
			o.name AS organization_name,
			o.slug AS organization_slug,
			o.type AS organization_type,
			m.role AS member_role,
			m.status AS member_status
		FROM memberships m
		INNER JOIN organizations o ON o.id = m.organization_id
		WHERE m.user_id = $1
		ORDER BY o.created_at ASC
	`
	var items []*organization.OrganizationSummary
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &items, query, userID); err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*organization.OrganizationSummary, 0)
	}
	return items, nil
}

func (r *OrganizationRepository) GetMembership(ctx context.Context, orgID, userID uuid.UUID) (*organization.Membership, error) {
	query := `
		SELECT id, organization_id, user_id, role, status, created_at
		FROM memberships
		WHERE organization_id = $1 AND user_id = $2
	`
	var m organization.Membership
	err := getExecutor(ctx, r.db).GetContext(ctx, &m, query, orgID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *OrganizationRepository) AddMember(ctx context.Context, m *organization.Membership) error {
	query := `
		INSERT INTO memberships (id, organization_id, user_id, role, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query,
		m.ID, m.OrganizationID, m.UserID, m.Role, m.Status, m.CreatedAt,
	)
	return err
}

func (r *OrganizationRepository) ListMembers(ctx context.Context, orgID uuid.UUID) ([]*organization.MemberSummary, error) {
	query := `
		SELECT m.id, m.user_id, u.full_name, u.email, m.role, m.status
		FROM memberships m
		INNER JOIN users u ON u.id = m.user_id
		WHERE m.organization_id = $1 AND u.deleted_at IS NULL
		ORDER BY u.full_name ASC
	`
	var items []*organization.MemberSummary
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &items, query, orgID); err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*organization.MemberSummary, 0)
	}
	return items, nil
}

func (r *OrganizationRepository) EnsureTrainerDefaultOrg(ctx context.Context, trainerUserID uuid.UUID, trainerName string) (*organization.Organization, error) {
	existingQuery := `
		SELECT o.id, o.name, o.slug, o.type, o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN memberships m ON m.organization_id = o.id
		WHERE m.user_id = $1 AND m.status = 'ACTIVE'
		ORDER BY o.created_at ASC
		LIMIT 1
	`
	var org organization.Organization
	err := getExecutor(ctx, r.db).GetContext(ctx, &org, existingQuery, trainerUserID)
	if err == nil {
		return &org, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	slug := "org-" + strings.ReplaceAll(trainerUserID.String(), "-", "")
	name := strings.TrimSpace(trainerName)
	if name == "" {
		name = "Personal"
	}
	name = name + " Studio"

	insertOrg := `
		INSERT INTO organizations (id, name, slug, type)
		VALUES ($1, $2, $3, 'SOLO')
		ON CONFLICT (slug) DO NOTHING
	`
	if _, err := getExecutor(ctx, r.db).ExecContext(ctx, insertOrg, uuid.New(), name, slug); err != nil {
		return nil, err
	}

	if err := getExecutor(ctx, r.db).GetContext(ctx, &org, `SELECT id, name, slug, type, created_at, updated_at FROM organizations WHERE slug = $1`, slug); err != nil {
		return nil, err
	}

	insertMem := `
		INSERT INTO memberships (id, organization_id, user_id, role, status)
		VALUES ($1, $2, $3, 'OWNER', 'ACTIVE')
		ON CONFLICT (organization_id, user_id) DO NOTHING
	`
	if _, err := getExecutor(ctx, r.db).ExecContext(ctx, insertMem, uuid.New(), org.ID, trainerUserID); err != nil {
		return nil, err
	}

	return &org, nil
}

func (r *OrganizationRepository) CreateRelationship(ctx context.Context, rel *organization.TrainerStudentRelationship) error {
	query := `
		INSERT INTO trainer_student_relationships (id, organization_id, trainer_user_id, student_id, status, started_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query,
		rel.ID, rel.OrganizationID, rel.TrainerUserID, rel.StudentID, rel.Status, rel.StartedAt,
	)
	return err
}

func (r *OrganizationRepository) HasActiveRelationship(ctx context.Context, trainerUserID, studentID uuid.UUID) (bool, error) {
	query := `
		SELECT (
			EXISTS (
				SELECT 1 FROM trainer_student_relationships r
				WHERE r.trainer_user_id = $1 AND r.student_id = $2 AND r.status = 'ACTIVE'
			)
			OR (
				EXISTS (
					SELECT 1 FROM students s
					WHERE s.id = $2 AND s.trainer_id = $1
				)
				AND NOT EXISTS (
					SELECT 1 FROM trainer_student_relationships r2
					WHERE r2.trainer_user_id = $1 AND r2.student_id = $2
				)
			)
		)
	`
	var ok bool
	err := getExecutor(ctx, r.db).GetContext(ctx, &ok, query, trainerUserID, studentID)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (r *OrganizationRepository) UpdateRelationshipStatus(ctx context.Context, orgID, trainerUserID, studentID uuid.UUID, status string) error {
	var endedAt *time.Time
	if status == organization.RelEnded {
		now := time.Now().UTC()
		endedAt = &now
	}
	query := `
		UPDATE trainer_student_relationships
		SET status = $1,
		    ended_at = $2
		WHERE organization_id = $3 AND trainer_user_id = $4 AND student_id = $5
	`
	res, err := getExecutor(ctx, r.db).ExecContext(ctx, query, status, endedAt, orgID, trainerUserID, studentID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *OrganizationRepository) EndAllByStudent(ctx context.Context, studentID uuid.UUID) error {
	query := `
		UPDATE trainer_student_relationships
		SET status = 'ENDED', ended_at = NOW()
		WHERE student_id = $1 AND status <> 'ENDED'
	`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query, studentID)
	return err
}
