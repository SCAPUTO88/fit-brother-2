package catalog

import (
	"context"

	"github.com/google/uuid"

	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
)

type Repository interface {
	ListMuscleGroups(ctx context.Context) ([]*MuscleGroup, error)
	ListEquipments(ctx context.Context) ([]*Equipment, error)

	UpsertGlobalExercise(ctx context.Context, ex *Exercise) (bool, error)
	CreateCustomExercise(ctx context.Context, ex *Exercise) error
	UpdateCustomExercise(ctx context.Context, ex *Exercise) error
	SoftDeleteCustomExercise(ctx context.Context, id uuid.UUID) error
	GetExerciseByID(ctx context.Context, id uuid.UUID) (*Exercise, error)
	SearchExercises(ctx context.Context, filter ExerciseFilter) ([]*Exercise, int, error)

	CreateMedia(ctx context.Context, media *ExerciseMedia) error
	GetMediaByID(ctx context.Context, id uuid.UUID) (*ExerciseMedia, error)
	GetMediaByKeyAndExercise(ctx context.Context, exerciseID uuid.UUID, storageKey string) (*ExerciseMedia, error)
	ConfirmMedia(ctx context.Context, mediaID uuid.UUID, sizeBytes int64, mimeType string) error
	ListMediaByExercise(ctx context.Context, exerciseID uuid.UUID, onlyConfirmed bool) ([]*ExerciseMedia, error)
	DeleteMedia(ctx context.Context, mediaID uuid.UUID) error
}

type OrgAccessPort interface {
	EnsureTrainerDefaultOrg(ctx context.Context, trainerUserID uuid.UUID, trainerName string) (*domainOrg.Organization, error)
	GetMembership(ctx context.Context, orgID, userID uuid.UUID) (*domainOrg.Membership, error)
	ListActiveOrgIDsForUser(ctx context.Context, userID uuid.UUID, role string) ([]uuid.UUID, error)
}
