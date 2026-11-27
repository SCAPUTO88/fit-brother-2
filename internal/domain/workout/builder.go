package workout

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
)

const (
	AssignmentStatusActive   = "ACTIVE"
	AssignmentStatusArchived = "ARCHIVED"
)

type TemplateExercise struct {
	ID                uuid.UUID  `json:"id"`
	CatalogExerciseID *uuid.UUID `json:"catalog_exercise_id,omitempty"`
	LegacyExerciseID  *uuid.UUID `json:"legacy_exercise_id,omitempty"`
	MediaID           *uuid.UUID `json:"media_id,omitempty"`
	Name              string     `json:"name"`
	PrimaryMuscle     string     `json:"primary_muscle,omitempty"`
	Equipment         string     `json:"equipment,omitempty"`
	Sets              int        `json:"sets"`
	Reps              string     `json:"reps"`
	RestSeconds       int        `json:"rest_seconds"`
	Notes             string     `json:"notes,omitempty"`
	Position          int        `json:"position"`
}

type TemplateSession struct {
	ID        uuid.UUID          `json:"id"`
	Title     string             `json:"title"`
	DayOfWeek int                `json:"day_of_week"`
	Notes     string             `json:"notes,omitempty"`
	Position  int                `json:"position"`
	Exercises []TemplateExercise `json:"exercises"`
}

type WorkoutStructure struct {
	Sessions []TemplateSession `json:"sessions"`
}

type WorkoutTemplate struct {
	ID                   uuid.UUID
	OrganizationID       uuid.UUID
	TrainerUserID        uuid.UUID
	Title                string
	Description          string
	Goal                 string
	DraftStructure       WorkoutStructure
	ClonedFromTemplateID *uuid.UUID
	IsArchived           bool
	LatestVersionNumber  int
	LatestVersionID      *uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type WorkoutVersion struct {
	ID                uuid.UUID
	TemplateID        uuid.UUID
	VersionNumber     int
	Title             string
	Description       string
	Goal              string
	SnapshotStructure WorkoutStructure
	PublishedByUserID *uuid.UUID
	PublishedAt       time.Time
}

type WorkoutAssignment struct {
	ID                  uuid.UUID
	OrganizationID      uuid.UUID
	TemplateID          uuid.UUID
	WorkoutVersionID    uuid.UUID
	StudentID           uuid.UUID
	StudentName         string
	TrainerUserID       uuid.UUID
	LegacyWorkoutPlanID *uuid.UUID
	ValidFrom           time.Time
	ValidUntil          *time.Time
	Status              string
	Notes               string
	TemplateTitle       string
	Goal                string
	VersionNumber       int
	SnapshotStructure   WorkoutStructure
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func NewWorkoutTemplate(orgID, trainerID uuid.UUID, title, description, goal string, structure WorkoutStructure) (*WorkoutTemplate, error) {
	if orgID == uuid.Nil {
		return nil, fmt.Errorf("%w: organization_id é obrigatório", errs.ErrInvalidInput)
	}
	if trainerID == uuid.Nil {
		return nil, fmt.Errorf("%w: trainer_user_id é obrigatório", errs.ErrInvalidInput)
	}
	trimmedTitle := strings.TrimSpace(title)
	if len(trimmedTitle) < 3 {
		return nil, fmt.Errorf("%w: o título do template deve ter ao menos 3 caracteres", errs.ErrInvalidInput)
	}
	normalized, err := NormalizeAndValidateStructure(structure)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &WorkoutTemplate{
		ID:             uuid.New(),
		OrganizationID: orgID,
		TrainerUserID:  trainerID,
		Title:          trimmedTitle,
		Description:    strings.TrimSpace(description),
		Goal:           strings.TrimSpace(goal),
		DraftStructure: normalized,
		IsArchived:     false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func NormalizeAndValidateStructure(structure WorkoutStructure) (WorkoutStructure, error) {
	sessions := make([]TemplateSession, 0, len(structure.Sessions))
	for sIdx, rawSess := range structure.Sessions {
		sessTitle := strings.TrimSpace(rawSess.Title)
		if len(sessTitle) < 2 {
			return WorkoutStructure{}, fmt.Errorf("%w: o título da sessão deve ter ao menos 2 caracteres", errs.ErrInvalidInput)
		}
		if rawSess.DayOfWeek < 0 || rawSess.DayOfWeek > 6 {
			return WorkoutStructure{}, fmt.Errorf("%w: dia da semana da sessão deve ser entre 0 e 6", errs.ErrInvalidInput)
		}

		sessID := rawSess.ID
		if sessID == uuid.Nil {
			sessID = uuid.New()
		}
		sessPos := rawSess.Position
		if sessPos <= 0 {
			sessPos = sIdx + 1
		}

		exercises := make([]TemplateExercise, 0, len(rawSess.Exercises))
		for eIdx, rawEx := range rawSess.Exercises {
			exName := strings.TrimSpace(rawEx.Name)
			if len(exName) < 2 {
				return WorkoutStructure{}, fmt.Errorf("%w: o nome do exercício deve ter ao menos 2 caracteres", errs.ErrInvalidInput)
			}
			if rawEx.Sets < 0 {
				return WorkoutStructure{}, fmt.Errorf("%w: número de séries não pode ser negativo", errs.ErrInvalidInput)
			}
			if rawEx.RestSeconds < 0 {
				return WorkoutStructure{}, fmt.Errorf("%w: descanso em segundos não pode ser negativo", errs.ErrInvalidInput)
			}

			exID := rawEx.ID
			if exID == uuid.Nil {
				exID = uuid.New()
			}
			exPos := rawEx.Position
			if exPos <= 0 {
				exPos = eIdx + 1
			}

			exercises = append(exercises, TemplateExercise{
				ID:                exID,
				CatalogExerciseID: copyUUIDPtr(rawEx.CatalogExerciseID),
				LegacyExerciseID:  copyUUIDPtr(rawEx.LegacyExerciseID),
				MediaID:           copyUUIDPtr(rawEx.MediaID),
				Name:              exName,
				PrimaryMuscle:     strings.TrimSpace(rawEx.PrimaryMuscle),
				Equipment:         strings.TrimSpace(rawEx.Equipment),
				Sets:              rawEx.Sets,
				Reps:              strings.TrimSpace(rawEx.Reps),
				RestSeconds:       rawEx.RestSeconds,
				Notes:             strings.TrimSpace(rawEx.Notes),
				Position:          exPos,
			})
		}

		sessions = append(sessions, TemplateSession{
			ID:        sessID,
			Title:     sessTitle,
			DayOfWeek: rawSess.DayOfWeek,
			Notes:     strings.TrimSpace(rawSess.Notes),
			Position:  sessPos,
			Exercises: exercises,
		})
	}
	return WorkoutStructure{Sessions: sessions}, nil
}

func CloneStructure(structure WorkoutStructure) WorkoutStructure {
	clonedSessions := make([]TemplateSession, 0, len(structure.Sessions))
	for _, s := range structure.Sessions {
		clonedExercises := make([]TemplateExercise, 0, len(s.Exercises))
		for _, e := range s.Exercises {
			clonedExercises = append(clonedExercises, TemplateExercise{
				ID:                uuid.New(),
				CatalogExerciseID: copyUUIDPtr(e.CatalogExerciseID),
				MediaID:           copyUUIDPtr(e.MediaID),
				Name:              e.Name,
				PrimaryMuscle:     e.PrimaryMuscle,
				Equipment:         e.Equipment,
				Sets:              e.Sets,
				Reps:              e.Reps,
				RestSeconds:       e.RestSeconds,
				Notes:             e.Notes,
				Position:          e.Position,
			})
		}
		clonedSessions = append(clonedSessions, TemplateSession{
			ID:        uuid.New(),
			Title:     s.Title,
			DayOfWeek: s.DayOfWeek,
			Notes:     s.Notes,
			Position:  s.Position,
			Exercises: clonedExercises,
		})
	}
	return WorkoutStructure{Sessions: clonedSessions}
}

func copyUUIDPtr(in *uuid.UUID) *uuid.UUID {
	if in == nil || *in == uuid.Nil {
		return nil
	}
	v := *in
	return &v
}
