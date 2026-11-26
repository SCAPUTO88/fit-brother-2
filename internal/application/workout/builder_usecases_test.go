package workout_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	appworkout "github.com/SCAPUTO88/fit-brother/internal/application/workout"
	domainCatalog "github.com/SCAPUTO88/fit-brother/internal/domain/catalog"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	domain "github.com/SCAPUTO88/fit-brother/internal/domain/workout"
)

type inMemoryBuilderRepo struct {
	templates   map[uuid.UUID]*domain.WorkoutTemplate
	versions    map[uuid.UUID]*domain.WorkoutVersion
	assignments map[uuid.UUID]*domain.WorkoutAssignment
	studNames   map[uuid.UUID]string
}

func newInMemoryBuilderRepo() *inMemoryBuilderRepo {
	return &inMemoryBuilderRepo{
		templates:   make(map[uuid.UUID]*domain.WorkoutTemplate),
		versions:    make(map[uuid.UUID]*domain.WorkoutVersion),
		assignments: make(map[uuid.UUID]*domain.WorkoutAssignment),
		studNames:   make(map[uuid.UUID]string),
	}
}

func (r *inMemoryBuilderRepo) CreateTemplate(_ context.Context, tpl *domain.WorkoutTemplate) error {
	copyTpl := *tpl
	copyTpl.DraftStructure = domain.CloneStructure(tpl.DraftStructure)
	r.templates[tpl.ID] = &copyTpl
	return nil
}

func (r *inMemoryBuilderRepo) GetTemplateByID(_ context.Context, id uuid.UUID) (*domain.WorkoutTemplate, error) {
	tpl, ok := r.templates[id]
	if !ok {
		return nil, nil
	}
	copyTpl := *tpl
	maxVer := 0
	var latestID *uuid.UUID
	for _, v := range r.versions {
		if v.TemplateID == id && v.VersionNumber > maxVer {
			maxVer = v.VersionNumber
			vid := v.ID
			latestID = &vid
		}
	}
	copyTpl.LatestVersionNumber = maxVer
	copyTpl.LatestVersionID = latestID
	return &copyTpl, nil
}

func (r *inMemoryBuilderRepo) ListTemplatesByOrganization(_ context.Context, orgID uuid.UUID, includeArchived bool) ([]*domain.WorkoutTemplate, error) {
	out := make([]*domain.WorkoutTemplate, 0)
	for id, tpl := range r.templates {
		if tpl.OrganizationID != orgID {
			continue
		}
		if !includeArchived && tpl.IsArchived {
			continue
		}
		loaded, _ := r.GetTemplateByID(context.Background(), id)
		out = append(out, loaded)
	}
	return out, nil
}

func (r *inMemoryBuilderRepo) UpdateTemplateDraft(_ context.Context, tpl *domain.WorkoutTemplate) error {
	copyTpl := *tpl
	copyTpl.DraftStructure = domain.CloneStructure(tpl.DraftStructure)
	r.templates[tpl.ID] = &copyTpl
	return nil
}

func (r *inMemoryBuilderRepo) SetTemplateArchived(_ context.Context, id uuid.UUID, archived bool) error {
	tpl, ok := r.templates[id]
	if !ok {
		return nil
	}
	tpl.IsArchived = archived
	return nil
}

func (r *inMemoryBuilderRepo) NextVersionNumber(_ context.Context, templateID uuid.UUID) (int, error) {
	maxVer := 0
	for _, v := range r.versions {
		if v.TemplateID == templateID && v.VersionNumber > maxVer {
			maxVer = v.VersionNumber
		}
	}
	return maxVer + 1, nil
}

func (r *inMemoryBuilderRepo) CreateVersion(_ context.Context, version *domain.WorkoutVersion) error {
	copyV := *version
	copyV.SnapshotStructure = domain.CloneStructure(version.SnapshotStructure)
	r.versions[version.ID] = &copyV
	return nil
}

func (r *inMemoryBuilderRepo) GetVersionByID(_ context.Context, id uuid.UUID) (*domain.WorkoutVersion, error) {
	v, ok := r.versions[id]
	if !ok {
		return nil, nil
	}
	copyV := *v
	return &copyV, nil
}

func (r *inMemoryBuilderRepo) ListVersionsByTemplate(_ context.Context, templateID uuid.UUID) ([]*domain.WorkoutVersion, error) {
	out := make([]*domain.WorkoutVersion, 0)
	for _, v := range r.versions {
		if v.TemplateID == templateID {
			copyV := *v
			out = append(out, &copyV)
		}
	}
	return out, nil
}

func (r *inMemoryBuilderRepo) CreateAssignment(_ context.Context, a *domain.WorkoutAssignment) error {
	copyA := *a
	r.assignments[a.ID] = &copyA
	return nil
}

func (r *inMemoryBuilderRepo) GetAssignmentByID(_ context.Context, id uuid.UUID) (*domain.WorkoutAssignment, error) {
	a, ok := r.assignments[id]
	if !ok {
		return nil, nil
	}
	return r.hydrateAssignment(a), nil
}

func (r *inMemoryBuilderRepo) ListAssignmentsByStudent(_ context.Context, studentID uuid.UUID) ([]*domain.WorkoutAssignment, error) {
	out := make([]*domain.WorkoutAssignment, 0)
	for _, a := range r.assignments {
		if a.StudentID == studentID {
			out = append(out, r.hydrateAssignment(a))
		}
	}
	return out, nil
}

func (r *inMemoryBuilderRepo) ListAssignmentsByOrganization(_ context.Context, orgID uuid.UUID, studentID *uuid.UUID) ([]*domain.WorkoutAssignment, error) {
	out := make([]*domain.WorkoutAssignment, 0)
	for _, a := range r.assignments {
		if a.OrganizationID != orgID {
			continue
		}
		if studentID != nil && a.StudentID != *studentID {
			continue
		}
		out = append(out, r.hydrateAssignment(a))
	}
	return out, nil
}

func (r *inMemoryBuilderRepo) UpdateAssignmentStatus(_ context.Context, id uuid.UUID, status string) error {
	if a, ok := r.assignments[id]; ok {
		a.Status = status
		a.UpdatedAt = time.Now().UTC()
	}
	return nil
}

func (r *inMemoryBuilderRepo) hydrateAssignment(a *domain.WorkoutAssignment) *domain.WorkoutAssignment {
	copyA := *a
	copyA.StudentName = r.studNames[a.StudentID]
	if v, ok := r.versions[a.WorkoutVersionID]; ok {
		copyA.TemplateTitle = v.Title
		copyA.Goal = v.Goal
		copyA.VersionNumber = v.VersionNumber
		copyA.SnapshotStructure = v.SnapshotStructure
	}
	return &copyA
}

type mockBuilderOrgPort struct {
	orgs          map[uuid.UUID]uuid.UUID
	relationships map[string]bool
}

func (m *mockBuilderOrgPort) EnsureTrainerDefaultOrg(_ context.Context, trainerUserID uuid.UUID, trainerName string) (*domainOrg.Organization, error) {
	if orgID, ok := m.orgs[trainerUserID]; ok {
		return &domainOrg.Organization{ID: orgID, Name: trainerName}, nil
	}
	orgID := uuid.New()
	m.orgs[trainerUserID] = orgID
	return &domainOrg.Organization{ID: orgID, Name: trainerName}, nil
}

func (m *mockBuilderOrgPort) GetMembership(_ context.Context, orgID, userID uuid.UUID) (*domainOrg.Membership, error) {
	if assigned, ok := m.orgs[userID]; ok && assigned == orgID {
		return &domainOrg.Membership{OrganizationID: orgID, UserID: userID, Role: "OWNER", Status: domainOrg.MembershipActive}, nil
	}
	return nil, nil
}

func (m *mockBuilderOrgPort) ListActiveOrgIDsForUser(_ context.Context, userID uuid.UUID, _ string) ([]uuid.UUID, error) {
	if orgID, ok := m.orgs[userID]; ok {
		return []uuid.UUID{orgID}, nil
	}
	return []uuid.UUID{}, nil
}

func (m *mockBuilderOrgPort) HasActiveRelationship(_ context.Context, trainerUserID, studentID uuid.UUID) (bool, error) {
	return m.relationships[trainerUserID.String()+"|"+studentID.String()], nil
}

type mockBuilderCatalogRepo struct {
	exercises map[uuid.UUID]*domainCatalog.Exercise
	media     map[uuid.UUID][]*domainCatalog.ExerciseMedia
}

func (m *mockBuilderCatalogRepo) ListMuscleGroups(context.Context) ([]*domainCatalog.MuscleGroup, error) {
	return nil, nil
}
func (m *mockBuilderCatalogRepo) ListEquipments(context.Context) ([]*domainCatalog.Equipment, error) {
	return nil, nil
}
func (m *mockBuilderCatalogRepo) UpsertGlobalExercise(context.Context, *domainCatalog.Exercise) (bool, error) {
	return false, nil
}
func (m *mockBuilderCatalogRepo) CreateCustomExercise(context.Context, *domainCatalog.Exercise) error {
	return nil
}
func (m *mockBuilderCatalogRepo) UpdateCustomExercise(context.Context, *domainCatalog.Exercise) error {
	return nil
}
func (m *mockBuilderCatalogRepo) SoftDeleteCustomExercise(context.Context, uuid.UUID) error {
	return nil
}
func (m *mockBuilderCatalogRepo) GetExerciseByID(_ context.Context, id uuid.UUID) (*domainCatalog.Exercise, error) {
	return m.exercises[id], nil
}
func (m *mockBuilderCatalogRepo) SearchExercises(context.Context, domainCatalog.ExerciseFilter) ([]*domainCatalog.Exercise, int, error) {
	return nil, 0, nil
}
func (m *mockBuilderCatalogRepo) CreateMedia(context.Context, *domainCatalog.ExerciseMedia) error {
	return nil
}
func (m *mockBuilderCatalogRepo) GetMediaByID(_ context.Context, id uuid.UUID) (*domainCatalog.ExerciseMedia, error) {
	for _, list := range m.media {
		for _, item := range list {
			if item.ID == id {
				return item, nil
			}
		}
	}
	return nil, nil
}
func (m *mockBuilderCatalogRepo) GetMediaByKeyAndExercise(context.Context, uuid.UUID, string) (*domainCatalog.ExerciseMedia, error) {
	return nil, nil
}
func (m *mockBuilderCatalogRepo) ConfirmMedia(context.Context, uuid.UUID, int64, string) error {
	return nil
}
func (m *mockBuilderCatalogRepo) ListMediaByExercise(_ context.Context, exID uuid.UUID, onlyConfirmed bool) ([]*domainCatalog.ExerciseMedia, error) {
	list := m.media[exID]
	out := make([]*domainCatalog.ExerciseMedia, 0, len(list))
	for _, item := range list {
		if onlyConfirmed && item.Status != domainCatalog.MediaStatusConfirmed {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}
func (m *mockBuilderCatalogRepo) DeleteMedia(context.Context, uuid.UUID) error { return nil }

type mockVideoChecker struct {
	confirmed map[uuid.UUID]bool
}

func (m *mockVideoChecker) HasConfirmedVideosByExerciseIDs(_ context.Context, exerciseIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	out := make(map[uuid.UUID]bool, len(exerciseIDs))
	for _, id := range exerciseIDs {
		if m.confirmed[id] {
			out[id] = true
		}
	}
	return out, nil
}

func TestBuilderUseCaseLifecycleAndImmutability(t *testing.T) {
	ctx := context.Background()

	trainerA := uuid.New()
	trainerB := uuid.New()
	orgA := uuid.New()
	orgB := uuid.New()
	studentAUserID := uuid.New()
	studentAID := uuid.New()
	otherStudentUserID := uuid.New()
	otherStudentID := uuid.New()

	globalCatalogID := uuid.New()
	orgBCustomID := uuid.New()
	globalMediaID := uuid.New()

	builderRepo := newInMemoryBuilderRepo()
	builderRepo.studNames[studentAID] = "Aluno A"

	orgPort := &mockBuilderOrgPort{
		orgs: map[uuid.UUID]uuid.UUID{
			trainerA:       orgA,
			trainerB:       orgB,
			studentAUserID: orgA,
		},
		relationships: map[string]bool{
			trainerA.String() + "|" + studentAID.String(): true,
		},
	}

	studRepo := new(MockWorkoutStudentRepo)
	studRepo.On("GetByID", ctx, studentAID).Return(&domainS.Student{ID: studentAID, UserID: studentAUserID}, nil)
	studRepo.On("GetByUserID", ctx, studentAUserID).Return(&domainS.Student{ID: studentAID, UserID: studentAUserID}, nil)
	studRepo.On("GetByUserID", ctx, otherStudentUserID).Return(&domainS.Student{ID: otherStudentID, UserID: otherStudentUserID}, nil)

	catRepo := &mockBuilderCatalogRepo{
		exercises: map[uuid.UUID]*domainCatalog.Exercise{
			globalCatalogID: {
				ID:            globalCatalogID,
				NamePT:        "Supino Reto com Barra",
				PrimaryMuscle: "chest",
				Equipment:     "barbell",
			},
			orgBCustomID: {
				ID:             orgBCustomID,
				OrganizationID: &orgB,
				NamePT:         "Exercício Exclusivo Org B",
				PrimaryMuscle:  "chest",
				Equipment:      "cable",
			},
		},
		media: map[uuid.UUID][]*domainCatalog.ExerciseMedia{
			globalCatalogID: {
				{
					ID:                globalMediaID,
					CatalogExerciseID: globalCatalogID,
					Status:            domainCatalog.MediaStatusConfirmed,
				},
			},
		},
	}

	sut := appworkout.NewBuilderUseCase(builderRepo, orgPort, studRepo, nil, catRepo, nil, nil)

	t.Run("Deve bloquear referência a exercício customizado de outra organização", func(t *testing.T) {
		_, err := sut.CreateTemplate(ctx, trainerA, domainU.RoleTrainer, appworkout.CreateTemplateInput{
			Title: "Template Indevido",
			Sessions: []appworkout.TemplateSessionInput{
				{
					Title:     "Treino A",
					DayOfWeek: 1,
					Exercises: []appworkout.TemplateExerciseInput{
						{CatalogExerciseID: orgBCustomID.String(), Sets: 3, Reps: "10"},
					},
				},
			},
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})

	t.Run("Deve criar template com enriquecimento do catálogo, publicar v1, editar v2 sem alterar v1, clonar e atribuir", func(t *testing.T) {
		tpl, err := sut.CreateTemplate(ctx, trainerA, domainU.RoleTrainer, appworkout.CreateTemplateInput{
			Title:       "Hipertrofia Push/Pull",
			Description: "Modelo base",
			Goal:        "Hipertrofia",
			Sessions: []appworkout.TemplateSessionInput{
				{
					Title:     "Treino A - Peito",
					DayOfWeek: 1,
					Exercises: []appworkout.TemplateExerciseInput{
						{
							CatalogExerciseID: globalCatalogID.String(),
							Sets:              4,
							Reps:              "8-10",
							RestSeconds:       90,
						},
					},
				},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, orgA.String(), tpl.OrganizationID)
		require.Len(t, tpl.DraftStructure.Sessions, 1)
		ex := tpl.DraftStructure.Sessions[0].Exercises[0]
		assert.Equal(t, "Supino Reto com Barra", ex.Name)
		require.NotNil(t, ex.MediaID)
		assert.Equal(t, globalMediaID.String(), *ex.MediaID)
		assert.True(t, ex.HasMedia)

		tplID := uuid.MustParse(tpl.ID)

		// Publica v1 e atribui para Aluno A
		assignOut, err := sut.AssignTemplate(ctx, trainerA, domainU.RoleTrainer, tplID, appworkout.AssignTemplateInput{
			StudentID: studentAID.String(),
			Notes:     "Ciclo 1",
		})
		require.NoError(t, err)
		assert.Equal(t, 1, assignOut.VersionNumber)
		assert.Equal(t, "Hipertrofia Push/Pull", assignOut.TemplateTitle)
		assert.Len(t, assignOut.SnapshotStructure.Sessions, 1)
		assert.Equal(t, 4, assignOut.SnapshotStructure.Sessions[0].Exercises[0].Sets)

		// Treinador edita o rascunho do template (muda título e séries para 5) e publica v2
		updatedTpl, err := sut.UpdateTemplate(ctx, trainerA, domainU.RoleTrainer, tplID, appworkout.UpdateTemplateInput{
			Title: "Hipertrofia Push/Pull Ajustado",
			Goal:  "Força",
			Sessions: []appworkout.TemplateSessionInput{
				{
					Title:     "Treino A - Peito Pesado",
					DayOfWeek: 1,
					Exercises: []appworkout.TemplateExerciseInput{
						{
							CatalogExerciseID: globalCatalogID.String(),
							Sets:              5,
							Reps:              "5-6",
							RestSeconds:       120,
						},
					},
				},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "Hipertrofia Push/Pull Ajustado", updatedTpl.Title)

		v2, err := sut.PublishVersion(ctx, trainerA, domainU.RoleTrainer, tplID, "")
		require.NoError(t, err)
		assert.Equal(t, 2, v2.VersionNumber)
		assert.Equal(t, 5, v2.SnapshotStructure.Sessions[0].Exercises[0].Sets)

		// Atribuição antiga do aluno continua intacta na v1!
		loadedAssign, err := sut.GetAssignment(ctx, studentAUserID, domainU.RoleStudent, uuid.MustParse(assignOut.ID))
		require.NoError(t, err)
		assert.Equal(t, 1, loadedAssign.VersionNumber)
		assert.Equal(t, "Hipertrofia Push/Pull", loadedAssign.TemplateTitle)
		assert.Equal(t, 4, loadedAssign.SnapshotStructure.Sessions[0].Exercises[0].Sets)

		// Outro aluno não acessa atribuição do Aluno A
		_, err = sut.GetAssignment(ctx, otherStudentUserID, domainU.RoleStudent, uuid.MustParse(assignOut.ID))
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrForbidden))

		// Clonar gera template independente sem versões publicadas
		cloned, err := sut.CloneTemplate(ctx, trainerA, domainU.RoleTrainer, tplID, appworkout.CloneTemplateInput{
			Title: "Push/Pull Feminino",
		})
		require.NoError(t, err)
		assert.NotEqual(t, tpl.ID, cloned.ID)
		assert.Equal(t, "Push/Pull Feminino", cloned.Title)
		assert.Equal(t, 0, cloned.LatestVersionNumber)
		require.NotNil(t, cloned.ClonedFromTemplateID)
		assert.Equal(t, tpl.ID, *cloned.ClonedFromTemplateID)
	})
}

func TestGetPlanDetailsHasVideoResolution(t *testing.T) {
	ctx := context.Background()
	repo := new(MockWorkoutRepo)
	studRepo := new(MockWorkoutStudentRepo)

	planID := uuid.New()
	sessionID := uuid.New()
	exerciseWithVideo := uuid.New()
	exerciseNoVideo := uuid.New()
	userID := uuid.New()
	studentID := uuid.New()

	studRepo.On("GetByUserID", ctx, userID).Return(&domainS.Student{ID: studentID, UserID: userID}, nil)
	repo.On("GetPlanByID", ctx, planID).Return(&domain.WorkoutPlan{
		ID:        planID,
		TrainerID: uuid.New(),
		StudentID: studentID,
		Title:     "Plano Legado",
	}, nil)
	repo.On("ListSessionsByPlan", ctx, planID).Return([]*domain.WorkoutSession{
		{ID: sessionID, WorkoutPlanID: planID, Title: "Sessão A", DayOfWeek: 1},
	}, nil)
	repo.On("ListExercisesBySession", ctx, sessionID).Return([]*domain.Exercise{
		{ID: exerciseWithVideo, WorkoutSessionID: sessionID, Name: "Supino", Sets: 3, Reps: "10", RestSeconds: 60},
		{ID: exerciseNoVideo, WorkoutSessionID: sessionID, Name: "Rosca", Sets: 3, Reps: "12", RestSeconds: 45},
	}, nil)

	checker := &mockVideoChecker{confirmed: map[uuid.UUID]bool{exerciseWithVideo: true}}
	sut := appworkout.NewGetPlanDetailsUseCase(repo, studRepo).WithVideoChecker(checker)

	out, err := sut.Execute(ctx, planID.String(), userID, domainU.RoleStudent)
	require.NoError(t, err)
	require.Len(t, out.Sessions, 1)
	require.Len(t, out.Sessions[0].Exercises, 2)
	assert.True(t, out.Sessions[0].Exercises[0].HasVideo)
	assert.False(t, out.Sessions[0].Exercises[1].HasVideo)
	mock.AssertExpectationsForObjects(t, repo, studRepo)
}
