package catalog_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/application/catalog"
	domainCatalog "github.com/SCAPUTO88/fit-brother/internal/domain/catalog"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	domainStorage "github.com/SCAPUTO88/fit-brother/internal/domain/storage"
	domainUser "github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

type memCatalogRepo struct {
	muscles    []*domainCatalog.MuscleGroup
	equipments []*domainCatalog.Equipment
	exercises  map[uuid.UUID]*domainCatalog.Exercise
	media      map[uuid.UUID]*domainCatalog.ExerciseMedia
}

func newMemCatalogRepo() *memCatalogRepo {
	return &memCatalogRepo{
		muscles: []*domainCatalog.MuscleGroup{
			{Code: "chest", NamePT: "Peitoral", NameEN: "Chest"},
			{Code: "lats", NamePT: "Dorsais", NameEN: "Lats"},
		},
		equipments: []*domainCatalog.Equipment{
			{Code: "barbell", NamePT: "Barra", NameEN: "Barbell"},
			{Code: "dumbbell", NamePT: "Halteres", NameEN: "Dumbbell"},
		},
		exercises: make(map[uuid.UUID]*domainCatalog.Exercise),
		media:     make(map[uuid.UUID]*domainCatalog.ExerciseMedia),
	}
}

func (m *memCatalogRepo) ListMuscleGroups(_ context.Context) ([]*domainCatalog.MuscleGroup, error) {
	return m.muscles, nil
}

func (m *memCatalogRepo) ListEquipments(_ context.Context) ([]*domainCatalog.Equipment, error) {
	return m.equipments, nil
}

func (m *memCatalogRepo) UpsertGlobalExercise(_ context.Context, ex *domainCatalog.Exercise) (bool, error) {
	for _, existing := range m.exercises {
		if existing.OrganizationID == nil && existing.ExternalSource == ex.ExternalSource &&
			existing.ExternalID != nil && ex.ExternalID != nil && *existing.ExternalID == *ex.ExternalID {
			ex.ID = existing.ID
			cp := *ex
			m.exercises[existing.ID] = &cp
			return false, nil
		}
	}
	if ex.ID == uuid.Nil {
		ex.ID = uuid.New()
	}
	cp := *ex
	m.exercises[ex.ID] = &cp
	return true, nil
}

func (m *memCatalogRepo) CreateCustomExercise(_ context.Context, ex *domainCatalog.Exercise) error {
	cp := *ex
	m.exercises[ex.ID] = &cp
	return nil
}

func (m *memCatalogRepo) UpdateCustomExercise(_ context.Context, ex *domainCatalog.Exercise) error {
	existing := m.exercises[ex.ID]
	if existing == nil || !existing.IsActive {
		return sql.ErrNoRows
	}
	cp := *ex
	m.exercises[ex.ID] = &cp
	return nil
}

func (m *memCatalogRepo) SoftDeleteCustomExercise(_ context.Context, id uuid.UUID) error {
	existing := m.exercises[id]
	if existing == nil || !existing.IsActive {
		return sql.ErrNoRows
	}
	existing.IsActive = false
	return nil
}

func (m *memCatalogRepo) GetExerciseByID(_ context.Context, id uuid.UUID) (*domainCatalog.Exercise, error) {
	ex := m.exercises[id]
	if ex == nil || !ex.IsActive {
		return nil, nil
	}
	cp := *ex
	return &cp, nil
}

func (m *memCatalogRepo) SearchExercises(_ context.Context, filter domainCatalog.ExerciseFilter) ([]*domainCatalog.Exercise, int, error) {
	allowedSet := make(map[uuid.UUID]bool, len(filter.AllowedOrg))
	for _, id := range filter.AllowedOrg {
		allowedSet[id] = true
	}

	var out []*domainCatalog.Exercise
	for _, ex := range m.exercises {
		if !ex.IsActive {
			continue
		}
		switch domainCatalog.NormalizeCode(filter.Scope) {
		case domainCatalog.ScopeGlobal:
			if ex.OrganizationID != nil {
				continue
			}
		case domainCatalog.ScopeCustom:
			if ex.OrganizationID == nil || !allowedSet[*ex.OrganizationID] {
				continue
			}
		default:
			if ex.OrganizationID != nil && !allowedSet[*ex.OrganizationID] {
				continue
			}
		}

		if filter.Query != "" {
			q := strings.ToLower(filter.Query)
			if !strings.Contains(strings.ToLower(ex.NamePT), q) && !strings.Contains(strings.ToLower(ex.NameEN), q) {
				continue
			}
		}
		if filter.Muscle != "" && strings.ToLower(ex.PrimaryMuscle) != strings.ToLower(filter.Muscle) {
			continue
		}
		if filter.Equipment != "" && strings.ToLower(ex.Equipment) != strings.ToLower(filter.Equipment) {
			continue
		}
		if filter.Category != "" && strings.ToLower(ex.Category) != strings.ToLower(filter.Category) {
			continue
		}
		cp := *ex
		out = append(out, &cp)
	}
	return out, len(out), nil
}

func (m *memCatalogRepo) CreateMedia(_ context.Context, media *domainCatalog.ExerciseMedia) error {
	cp := *media
	m.media[media.ID] = &cp
	return nil
}

func (m *memCatalogRepo) GetMediaByID(_ context.Context, id uuid.UUID) (*domainCatalog.ExerciseMedia, error) {
	med := m.media[id]
	if med == nil {
		return nil, nil
	}
	cp := *med
	return &cp, nil
}

func (m *memCatalogRepo) GetMediaByKeyAndExercise(_ context.Context, exerciseID uuid.UUID, storageKey string) (*domainCatalog.ExerciseMedia, error) {
	for _, med := range m.media {
		if med.CatalogExerciseID == exerciseID && med.StorageKey == storageKey {
			cp := *med
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *memCatalogRepo) ConfirmMedia(_ context.Context, mediaID uuid.UUID, sizeBytes int64, mimeType string) error {
	med := m.media[mediaID]
	if med == nil || med.Status != domainCatalog.MediaStatusPending {
		return sql.ErrNoRows
	}
	med.Status = domainCatalog.MediaStatusConfirmed
	med.SizeBytes = sizeBytes
	med.MimeType = mimeType
	return nil
}

func (m *memCatalogRepo) ListMediaByExercise(_ context.Context, exerciseID uuid.UUID, onlyConfirmed bool) ([]*domainCatalog.ExerciseMedia, error) {
	var out []*domainCatalog.ExerciseMedia
	for _, med := range m.media {
		if med.CatalogExerciseID != exerciseID {
			continue
		}
		if onlyConfirmed && med.Status != domainCatalog.MediaStatusConfirmed {
			continue
		}
		cp := *med
		out = append(out, &cp)
	}
	return out, nil
}

func (m *memCatalogRepo) DeleteMedia(_ context.Context, mediaID uuid.UUID) error {
	delete(m.media, mediaID)
	return nil
}

type memOrgAccess struct {
	orgs        map[uuid.UUID]*domainOrg.Organization
	memberships map[string]*domainOrg.Membership
	studentOrgs map[uuid.UUID][]uuid.UUID
}

func newMemOrgAccess() *memOrgAccess {
	return &memOrgAccess{
		orgs:        make(map[uuid.UUID]*domainOrg.Organization),
		memberships: make(map[string]*domainOrg.Membership),
		studentOrgs: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *memOrgAccess) addMember(orgID, userID uuid.UUID, role string) {
	m.orgs[orgID] = &domainOrg.Organization{ID: orgID, Name: "Studio", Type: domainOrg.TypeSolo}
	m.memberships[orgID.String()+":"+userID.String()] = &domainOrg.Membership{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
		Status:         domainOrg.MembershipActive,
	}
}

func (m *memOrgAccess) EnsureTrainerDefaultOrg(_ context.Context, trainerUserID uuid.UUID, trainerName string) (*domainOrg.Organization, error) {
	for _, mem := range m.memberships {
		if mem.UserID == trainerUserID && mem.Status == domainOrg.MembershipActive {
			return m.orgs[mem.OrganizationID], nil
		}
	}
	orgID := uuid.New()
	m.addMember(orgID, trainerUserID, domainOrg.RoleOwner)
	m.orgs[orgID].Name = trainerName + " Studio"
	return m.orgs[orgID], nil
}

func (m *memOrgAccess) GetMembership(_ context.Context, orgID, userID uuid.UUID) (*domainOrg.Membership, error) {
	return m.memberships[orgID.String()+":"+userID.String()], nil
}

func (m *memOrgAccess) ListActiveOrgIDsForUser(_ context.Context, userID uuid.UUID, role string) ([]uuid.UUID, error) {
	if role == domainUser.RoleStudent {
		return m.studentOrgs[userID], nil
	}
	var out []uuid.UUID
	for _, mem := range m.memberships {
		if mem.UserID == userID && mem.Status == domainOrg.MembershipActive {
			out = append(out, mem.OrganizationID)
		}
	}
	return out, nil
}

type memCatalogStorage struct {
	objects map[string]domainStorage.ObjectMetadata
	deleted []string
}

func newMemCatalogStorage() *memCatalogStorage {
	return &memCatalogStorage{objects: make(map[string]domainStorage.ObjectMetadata)}
}

func (s *memCatalogStorage) GeneratePresignedUploadURL(_ context.Context, key, _ string, _ time.Duration) (string, error) {
	return "https://storage.local/upload/" + key, nil
}

func (s *memCatalogStorage) HeadObject(_ context.Context, key string) (*domainStorage.ObjectMetadata, error) {
	meta, ok := s.objects[key]
	if !ok {
		return nil, domainStorage.ErrObjectNotFound
	}
	return &meta, nil
}

func (s *memCatalogStorage) GenerateSignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://storage.local/signed/" + key, nil
}

func (s *memCatalogStorage) DeleteObject(_ context.Context, key string) error {
	delete(s.objects, key)
	s.deleted = append(s.deleted, key)
	return nil
}

type memCatalogPrivacyRepo struct {
	events []*domainPrivacy.AuditEvent
}

func (m *memCatalogPrivacyRepo) CreateConsent(_ context.Context, _ *domainPrivacy.ConsentRecord) error {
	return nil
}

func (m *memCatalogPrivacyRepo) ListConsentsByUser(_ context.Context, _ uuid.UUID) ([]*domainPrivacy.ConsentRecord, error) {
	return nil, nil
}

func (m *memCatalogPrivacyRepo) CreatePrivacyRequest(_ context.Context, _ *domainPrivacy.PrivacyRequest) error {
	return nil
}

func (m *memCatalogPrivacyRepo) ListPrivacyRequestsByUser(_ context.Context, _ uuid.UUID) ([]*domainPrivacy.PrivacyRequest, error) {
	return nil, nil
}

func (m *memCatalogPrivacyRepo) CompletePrivacyRequestsByUserAndType(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (m *memCatalogPrivacyRepo) RecordAuditEvent(_ context.Context, evt *domainPrivacy.AuditEvent) error {
	m.events = append(m.events, evt)
	return nil
}

func (m *memCatalogPrivacyRepo) ListAuditEventsForUser(_ context.Context, _ uuid.UUID, _ int) ([]*domainPrivacy.AuditEvent, error) {
	return m.events, nil
}

func TestCatalogMultiTenantIsolationAndMediaFlow(t *testing.T) {
	ctx := context.Background()
	repo := newMemCatalogRepo()
	orgAccess := newMemOrgAccess()
	storage := newMemCatalogStorage()
	priv := &memCatalogPrivacyRepo{}

	trainerA := uuid.New()
	trainerB := uuid.New()
	linkedStudent := uuid.New()
	otherStudent := uuid.New()

	orgA := uuid.New()
	orgB := uuid.New()
	orgAccess.addMember(orgA, trainerA, domainOrg.RoleOwner)
	orgAccess.addMember(orgB, trainerB, domainOrg.RoleOwner)
	orgAccess.studentOrgs[linkedStudent] = []uuid.UUID{orgA}

	globalID := uuid.New()
	extID := "Barbell_Bench_Press_-_Medium_Grip"
	repo.exercises[globalID] = &domainCatalog.Exercise{
		ID:             globalID,
		ExternalSource: domainCatalog.SourceFreeExerciseDB,
		ExternalID:     &extID,
		NamePT:         "Supino Reto com Barra",
		NameEN:         "Barbell Bench Press - Medium Grip",
		PrimaryMuscle:  "chest",
		Equipment:      "barbell",
		Category:       "strength",
		Level:          "beginner",
		LicenseInfo:    domainCatalog.LicenseUnlicense,
		IsActive:       true,
	}

	uc := catalog.NewUseCase(repo, orgAccess, nil, priv, storage, 5*time.Minute)

	// Aluno não pode criar exercício customizado
	_, err := uc.CreateCustomExercise(ctx, linkedStudent, domainUser.RoleStudent, catalog.CreateExerciseInput{
		NamePT:        "Exercício Proibido",
		PrimaryMuscle: "chest",
		Equipment:     "barbell",
		Category:      "strength",
	})
	require.ErrorIs(t, err, errs.ErrForbidden)

	// Treinador A cria exercício customizado na Org A
	created, err := uc.CreateCustomExercise(ctx, trainerA, domainUser.RoleTrainer, catalog.CreateExerciseInput{
		OrganizationID: orgA.String(),
		NamePT:         "Supino Guiado Custom FitBrother",
		PrimaryMuscle:  "chest",
		Equipment:      "machine",
		Category:       "strength",
		Level:          "intermediate",
		InstructionsPT: []string{"Ajustar o banco", "Executar com controle"},
	})
	require.NoError(t, err)
	customID := uuid.MustParse(created.ID)

	// Treinador B não pode buscar, ler ou editar exercício da Org A
	searchB, err := uc.SearchExercises(ctx, trainerB, domainUser.RoleTrainer, domainCatalog.ExerciseFilter{})
	require.NoError(t, err)
	assert.Len(t, searchB.Items, 1)
	assert.Equal(t, globalID.String(), searchB.Items[0].ID)

	_, err = uc.GetExercise(ctx, trainerB, domainUser.RoleTrainer, customID)
	require.ErrorIs(t, err, errs.ErrForbidden)

	_, err = uc.UpdateCustomExercise(ctx, trainerB, domainUser.RoleTrainer, customID, catalog.UpdateExerciseInput{
		NamePT:        "Tentativa BOLA",
		PrimaryMuscle: "chest",
		Equipment:     "machine",
		Category:      "strength",
	})
	require.ErrorIs(t, err, errs.ErrForbidden)

	// Treinador A não pode alterar exercício global do catálogo público
	_, err = uc.UpdateCustomExercise(ctx, trainerA, domainUser.RoleTrainer, globalID, catalog.UpdateExerciseInput{
		NamePT:        "Editar Global",
		PrimaryMuscle: "chest",
		Equipment:     "barbell",
		Category:      "strength",
	})
	require.ErrorIs(t, err, errs.ErrForbidden)

	// Upload seguro de imagem customizada: bloqueia sem arquivo no storage e MIME inválido
	_, err = uc.RequestMediaUploadURL(ctx, trainerA, domainUser.RoleTrainer, customID, catalog.RequestMediaUploadInput{
		Filename: "exec.exe",
		MimeType: "application/octet-stream",
	})
	require.ErrorIs(t, err, errs.ErrInvalidInput)

	upOut, err := uc.RequestMediaUploadURL(ctx, trainerA, domainUser.RoleTrainer, customID, catalog.RequestMediaUploadInput{
		Filename:  "demonstracao.png",
		MimeType:  "image/png",
		MediaType: domainCatalog.MediaTypeImage,
	})
	require.NoError(t, err)
	assert.Contains(t, upOut.StorageKey, "catalog/"+orgA.String()+"/"+customID.String()+"/")

	// Antes de enviar ao storage, ConfirmMediaUpload falha
	_, err = uc.ConfirmMediaUpload(ctx, trainerA, domainUser.RoleTrainer, customID, catalog.ConfirmMediaUploadInput{
		StorageKey: upOut.StorageKey,
	})
	require.ErrorIs(t, err, errs.ErrInvalidInput)

	// Antes de confirmar, SignedURL é negado
	mediaID := uuid.MustParse(upOut.MediaID)
	_, err = uc.GetMediaSignedURL(ctx, trainerA, domainUser.RoleTrainer, customID, mediaID)
	require.ErrorIs(t, err, errs.ErrNotFound)

	// Simula upload acima do limite de imagem (>10MB)
	storage.objects[upOut.StorageKey] = domainStorage.ObjectMetadata{
		SizeBytes: catalog.MaxImageMediaBytes + 1,
		MimeType:  "image/png",
	}
	_, err = uc.ConfirmMediaUpload(ctx, trainerA, domainUser.RoleTrainer, customID, catalog.ConfirmMediaUploadInput{
		StorageKey: upOut.StorageKey,
	})
	require.ErrorIs(t, err, errs.ErrInvalidInput)

	// Simula upload válido e confirma
	storage.objects[upOut.StorageKey] = domainStorage.ObjectMetadata{
		SizeBytes: 204800,
		MimeType:  "image/png",
	}
	confirmed, err := uc.ConfirmMediaUpload(ctx, trainerA, domainUser.RoleTrainer, customID, catalog.ConfirmMediaUploadInput{
		StorageKey: upOut.StorageKey,
	})
	require.NoError(t, err)
	assert.Equal(t, domainCatalog.MediaStatusConfirmed, confirmed.Status)

	// Aluno vinculado acessa exercício customizado e obtém URL assinada da mídia
	detailStud, err := uc.GetExercise(ctx, linkedStudent, domainUser.RoleStudent, customID)
	require.NoError(t, err)
	assert.Len(t, detailStud.Media, 1)

	signedOut, err := uc.GetMediaSignedURL(ctx, linkedStudent, domainUser.RoleStudent, customID, mediaID)
	require.NoError(t, err)
	assert.Contains(t, signedOut.SignedURL, upOut.StorageKey)

	// Aluno de outro treinador recebe 403
	_, err = uc.GetMediaSignedURL(ctx, otherStudent, domainUser.RoleStudent, customID, mediaID)
	require.ErrorIs(t, err, errs.ErrForbidden)

	// Remoção limpa objeto do storage e desativa exercício
	require.NoError(t, uc.DeleteCustomExercise(ctx, trainerA, domainUser.RoleTrainer, customID, "127.0.0.1"))
	assert.Contains(t, storage.deleted, upOut.StorageKey)
	_, err = uc.GetExercise(ctx, trainerA, domainUser.RoleTrainer, customID)
	require.True(t, errors.Is(err, errs.ErrNotFound))
}
