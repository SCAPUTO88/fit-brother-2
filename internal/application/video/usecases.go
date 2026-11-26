package video

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainStorage "github.com/SCAPUTO88/fit-brother/internal/domain/storage"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	"github.com/SCAPUTO88/fit-brother/internal/domain/user"
	domainW "github.com/SCAPUTO88/fit-brother/internal/domain/workout"
)

const (
	StatusPending   = "PENDING"
	StatusConfirmed = "CONFIRMED"

	MaxVideoSizeBytes int64 = 50 * 1024 * 1024
)

var allowedVideoMimeExtensions = map[string]string{
	"video/mp4":       ".mp4",
	"video/webm":      ".webm",
	"video/quicktime": ".mov",
}

type StorageClient = domainStorage.Port

type VideoRecord struct {
	ID         uuid.UUID
	ExerciseID uuid.UUID
	UploaderID uuid.UUID
	R2Key      string
	Filename   string
	SizeBytes  int64
	MimeType   string
	Status     string
	CreatedAt  time.Time
}

type Repository interface {
	Save(ctx context.Context, v *VideoRecord) error
	GetByExerciseID(ctx context.Context, exerciseID uuid.UUID) (*VideoRecord, error)
	GetByKeyAndExercise(ctx context.Context, exerciseID uuid.UUID, r2Key string) (*VideoRecord, error)
	ConfirmPending(ctx context.Context, id uuid.UUID, sizeBytes int64, mimeType string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type RequestUploadURLUseCase struct {
	repo         Repository
	workoutRepo  domainW.Repository
	storage      domainStorage.Port
	signedURLTTL time.Duration
}

func NewRequestUploadURLUseCase(
	repo Repository,
	workoutRepo domainW.Repository,
	s domainStorage.Port,
	signedURLTTL time.Duration,
) *RequestUploadURLUseCase {
	return &RequestUploadURLUseCase{
		repo:         repo,
		workoutRepo:  workoutRepo,
		storage:      s,
		signedURLTTL: signedURLTTL,
	}
}

func (uc *RequestUploadURLUseCase) Execute(ctx context.Context, uploaderID uuid.UUID, input RequestUploadURLInput) (*RequestUploadURLOutput, error) {
	exerciseID, err := uuid.Parse(input.ExerciseID)
	if err != nil {
		return nil, fmt.Errorf("%w: exercise_id inválido", errs.ErrInvalidInput)
	}

	mimeType := normalizeMIME(input.MimeType)
	ext, ok := allowedVideoMimeExtensions[mimeType]
	if !ok {
		return nil, fmt.Errorf("%w: tipo de vídeo não suportado; use MP4, WebM ou MOV", errs.ErrInvalidInput)
	}

	safeFilename, err := sanitizeFilename(input.Filename)
	if err != nil {
		return nil, fmt.Errorf("%w: nome de arquivo inválido", errs.ErrInvalidInput)
	}

	if err := verifyExerciseTrainerOwnership(ctx, uc.workoutRepo, exerciseID, uploaderID); err != nil {
		return nil, err
	}

	r2Key := fmt.Sprintf("exercises/%s/%s%s", exerciseID, uuid.NewString(), ext)

	ttl := uc.signedURLTTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	presignedURL, err := uc.storage.GeneratePresignedUploadURL(ctx, r2Key, mimeType, ttl)
	if err != nil {
		if errors.Is(err, domainStorage.ErrStorageDisabled) {
			return nil, fmt.Errorf("%w: módulo de vídeo desabilitado neste ambiente", errs.ErrServiceUnavailable)
		}
		return nil, errors.New("falha ao gerar URL de upload")
	}

	record := &VideoRecord{
		ID:         uuid.New(),
		ExerciseID: exerciseID,
		UploaderID: uploaderID,
		R2Key:      r2Key,
		Filename:   safeFilename,
		SizeBytes:  0,
		MimeType:   mimeType,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
	}

	if err := uc.repo.Save(ctx, record); err != nil {
		return nil, errors.New("falha ao registrar intenção de upload")
	}

	return &RequestUploadURLOutput{
		PresignedURL: presignedURL,
		R2Key:        r2Key,
		ExpiresIn:    int(ttl.Seconds()),
	}, nil
}

type ConfirmUploadUseCase struct {
	repo        Repository
	workoutRepo domainW.Repository
	storage     domainStorage.Port
}

func NewConfirmUploadUseCase(
	r Repository,
	workoutRepo domainW.Repository,
	s domainStorage.Port,
) *ConfirmUploadUseCase {
	return &ConfirmUploadUseCase{
		repo:        r,
		workoutRepo: workoutRepo,
		storage:     s,
	}
}

func (uc *ConfirmUploadUseCase) Execute(ctx context.Context, uploaderID uuid.UUID, input ConfirmUploadInput) (*ConfirmUploadOutput, error) {
	exerciseID, err := uuid.Parse(input.ExerciseID)
	if err != nil {
		return nil, fmt.Errorf("%w: exercise_id inválido", errs.ErrInvalidInput)
	}

	if strings.TrimSpace(input.R2Key) == "" {
		return nil, fmt.Errorf("%w: r2_key inválido", errs.ErrInvalidInput)
	}

	if err := verifyExerciseTrainerOwnership(ctx, uc.workoutRepo, exerciseID, uploaderID); err != nil {
		return nil, err
	}

	pending, err := uc.repo.GetByKeyAndExercise(ctx, exerciseID, input.R2Key)
	if err != nil {
		return nil, errors.New("falha ao consultar intenção de upload")
	}
	if pending == nil {
		return nil, fmt.Errorf("%w: intenção de upload não encontrada para este exercício", errs.ErrNotFound)
	}
	if pending.UploaderID != uploaderID {
		return nil, fmt.Errorf("%w: você não iniciou este upload", errs.ErrForbidden)
	}
	if pending.Status == StatusConfirmed {
		return nil, fmt.Errorf("%w: upload já confirmado", errs.ErrConflict)
	}

	meta, err := uc.storage.HeadObject(ctx, input.R2Key)
	if err != nil {
		if errors.Is(err, domainStorage.ErrStorageDisabled) {
			return nil, fmt.Errorf("%w: módulo de vídeo desabilitado neste ambiente", errs.ErrServiceUnavailable)
		}
		if errors.Is(err, domainStorage.ErrObjectNotFound) {
			return nil, fmt.Errorf("%w: arquivo ainda não foi enviado ao armazenamento", errs.ErrInvalidInput)
		}
		return nil, errors.New("falha ao verificar arquivo no armazenamento")
	}

	if meta.SizeBytes <= 0 || meta.SizeBytes > MaxVideoSizeBytes {
		return nil, fmt.Errorf("%w: tamanho do vídeo deve estar entre 1 byte e 50MB", errs.ErrInvalidInput)
	}

	resolvedMIME := pending.MimeType
	if meta.MimeType != "" {
		normalizedMetaMIME := normalizeMIME(meta.MimeType)
		if _, ok := allowedVideoMimeExtensions[normalizedMetaMIME]; !ok || (pending.MimeType != "" && normalizedMetaMIME != pending.MimeType) {
			return nil, fmt.Errorf("%w: tipo MIME do arquivo enviado não corresponde ao formato autorizado", errs.ErrInvalidInput)
		}
		resolvedMIME = normalizedMetaMIME
	}

	if err := uc.repo.ConfirmPending(ctx, pending.ID, meta.SizeBytes, resolvedMIME); err != nil {
		return nil, errors.New("falha ao confirmar metadados do vídeo")
	}

	return &ConfirmUploadOutput{VideoID: pending.ID.String()}, nil
}

type GetSignedURLUseCase struct {
	repo         Repository
	workoutRepo  domainW.Repository
	studRepo     domainS.Repository
	storage      domainStorage.Port
	signedURLTTL time.Duration
}

func NewGetSignedURLUseCase(
	r Repository,
	workoutRepo domainW.Repository,
	studRepo domainS.Repository,
	s domainStorage.Port,
	signedURLTTL time.Duration,
) *GetSignedURLUseCase {
	return &GetSignedURLUseCase{
		repo:         r,
		workoutRepo:  workoutRepo,
		studRepo:     studRepo,
		storage:      s,
		signedURLTTL: signedURLTTL,
	}
}

func (uc *GetSignedURLUseCase) Execute(ctx context.Context, exerciseID, requesterID uuid.UUID, role string) (*GetSignedURLOutput, error) {
	if err := verifyExerciseReadAccess(ctx, uc.workoutRepo, uc.studRepo, exerciseID, requesterID, role); err != nil {
		return nil, err
	}

	video, err := uc.repo.GetByExerciseID(ctx, exerciseID)
	if err != nil {
		return nil, errors.New("falha ao consultar vídeo do exercício")
	}
	if video == nil {
		return nil, fmt.Errorf("%w: vídeo não encontrado para este exercício", errs.ErrNotFound)
	}

	ttl := uc.signedURLTTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	signedURL, err := uc.storage.GenerateSignedURL(ctx, video.R2Key, ttl)
	if err != nil {
		if errors.Is(err, domainStorage.ErrStorageDisabled) {
			return nil, fmt.Errorf("%w: módulo de vídeo desabilitado neste ambiente", errs.ErrServiceUnavailable)
		}
		return nil, errors.New("falha ao gerar URL de acesso ao vídeo")
	}

	return &GetSignedURLOutput{
		SignedURL: signedURL,
		ExpiresIn: int(ttl.Seconds()),
	}, nil
}

func verifyExerciseTrainerOwnership(ctx context.Context, workoutRepo domainW.Repository, exerciseID, trainerID uuid.UUID) error {
	plan, err := loadPlanForExercise(ctx, workoutRepo, exerciseID)
	if err != nil {
		return err
	}
	if plan.TrainerID != trainerID {
		return fmt.Errorf("%w: você não possui permissão sobre este exercício", errs.ErrForbidden)
	}
	return nil
}

func verifyExerciseReadAccess(
	ctx context.Context,
	workoutRepo domainW.Repository,
	studRepo domainS.Repository,
	exerciseID uuid.UUID,
	requesterID uuid.UUID,
	role string,
) error {
	plan, err := loadPlanForExercise(ctx, workoutRepo, exerciseID)
	if err != nil {
		return err
	}

	switch role {
	case user.RoleTrainer:
		if plan.TrainerID != requesterID {
			return fmt.Errorf("%w: você não possui permissão para acessar este vídeo", errs.ErrForbidden)
		}
		return nil
	case user.RoleStudent:
		stud, err := studRepo.GetByUserID(ctx, requesterID)
		if err != nil {
			return errors.New("falha ao verificar vínculo do aluno")
		}
		if stud == nil || plan.StudentID != stud.ID {
			return fmt.Errorf("%w: você não possui permissão para acessar este vídeo", errs.ErrForbidden)
		}
		return nil
	default:
		return fmt.Errorf("%w: acesso negado", errs.ErrForbidden)
	}
}

func loadPlanForExercise(ctx context.Context, workoutRepo domainW.Repository, exerciseID uuid.UUID) (*domainW.WorkoutPlan, error) {
	exercise, err := workoutRepo.GetExerciseByID(ctx, exerciseID)
	if err != nil {
		return nil, errors.New("falha ao carregar exercício")
	}
	if exercise == nil {
		return nil, fmt.Errorf("%w: exercício não encontrado", errs.ErrNotFound)
	}

	session, err := workoutRepo.GetSessionByID(ctx, exercise.WorkoutSessionID)
	if err != nil {
		return nil, errors.New("falha ao carregar sessão de treino")
	}
	if session == nil {
		return nil, fmt.Errorf("%w: sessão não encontrada", errs.ErrNotFound)
	}

	plan, err := workoutRepo.GetPlanByID(ctx, session.WorkoutPlanID)
	if err != nil {
		return nil, errors.New("falha ao carregar plano de treino")
	}
	if plan == nil {
		return nil, fmt.Errorf("%w: plano de treino não encontrado", errs.ErrNotFound)
	}

	return plan, nil
}

func normalizeMIME(raw string) string {
	mime := strings.ToLower(strings.TrimSpace(raw))
	if idx := strings.Index(mime, ";"); idx != -1 {
		mime = strings.TrimSpace(mime[:idx])
	}
	return mime
}

func sanitizeFilename(raw string) (string, error) {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return "", errors.New("nome vazio")
	}
	cleaned = strings.ReplaceAll(cleaned, "\\", "/")
	base := filepath.Base(cleaned)
	if base == "" || base == "." || base == ".." || base == "/" {
		return "", errors.New("nome inválido")
	}
	return base, nil
}
