package video

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/adapters/storage"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
)

type VideoRecord struct {
	ID         uuid.UUID
	ExerciseID uuid.UUID
	UploaderID uuid.UUID
	R2Key      string
	Filename   string
	SizeBytes  int64
	MimeType   string
	CreatedAt  time.Time
}

type Repository interface {
	Save(ctx context.Context, v *VideoRecord) error
	GetByExerciseID(ctx context.Context, exerciseID uuid.UUID) (*VideoRecord, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type RequestUploadURLUseCase struct {
	storage *storage.R2Storage
	cfg     *config.R2Config
}

func NewRequestUploadURLUseCase(s *storage.R2Storage, cfg *config.R2Config) *RequestUploadURLUseCase {
	return &RequestUploadURLUseCase{storage: s, cfg: cfg}
}

func (uc *RequestUploadURLUseCase) Execute(ctx context.Context, uploaderID uuid.UUID, input RequestUploadURLInput) (*RequestUploadURLOutput, error) {
	exerciseID, err := uuid.Parse(input.ExerciseID)
	if err != nil {
		return nil, errors.New("exercise_id inválido")
	}

	r2Key := fmt.Sprintf("exercises/%s/%s-%s", exerciseID, uuid.NewString(), input.Filename)

	ttl := uc.cfg.SignedURLTTL
	if ttl == 0 {
		ttl = 10 * time.Minute
	}

	presignedURL, err := uc.storage.GeneratePresignedUploadURL(ctx, r2Key, ttl)
	if err != nil {
		return nil, errors.New("falha ao gerar URL de upload")
	}

	return &RequestUploadURLOutput{
		PresignedURL: presignedURL,
		R2Key:        r2Key,
		ExpiresIn:    int(ttl.Seconds()),
	}, nil
}

type ConfirmUploadUseCase struct {
	repo    Repository
	storage *storage.R2Storage
}

func NewConfirmUploadUseCase(r Repository, s *storage.R2Storage) *ConfirmUploadUseCase {
	return &ConfirmUploadUseCase{repo: r, storage: s}
}

func (uc *ConfirmUploadUseCase) Execute(ctx context.Context, uploaderID uuid.UUID, input ConfirmUploadInput) (*ConfirmUploadOutput, error) {
	exerciseID, err := uuid.Parse(input.ExerciseID)
	if err != nil {
		return nil, errors.New("exercise_id inválido")
	}

	record := &VideoRecord{
		ID:         uuid.New(),
		ExerciseID: exerciseID,
		UploaderID: uploaderID,
		R2Key:      input.R2Key,
		Filename:   input.Filename,
		SizeBytes:  input.SizeBytes,
		MimeType:   input.MimeType,
		CreatedAt:  time.Now(),
	}

	if err := uc.repo.Save(ctx, record); err != nil {
		return nil, errors.New("falha ao registrar metadados do vídeo")
	}

	return &ConfirmUploadOutput{VideoID: record.ID.String()}, nil
}

type GetSignedURLUseCase struct {
	repo    Repository
	storage *storage.R2Storage
	cfg     *config.R2Config
}

func NewGetSignedURLUseCase(r Repository, s *storage.R2Storage, cfg *config.R2Config) *GetSignedURLUseCase {
	return &GetSignedURLUseCase{repo: r, storage: s, cfg: cfg}
}

func (uc *GetSignedURLUseCase) Execute(ctx context.Context, exerciseID uuid.UUID) (*GetSignedURLOutput, error) {
	video, err := uc.repo.GetByExerciseID(ctx, exerciseID)
	if err != nil || video == nil {
		return nil, errors.New("vídeo não encontrado para este exercício")
	}

	ttl := uc.cfg.SignedURLTTL
	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	signedURL, err := uc.storage.GenerateSignedURL(ctx, video.R2Key, ttl)
	if err != nil {
		return nil, errors.New("falha ao gerar URL de acesso ao vídeo")
	}

	return &GetSignedURLOutput{
		SignedURL: signedURL,
		ExpiresIn: int(ttl.Seconds()),
	}, nil
}
