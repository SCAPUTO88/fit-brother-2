package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	domain "github.com/SCAPUTO88/fit-brother/internal/domain/workout"
)

type WorkoutBuilderRepository struct {
	db *sqlx.DB
}

func NewWorkoutBuilderRepository(db *sqlx.DB) *WorkoutBuilderRepository {
	return &WorkoutBuilderRepository{db: db}
}

type templateRow struct {
	ID                   uuid.UUID  `db:"id"`
	OrganizationID       uuid.UUID  `db:"organization_id"`
	TrainerUserID        uuid.UUID  `db:"trainer_user_id"`
	Title                string     `db:"title"`
	Description          string     `db:"description"`
	Goal                 string     `db:"goal"`
	DraftStructure       []byte     `db:"draft_structure"`
	ClonedFromTemplateID *uuid.UUID `db:"cloned_from_template_id"`
	IsArchived           bool       `db:"is_archived"`
	LatestVersionNumber  int        `db:"latest_version_number"`
	LatestVersionID      *uuid.UUID `db:"latest_version_id"`
	CreatedAt            time.Time  `db:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at"`
}

func (row *templateRow) toEntity() (*domain.WorkoutTemplate, error) {
	var st domain.WorkoutStructure
	if len(row.DraftStructure) > 0 {
		if err := json.Unmarshal(row.DraftStructure, &st); err != nil {
			return nil, err
		}
	}
	if st.Sessions == nil {
		st.Sessions = []domain.TemplateSession{}
	}
	return &domain.WorkoutTemplate{
		ID:                   row.ID,
		OrganizationID:       row.OrganizationID,
		TrainerUserID:        row.TrainerUserID,
		Title:                row.Title,
		Description:          row.Description,
		Goal:                 row.Goal,
		DraftStructure:       st,
		ClonedFromTemplateID: row.ClonedFromTemplateID,
		IsArchived:           row.IsArchived,
		LatestVersionNumber:  row.LatestVersionNumber,
		LatestVersionID:      row.LatestVersionID,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}, nil
}

type versionRow struct {
	ID                uuid.UUID  `db:"id"`
	TemplateID        uuid.UUID  `db:"template_id"`
	VersionNumber     int        `db:"version_number"`
	Title             string     `db:"title"`
	Description       string     `db:"description"`
	Goal              string     `db:"goal"`
	SnapshotStructure []byte     `db:"snapshot_structure"`
	PublishedByUserID *uuid.UUID `db:"published_by_user_id"`
	PublishedAt       time.Time  `db:"published_at"`
}

func (row *versionRow) toEntity() (*domain.WorkoutVersion, error) {
	var st domain.WorkoutStructure
	if len(row.SnapshotStructure) > 0 {
		if err := json.Unmarshal(row.SnapshotStructure, &st); err != nil {
			return nil, err
		}
	}
	if st.Sessions == nil {
		st.Sessions = []domain.TemplateSession{}
	}
	return &domain.WorkoutVersion{
		ID:                row.ID,
		TemplateID:        row.TemplateID,
		VersionNumber:     row.VersionNumber,
		Title:             row.Title,
		Description:       row.Description,
		Goal:              row.Goal,
		SnapshotStructure: st,
		PublishedByUserID: row.PublishedByUserID,
		PublishedAt:       row.PublishedAt,
	}, nil
}

type assignmentRow struct {
	ID                  uuid.UUID  `db:"id"`
	OrganizationID      uuid.UUID  `db:"organization_id"`
	TemplateID          uuid.UUID  `db:"template_id"`
	WorkoutVersionID    uuid.UUID  `db:"workout_version_id"`
	StudentID           uuid.UUID  `db:"student_id"`
	StudentName         string     `db:"student_name"`
	TrainerUserID       uuid.UUID  `db:"trainer_user_id"`
	LegacyWorkoutPlanID *uuid.UUID `db:"legacy_workout_plan_id"`
	ValidFrom           time.Time  `db:"valid_from"`
	ValidUntil          *time.Time `db:"valid_until"`
	Status              string     `db:"status"`
	Notes               string     `db:"notes"`
	TemplateTitle       string     `db:"template_title"`
	Goal                string     `db:"goal"`
	VersionNumber       int        `db:"version_number"`
	SnapshotStructure   []byte     `db:"snapshot_structure"`
	CreatedAt           time.Time  `db:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"`
}

func (row *assignmentRow) toEntity() (*domain.WorkoutAssignment, error) {
	var st domain.WorkoutStructure
	if len(row.SnapshotStructure) > 0 {
		if err := json.Unmarshal(row.SnapshotStructure, &st); err != nil {
			return nil, err
		}
	}
	if st.Sessions == nil {
		st.Sessions = []domain.TemplateSession{}
	}
	return &domain.WorkoutAssignment{
		ID:                  row.ID,
		OrganizationID:      row.OrganizationID,
		TemplateID:          row.TemplateID,
		WorkoutVersionID:    row.WorkoutVersionID,
		StudentID:           row.StudentID,
		StudentName:         row.StudentName,
		TrainerUserID:       row.TrainerUserID,
		LegacyWorkoutPlanID: row.LegacyWorkoutPlanID,
		ValidFrom:           row.ValidFrom,
		ValidUntil:          row.ValidUntil,
		Status:              row.Status,
		Notes:               row.Notes,
		TemplateTitle:       row.TemplateTitle,
		Goal:                row.Goal,
		VersionNumber:       row.VersionNumber,
		SnapshotStructure:   st,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}, nil
}

func marshalWorkoutStructure(st domain.WorkoutStructure) (string, error) {
	if st.Sessions == nil {
		st.Sessions = []domain.TemplateSession{}
	}
	raw, err := json.Marshal(st)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (r *WorkoutBuilderRepository) CreateTemplate(ctx context.Context, tpl *domain.WorkoutTemplate) error {
	rawStruct, err := marshalWorkoutStructure(tpl.DraftStructure)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO workout_templates (
			id, organization_id, trainer_user_id, title, description, goal,
			draft_structure, cloned_from_template_id, is_archived, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10, $11)
	`
	_, err = getExecutor(ctx, r.db).ExecContext(
		ctx,
		query,
		tpl.ID,
		tpl.OrganizationID,
		tpl.TrainerUserID,
		tpl.Title,
		tpl.Description,
		tpl.Goal,
		rawStruct,
		tpl.ClonedFromTemplateID,
		tpl.IsArchived,
		tpl.CreatedAt,
		tpl.UpdatedAt,
	)
	return err
}

func (r *WorkoutBuilderRepository) GetTemplateByID(ctx context.Context, id uuid.UUID) (*domain.WorkoutTemplate, error) {
	query := `
		SELECT wt.id, wt.organization_id, wt.trainer_user_id, wt.title, wt.description, wt.goal,
		       wt.draft_structure, wt.cloned_from_template_id, wt.is_archived,
		       COALESCE((SELECT MAX(wv.version_number) FROM workout_versions wv WHERE wv.template_id = wt.id), 0) AS latest_version_number,
		       (SELECT wv.id FROM workout_versions wv WHERE wv.template_id = wt.id ORDER BY wv.version_number DESC LIMIT 1) AS latest_version_id,
		       wt.created_at, wt.updated_at
		FROM workout_templates wt
		WHERE wt.id = $1
	`
	var row templateRow
	if err := getExecutor(ctx, r.db).GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return row.toEntity()
}

func (r *WorkoutBuilderRepository) ListTemplatesByOrganization(ctx context.Context, orgID uuid.UUID, includeArchived bool) ([]*domain.WorkoutTemplate, error) {
	query := `
		SELECT wt.id, wt.organization_id, wt.trainer_user_id, wt.title, wt.description, wt.goal,
		       wt.draft_structure, wt.cloned_from_template_id, wt.is_archived,
		       COALESCE((SELECT MAX(wv.version_number) FROM workout_versions wv WHERE wv.template_id = wt.id), 0) AS latest_version_number,
		       (SELECT wv.id FROM workout_versions wv WHERE wv.template_id = wt.id ORDER BY wv.version_number DESC LIMIT 1) AS latest_version_id,
		       wt.created_at, wt.updated_at
		FROM workout_templates wt
		WHERE wt.organization_id = $1
		  AND ($2::boolean = TRUE OR wt.is_archived = FALSE)
		ORDER BY wt.updated_at DESC
	`
	var rows []templateRow
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &rows, query, orgID, includeArchived); err != nil {
		return nil, err
	}
	out := make([]*domain.WorkoutTemplate, 0, len(rows))
	for i := range rows {
		ent, err := rows[i].toEntity()
		if err != nil {
			return nil, err
		}
		out = append(out, ent)
	}
	return out, nil
}

func (r *WorkoutBuilderRepository) UpdateTemplateDraft(ctx context.Context, tpl *domain.WorkoutTemplate) error {
	rawStruct, err := marshalWorkoutStructure(tpl.DraftStructure)
	if err != nil {
		return err
	}
	query := `
		UPDATE workout_templates
		SET title = $1,
		    description = $2,
		    goal = $3,
		    draft_structure = $4::jsonb,
		    is_archived = $5,
		    updated_at = $6
		WHERE id = $7
	`
	res, err := getExecutor(ctx, r.db).ExecContext(
		ctx,
		query,
		tpl.Title,
		tpl.Description,
		tpl.Goal,
		rawStruct,
		tpl.IsArchived,
		tpl.UpdatedAt,
		tpl.ID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *WorkoutBuilderRepository) SetTemplateArchived(ctx context.Context, id uuid.UUID, archived bool) error {
	query := `
		UPDATE workout_templates
		SET is_archived = $1, updated_at = NOW()
		WHERE id = $2
	`
	res, err := getExecutor(ctx, r.db).ExecContext(ctx, query, archived, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *WorkoutBuilderRepository) NextVersionNumber(ctx context.Context, templateID uuid.UUID) (int, error) {
	var next int
	query := `SELECT COALESCE(MAX(version_number), 0) + 1 FROM workout_versions WHERE template_id = $1`
	if err := getExecutor(ctx, r.db).QueryRowContext(ctx, query, templateID).Scan(&next); err != nil {
		return 0, err
	}
	return next, nil
}

func (r *WorkoutBuilderRepository) CreateVersion(ctx context.Context, version *domain.WorkoutVersion) error {
	rawStruct, err := marshalWorkoutStructure(version.SnapshotStructure)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO workout_versions (
			id, template_id, version_number, title, description, goal,
			snapshot_structure, published_by_user_id, published_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9)
	`
	_, err = getExecutor(ctx, r.db).ExecContext(
		ctx,
		query,
		version.ID,
		version.TemplateID,
		version.VersionNumber,
		version.Title,
		version.Description,
		version.Goal,
		rawStruct,
		version.PublishedByUserID,
		version.PublishedAt,
	)
	return err
}

func (r *WorkoutBuilderRepository) GetVersionByID(ctx context.Context, id uuid.UUID) (*domain.WorkoutVersion, error) {
	query := `
		SELECT id, template_id, version_number, title, description, goal,
		       snapshot_structure, published_by_user_id, published_at
		FROM workout_versions
		WHERE id = $1
	`
	var row versionRow
	if err := getExecutor(ctx, r.db).GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return row.toEntity()
}

func (r *WorkoutBuilderRepository) ListVersionsByTemplate(ctx context.Context, templateID uuid.UUID) ([]*domain.WorkoutVersion, error) {
	query := `
		SELECT id, template_id, version_number, title, description, goal,
		       snapshot_structure, published_by_user_id, published_at
		FROM workout_versions
		WHERE template_id = $1
		ORDER BY version_number DESC
	`
	var rows []versionRow
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &rows, query, templateID); err != nil {
		return nil, err
	}
	out := make([]*domain.WorkoutVersion, 0, len(rows))
	for i := range rows {
		ent, err := rows[i].toEntity()
		if err != nil {
			return nil, err
		}
		out = append(out, ent)
	}
	return out, nil
}

func (r *WorkoutBuilderRepository) CreateAssignment(ctx context.Context, a *domain.WorkoutAssignment) error {
	query := `
		INSERT INTO workout_assignments (
			id, organization_id, template_id, workout_version_id, student_id,
			trainer_user_id, legacy_workout_plan_id, valid_from, valid_until,
			status, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := getExecutor(ctx, r.db).ExecContext(
		ctx,
		query,
		a.ID,
		a.OrganizationID,
		a.TemplateID,
		a.WorkoutVersionID,
		a.StudentID,
		a.TrainerUserID,
		a.LegacyWorkoutPlanID,
		a.ValidFrom,
		a.ValidUntil,
		a.Status,
		a.Notes,
		a.CreatedAt,
		a.UpdatedAt,
	)
	return err
}

func (r *WorkoutBuilderRepository) GetAssignmentByID(ctx context.Context, id uuid.UUID) (*domain.WorkoutAssignment, error) {
	query := `
		SELECT wa.id, wa.organization_id, wa.template_id, wa.workout_version_id,
		       wa.student_id, COALESCE(u.full_name, '') AS student_name,
		       wa.trainer_user_id, wa.legacy_workout_plan_id,
		       wa.valid_from, wa.valid_until, wa.status, wa.notes,
		       wv.title AS template_title, wv.goal, wv.version_number,
		       wv.snapshot_structure, wa.created_at, wa.updated_at
		FROM workout_assignments wa
		INNER JOIN workout_versions wv ON wv.id = wa.workout_version_id
		INNER JOIN students s ON s.id = wa.student_id
		INNER JOIN users u ON u.id = s.user_id
		WHERE wa.id = $1
	`
	var row assignmentRow
	if err := getExecutor(ctx, r.db).GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return row.toEntity()
}

func (r *WorkoutBuilderRepository) ListAssignmentsByStudent(ctx context.Context, studentID uuid.UUID) ([]*domain.WorkoutAssignment, error) {
	query := `
		SELECT wa.id, wa.organization_id, wa.template_id, wa.workout_version_id,
		       wa.student_id, COALESCE(u.full_name, '') AS student_name,
		       wa.trainer_user_id, wa.legacy_workout_plan_id,
		       wa.valid_from, wa.valid_until, wa.status, wa.notes,
		       wv.title AS template_title, wv.goal, wv.version_number,
		       wv.snapshot_structure, wa.created_at, wa.updated_at
		FROM workout_assignments wa
		INNER JOIN workout_versions wv ON wv.id = wa.workout_version_id
		INNER JOIN students s ON s.id = wa.student_id
		INNER JOIN users u ON u.id = s.user_id
		WHERE wa.student_id = $1
		ORDER BY wa.created_at DESC
	`
	var rows []assignmentRow
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &rows, query, studentID); err != nil {
		return nil, err
	}
	out := make([]*domain.WorkoutAssignment, 0, len(rows))
	for i := range rows {
		ent, err := rows[i].toEntity()
		if err != nil {
			return nil, err
		}
		out = append(out, ent)
	}
	return out, nil
}

func (r *WorkoutBuilderRepository) ListAssignmentsByOrganization(ctx context.Context, orgID uuid.UUID, studentID *uuid.UUID) ([]*domain.WorkoutAssignment, error) {
	var query string
	var args []any
	if studentID != nil && *studentID != uuid.Nil {
		query = `
			SELECT wa.id, wa.organization_id, wa.template_id, wa.workout_version_id,
			       wa.student_id, COALESCE(u.full_name, '') AS student_name,
			       wa.trainer_user_id, wa.legacy_workout_plan_id,
			       wa.valid_from, wa.valid_until, wa.status, wa.notes,
			       wv.title AS template_title, wv.goal, wv.version_number,
			       wv.snapshot_structure, wa.created_at, wa.updated_at
			FROM workout_assignments wa
			INNER JOIN workout_versions wv ON wv.id = wa.workout_version_id
			INNER JOIN students s ON s.id = wa.student_id
			INNER JOIN users u ON u.id = s.user_id
			WHERE wa.organization_id = $1 AND wa.student_id = $2
			ORDER BY wa.created_at DESC
		`
		args = []any{orgID, *studentID}
	} else {
		query = `
			SELECT wa.id, wa.organization_id, wa.template_id, wa.workout_version_id,
			       wa.student_id, COALESCE(u.full_name, '') AS student_name,
			       wa.trainer_user_id, wa.legacy_workout_plan_id,
			       wa.valid_from, wa.valid_until, wa.status, wa.notes,
			       wv.title AS template_title, wv.goal, wv.version_number,
			       wv.snapshot_structure, wa.created_at, wa.updated_at
			FROM workout_assignments wa
			INNER JOIN workout_versions wv ON wv.id = wa.workout_version_id
			INNER JOIN students s ON s.id = wa.student_id
			INNER JOIN users u ON u.id = s.user_id
			WHERE wa.organization_id = $1
			ORDER BY wa.created_at DESC
		`
		args = []any{orgID}
	}

	var rows []assignmentRow
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	out := make([]*domain.WorkoutAssignment, 0, len(rows))
	for i := range rows {
		ent, err := rows[i].toEntity()
		if err != nil {
			return nil, err
		}
		out = append(out, ent)
	}
	return out, nil
}

func (r *WorkoutBuilderRepository) UpdateAssignmentStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE workout_assignments
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	res, err := getExecutor(ctx, r.db).ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
