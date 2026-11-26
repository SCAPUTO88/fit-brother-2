package organization

import (
	"context"

	"github.com/google/uuid"
)

// Repository define o Port de persistência multi-tenant (organizações, memberships e vínculos).
type Repository interface {
	CreateOrganizationWithOwner(ctx context.Context, org *Organization, ownerMembership *Membership) error
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	ListOrganizationsByUser(ctx context.Context, userID uuid.UUID) ([]*OrganizationSummary, error)
	GetMembership(ctx context.Context, orgID, userID uuid.UUID) (*Membership, error)
	AddMember(ctx context.Context, m *Membership) error
	ListMembers(ctx context.Context, orgID uuid.UUID) ([]*MemberSummary, error)
	EnsureTrainerDefaultOrg(ctx context.Context, trainerUserID uuid.UUID, trainerName string) (*Organization, error)

	CreateRelationship(ctx context.Context, rel *TrainerStudentRelationship) error
	HasActiveRelationship(ctx context.Context, trainerUserID, studentID uuid.UUID) (bool, error)
	UpdateRelationshipStatus(ctx context.Context, orgID, trainerUserID, studentID uuid.UUID, status string) error
	EndAllByStudent(ctx context.Context, studentID uuid.UUID) error
}
