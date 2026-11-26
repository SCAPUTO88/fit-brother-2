package privacy_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/application/privacy"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	domainW "github.com/SCAPUTO88/fit-brother/internal/domain/workout"
)

type memPrivRepo struct {
	consents []*domainPrivacy.ConsentRecord
	reqs     []*domainPrivacy.PrivacyRequest
	events   []*domainPrivacy.AuditEvent
}

func (m *memPrivRepo) CreateConsent(_ context.Context, r *domainPrivacy.ConsentRecord) error {
	m.consents = append(m.consents, r)
	return nil
}
func (m *memPrivRepo) ListConsentsByUser(_ context.Context, userID uuid.UUID) ([]*domainPrivacy.ConsentRecord, error) {
	var out []*domainPrivacy.ConsentRecord
	for _, c := range m.consents {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}
func (m *memPrivRepo) CreatePrivacyRequest(_ context.Context, r *domainPrivacy.PrivacyRequest) error {
	m.reqs = append(m.reqs, r)
	return nil
}
func (m *memPrivRepo) ListPrivacyRequestsByUser(_ context.Context, userID uuid.UUID) ([]*domainPrivacy.PrivacyRequest, error) {
	var out []*domainPrivacy.PrivacyRequest
	for _, r := range m.reqs {
		if r.UserID == userID {
			out = append(out, r)
		}
	}
	return out, nil
}
func (m *memPrivRepo) CompletePrivacyRequestsByUserAndType(_ context.Context, userID uuid.UUID, reqType string) error {
	now := time.Now().UTC()
	for _, r := range m.reqs {
		if r.UserID == userID && r.RequestType == reqType && r.Status == domainPrivacy.StatusOpen {
			r.Status = domainPrivacy.StatusCompleted
			r.CompletedAt = &now
		}
	}
	return nil
}
func (m *memPrivRepo) RecordAuditEvent(_ context.Context, e *domainPrivacy.AuditEvent) error {
	m.events = append(m.events, e)
	return nil
}
func (m *memPrivRepo) ListAuditEventsForUser(_ context.Context, userID uuid.UUID, _ int) ([]*domainPrivacy.AuditEvent, error) {
	var out []*domainPrivacy.AuditEvent
	for _, e := range m.events {
		if e.ActorUserID != nil && *e.ActorUserID == userID {
			out = append(out, e)
		}
	}
	return out, nil
}

type memUserRepo struct {
	users   map[uuid.UUID]*domainU.User
	revoked map[uuid.UUID]bool
}

func (m *memUserRepo) Create(_ context.Context, u *domainU.User) error { m.users[u.ID] = u; return nil }
func (m *memUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domainU.User, error) {
	return m.users[id], nil
}
func (m *memUserRepo) GetByEmail(_ context.Context, _ string) (*domainU.User, error) { return nil, nil }
func (m *memUserRepo) Update(_ context.Context, u *domainU.User) error { m.users[u.ID] = u; return nil }
func (m *memUserRepo) CreateRefreshToken(_ context.Context, _ *domainU.RefreshToken) error {
	return nil
}
func (m *memUserRepo) GetRefreshTokenByHash(_ context.Context, _ string) (*domainU.RefreshToken, error) {
	return nil, nil
}
func (m *memUserRepo) RevokeRefreshToken(_ context.Context, _ uuid.UUID) error { return nil }
func (m *memUserRepo) RevokeTokenFamily(_ context.Context, _ uuid.UUID) error  { return nil }
func (m *memUserRepo) RevokeAllByUserID(_ context.Context, id uuid.UUID) error {
	m.revoked[id] = true
	return nil
}

type memStudRepo struct {
	studs map[uuid.UUID]*domainS.Student
}

func (m *memStudRepo) Create(_ context.Context, s *domainS.Student) error { m.studs[s.ID] = s; return nil }
func (m *memStudRepo) GetByID(_ context.Context, id uuid.UUID) (*domainS.Student, error) {
	return m.studs[id], nil
}
func (m *memStudRepo) GetByUserID(_ context.Context, uid uuid.UUID) (*domainS.Student, error) {
	for _, s := range m.studs {
		if s.UserID == uid {
			return s, nil
		}
	}
	return nil, nil
}
func (m *memStudRepo) GetByCPFHash(_ context.Context, _ string) (*domainS.Student, error) {
	return nil, nil
}
func (m *memStudRepo) ListByTrainer(_ context.Context, _ uuid.UUID) ([]*domainS.StudentBaseInfo, error) {
	return nil, nil
}
func (m *memStudRepo) Delete(_ context.Context, id uuid.UUID) error { delete(m.studs, id); return nil }
func (m *memStudRepo) SaveBiometric(_ context.Context, _ *domainS.BiometricData) error { return nil }

type memOrgRepo struct {
	ended bool
}

func (m *memOrgRepo) CreateOrganizationWithOwner(_ context.Context, _ *domainOrg.Organization, _ *domainOrg.Membership) error {
	return nil
}
func (m *memOrgRepo) GetOrganizationByID(_ context.Context, _ uuid.UUID) (*domainOrg.Organization, error) {
	return nil, nil
}
func (m *memOrgRepo) ListOrganizationsByUser(_ context.Context, _ uuid.UUID) ([]*domainOrg.OrganizationSummary, error) {
	return []*domainOrg.OrganizationSummary{}, nil
}
func (m *memOrgRepo) GetMembership(_ context.Context, _, _ uuid.UUID) (*domainOrg.Membership, error) {
	return nil, nil
}
func (m *memOrgRepo) AddMember(_ context.Context, _ *domainOrg.Membership) error { return nil }
func (m *memOrgRepo) ListMembers(_ context.Context, _ uuid.UUID) ([]*domainOrg.MemberSummary, error) {
	return nil, nil
}
func (m *memOrgRepo) EnsureTrainerDefaultOrg(_ context.Context, _ uuid.UUID, _ string) (*domainOrg.Organization, error) {
	return nil, nil
}
func (m *memOrgRepo) CreateRelationship(_ context.Context, _ *domainOrg.TrainerStudentRelationship) error {
	return nil
}
func (m *memOrgRepo) HasActiveRelationship(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return true, nil
}
func (m *memOrgRepo) UpdateRelationshipStatus(_ context.Context, _, _, _ uuid.UUID, _ string) error {
	return nil
}
func (m *memOrgRepo) EndAllByStudent(_ context.Context, _ uuid.UUID) error { m.ended = true; return nil }

type memWorkoutRepo struct {
	plans []*domainW.WorkoutPlan
}

func (m *memWorkoutRepo) CreatePlan(_ context.Context, _ *domainW.WorkoutPlan) error { return nil }
func (m *memWorkoutRepo) GetPlanByID(_ context.Context, _ uuid.UUID) (*domainW.WorkoutPlan, error) {
	return nil, nil
}
func (m *memWorkoutRepo) ListPlansByStudent(_ context.Context, _ uuid.UUID) ([]*domainW.WorkoutPlan, error) {
	return m.plans, nil
}
func (m *memWorkoutRepo) ListPlansByTrainer(_ context.Context, _ uuid.UUID) ([]*domainW.WorkoutPlan, error) {
	return nil, nil
}
func (m *memWorkoutRepo) UpdatePlan(_ context.Context, _ *domainW.WorkoutPlan) error { return nil }
func (m *memWorkoutRepo) CreateSession(_ context.Context, _ *domainW.WorkoutSession) error {
	return nil
}
func (m *memWorkoutRepo) GetSessionByID(_ context.Context, _ uuid.UUID) (*domainW.WorkoutSession, error) {
	return nil, nil
}
func (m *memWorkoutRepo) ListSessionsByPlan(_ context.Context, _ uuid.UUID) ([]*domainW.WorkoutSession, error) {
	return nil, nil
}
func (m *memWorkoutRepo) CreateExercise(_ context.Context, _ *domainW.Exercise) error { return nil }
func (m *memWorkoutRepo) GetExerciseByID(_ context.Context, _ uuid.UUID) (*domainW.Exercise, error) {
	return nil, nil
}
func (m *memWorkoutRepo) ListExercisesBySession(_ context.Context, _ uuid.UUID) ([]*domainW.Exercise, error) {
	return nil, nil
}
func (m *memWorkoutRepo) DeleteExercise(_ context.Context, _ uuid.UUID) error { return nil }

func TestPrivacyExportAndSelfAnonymize(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	trainerID := uuid.New()
	studID := uuid.New()
	encCPF := "secret-cipher"
	hashCPF := "secret-hash"

	users := map[uuid.UUID]*domainU.User{
		userID: {
			ID:           userID,
			Email:        "aluno.lgpd@fit.test",
			FullName:     "Aluno LGPD",
			PasswordHash: "secret-argon2-hash",
			Role:         domainU.RoleStudent,
			IsActive:     true,
			CreatedAt:    time.Now().UTC(),
		},
	}
	studs := map[uuid.UUID]*domainS.Student{
		studID: {
			ID:           studID,
			UserID:       userID,
			TrainerID:    &trainerID,
			CPFEncrypted: &encCPF,
			CPFHash:      &hashCPF,
			CreatedAt:    time.Now().UTC(),
		},
	}

	privRepo := &memPrivRepo{}
	userRepo := &memUserRepo{users: users, revoked: make(map[uuid.UUID]bool)}
	studRepo := &memStudRepo{studs: studs}
	orgRepo := &memOrgRepo{}
	workoutRepo := &memWorkoutRepo{
		plans: []*domainW.WorkoutPlan{
			{ID: uuid.New(), Title: "Treino Base", IsActive: true, CreatedAt: time.Now().UTC()},
		},
	}

	uc := privacy.NewUseCase(privRepo, userRepo, nil, studRepo, orgRepo, workoutRepo, nil)

	_, err := uc.RecordConsent(ctx, userID, privacy.RecordConsentInput{
		Purpose:       domainPrivacy.PurposeHealthData,
		PolicyVersion: "2026-11",
		Granted:       true,
		ClientIP:      "127.0.0.1",
	})
	require.NoError(t, err)

	req, err := uc.CreateRequest(ctx, userID, privacy.CreatePrivacyRequestInput{
		RequestType: domainPrivacy.RequestAnonymize,
		Details:     "Solicitação de exclusão LGPD",
	})
	require.NoError(t, err)
	assert.Equal(t, domainPrivacy.StatusOpen, req.Status)

	exported, err := uc.ExportMyData(ctx, userID, "127.0.0.1")
	require.NoError(t, err)
	rawJSON, err := json.Marshal(exported)
	require.NoError(t, err)
	assert.NotContains(t, string(rawJSON), "secret-argon2-hash")
	assert.NotContains(t, string(rawJSON), "secret-cipher")
	assert.NotContains(t, string(rawJSON), "secret-hash")
	assert.Len(t, exported.Workouts, 1)

	// Aluno não pode listar audit_events de estúdio
	_, err = uc.ListAuditEvents(ctx, userID, domainU.RoleStudent)
	require.ErrorIs(t, err, errs.ErrForbidden)

	// Auto-anonimização LGPD encerra vínculos, revoga tokens e conclui solicitações
	require.NoError(t, uc.AnonymizeSelf(ctx, userID, "127.0.0.1"))
	assert.True(t, orgRepo.ended)
	assert.True(t, userRepo.revoked[userID])
	assert.False(t, users[userID].IsActive)
	assert.Equal(t, "Usuário Anonimizado", users[userID].FullName)
	assert.Empty(t, studs)
	assert.Equal(t, domainPrivacy.StatusCompleted, req.Status)
}
