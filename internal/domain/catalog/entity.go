package catalog

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
)

const (
	SourceFreeExerciseDB = "free-exercise-db"
	SourceCustom         = "custom"

	LicenseUnlicense = "The Unlicense (Public Domain) - yuhonas/free-exercise-db"
	LicenseCustom    = "Organização / Conteúdo Próprio"

	MediaTypeVideo = "VIDEO"
	MediaTypeImage = "IMAGE"
	MediaTypeGIF   = "GIF"

	MediaStatusPending   = "PENDING"
	MediaStatusConfirmed = "CONFIRMED"

	ScopeAll    = "all"
	ScopeGlobal = "global"
	ScopeCustom = "custom"
)

type MuscleGroup struct {
	Code   string `db:"code" json:"code"`
	NamePT string `db:"name_pt" json:"name_pt"`
	NameEN string `db:"name_en" json:"name_en"`
}

type Equipment struct {
	Code   string `db:"code" json:"code"`
	NamePT string `db:"name_pt" json:"name_pt"`
	NameEN string `db:"name_en" json:"name_en"`
}

type CategoryOption struct {
	Code   string `json:"code"`
	NamePT string `json:"name_pt"`
}

type Exercise struct {
	ID               uuid.UUID  `json:"id"`
	OrganizationID   *uuid.UUID `json:"organization_id,omitempty"`
	CreatedByUserID  *uuid.UUID `json:"created_by_user_id,omitempty"`
	ExternalSource   string     `json:"external_source"`
	ExternalID       *string    `json:"external_id,omitempty"`
	NamePT           string     `json:"name_pt"`
	NameEN           string     `json:"name_en,omitempty"`
	PrimaryMuscle    string     `json:"primary_muscle"`
	SecondaryMuscles []string   `json:"secondary_muscles"`
	Equipment        string     `json:"equipment"`
	Category         string     `json:"category"`
	Level            string     `json:"level"`
	Force            string     `json:"force,omitempty"`
	Mechanic         string     `json:"mechanic,omitempty"`
	InstructionsPT   []string   `json:"instructions_pt"`
	InstructionsEN   []string   `json:"instructions_en,omitempty"`
	LicenseInfo      string     `json:"license_info"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type ExerciseMedia struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	CatalogExerciseID uuid.UUID  `db:"catalog_exercise_id" json:"catalog_exercise_id"`
	OrganizationID    *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`
	UploaderID        *uuid.UUID `db:"uploader_id" json:"uploader_id,omitempty"`
	StorageKey        string     `db:"storage_key" json:"storage_key"`
	Filename          string     `db:"filename" json:"filename"`
	MediaType         string     `db:"media_type" json:"media_type"`
	MimeType          string     `db:"mime_type" json:"mime_type"`
	SizeBytes         int64      `db:"size_bytes" json:"size_bytes"`
	Status            string     `db:"status" json:"status"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

type ExerciseFilter struct {
	Query      string
	Muscle     string
	Equipment  string
	Category   string
	Scope      string
	AllowedOrg []uuid.UUID
	Limit      int
	Offset     int
}

func AllowedCategories() []CategoryOption {
	return []CategoryOption{
		{Code: "strength", NamePT: "Força / Hipertrofia"},
		{Code: "stretching", NamePT: "Alongamento"},
		{Code: "cardio", NamePT: "Cardio"},
		{Code: "plyometrics", NamePT: "Pliometria"},
		{Code: "powerlifting", NamePT: "Powerlifting"},
		{Code: "olympic weightlifting", NamePT: "Levantamento Olímpico"},
		{Code: "strongman", NamePT: "Strongman"},
	}
}

func NormalizeCode(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func ValidateExerciseFields(namePT, primaryMuscle, equipment, category, level string) error {
	trimmedName := strings.TrimSpace(namePT)
	if len(trimmedName) < 2 || len(trimmedName) > 255 {
		return fmt.Errorf("%w: nome do exercício deve ter entre 2 e 255 caracteres", errs.ErrInvalidInput)
	}
	if strings.TrimSpace(primaryMuscle) == "" {
		return fmt.Errorf("%w: grupo muscular principal é obrigatório", errs.ErrInvalidInput)
	}
	if strings.TrimSpace(equipment) == "" {
		return fmt.Errorf("%w: equipamento é obrigatório", errs.ErrInvalidInput)
	}
	if strings.TrimSpace(category) == "" {
		return fmt.Errorf("%w: categoria é obrigatória", errs.ErrInvalidInput)
	}
	switch NormalizeCode(level) {
	case "", "beginner", "intermediate", "expert":
		return nil
	default:
		return fmt.Errorf("%w: nível inválido; use beginner, intermediate ou expert", errs.ErrInvalidInput)
	}
}
