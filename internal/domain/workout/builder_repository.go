package workout

import (
	"context"

	"github.com/google/uuid"
)

type BuilderRepository interface {
	CreateTemplate(ctx context.Context, tpl *WorkoutTemplate) error
	GetTemplateByID(ctx context.Context, id uuid.UUID) (*WorkoutTemplate, error)
	ListTemplatesByOrganization(ctx context.Context, orgID uuid.UUID, includeArchived bool) ([]*WorkoutTemplate, error)
	UpdateTemplateDraft(ctx context.Context, tpl *WorkoutTemplate) error
	SetTemplateArchived(ctx context.Context, id uuid.UUID, archived bool) error

	NextVersionNumber(ctx context.Context, templateID uuid.UUID) (int, error)
	CreateVersion(ctx context.Context, version *WorkoutVersion) error
	GetVersionByID(ctx context.Context, id uuid.UUID) (*WorkoutVersion, error)
	ListVersionsByTemplate(ctx context.Context, templateID uuid.UUID) ([]*WorkoutVersion, error)

	CreateAssignment(ctx context.Context, assignment *WorkoutAssignment) error
	GetAssignmentByID(ctx context.Context, id uuid.UUID) (*WorkoutAssignment, error)
	ListAssignmentsByStudent(ctx context.Context, studentID uuid.UUID) ([]*WorkoutAssignment, error)
	ListAssignmentsByOrganization(ctx context.Context, orgID uuid.UUID, studentID *uuid.UUID) ([]*WorkoutAssignment, error)
	UpdateAssignmentStatus(ctx context.Context, id uuid.UUID, status string) error
}
