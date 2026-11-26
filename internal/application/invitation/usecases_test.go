package invitation_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/application/invitation"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainInv "github.com/SCAPUTO88/fit-brother/internal/domain/invitation"
	domainMailer "github.com/SCAPUTO88/fit-brother/internal/domain/mailer"
	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

type memInvRepo struct {
	items map[uuid.UUID]*domainInv.Invitation
}

func newMemInvRepo() *memInvRepo {
	return &memInvRepo{items: make(map[uuid.UUID]*domainInv.Invitation)}
}

func (m *memInvRepo) Create(_ context.Context, inv *domainInv.Invitation) error {
	cp := *inv
	m.items[inv.ID] = &cp
	return nil
}

func (m *memInvRepo) GetByID(_ context.Context, id uuid.UUID) (*domainInv.Invitation, error) {
	inv := m.items[id]
	if inv == nil {
		return nil, nil
	}
	cp := *inv
	return &cp, nil
}

func (m *memInvRepo) GetByTokenHash(_ context.Context, tokenHash string) (*domainInv.Invitation, error) {
	for _, inv := range m.items {
		if inv.TokenHash == tokenHash {
			cp := *inv
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memInvRepo) ListByTrainer(_ context.Context, trainerUserID uuid.UUID) ([]*domainInv.Invitation, error) {
	var out []*domainInv.Invitation
	for _, inv := range m.items {
		if inv.TrainerUserID == trainerUserID {
			cp := *inv
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *memInvRepo) RevokePendingByEmailAndOrg(_ context.Context, orgID uuid.UUID, email string) error {
	for _, inv := range m.items {
		if inv.OrganizationID == orgID && inv.Email == email && inv.Status == domainInv.StatusPending {
			inv.Status = domainInv.StatusRevoked
		}
	}
	return nil
}

func (m *memInvRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string) error {
	if inv := m.items[id]; inv != nil {
		inv.Status = status
	}
	return nil
}

func (m *memInvRepo) MarkAccepted(_ context.Context, id uuid.UUID) error {
	if inv := m.items[id]; inv != nil {
		now := time.Now().UTC()
		inv.Status = domainInv.StatusAccepted
		inv.AcceptedAt = &now
	}
	return nil
}

type memOrgRepo struct {
	orgs map[uuid.UUID]*domainOrg.Organization
	rels []*domainOrg.TrainerStudentRelationship
}

func newMemOrgRepo() *memOrgRepo {
	return &memOrgRepo{orgs: make(map[uuid.UUID]*domainOrg.Organization)}
}

func (m *memOrgRepo) CreateOrganizationWithOwner(_ context.Context, org *domainOrg.Organization, _ *domainOrg.Membership) error {
	m.orgs[org.ID] = org
	return nil
}
func (m *memOrgRepo) GetOrganizationByID(_ context.Context, id uuid.UUID) (*domainOrg.Organization, error) {
	return m.orgs[id], nil
}
func (m *memOrgRepo) ListOrganizationsByUser(_ context.Context, _ uuid.UUID) ([]*domainOrg.OrganizationSummary, error) {
	return nil, nil
}
func (m *memOrgRepo) GetMembership(_ context.Context, _, _ uuid.UUID) (*domainOrg.Membership, error) {
	return nil, nil
}
func (m *memOrgRepo) AddMember(_ context.Context, _ *domainOrg.Membership) error { return nil }
func (m *memOrgRepo) ListMembers(_ context.Context, _ uuid.UUID) ([]*domainOrg.MemberSummary, error) {
	return nil, nil
}
func (m *memOrgRepo) EnsureTrainerDefaultOrg(_ context.Context, trainerUserID uuid.UUID, trainerName string) (*domainOrg.Organization, error) {
	for _, o := range m.orgs {
		return o, nil
	}
	org := &domainOrg.Organization{
		ID:   uuid.New(),
		Name: trainerName + " Studio",
		Slug: "org-" + trainerUserID.String(),
		Type: domainOrg.TypeSolo,
	}
	m.orgs[org.ID] = org
	return org, nil
}
func (m *memOrgRepo) CreateRelationship(_ context.Context, rel *domainOrg.TrainerStudentRelationship) error {
	m.rels = append(m.rels, rel)
	return nil
}
func (m *memOrgRepo) HasActiveRelationship(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return true, nil
}
func (m *memOrgRepo) UpdateRelationshipStatus(_ context.Context, _, _, _ uuid.UUID, _ string) error {
	return nil
}
func (m *memOrgRepo) EndAllByStudent(_ context.Context, _ uuid.UUID) error { return nil }

type memUserRepo struct {
	users map[uuid.UUID]*domainU.User
}

func (m *memUserRepo) Create(_ context.Context, u *domainU.User) error {
	cp := *u
	m.users[u.ID] = &cp
	return nil
}
func (m *memUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domainU.User, error) {
	return m.users[id], nil
}
func (m *memUserRepo) GetByEmail(_ context.Context, email string) (*domainU.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}
func (m *memUserRepo) Update(_ context.Context, u *domainU.User) error {
	m.users[u.ID] = u
	return nil
}
func (m *memUserRepo) CreateRefreshToken(_ context.Context, _ *domainU.RefreshToken) error { return nil }
func (m *memUserRepo) GetRefreshTokenByHash(_ context.Context, _ string) (*domainU.RefreshToken, error) {
	return nil, nil
}
func (m *memUserRepo) RevokeRefreshToken(_ context.Context, _ uuid.UUID) error { return nil }
func (m *memUserRepo) RevokeTokenFamily(_ context.Context, _ uuid.UUID) error  { return nil }
func (m *memUserRepo) RevokeAllByUserID(_ context.Context, _ uuid.UUID) error { return nil }

type memStudRepo struct {
	studs map[uuid.UUID]*domainS.Student
}

func (m *memStudRepo) Create(_ context.Context, s *domainS.Student) error {
	m.studs[s.ID] = s
	return nil
}
func (m *memStudRepo) GetByID(_ context.Context, id uuid.UUID) (*domainS.Student, error) {
	return m.studs[id], nil
}
func (m *memStudRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*domainS.Student, error) {
	for _, s := range m.studs {
		if s.UserID == userID {
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
func (m *memStudRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.studs, id)
	return nil
}
func (m *memStudRepo) SaveBiometric(_ context.Context, _ *domainS.BiometricData) error { return nil }

type memPrivacyRepo struct {
	consents []*domainPrivacy.ConsentRecord
	events   []*domainPrivacy.AuditEvent
}

func (m *memPrivacyRepo) CreateConsent(_ context.Context, record *domainPrivacy.ConsentRecord) error {
	m.consents = append(m.consents, record)
	return nil
}
func (m *memPrivacyRepo) ListConsentsByUser(_ context.Context, _ uuid.UUID) ([]*domainPrivacy.ConsentRecord, error) {
	return m.consents, nil
}
func (m *memPrivacyRepo) CreatePrivacyRequest(_ context.Context, _ *domainPrivacy.PrivacyRequest) error {
	return nil
}
func (m *memPrivacyRepo) ListPrivacyRequestsByUser(_ context.Context, _ uuid.UUID) ([]*domainPrivacy.PrivacyRequest, error) {
	return nil, nil
}
func (m *memPrivacyRepo) CompletePrivacyRequestsByUserAndType(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (m *memPrivacyRepo) RecordAuditEvent(_ context.Context, evt *domainPrivacy.AuditEvent) error {
	m.events = append(m.events, evt)
	return nil
}
func (m *memPrivacyRepo) ListAuditEventsForUser(_ context.Context, _ uuid.UUID, _ int) ([]*domainPrivacy.AuditEvent, error) {
	return m.events, nil
}

type fakeMailer struct {
	tokens []string
}

func (f *fakeMailer) SendInvitation(_ context.Context, msg domainMailer.InvitationEmail) error {
	f.tokens = append(f.tokens, msg.RawToken)
	return nil
}
func (f *fakeMailer) SendPasswordReset(_ context.Context, _ domainMailer.PasswordResetEmail) error {
	return nil
}

func newTestJWTService(t *testing.T) *jwt.Service {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	svc, err := jwt.NewServiceFromPEM(string(privPEM), string(pubPEM), "fit-test")
	require.NoError(t, err)
	return svc
}

func TestInvitationCreateRevokeAndAcceptWithoutCPF(t *testing.T) {
	ctx := context.Background()
	trainerID := uuid.New()
	users := map[uuid.UUID]*domainU.User{
		trainerID: {ID: trainerID, Email: "trainer@fit.test", FullName: "Treinador", Role: domainU.RoleTrainer, IsActive: true},
	}

	invRepo := newMemInvRepo()
	orgRepo := newMemOrgRepo()
	userRepo := &memUserRepo{users: users}
	studRepo := &memStudRepo{studs: make(map[uuid.UUID]*domainS.Student)}
	privRepo := &memPrivacyRepo{}
	passSvc, _ := crypto.NewPasswordService("pepper")
	jwtSvc := newTestJWTService(t)
	mailer := &fakeMailer{}

	uc := invitation.NewUseCase(invRepo, orgRepo, userRepo, studRepo, privRepo, nil, passSvc, jwtSvc, mailer, 15*time.Minute, 24*time.Hour)

	// Primeiro convite criado e depois substituído ao reenviar para o mesmo e-mail
	first, err := uc.Create(ctx, domainU.RoleTrainer, invitation.CreateInvitationInput{
		Email:         "novo@fit.test",
		FullName:      "Aluno Novo",
		TrainerUserID: trainerID,
	})
	require.NoError(t, err)
	second, err := uc.Create(ctx, domainU.RoleTrainer, invitation.CreateInvitationInput{
		Email:         "novo@fit.test",
		FullName:      "Aluno Novo",
		TrainerUserID: trainerID,
	})
	require.NoError(t, err)
	require.Len(t, mailer.tokens, 2)

	oldInv, _ := invRepo.GetByID(ctx, uuid.MustParse(first.ID))
	assert.Equal(t, domainInv.StatusRevoked, oldInv.Status)

	// Token antigo revogado deve ser recusado
	_, err = uc.Accept(ctx, invitation.AcceptInvitationInput{
		Token:             mailer.tokens[0],
		Password:          "StrongPass!2026",
		ConsentTerms:      true,
		ConsentHealthData: true,
	})
	require.ErrorIs(t, err, errs.ErrUnauthorized)

	// Token atual sem aceite LGPD deve ser recusado
	_, err = uc.Accept(ctx, invitation.AcceptInvitationInput{
		Token:             mailer.tokens[1],
		Password:          "StrongPass!2026",
		ConsentTerms:      true,
		ConsentHealthData: false,
	})
	require.ErrorIs(t, err, errs.ErrInvalidInput)

	// Aceite válido cria aluno sem CPF, vínculo e 2 consentimentos
	out, err := uc.Accept(ctx, invitation.AcceptInvitationInput{
		Token:             mailer.tokens[1],
		Password:          "StrongPass!2026",
		ConsentTerms:      true,
		ConsentHealthData: true,
		ClientIP:          "127.0.0.1",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, out.Session.AccessToken)
	assert.Equal(t, domainU.RoleStudent, out.Session.Role)

	stud, _ := studRepo.GetByID(ctx, uuid.MustParse(out.StudentID))
	require.NotNil(t, stud)
	assert.Nil(t, stud.CPFEncrypted)
	assert.Nil(t, stud.CPFHash)
	assert.Len(t, privRepo.consents, 2)
	assert.Len(t, orgRepo.rels, 1)

	acceptedInv, _ := invRepo.GetByID(ctx, uuid.MustParse(second.ID))
	assert.Equal(t, domainInv.StatusAccepted, acceptedInv.Status)

	// Reuso do mesmo token deve ser recusado
	_, err = uc.Accept(ctx, invitation.AcceptInvitationInput{
		Token:             mailer.tokens[1],
		Password:          "AnotherPass!2026",
		ConsentTerms:      true,
		ConsentHealthData: true,
	})
	require.ErrorIs(t, err, errs.ErrUnauthorized)
}
