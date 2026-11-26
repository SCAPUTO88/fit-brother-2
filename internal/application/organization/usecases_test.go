package organization_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/application/organization"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

type memOrgRepo struct {
	orgs    map[uuid.UUID]*domainOrg.Organization
	members []*domainOrg.Membership
	rels    []*domainOrg.TrainerStudentRelationship
	users   map[uuid.UUID]*domainU.User
}

func newMemOrgRepo(users map[uuid.UUID]*domainU.User) *memOrgRepo {
	return &memOrgRepo{
		orgs:  make(map[uuid.UUID]*domainOrg.Organization),
		users: users,
	}
}

func (m *memOrgRepo) CreateOrganizationWithOwner(_ context.Context, org *domainOrg.Organization, owner *domainOrg.Membership) error {
	for _, existing := range m.orgs {
		if existing.Slug == org.Slug {
			return errors.New("duplicate key value violates unique constraint")
		}
	}
	cpOrg := *org
	cpMem := *owner
	m.orgs[org.ID] = &cpOrg
	m.members = append(m.members, &cpMem)
	return nil
}

func (m *memOrgRepo) GetOrganizationByID(_ context.Context, id uuid.UUID) (*domainOrg.Organization, error) {
	org := m.orgs[id]
	if org == nil {
		return nil, nil
	}
	cp := *org
	return &cp, nil
}

func (m *memOrgRepo) ListOrganizationsByUser(_ context.Context, userID uuid.UUID) ([]*domainOrg.OrganizationSummary, error) {
	var out []*domainOrg.OrganizationSummary
	for _, mem := range m.members {
		if mem.UserID == userID {
			org := m.orgs[mem.OrganizationID]
			if org != nil {
				out = append(out, &domainOrg.OrganizationSummary{
					OrganizationID:   org.ID,
					OrganizationName: org.Name,
					OrganizationSlug: org.Slug,
					OrganizationType: org.Type,
					MemberRole:       mem.Role,
					MemberStatus:     mem.Status,
				})
			}
		}
	}
	return out, nil
}

func (m *memOrgRepo) GetMembership(_ context.Context, orgID, userID uuid.UUID) (*domainOrg.Membership, error) {
	for _, mem := range m.members {
		if mem.OrganizationID == orgID && mem.UserID == userID {
			cp := *mem
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memOrgRepo) AddMember(_ context.Context, mem *domainOrg.Membership) error {
	cp := *mem
	m.members = append(m.members, &cp)
	return nil
}

func (m *memOrgRepo) ListMembers(_ context.Context, orgID uuid.UUID) ([]*domainOrg.MemberSummary, error) {
	var out []*domainOrg.MemberSummary
	for _, mem := range m.members {
		if mem.OrganizationID == orgID {
			usr := m.users[mem.UserID]
			if usr != nil {
				out = append(out, &domainOrg.MemberSummary{
					ID:       mem.ID,
					UserID:   usr.ID,
					FullName: usr.FullName,
					Email:    usr.Email,
					Role:     mem.Role,
					Status:   mem.Status,
				})
			}
		}
	}
	return out, nil
}

func (m *memOrgRepo) EnsureTrainerDefaultOrg(ctx context.Context, trainerUserID uuid.UUID, trainerName string) (*domainOrg.Organization, error) {
	list, _ := m.ListOrganizationsByUser(ctx, trainerUserID)
	if len(list) > 0 {
		return m.orgs[list[0].OrganizationID], nil
	}
	org := &domainOrg.Organization{
		ID:        uuid.New(),
		Name:      trainerName + " Studio",
		Slug:      "org-" + trainerUserID.String(),
		Type:      domainOrg.TypeSolo,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	mem := &domainOrg.Membership{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		UserID:         trainerUserID,
		Role:           domainOrg.RoleOwner,
		Status:         domainOrg.MembershipActive,
		CreatedAt:      time.Now().UTC(),
	}
	_ = m.CreateOrganizationWithOwner(ctx, org, mem)
	return org, nil
}

func (m *memOrgRepo) CreateRelationship(_ context.Context, rel *domainOrg.TrainerStudentRelationship) error {
	cp := *rel
	m.rels = append(m.rels, &cp)
	return nil
}

func (m *memOrgRepo) HasActiveRelationship(_ context.Context, trainerUserID, studentID uuid.UUID) (bool, error) {
	for _, r := range m.rels {
		if r.TrainerUserID == trainerUserID && r.StudentID == studentID && r.Status == domainOrg.RelActive {
			return true, nil
		}
	}
	return false, nil
}

func (m *memOrgRepo) UpdateRelationshipStatus(_ context.Context, orgID, trainerUserID, studentID uuid.UUID, status string) error {
	for _, r := range m.rels {
		if r.OrganizationID == orgID && r.TrainerUserID == trainerUserID && r.StudentID == studentID {
			r.Status = status
			if status == domainOrg.RelEnded {
				now := time.Now().UTC()
				r.EndedAt = &now
			} else {
				r.EndedAt = nil
			}
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *memOrgRepo) EndAllByStudent(_ context.Context, studentID uuid.UUID) error {
	now := time.Now().UTC()
	for _, r := range m.rels {
		if r.StudentID == studentID && r.Status != domainOrg.RelEnded {
			r.Status = domainOrg.RelEnded
			r.EndedAt = &now
		}
	}
	return nil
}

type memUserRepo struct {
	users map[uuid.UUID]*domainU.User
}

func (m *memUserRepo) Create(_ context.Context, u *domainU.User) error {
	cp := *u
	m.users[u.ID] = &cp
	return nil
}
func (m *memUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domainU.User, error) {
	u := m.users[id]
	if u == nil {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}
func (m *memUserRepo) GetByEmail(_ context.Context, email string) (*domainU.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, nil
}
func (m *memUserRepo) Update(_ context.Context, u *domainU.User) error {
	cp := *u
	m.users[u.ID] = &cp
	return nil
}
func (m *memUserRepo) CreateRefreshToken(_ context.Context, _ *domainU.RefreshToken) error { return nil }
func (m *memUserRepo) GetRefreshTokenByHash(_ context.Context, _ string) (*domainU.RefreshToken, error) {
	return nil, nil
}
func (m *memUserRepo) RevokeRefreshToken(_ context.Context, _ uuid.UUID) error { return nil }
func (m *memUserRepo) RevokeTokenFamily(_ context.Context, _ uuid.UUID) error { return nil }
func (m *memUserRepo) RevokeAllByUserID(_ context.Context, _ uuid.UUID) error { return nil }

type memPrivacyAudit struct {
	events []*domainPrivacy.AuditEvent
}

func (m *memPrivacyAudit) CreateConsent(_ context.Context, _ *domainPrivacy.ConsentRecord) error {
	return nil
}
func (m *memPrivacyAudit) ListConsentsByUser(_ context.Context, _ uuid.UUID) ([]*domainPrivacy.ConsentRecord, error) {
	return nil, nil
}
func (m *memPrivacyAudit) CreatePrivacyRequest(_ context.Context, _ *domainPrivacy.PrivacyRequest) error {
	return nil
}
func (m *memPrivacyAudit) ListPrivacyRequestsByUser(_ context.Context, _ uuid.UUID) ([]*domainPrivacy.PrivacyRequest, error) {
	return nil, nil
}
func (m *memPrivacyAudit) CompletePrivacyRequestsByUserAndType(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (m *memPrivacyAudit) RecordAuditEvent(_ context.Context, evt *domainPrivacy.AuditEvent) error {
	m.events = append(m.events, evt)
	return nil
}
func (m *memPrivacyAudit) ListAuditEventsForUser(_ context.Context, _ uuid.UUID, _ int) ([]*domainPrivacy.AuditEvent, error) {
	return m.events, nil
}

func TestOrganizationLifecycleMembersAndRelationships(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	trainer2ID := uuid.New()
	studentUserID := uuid.New()
	studentID := uuid.New()

	users := map[uuid.UUID]*domainU.User{
		ownerID: {
			ID: ownerID, Email: "owner@fit.test", FullName: "Owner Trainer", Role: domainU.RoleTrainer, IsActive: true,
		},
		trainer2ID: {
			ID: trainer2ID, Email: "t2@fit.test", FullName: "Second Trainer", Role: domainU.RoleTrainer, IsActive: true,
		},
		studentUserID: {
			ID: studentUserID, Email: "stud@fit.test", FullName: "Aluno Teste", Role: domainU.RoleStudent, IsActive: true,
		},
	}

	orgRepo := newMemOrgRepo(users)
	userRepo := &memUserRepo{users: users}
	auditRepo := &memPrivacyAudit{}
	uc := organization.NewUseCase(orgRepo, userRepo, auditRepo, nil)

	// Aluno não pode criar organização
	_, err := uc.CreateOrganization(ctx, studentUserID, domainU.RoleStudent, organization.CreateOrganizationInput{
		Name: "Fake Studio",
	})
	require.ErrorIs(t, err, errs.ErrForbidden)

	// Treinador cria Studio
	org, err := uc.CreateOrganization(ctx, ownerID, domainU.RoleTrainer, organization.CreateOrganizationInput{
		Name: "Studio Iron",
		Type: domainOrg.TypeStudio,
	})
	require.NoError(t, err)
	assert.Equal(t, "Studio Iron", org.Name)

	// Treinador 2 não é membro ainda -> ListMembers retorna Forbidden
	_, err = uc.ListMembers(ctx, org.ID, trainer2ID)
	require.ErrorIs(t, err, errs.ErrForbidden)

	// Aluno não pode ser adicionado como membro de estúdio
	_, err = uc.AddMember(ctx, org.ID, ownerID, organization.AddMemberInput{
		Email: "stud@fit.test",
		Role:  domainOrg.RoleTrainer,
	})
	require.ErrorIs(t, err, errs.ErrInvalidInput)

	// Owner adiciona segundo treinador
	added, err := uc.AddMember(ctx, org.ID, ownerID, organization.AddMemberInput{
		Email: "t2@fit.test",
		Role:  domainOrg.RoleTrainer,
	})
	require.NoError(t, err)
	assert.Equal(t, domainOrg.RoleTrainer, added.Role)

	// Treinador 2 agora vê membros, mas não pode adicionar membros (apenas OWNER/ADMIN)
	members, err := uc.ListMembers(ctx, org.ID, trainer2ID)
	require.NoError(t, err)
	assert.Len(t, members, 2)

	_, err = uc.AddMember(ctx, org.ID, trainer2ID, organization.AddMemberInput{
		Email: "owner@fit.test",
		Role:  domainOrg.RoleAdmin,
	})
	require.ErrorIs(t, err, errs.ErrForbidden)

	// Vínculo treinador-aluno pode ser pausado e encerrado
	require.NoError(t, orgRepo.CreateRelationship(ctx, &domainOrg.TrainerStudentRelationship{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		TrainerUserID:  ownerID,
		StudentID:      studentID,
		Status:         domainOrg.RelActive,
		StartedAt:      time.Now().UTC(),
	}))

	require.NoError(t, uc.UpdateRelationshipStatus(ctx, org.ID, ownerID, domainU.RoleTrainer, studentID, organization.UpdateRelationshipStatusInput{
		Status: domainOrg.RelEnded,
	}))

	active, err := orgRepo.HasActiveRelationship(ctx, ownerID, studentID)
	require.NoError(t, err)
	assert.False(t, active)
	assert.GreaterOrEqual(t, len(auditRepo.events), 3)
}
