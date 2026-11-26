package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	domainCatalog "github.com/SCAPUTO88/fit-brother/internal/domain/catalog"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	domainStorage "github.com/SCAPUTO88/fit-brother/internal/domain/storage"
	domainUser "github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

const (
	MaxVideoMediaBytes int64 = 50 * 1024 * 1024
	MaxImageMediaBytes int64 = 10 * 1024 * 1024
	MaxGIFMediaBytes   int64 = 15 * 1024 * 1024
)

type mediaMimeSpec struct {
	mediaType string
	ext       string
	maxBytes  int64
}

var allowedMediaMimeSpecs = map[string]mediaMimeSpec{
	"video/mp4":       {mediaType: domainCatalog.MediaTypeVideo, ext: ".mp4", maxBytes: MaxVideoMediaBytes},
	"video/webm":      {mediaType: domainCatalog.MediaTypeVideo, ext: ".webm", maxBytes: MaxVideoMediaBytes},
	"video/quicktime": {mediaType: domainCatalog.MediaTypeVideo, ext: ".mov", maxBytes: MaxVideoMediaBytes},
	"image/jpeg":      {mediaType: domainCatalog.MediaTypeImage, ext: ".jpg", maxBytes: MaxImageMediaBytes},
	"image/png":       {mediaType: domainCatalog.MediaTypeImage, ext: ".png", maxBytes: MaxImageMediaBytes},
	"image/webp":      {mediaType: domainCatalog.MediaTypeImage, ext: ".webp", maxBytes: MaxImageMediaBytes},
	"image/gif":       {mediaType: domainCatalog.MediaTypeGIF, ext: ".gif", maxBytes: MaxGIFMediaBytes},
}

type UseCase struct {
	repo         domainCatalog.Repository
	orgAccess    domainCatalog.OrgAccessPort
	userRepo     domainUser.Repository
	privacyRepo  domainPrivacy.Repository
	storage      domainStorage.Port
	signedURLTTL time.Duration
}

func NewUseCase(
	repo domainCatalog.Repository,
	orgAccess domainCatalog.OrgAccessPort,
	userRepo domainUser.Repository,
	privacyRepo domainPrivacy.Repository,
	storage domainStorage.Port,
	signedURLTTL time.Duration,
) *UseCase {
	return &UseCase{
		repo:         repo,
		orgAccess:    orgAccess,
		userRepo:     userRepo,
		privacyRepo:  privacyRepo,
		storage:      storage,
		signedURLTTL: signedURLTTL,
	}
}

func (u *UseCase) ListFilters(ctx context.Context) (*FiltersOutput, error) {
	muscles, err := u.repo.ListMuscleGroups(ctx)
	if err != nil {
		return nil, errors.New("falha ao listar grupos musculares")
	}
	equipments, err := u.repo.ListEquipments(ctx)
	if err != nil {
		return nil, errors.New("falha ao listar equipamentos")
	}
	return &FiltersOutput{
		MuscleGroups: muscles,
		Equipments:   equipments,
		Categories:   domainCatalog.AllowedCategories(),
	}, nil
}

func (u *UseCase) SearchExercises(ctx context.Context, requesterID uuid.UUID, role string, filter domainCatalog.ExerciseFilter) (*SearchResultOutput, error) {
	allowedOrgs, err := u.orgAccess.ListActiveOrgIDsForUser(ctx, requesterID, role)
	if err != nil {
		return nil, errors.New("falha ao resolver organizações autorizadas")
	}
	filter.AllowedOrg = allowedOrgs

	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 30
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	items, total, err := u.repo.SearchExercises(ctx, filter)
	if err != nil {
		return nil, errors.New("falha ao pesquisar biblioteca de exercícios")
	}

	dtos := make([]ExerciseDTO, 0, len(items))
	for _, ex := range items {
		dtos = append(dtos, toExerciseDTO(ex))
	}

	return &SearchResultOutput{
		Items:  dtos,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

func (u *UseCase) GetExercise(ctx context.Context, requesterID uuid.UUID, role string, exerciseID uuid.UUID) (*ExerciseDetailOutput, error) {
	ex, err := u.repo.GetExerciseByID(ctx, exerciseID)
	if err != nil {
		return nil, errors.New("falha ao consultar exercício")
	}
	if ex == nil {
		return nil, fmt.Errorf("%w: exercício não encontrado", errs.ErrNotFound)
	}

	if err := u.ensureReadAccess(ctx, requesterID, role, ex); err != nil {
		return nil, err
	}

	mediaRows, err := u.repo.ListMediaByExercise(ctx, ex.ID, true)
	if err != nil {
		return nil, errors.New("falha ao listar mídias do exercício")
	}

	mediaDTOs := make([]MediaDTO, 0, len(mediaRows))
	for _, m := range mediaRows {
		mediaDTOs = append(mediaDTOs, MediaDTO{
			ID:        m.ID.String(),
			MediaType: m.MediaType,
			MimeType:  m.MimeType,
			Filename:  m.Filename,
			SizeBytes: m.SizeBytes,
			Status:    m.Status,
			CreatedAt: m.CreatedAt,
		})
	}

	return &ExerciseDetailOutput{
		ExerciseDTO:    toExerciseDTO(ex),
		InstructionsPT: ex.InstructionsPT,
		InstructionsEN: ex.InstructionsEN,
		Media:          mediaDTOs,
	}, nil
}

func (u *UseCase) CreateCustomExercise(ctx context.Context, requesterID uuid.UUID, role string, input CreateExerciseInput) (*ExerciseDetailOutput, error) {
	if role != domainUser.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem criar exercícios customizados", errs.ErrForbidden)
	}

	if err := domainCatalog.ValidateExerciseFields(input.NamePT, input.PrimaryMuscle, input.Equipment, input.Category, input.Level); err != nil {
		return nil, err
	}

	orgID, err := u.resolveTrainerOrg(ctx, requesterID, input.OrganizationID)
	if err != nil {
		return nil, err
	}

	level := domainCatalog.NormalizeCode(input.Level)
	if level == "" {
		level = "intermediate"
	}

	now := time.Now().UTC()
	ex := &domainCatalog.Exercise{
		ID:               uuid.New(),
		OrganizationID:   &orgID,
		CreatedByUserID:  &requesterID,
		ExternalSource:   domainCatalog.SourceCustom,
		NamePT:           strings.TrimSpace(input.NamePT),
		NameEN:           strings.TrimSpace(input.NameEN),
		PrimaryMuscle:    domainCatalog.NormalizeCode(input.PrimaryMuscle),
		SecondaryMuscles: normalizeStringList(input.SecondaryMuscles, true),
		Equipment:        domainCatalog.NormalizeCode(input.Equipment),
		Category:         domainCatalog.NormalizeCode(input.Category),
		Level:            level,
		Force:            domainCatalog.NormalizeCode(input.Force),
		Mechanic:         domainCatalog.NormalizeCode(input.Mechanic),
		InstructionsPT:   normalizeStringList(input.InstructionsPT, false),
		InstructionsEN:   []string{},
		LicenseInfo:      domainCatalog.LicenseCustom,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := u.repo.CreateCustomExercise(ctx, ex); err != nil {
		return nil, errors.New("falha ao salvar exercício customizado")
	}

	u.recordAudit(ctx, &requesterID, &orgID, "CATALOG_EXERCISE_CREATED", "exercise_catalog", &ex.ID, map[string]any{
		"name_pt":        ex.NamePT,
		"primary_muscle": ex.PrimaryMuscle,
	}, input.ClientIP)

	return &ExerciseDetailOutput{
		ExerciseDTO:    toExerciseDTO(ex),
		InstructionsPT: ex.InstructionsPT,
		Media:          []MediaDTO{},
	}, nil
}

func (u *UseCase) UpdateCustomExercise(ctx context.Context, requesterID uuid.UUID, role string, exerciseID uuid.UUID, input UpdateExerciseInput) (*ExerciseDetailOutput, error) {
	if role != domainUser.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem editar exercícios customizados", errs.ErrForbidden)
	}
	if err := domainCatalog.ValidateExerciseFields(input.NamePT, input.PrimaryMuscle, input.Equipment, input.Category, input.Level); err != nil {
		return nil, err
	}

	ex, err := u.loadMutableCustomExercise(ctx, requesterID, exerciseID)
	if err != nil {
		return nil, err
	}

	ex.NamePT = strings.TrimSpace(input.NamePT)
	ex.PrimaryMuscle = domainCatalog.NormalizeCode(input.PrimaryMuscle)
	ex.SecondaryMuscles = normalizeStringList(input.SecondaryMuscles, true)
	ex.Equipment = domainCatalog.NormalizeCode(input.Equipment)
	ex.Category = domainCatalog.NormalizeCode(input.Category)
	if lvl := domainCatalog.NormalizeCode(input.Level); lvl != "" {
		ex.Level = lvl
	}
	ex.Force = domainCatalog.NormalizeCode(input.Force)
	ex.Mechanic = domainCatalog.NormalizeCode(input.Mechanic)
	ex.InstructionsPT = normalizeStringList(input.InstructionsPT, false)
	ex.UpdatedAt = time.Now().UTC()

	if err := u.repo.UpdateCustomExercise(ctx, ex); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: exercício não encontrado", errs.ErrNotFound)
		}
		return nil, errors.New("falha ao atualizar exercício customizado")
	}

	u.recordAudit(ctx, &requesterID, ex.OrganizationID, "CATALOG_EXERCISE_UPDATED", "exercise_catalog", &ex.ID, map[string]any{
		"name_pt": ex.NamePT,
	}, input.ClientIP)

	return u.GetExercise(ctx, requesterID, role, ex.ID)
}

func (u *UseCase) DeleteCustomExercise(ctx context.Context, requesterID uuid.UUID, role string, exerciseID uuid.UUID, clientIP string) error {
	if role != domainUser.RoleTrainer {
		return fmt.Errorf("%w: apenas treinadores podem remover exercícios customizados", errs.ErrForbidden)
	}

	ex, err := u.loadMutableCustomExercise(ctx, requesterID, exerciseID)
	if err != nil {
		return err
	}

	mediaList, err := u.repo.ListMediaByExercise(ctx, ex.ID, false)
	if err == nil && u.storage != nil {
		for _, m := range mediaList {
			_ = u.storage.DeleteObject(ctx, m.StorageKey)
			_ = u.repo.DeleteMedia(ctx, m.ID)
		}
	}

	if err := u.repo.SoftDeleteCustomExercise(ctx, ex.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: exercício não encontrado", errs.ErrNotFound)
		}
		return errors.New("falha ao remover exercício customizado")
	}

	u.recordAudit(ctx, &requesterID, ex.OrganizationID, "CATALOG_EXERCISE_DELETED", "exercise_catalog", &ex.ID, map[string]any{
		"name_pt": ex.NamePT,
	}, clientIP)
	return nil
}

func (u *UseCase) RequestMediaUploadURL(ctx context.Context, requesterID uuid.UUID, role string, exerciseID uuid.UUID, input RequestMediaUploadInput) (*RequestMediaUploadOutput, error) {
	if role != domainUser.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem enviar mídias", errs.ErrForbidden)
	}
	if u.storage == nil {
		return nil, fmt.Errorf("%w: storage não configurado", errs.ErrServiceUnavailable)
	}

	ex, err := u.loadMutableCustomExercise(ctx, requesterID, exerciseID)
	if err != nil {
		return nil, err
	}

	mimeType := normalizeMIMEType(input.MimeType)
	spec, ok := allowedMediaMimeSpecs[mimeType]
	if !ok {
		return nil, fmt.Errorf("%w: formato de mídia não suportado; use MP4, WebM, MOV, JPG, PNG, WebP ou GIF", errs.ErrInvalidInput)
	}
	if requestedType := strings.ToUpper(strings.TrimSpace(input.MediaType)); requestedType != "" && requestedType != spec.mediaType {
		return nil, fmt.Errorf("%w: media_type não corresponde ao MIME informado", errs.ErrInvalidInput)
	}

	safeFilename, err := sanitizeMediaFilename(input.Filename)
	if err != nil {
		return nil, fmt.Errorf("%w: nome de arquivo inválido", errs.ErrInvalidInput)
	}

	mediaID := uuid.New()
	storageKey := fmt.Sprintf("catalog/%s/%s/%s%s", ex.OrganizationID.String(), ex.ID.String(), mediaID.String(), spec.ext)

	ttl := u.signedURLTTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	presignedURL, err := u.storage.GeneratePresignedUploadURL(ctx, storageKey, mimeType, ttl)
	if err != nil {
		if errors.Is(err, domainStorage.ErrStorageDisabled) {
			return nil, fmt.Errorf("%w: módulo de mídia desabilitado neste ambiente", errs.ErrServiceUnavailable)
		}
		return nil, errors.New("falha ao gerar URL de upload")
	}

	now := time.Now().UTC()
	media := &domainCatalog.ExerciseMedia{
		ID:                mediaID,
		CatalogExerciseID: ex.ID,
		OrganizationID:    ex.OrganizationID,
		UploaderID:        &requesterID,
		StorageKey:        storageKey,
		Filename:          safeFilename,
		MediaType:         spec.mediaType,
		MimeType:          mimeType,
		SizeBytes:         0,
		Status:            domainCatalog.MediaStatusPending,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := u.repo.CreateMedia(ctx, media); err != nil {
		return nil, errors.New("falha ao registrar intenção de upload de mídia")
	}

	return &RequestMediaUploadOutput{
		MediaID:      media.ID.String(),
		PresignedURL: presignedURL,
		StorageKey:   storageKey,
		MediaType:    spec.mediaType,
		ExpiresIn:    int(ttl.Seconds()),
	}, nil
}

func (u *UseCase) ConfirmMediaUpload(ctx context.Context, requesterID uuid.UUID, role string, exerciseID uuid.UUID, input ConfirmMediaUploadInput) (*MediaDTO, error) {
	if role != domainUser.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem confirmar mídias", errs.ErrForbidden)
	}
	if u.storage == nil {
		return nil, fmt.Errorf("%w: storage não configurado", errs.ErrServiceUnavailable)
	}

	if strings.TrimSpace(input.StorageKey) == "" {
		return nil, fmt.Errorf("%w: storage_key é obrigatório", errs.ErrInvalidInput)
	}

	ex, err := u.loadMutableCustomExercise(ctx, requesterID, exerciseID)
	if err != nil {
		return nil, err
	}

	pending, err := u.repo.GetMediaByKeyAndExercise(ctx, ex.ID, input.StorageKey)
	if err != nil {
		return nil, errors.New("falha ao consultar intenção de upload")
	}
	if pending == nil {
		return nil, fmt.Errorf("%w: intenção de upload não encontrada para este exercício", errs.ErrNotFound)
	}
	if pending.UploaderID == nil || *pending.UploaderID != requesterID {
		return nil, fmt.Errorf("%w: você não iniciou este upload", errs.ErrForbidden)
	}
	if pending.Status == domainCatalog.MediaStatusConfirmed {
		return nil, fmt.Errorf("%w: upload de mídia já confirmado", errs.ErrConflict)
	}

	meta, err := u.storage.HeadObject(ctx, input.StorageKey)
	if err != nil {
		if errors.Is(err, domainStorage.ErrStorageDisabled) {
			return nil, fmt.Errorf("%w: módulo de mídia desabilitado neste ambiente", errs.ErrServiceUnavailable)
		}
		if errors.Is(err, domainStorage.ErrObjectNotFound) {
			return nil, fmt.Errorf("%w: arquivo ainda não foi enviado ao armazenamento", errs.ErrInvalidInput)
		}
		return nil, errors.New("falha ao validar arquivo no armazenamento")
	}

	spec, ok := allowedMediaMimeSpecs[pending.MimeType]
	if !ok {
		return nil, fmt.Errorf("%w: formato de mídia inválido", errs.ErrInvalidInput)
	}

	if meta.SizeBytes <= 0 || meta.SizeBytes > spec.maxBytes {
		return nil, fmt.Errorf("%w: tamanho da mídia fora do limite permitido para %s", errs.ErrInvalidInput, pending.MediaType)
	}

	resolvedMIME := pending.MimeType
	if meta.MimeType != "" {
		metaNorm := normalizeMIMEType(meta.MimeType)
		if _, allowed := allowedMediaMimeSpecs[metaNorm]; !allowed || metaNorm != pending.MimeType {
			return nil, fmt.Errorf("%w: tipo MIME do arquivo enviado não corresponde ao formato autorizado", errs.ErrInvalidInput)
		}
		resolvedMIME = metaNorm
	}

	if err := u.repo.ConfirmMedia(ctx, pending.ID, meta.SizeBytes, resolvedMIME); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: upload já confirmado", errs.ErrConflict)
		}
		return nil, errors.New("falha ao confirmar metadados da mídia")
	}

	u.recordAudit(ctx, &requesterID, ex.OrganizationID, "CATALOG_MEDIA_CONFIRMED", "exercise_media", &pending.ID, map[string]any{
		"exercise_id": ex.ID.String(),
		"media_type":  pending.MediaType,
		"size_bytes":  meta.SizeBytes,
	}, input.ClientIP)

	return &MediaDTO{
		ID:        pending.ID.String(),
		MediaType: pending.MediaType,
		MimeType:  resolvedMIME,
		Filename:  pending.Filename,
		SizeBytes: meta.SizeBytes,
		Status:    domainCatalog.MediaStatusConfirmed,
		CreatedAt: pending.CreatedAt,
	}, nil
}

func (u *UseCase) GetMediaSignedURL(ctx context.Context, requesterID uuid.UUID, role string, exerciseID, mediaID uuid.UUID) (*SignedMediaURLOutput, error) {
	if u.storage == nil {
		return nil, fmt.Errorf("%w: storage não configurado", errs.ErrServiceUnavailable)
	}

	ex, err := u.repo.GetExerciseByID(ctx, exerciseID)
	if err != nil {
		return nil, errors.New("falha ao consultar exercício")
	}
	if ex == nil {
		return nil, fmt.Errorf("%w: exercício não encontrado", errs.ErrNotFound)
	}

	if err := u.ensureReadAccess(ctx, requesterID, role, ex); err != nil {
		return nil, err
	}

	media, err := u.repo.GetMediaByID(ctx, mediaID)
	if err != nil {
		return nil, errors.New("falha ao consultar mídia")
	}
	if media == nil || media.CatalogExerciseID != ex.ID {
		return nil, fmt.Errorf("%w: mídia não encontrada para este exercício", errs.ErrNotFound)
	}
	if media.Status != domainCatalog.MediaStatusConfirmed {
		return nil, fmt.Errorf("%w: mídia ainda não confirmada", errs.ErrNotFound)
	}

	ttl := u.signedURLTTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	signedURL, err := u.storage.GenerateSignedURL(ctx, media.StorageKey, ttl)
	if err != nil {
		if errors.Is(err, domainStorage.ErrStorageDisabled) {
			return nil, fmt.Errorf("%w: módulo de mídia desabilitado neste ambiente", errs.ErrServiceUnavailable)
		}
		return nil, errors.New("falha ao gerar URL assinada da mídia")
	}

	return &SignedMediaURLOutput{
		SignedURL: signedURL,
		MediaType: media.MediaType,
		ExpiresIn: int(ttl.Seconds()),
	}, nil
}

func (u *UseCase) ensureReadAccess(ctx context.Context, requesterID uuid.UUID, role string, ex *domainCatalog.Exercise) error {
	if ex.OrganizationID == nil {
		return nil
	}

	allowedOrgs, err := u.orgAccess.ListActiveOrgIDsForUser(ctx, requesterID, role)
	if err != nil {
		return errors.New("falha ao validar permissão organizacional")
	}
	for _, oid := range allowedOrgs {
		if oid == *ex.OrganizationID {
			return nil
		}
	}
	return fmt.Errorf("%w: você não possui permissão para acessar este exercício", errs.ErrForbidden)
}

func (u *UseCase) loadMutableCustomExercise(ctx context.Context, trainerUserID, exerciseID uuid.UUID) (*domainCatalog.Exercise, error) {
	ex, err := u.repo.GetExerciseByID(ctx, exerciseID)
	if err != nil {
		return nil, errors.New("falha ao consultar exercício")
	}
	if ex == nil {
		return nil, fmt.Errorf("%w: exercício não encontrado", errs.ErrNotFound)
	}
	if ex.OrganizationID == nil {
		return nil, fmt.Errorf("%w: exercícios públicos do catálogo global não podem ser modificados", errs.ErrForbidden)
	}

	mem, err := u.orgAccess.GetMembership(ctx, *ex.OrganizationID, trainerUserID)
	if err != nil {
		return nil, errors.New("falha ao validar vínculo organizacional")
	}
	if mem == nil || mem.Status != "ACTIVE" {
		return nil, fmt.Errorf("%w: você não pertence à organização proprietária deste exercício", errs.ErrForbidden)
	}
	return ex, nil
}

func (u *UseCase) resolveTrainerOrg(ctx context.Context, trainerUserID uuid.UUID, rawOrgID string) (uuid.UUID, error) {
	if trimmed := strings.TrimSpace(rawOrgID); trimmed != "" {
		orgID, err := uuid.Parse(trimmed)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%w: organization_id inválido", errs.ErrInvalidInput)
		}
		mem, err := u.orgAccess.GetMembership(ctx, orgID, trainerUserID)
		if err != nil {
			return uuid.Nil, errors.New("falha ao validar organização")
		}
		if mem == nil || mem.Status != "ACTIVE" {
			return uuid.Nil, fmt.Errorf("%w: você não é membro ativo desta organização", errs.ErrForbidden)
		}
		return orgID, nil
	}

	trainerName := "Personal"
	if u.userRepo != nil {
		if usr, err := u.userRepo.GetByID(ctx, trainerUserID); err == nil && usr != nil {
			trainerName = usr.FullName
		}
	}
	org, err := u.orgAccess.EnsureTrainerDefaultOrg(ctx, trainerUserID, trainerName)
	if err != nil {
		return uuid.Nil, errors.New("falha ao resolver organização padrão do treinador")
	}
	return org.ID, nil
}

func (u *UseCase) recordAudit(
	ctx context.Context,
	actorID *uuid.UUID,
	orgID *uuid.UUID,
	action, targetType string,
	targetID *uuid.UUID,
	meta map[string]any,
	clientIP string,
) {
	if u.privacyRepo == nil {
		return
	}
	rawMeta, err := json.Marshal(meta)
	if err != nil {
		rawMeta = []byte("{}")
	}
	var ipPtr *string
	if trimmed := strings.TrimSpace(clientIP); trimmed != "" {
		ipPtr = &trimmed
	}
	_ = u.privacyRepo.RecordAuditEvent(ctx, &domainPrivacy.AuditEvent{
		ID:             uuid.New(),
		ActorUserID:    actorID,
		OrganizationID: orgID,
		Action:         action,
		TargetType:     targetType,
		TargetID:       targetID,
		MetadataJSON:   string(rawMeta),
		IPAddress:      ipPtr,
		CreatedAt:      time.Now().UTC(),
	})
}

func toExerciseDTO(ex *domainCatalog.Exercise) ExerciseDTO {
	var orgStr *string
	if ex.OrganizationID != nil {
		s := ex.OrganizationID.String()
		orgStr = &s
	}
	return ExerciseDTO{
		ID:               ex.ID.String(),
		OrganizationID:   orgStr,
		ExternalSource:   ex.ExternalSource,
		ExternalID:       ex.ExternalID,
		NamePT:           ex.NamePT,
		NameEN:           ex.NameEN,
		PrimaryMuscle:    ex.PrimaryMuscle,
		PrimaryMusclePT:  TranslateMusclePT(ex.PrimaryMuscle),
		SecondaryMuscles: ex.SecondaryMuscles,
		Equipment:        ex.Equipment,
		EquipmentPT:      TranslateEquipmentPT(ex.Equipment),
		Category:         ex.Category,
		CategoryPT:       TranslateCategoryPT(ex.Category),
		Level:            ex.Level,
		Force:            ex.Force,
		Mechanic:         ex.Mechanic,
		LicenseInfo:      ex.LicenseInfo,
		IsCustom:         ex.OrganizationID != nil,
	}
}

func normalizeStringList(items []string, lowercase bool) []string {
	out := make([]string, 0, len(items))
	for _, s := range items {
		trimmed := strings.TrimSpace(s)
		if trimmed == "" {
			continue
		}
		if lowercase {
			trimmed = strings.ToLower(trimmed)
		}
		out = append(out, trimmed)
	}
	return out
}

func normalizeMIMEType(raw string) string {
	m := strings.ToLower(strings.TrimSpace(raw))
	if idx := strings.Index(m, ";"); idx != -1 {
		m = strings.TrimSpace(m[:idx])
	}
	return m
}

func sanitizeMediaFilename(raw string) (string, error) {
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
