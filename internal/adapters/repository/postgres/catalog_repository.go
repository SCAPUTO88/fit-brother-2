package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/SCAPUTO88/fit-brother/internal/domain/catalog"
)

type CatalogRepository struct {
	db *sqlx.DB
}

func NewCatalogRepository(db *sqlx.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

type exerciseRow struct {
	ID               uuid.UUID  `db:"id"`
	OrganizationID   *uuid.UUID `db:"organization_id"`
	CreatedByUserID  *uuid.UUID `db:"created_by_user_id"`
	ExternalSource   string     `db:"external_source"`
	ExternalID       *string    `db:"external_id"`
	NamePT           string     `db:"name_pt"`
	NameEN           string     `db:"name_en"`
	PrimaryMuscle    string     `db:"primary_muscle"`
	SecondaryMuscles []byte     `db:"secondary_muscles"`
	Equipment        string     `db:"equipment"`
	Category         string     `db:"category"`
	Level            string     `db:"level"`
	Force            string     `db:"force"`
	Mechanic         string     `db:"mechanic"`
	InstructionsPT   []byte     `db:"instructions_pt"`
	InstructionsEN   []byte     `db:"instructions_en"`
	LicenseInfo      string     `db:"license_info"`
	IsActive         bool       `db:"is_active"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}

func (row *exerciseRow) toEntity() (*catalog.Exercise, error) {
	sec := make([]string, 0)
	if len(row.SecondaryMuscles) > 0 {
		if err := json.Unmarshal(row.SecondaryMuscles, &sec); err != nil {
			return nil, err
		}
	}
	instPT := make([]string, 0)
	if len(row.InstructionsPT) > 0 {
		if err := json.Unmarshal(row.InstructionsPT, &instPT); err != nil {
			return nil, err
		}
	}
	instEN := make([]string, 0)
	if len(row.InstructionsEN) > 0 {
		if err := json.Unmarshal(row.InstructionsEN, &instEN); err != nil {
			return nil, err
		}
	}
	return &catalog.Exercise{
		ID:               row.ID,
		OrganizationID:   row.OrganizationID,
		CreatedByUserID:  row.CreatedByUserID,
		ExternalSource:   row.ExternalSource,
		ExternalID:       row.ExternalID,
		NamePT:           row.NamePT,
		NameEN:           row.NameEN,
		PrimaryMuscle:    row.PrimaryMuscle,
		SecondaryMuscles: sec,
		Equipment:        row.Equipment,
		Category:         row.Category,
		Level:            row.Level,
		Force:            row.Force,
		Mechanic:         row.Mechanic,
		InstructionsPT:   instPT,
		InstructionsEN:   instEN,
		LicenseInfo:      row.LicenseInfo,
		IsActive:         row.IsActive,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}, nil
}

func marshalStringSlice(items []string) ([]byte, error) {
	if items == nil {
		items = []string{}
	}
	return json.Marshal(items)
}

func (r *CatalogRepository) ListMuscleGroups(ctx context.Context) ([]*catalog.MuscleGroup, error) {
	query := `SELECT code, name_pt, name_en FROM muscle_groups ORDER BY name_pt ASC`
	var list []*catalog.MuscleGroup
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &list, query); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *CatalogRepository) ListEquipments(ctx context.Context) ([]*catalog.Equipment, error) {
	query := `SELECT code, name_pt, name_en FROM equipments ORDER BY name_pt ASC`
	var list []*catalog.Equipment
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &list, query); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *CatalogRepository) UpsertGlobalExercise(ctx context.Context, ex *catalog.Exercise) (bool, error) {
	secJSON, err := marshalStringSlice(ex.SecondaryMuscles)
	if err != nil {
		return false, err
	}
	instPTJSON, err := marshalStringSlice(ex.InstructionsPT)
	if err != nil {
		return false, err
	}
	instENJSON, err := marshalStringSlice(ex.InstructionsEN)
	if err != nil {
		return false, err
	}

	if ex.ID == uuid.Nil {
		ex.ID = uuid.New()
	}

	query := `
		INSERT INTO exercise_catalog (
			id, organization_id, created_by_user_id, external_source, external_id,
			name_pt, name_en, primary_muscle, secondary_muscles, equipment, category,
			level, force, mechanic, instructions_pt, instructions_en, license_info, is_active
		) VALUES (
			$1, NULL, NULL, $2, $3,
			$4, $5, $6, $7::jsonb, $8, $9,
			$10, $11, $12, $13::jsonb, $14::jsonb, $15, TRUE
		)
		ON CONFLICT (external_source, external_id)
		WHERE organization_id IS NULL AND external_id IS NOT NULL
		DO UPDATE SET
			name_pt = EXCLUDED.name_pt,
			name_en = EXCLUDED.name_en,
			primary_muscle = EXCLUDED.primary_muscle,
			secondary_muscles = EXCLUDED.secondary_muscles,
			equipment = EXCLUDED.equipment,
			category = EXCLUDED.category,
			level = EXCLUDED.level,
			force = EXCLUDED.force,
			mechanic = EXCLUDED.mechanic,
			instructions_pt = EXCLUDED.instructions_pt,
			instructions_en = EXCLUDED.instructions_en,
			license_info = EXCLUDED.license_info,
			is_active = TRUE,
			updated_at = NOW()
		RETURNING id, (xmax = 0) AS inserted
	`

	var returnedID uuid.UUID
	var inserted bool
	err = getExecutor(ctx, r.db).QueryRowContext(
		ctx,
		query,
		ex.ID,
		ex.ExternalSource,
		ex.ExternalID,
		ex.NamePT,
		ex.NameEN,
		ex.PrimaryMuscle,
		string(secJSON),
		ex.Equipment,
		ex.Category,
		ex.Level,
		ex.Force,
		ex.Mechanic,
		string(instPTJSON),
		string(instENJSON),
		ex.LicenseInfo,
	).Scan(&returnedID, &inserted)
	if err != nil {
		return false, err
	}
	ex.ID = returnedID
	return inserted, nil
}

func (r *CatalogRepository) CreateCustomExercise(ctx context.Context, ex *catalog.Exercise) error {
	secJSON, err := marshalStringSlice(ex.SecondaryMuscles)
	if err != nil {
		return err
	}
	instPTJSON, err := marshalStringSlice(ex.InstructionsPT)
	if err != nil {
		return err
	}
	instENJSON, err := marshalStringSlice(ex.InstructionsEN)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO exercise_catalog (
			id, organization_id, created_by_user_id, external_source, external_id,
			name_pt, name_en, primary_muscle, secondary_muscles, equipment, category,
			level, force, mechanic, instructions_pt, instructions_en, license_info, is_active,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9::jsonb, $10, $11,
			$12, $13, $14, $15::jsonb, $16::jsonb, $17, $18,
			$19, $20
		)
	`
	_, err = getExecutor(ctx, r.db).ExecContext(
		ctx,
		query,
		ex.ID,
		ex.OrganizationID,
		ex.CreatedByUserID,
		ex.ExternalSource,
		ex.ExternalID,
		ex.NamePT,
		ex.NameEN,
		ex.PrimaryMuscle,
		string(secJSON),
		ex.Equipment,
		ex.Category,
		ex.Level,
		ex.Force,
		ex.Mechanic,
		string(instPTJSON),
		string(instENJSON),
		ex.LicenseInfo,
		ex.IsActive,
		ex.CreatedAt,
		ex.UpdatedAt,
	)
	return err
}

func (r *CatalogRepository) UpdateCustomExercise(ctx context.Context, ex *catalog.Exercise) error {
	secJSON, err := marshalStringSlice(ex.SecondaryMuscles)
	if err != nil {
		return err
	}
	instPTJSON, err := marshalStringSlice(ex.InstructionsPT)
	if err != nil {
		return err
	}

	query := `
		UPDATE exercise_catalog
		SET name_pt = $1,
		    primary_muscle = $2,
		    secondary_muscles = $3::jsonb,
		    equipment = $4,
		    category = $5,
		    level = $6,
		    force = $7,
		    mechanic = $8,
		    instructions_pt = $9::jsonb,
		    updated_at = NOW()
		WHERE id = $10 AND organization_id IS NOT NULL AND is_active = TRUE
	`
	res, err := getExecutor(ctx, r.db).ExecContext(
		ctx,
		query,
		ex.NamePT,
		ex.PrimaryMuscle,
		string(secJSON),
		ex.Equipment,
		ex.Category,
		ex.Level,
		ex.Force,
		ex.Mechanic,
		string(instPTJSON),
		ex.ID,
	)
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

func (r *CatalogRepository) SoftDeleteCustomExercise(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE exercise_catalog
		SET is_active = FALSE, updated_at = NOW()
		WHERE id = $1 AND organization_id IS NOT NULL AND is_active = TRUE
	`
	res, err := getExecutor(ctx, r.db).ExecContext(ctx, query, id)
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

func (r *CatalogRepository) GetExerciseByID(ctx context.Context, id uuid.UUID) (*catalog.Exercise, error) {
	query := `
		SELECT id, organization_id, created_by_user_id, external_source, external_id,
		       name_pt, name_en, primary_muscle, secondary_muscles, equipment, category,
		       level, force, mechanic, instructions_pt, instructions_en, license_info,
		       is_active, created_at, updated_at
		FROM exercise_catalog
		WHERE id = $1 AND is_active = TRUE
	`
	var row exerciseRow
	err := getExecutor(ctx, r.db).GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return row.toEntity()
}

func (r *CatalogRepository) SearchExercises(ctx context.Context, filter catalog.ExerciseFilter) ([]*catalog.Exercise, int, error) {
	clauses := []string{"is_active = TRUE"}
	args := make([]any, 0, 12)

	addArg := func(v any) string {
		args = append(args, v)
		return "$" + strconv.Itoa(len(args))
	}

	switch catalog.NormalizeCode(filter.Scope) {
	case catalog.ScopeGlobal:
		clauses = append(clauses, "organization_id IS NULL")
	case catalog.ScopeCustom:
		if len(filter.AllowedOrg) == 0 {
			return []*catalog.Exercise{}, 0, nil
		}
		placeholders := make([]string, 0, len(filter.AllowedOrg))
		for _, orgID := range filter.AllowedOrg {
			placeholders = append(placeholders, addArg(orgID))
		}
		clauses = append(clauses, "organization_id IN ("+strings.Join(placeholders, ", ")+")")
	default:
		if len(filter.AllowedOrg) == 0 {
			clauses = append(clauses, "organization_id IS NULL")
		} else {
			placeholders := make([]string, 0, len(filter.AllowedOrg))
			for _, orgID := range filter.AllowedOrg {
				placeholders = append(placeholders, addArg(orgID))
			}
			clauses = append(clauses, "(organization_id IS NULL OR organization_id IN ("+strings.Join(placeholders, ", ")+"))")
		}
	}

	if q := strings.TrimSpace(filter.Query); q != "" {
		likePattern := "%" + q + "%"
		p1 := addArg(likePattern)
		p2 := addArg(likePattern)
		clauses = append(clauses, "(name_pt ILIKE "+p1+" OR name_en ILIKE "+p2+")")
	}

	if m := catalog.NormalizeCode(filter.Muscle); m != "" {
		p1 := addArg(m)
		secBytes, _ := json.Marshal([]string{m})
		p2 := addArg(string(secBytes))
		clauses = append(clauses, "(LOWER(primary_muscle) = "+p1+" OR secondary_muscles @> "+p2+"::jsonb)")
	}

	if eq := catalog.NormalizeCode(filter.Equipment); eq != "" {
		p1 := addArg(eq)
		clauses = append(clauses, "LOWER(equipment) = "+p1)
	}

	if cat := catalog.NormalizeCode(filter.Category); cat != "" {
		p1 := addArg(cat)
		clauses = append(clauses, "LOWER(category) = "+p1)
	}

	whereSQL := strings.Join(clauses, " AND ")

	countSQL := "SELECT COUNT(*) FROM exercise_catalog WHERE " + whereSQL
	var total int
	if err := getExecutor(ctx, r.db).GetContext(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	pLimit := addArg(limit)
	pOffset := addArg(offset)
	selectSQL := "SELECT id, organization_id, created_by_user_id, external_source, external_id, " +
		"name_pt, name_en, primary_muscle, secondary_muscles, equipment, category, " +
		"level, force, mechanic, instructions_pt, instructions_en, license_info, " +
		"is_active, created_at, updated_at " +
		"FROM exercise_catalog WHERE " + whereSQL +
		" ORDER BY (organization_id IS NULL) ASC, name_pt ASC LIMIT " + pLimit + " OFFSET " + pOffset

	var rows []exerciseRow
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &rows, selectSQL, args...); err != nil {
		return nil, 0, err
	}

	out := make([]*catalog.Exercise, 0, len(rows))
	for i := range rows {
		ent, err := rows[i].toEntity()
		if err != nil {
			return nil, 0, err
		}
		out = append(out, ent)
	}
	return out, total, nil
}

func (r *CatalogRepository) CreateMedia(ctx context.Context, media *catalog.ExerciseMedia) error {
	query := `
		INSERT INTO exercise_media (
			id, catalog_exercise_id, organization_id, uploader_id,
			storage_key, filename, media_type, mime_type, size_bytes, status,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := getExecutor(ctx, r.db).ExecContext(
		ctx,
		query,
		media.ID,
		media.CatalogExerciseID,
		media.OrganizationID,
		media.UploaderID,
		media.StorageKey,
		media.Filename,
		media.MediaType,
		media.MimeType,
		media.SizeBytes,
		media.Status,
		media.CreatedAt,
		media.UpdatedAt,
	)
	return err
}

func (r *CatalogRepository) GetMediaByID(ctx context.Context, id uuid.UUID) (*catalog.ExerciseMedia, error) {
	query := `
		SELECT id, catalog_exercise_id, organization_id, uploader_id,
		       storage_key, filename, media_type, mime_type, size_bytes, status,
		       created_at, updated_at
		FROM exercise_media
		WHERE id = $1
	`
	var m catalog.ExerciseMedia
	err := getExecutor(ctx, r.db).GetContext(ctx, &m, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *CatalogRepository) GetMediaByKeyAndExercise(ctx context.Context, exerciseID uuid.UUID, storageKey string) (*catalog.ExerciseMedia, error) {
	query := `
		SELECT id, catalog_exercise_id, organization_id, uploader_id,
		       storage_key, filename, media_type, mime_type, size_bytes, status,
		       created_at, updated_at
		FROM exercise_media
		WHERE catalog_exercise_id = $1 AND storage_key = $2
	`
	var m catalog.ExerciseMedia
	err := getExecutor(ctx, r.db).GetContext(ctx, &m, query, exerciseID, storageKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *CatalogRepository) ConfirmMedia(ctx context.Context, mediaID uuid.UUID, sizeBytes int64, mimeType string) error {
	query := `
		UPDATE exercise_media
		SET status = 'CONFIRMED', size_bytes = $1, mime_type = $2, updated_at = NOW()
		WHERE id = $3 AND status = 'PENDING'
	`
	res, err := getExecutor(ctx, r.db).ExecContext(ctx, query, sizeBytes, mimeType, mediaID)
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

func (r *CatalogRepository) ListMediaByExercise(ctx context.Context, exerciseID uuid.UUID, onlyConfirmed bool) ([]*catalog.ExerciseMedia, error) {
	query := `
		SELECT id, catalog_exercise_id, organization_id, uploader_id,
		       storage_key, filename, media_type, mime_type, size_bytes, status,
		       created_at, updated_at
		FROM exercise_media
		WHERE catalog_exercise_id = $1
	`
	if onlyConfirmed {
		query += " AND status = 'CONFIRMED'"
	}
	query += " ORDER BY created_at DESC"

	var list []*catalog.ExerciseMedia
	if err := getExecutor(ctx, r.db).SelectContext(ctx, &list, query, exerciseID); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *CatalogRepository) DeleteMedia(ctx context.Context, mediaID uuid.UUID) error {
	query := `DELETE FROM exercise_media WHERE id = $1`
	_, err := getExecutor(ctx, r.db).ExecContext(ctx, query, mediaID)
	return err
}
