package video_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/storage"
	appvideo "github.com/SCAPUTO88/fit-brother/internal/application/video"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	domainW "github.com/SCAPUTO88/fit-brother/internal/domain/workout"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
)

type fakeVideoRepo struct {
	records map[uuid.UUID]*appvideo.VideoRecord
}

func newFakeVideoRepo() *fakeVideoRepo {
	return &fakeVideoRepo{records: make(map[uuid.UUID]*appvideo.VideoRecord)}
}

func (r *fakeVideoRepo) Save(_ context.Context, v *appvideo.VideoRecord) error {
	cp := *v
	r.records[v.ID] = &cp
	return nil
}

func (r *fakeVideoRepo) GetByExerciseID(_ context.Context, exerciseID uuid.UUID) (*appvideo.VideoRecord, error) {
	for _, rec := range r.records {
		if rec.ExerciseID == exerciseID && rec.Status == appvideo.StatusConfirmed {
			cp := *rec
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *fakeVideoRepo) GetByKeyAndExercise(_ context.Context, exerciseID uuid.UUID, r2Key string) (*appvideo.VideoRecord, error) {
	for _, rec := range r.records {
		if rec.ExerciseID == exerciseID && rec.R2Key == r2Key {
			cp := *rec
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *fakeVideoRepo) ConfirmPending(_ context.Context, id uuid.UUID, sizeBytes int64, mimeType string) error {
	rec, ok := r.records[id]
	if !ok || rec.Status != appvideo.StatusPending {
		return errors.New("not pending")
	}
	rec.Status = appvideo.StatusConfirmed
	rec.SizeBytes = sizeBytes
	rec.MimeType = mimeType
	return nil
}

func (r *fakeVideoRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.records, id)
	return nil
}

type fakeStorage struct {
	objects map[string]storage.ObjectMetadata
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{objects: make(map[string]storage.ObjectMetadata)}
}

func (s *fakeStorage) GeneratePresignedUploadURL(_ context.Context, key, _ string, _ time.Duration) (string, error) {
	return "https://r2.example/upload/" + key, nil
}

func (s *fakeStorage) HeadObject(_ context.Context, key string) (*storage.ObjectMetadata, error) {
	meta, ok := s.objects[key]
	if !ok {
		return nil, storage.ErrObjectNotFound
	}
	return &meta, nil
}

func (s *fakeStorage) GenerateSignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://r2.example/signed/" + key, nil
}

func (s *fakeStorage) DeleteObject(_ context.Context, key string) error {
	delete(s.objects, key)
	return nil
}

type fakeWorkoutRepo struct {
	plan     *domainW.WorkoutPlan
	session  *domainW.WorkoutSession
	exercise *domainW.Exercise
}

func (r *fakeWorkoutRepo) CreatePlan(context.Context, *domainW.WorkoutPlan) error { return nil }
func (r *fakeWorkoutRepo) GetPlanByID(context.Context, uuid.UUID) (*domainW.WorkoutPlan, error) {
	return r.plan, nil
}
func (r *fakeWorkoutRepo) ListPlansByStudent(context.Context, uuid.UUID) ([]*domainW.WorkoutPlan, error) {
	return nil, nil
}
func (r *fakeWorkoutRepo) ListPlansByTrainer(context.Context, uuid.UUID) ([]*domainW.WorkoutPlan, error) {
	return nil, nil
}
func (r *fakeWorkoutRepo) UpdatePlan(context.Context, *domainW.WorkoutPlan) error      { return nil }
func (r *fakeWorkoutRepo) CreateSession(context.Context, *domainW.WorkoutSession) error { return nil }
func (r *fakeWorkoutRepo) GetSessionByID(context.Context, uuid.UUID) (*domainW.WorkoutSession, error) {
	return r.session, nil
}
func (r *fakeWorkoutRepo) ListSessionsByPlan(context.Context, uuid.UUID) ([]*domainW.WorkoutSession, error) {
	return nil, nil
}
func (r *fakeWorkoutRepo) CreateExercise(context.Context, *domainW.Exercise) error { return nil }
func (r *fakeWorkoutRepo) GetExerciseByID(context.Context, uuid.UUID) (*domainW.Exercise, error) {
	return r.exercise, nil
}
func (r *fakeWorkoutRepo) ListExercisesBySession(context.Context, uuid.UUID) ([]*domainW.Exercise, error) {
	return nil, nil
}
func (r *fakeWorkoutRepo) DeleteExercise(context.Context, uuid.UUID) error { return nil }

type fakeStudentRepo struct {
	byUser map[uuid.UUID]*domainS.Student
}

func (r *fakeStudentRepo) Create(context.Context, *domainS.Student) error { return nil }
func (r *fakeStudentRepo) GetByID(context.Context, uuid.UUID) (*domainS.Student, error) {
	return nil, nil
}
func (r *fakeStudentRepo) GetByUserID(_ context.Context, userID uuid.UUID) (*domainS.Student, error) {
	return r.byUser[userID], nil
}
func (r *fakeStudentRepo) GetByCPFHash(context.Context, string) (*domainS.Student, error) {
	return nil, nil
}
func (r *fakeStudentRepo) ListByTrainer(context.Context, uuid.UUID) ([]*domainS.StudentBaseInfo, error) {
	return nil, nil
}
func (r *fakeStudentRepo) Delete(context.Context, uuid.UUID) error                  { return nil }
func (r *fakeStudentRepo) SaveBiometric(context.Context, *domainS.BiometricData) error { return nil }

func TestVideoHardeningUseCases(t *testing.T) {
	ctx := context.Background()
	trainerA := uuid.New()
	trainerB := uuid.New()
	studentUser1 := uuid.New()
	studentUser2 := uuid.New()
	student1ID := uuid.New()
	student2ID := uuid.New()
	planID := uuid.New()
	sessionID := uuid.New()
	exerciseID := uuid.New()

	wRepo := &fakeWorkoutRepo{
		plan:     &domainW.WorkoutPlan{ID: planID, TrainerID: trainerA, StudentID: student1ID},
		session:  &domainW.WorkoutSession{ID: sessionID, WorkoutPlanID: planID},
		exercise: &domainW.Exercise{ID: exerciseID, WorkoutSessionID: sessionID},
	}
	sRepo := &fakeStudentRepo{
		byUser: map[uuid.UUID]*domainS.Student{
			studentUser1: {ID: student1ID, UserID: studentUser1, TrainerID: &trainerA},
			studentUser2: {ID: student2ID, UserID: studentUser2, TrainerID: &trainerB},
		},
	}
	cfg := &config.R2Config{SignedURLTTL: 5 * time.Minute}

	t.Run("RequestUploadURL bloqueia outro treinador, valida MIME, sanitiza chave e cria PENDING", func(t *testing.T) {
		vRepo := newFakeVideoRepo()
		store := newFakeStorage()
		uc := appvideo.NewRequestUploadURLUseCase(vRepo, wRepo, store, cfg)

		_, err := uc.Execute(ctx, trainerB, appvideo.RequestUploadURLInput{
			ExerciseID: exerciseID.String(),
			Filename:   "supino.mp4",
			MimeType:   "video/mp4",
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrForbidden))

		_, err = uc.Execute(ctx, trainerA, appvideo.RequestUploadURLInput{
			ExerciseID: exerciseID.String(),
			Filename:   "payload.sh",
			MimeType:   "application/x-sh",
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrInvalidInput))

		out, err := uc.Execute(ctx, trainerA, appvideo.RequestUploadURLInput{
			ExerciseID: exerciseID.String(),
			Filename:   "../../malicious/path/demo.mp4",
			MimeType:   "video/mp4",
		})
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(out.R2Key, "exercises/"+exerciseID.String()+"/"))
		assert.NotContains(t, out.R2Key, "..")
		assert.True(t, strings.HasSuffix(out.R2Key, ".mp4"))

		rec, err := vRepo.GetByKeyAndExercise(ctx, exerciseID, out.R2Key)
		require.NoError(t, err)
		require.NotNil(t, rec)
		assert.Equal(t, appvideo.StatusPending, rec.Status)
		assert.Equal(t, "demo.mp4", rec.Filename)
	})

	t.Run("ConfirmUpload exige objeto real no storage dentro do limite de 50MB", func(t *testing.T) {
		vRepo := newFakeVideoRepo()
		store := newFakeStorage()
		reqUC := appvideo.NewRequestUploadURLUseCase(vRepo, wRepo, store, cfg)
		confUC := appvideo.NewConfirmUploadUseCase(vRepo, wRepo, store)

		reqOut, err := reqUC.Execute(ctx, trainerA, appvideo.RequestUploadURLInput{
			ExerciseID: exerciseID.String(),
			Filename:   "treino.mp4",
			MimeType:   "video/mp4",
		})
		require.NoError(t, err)

		// Sem upload no storage ainda
		_, err = confUC.Execute(ctx, trainerA, appvideo.ConfirmUploadInput{
			ExerciseID: exerciseID.String(),
			R2Key:      reqOut.R2Key,
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrInvalidInput))

		// Arquivo maior que 50MB
		store.objects[reqOut.R2Key] = storage.ObjectMetadata{
			SizeBytes: appvideo.MaxVideoSizeBytes + 1,
			MimeType:  "video/mp4",
		}
		_, err = confUC.Execute(ctx, trainerA, appvideo.ConfirmUploadInput{
			ExerciseID: exerciseID.String(),
			R2Key:      reqOut.R2Key,
		})
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrInvalidInput))

		// Arquivo valido no storage -> confirma usando tamanho real do HeadObject
		store.objects[reqOut.R2Key] = storage.ObjectMetadata{
			SizeBytes: 2048,
			MimeType:  "video/mp4",
		}
		confOut, err := confUC.Execute(ctx, trainerA, appvideo.ConfirmUploadInput{
			ExerciseID: exerciseID.String(),
			R2Key:      reqOut.R2Key,
			SizeBytes:  999999999, // Tentativa de falsificar tamanho pelo cliente é ignorada
		})
		require.NoError(t, err)
		assert.NotEmpty(t, confOut.VideoID)

		confirmed, err := vRepo.GetByExerciseID(ctx, exerciseID)
		require.NoError(t, err)
		require.NotNil(t, confirmed)
		assert.Equal(t, int64(2048), confirmed.SizeBytes)
		assert.Equal(t, appvideo.StatusConfirmed, confirmed.Status)
	})

	t.Run("GetSignedURL valida ownership de treinador e aluno", func(t *testing.T) {
		vRepo := newFakeVideoRepo()
		store := newFakeStorage()
		reqUC := appvideo.NewRequestUploadURLUseCase(vRepo, wRepo, store, cfg)
		confUC := appvideo.NewConfirmUploadUseCase(vRepo, wRepo, store)
		getUC := appvideo.NewGetSignedURLUseCase(vRepo, wRepo, sRepo, store, cfg)

		reqOut, err := reqUC.Execute(ctx, trainerA, appvideo.RequestUploadURLInput{
			ExerciseID: exerciseID.String(),
			Filename:   "treino.mp4",
			MimeType:   "video/mp4",
		})
		require.NoError(t, err)

		// Ainda PENDING -> não retorna URL assinada
		_, err = getUC.Execute(ctx, exerciseID, trainerA, domainU.RoleTrainer)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrNotFound))

		store.objects[reqOut.R2Key] = storage.ObjectMetadata{SizeBytes: 1024, MimeType: "video/mp4"}
		_, err = confUC.Execute(ctx, trainerA, appvideo.ConfirmUploadInput{
			ExerciseID: exerciseID.String(),
			R2Key:      reqOut.R2Key,
		})
		require.NoError(t, err)

		// Trainer A e Student 1 podem acessar; Trainer B e Student 2 são bloqueados
		signed, err := getUC.Execute(ctx, exerciseID, studentUser1, domainU.RoleStudent)
		require.NoError(t, err)
		assert.Contains(t, signed.SignedURL, reqOut.R2Key)

		_, err = getUC.Execute(ctx, exerciseID, studentUser2, domainU.RoleStudent)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrForbidden))

		_, err = getUC.Execute(ctx, exerciseID, trainerB, domainU.RoleTrainer)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errs.ErrForbidden))
	})
}
