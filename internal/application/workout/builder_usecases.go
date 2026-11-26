package workout

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	domainCatalog "github.com/SCAPUTO88/fit-brother/internal/domain/catalog"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	"github.com/SCAPUTO88/fit-brother/internal/domain/tx"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	domain "github.com/SCAPUTO88/fit-brother/internal/domain/workout"
)

type BuilderOrgPort interface {
	EnsureTrainerDefaultOrg(ctx context.Context, trainerUserID uuid.UUID, trainerName string) (*domainOrg.Organization, error)
	GetMembership(ctx context.Context, orgID, userID uuid.UUID) (*domainOrg.Membership, error)
	ListActiveOrgIDsForUser(ctx context.Context, userID uuid.UUID, role string) ([]uuid.UUID, error)
	HasActiveRelationship(ctx context.Context, trainerUserID, studentID uuid.UUID) (bool, error)
}

type BuilderUseCase struct {
	repo        domain.BuilderRepository
	orgPort     BuilderOrgPort
	studRepo    domainS.Repository
	userRepo    domainU.Repository
	catalogRepo domainCatalog.Repository
	privacyRepo domainPrivacy.Repository
	txManager   tx.Manager
}

func NewBuilderUseCase(
	repo domain.BuilderRepository,
	orgPort BuilderOrgPort,
	studRepo domainS.Repository,
	userRepo domainU.Repository,
	catalogRepo domainCatalog.Repository,
	privacyRepo domainPrivacy.Repository,
	txManager tx.Manager,
) *BuilderUseCase {
	return &BuilderUseCase{
		repo:        repo,
		orgPort:     orgPort,
		studRepo:    studRepo,
		userRepo:    userRepo,
		catalogRepo: catalogRepo,
		privacyRepo: privacyRepo,
		txManager:   txManager,
	}
}

func (u *BuilderUseCase) CreateTemplate(ctx context.Context, requesterID uuid.UUID, role string, input CreateTemplateInput) (*TemplateDetailOutput, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem criar templates de treino", errs.ErrForbidden)
	}

	orgID, err := u.resolveTrainerOrg(ctx, requesterID, input.OrganizationID)
	if err != nil {
		return nil, err
	}

	structure, err := u.buildAndValidateStructure(ctx, orgID, input.Sessions)
	if err != nil {
		return nil, err
	}

	tpl, err := domain.NewWorkoutTemplate(orgID, requesterID, input.Title, input.Description, input.Goal, structure)
	if err != nil {
		return nil, err
	}

	if err := u.repo.CreateTemplate(ctx, tpl); err != nil {
		return nil, errors.New("falha ao salvar template de treino")
	}

	u.recordAudit(ctx, &requesterID, &orgID, "WORKOUT_TEMPLATE_CREATED", "workout_template", &tpl.ID, map[string]any{
		"title": tpl.Title,
	}, input.ClientIP)

	return toTemplateDetailOutput(tpl, []*domain.WorkoutVersion{}), nil
}

func (u *BuilderUseCase) ListTemplates(ctx context.Context, requesterID uuid.UUID, role string, includeArchived bool) ([]TemplateSummaryOutput, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem listar templates", errs.ErrForbidden)
	}

	orgIDs, err := u.orgPort.ListActiveOrgIDsForUser(ctx, requesterID, role)
	if err != nil {
		return nil, errors.New("falha ao resolver organizações do treinador")
	}

	out := make([]TemplateSummaryOutput, 0)
	for _, orgID := range orgIDs {
		items, err := u.repo.ListTemplatesByOrganization(ctx, orgID, includeArchived)
		if err != nil {
			return nil, errors.New("falha ao listar templates de treino")
		}
		for _, tpl := range items {
			out = append(out, toTemplateSummaryOutput(tpl))
		}
	}
	return out, nil
}

func (u *BuilderUseCase) GetTemplate(ctx context.Context, requesterID uuid.UUID, role string, templateID uuid.UUID) (*TemplateDetailOutput, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem consultar templates", errs.ErrForbidden)
	}

	tpl, err := u.loadAuthorizedTemplate(ctx, requesterID, templateID)
	if err != nil {
		return nil, err
	}

	versions, err := u.repo.ListVersionsByTemplate(ctx, tpl.ID)
	if err != nil {
		return nil, errors.New("falha ao listar versões do template")
	}

	return toTemplateDetailOutput(tpl, versions), nil
}

func (u *BuilderUseCase) UpdateTemplate(ctx context.Context, requesterID uuid.UUID, role string, templateID uuid.UUID, input UpdateTemplateInput) (*TemplateDetailOutput, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem editar templates", errs.ErrForbidden)
	}

	tpl, err := u.loadAuthorizedTemplate(ctx, requesterID, templateID)
	if err != nil {
		return nil, err
	}

	trimmedTitle := strings.TrimSpace(input.Title)
	if len(trimmedTitle) < 3 {
		return nil, fmt.Errorf("%w: o título do template deve ter ao menos 3 caracteres", errs.ErrInvalidInput)
	}

	structure, err := u.buildAndValidateStructure(ctx, tpl.OrganizationID, input.Sessions)
	if err != nil {
		return nil, err
	}

	tpl.Title = trimmedTitle
	tpl.Description = strings.TrimSpace(input.Description)
	tpl.Goal = strings.TrimSpace(input.Goal)
	tpl.DraftStructure = structure
	if input.IsArchived != nil {
		tpl.IsArchived = *input.IsArchived
	}
	tpl.UpdatedAt = time.Now().UTC()

	if err := u.repo.UpdateTemplateDraft(ctx, tpl); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: template não encontrado", errs.ErrNotFound)
		}
		return nil, errors.New("falha ao atualizar rascunho do template")
	}

	versions, err := u.repo.ListVersionsByTemplate(ctx, tpl.ID)
	if err != nil {
		return nil, errors.New("falha ao consultar versões do template")
	}

	u.recordAudit(ctx, &requesterID, &tpl.OrganizationID, "WORKOUT_TEMPLATE_UPDATED", "workout_template", &tpl.ID, map[string]any{
		"title": tpl.Title,
	}, input.ClientIP)

	return toTemplateDetailOutput(tpl, versions), nil
}

func (u *BuilderUseCase) CloneTemplate(ctx context.Context, requesterID uuid.UUID, role string, sourceID uuid.UUID, input CloneTemplateInput) (*TemplateDetailOutput, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem clonar templates", errs.ErrForbidden)
	}

	orig, err := u.loadAuthorizedTemplate(ctx, requesterID, sourceID)
	if err != nil {
		return nil, err
	}

	newTitle := strings.TrimSpace(input.Title)
	if newTitle == "" {
		newTitle = "Cópia de " + orig.Title
	}
	if len(newTitle) < 3 {
		return nil, fmt.Errorf("%w: o título do template clonado deve ter ao menos 3 caracteres", errs.ErrInvalidInput)
	}

	clonedStruct := domain.CloneStructure(orig.DraftStructure)
	now := time.Now().UTC()
	cloned := &domain.WorkoutTemplate{
		ID:                   uuid.New(),
		OrganizationID:       orig.OrganizationID,
		TrainerUserID:        requesterID,
		Title:                newTitle,
		Description:          orig.Description,
		Goal:                 orig.Goal,
		DraftStructure:       clonedStruct,
		ClonedFromTemplateID: &orig.ID,
		IsArchived:           false,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := u.repo.CreateTemplate(ctx, cloned); err != nil {
		return nil, errors.New("falha ao clonar template de treino")
	}

	u.recordAudit(ctx, &requesterID, &cloned.OrganizationID, "WORKOUT_TEMPLATE_CLONED", "workout_template", &cloned.ID, map[string]any{
		"cloned_from": orig.ID.String(),
		"title":       cloned.Title,
	}, input.ClientIP)

	return toTemplateDetailOutput(cloned, []*domain.WorkoutVersion{}), nil
}

func (u *BuilderUseCase) PublishVersion(ctx context.Context, requesterID uuid.UUID, role string, templateID uuid.UUID, clientIP string) (*VersionDetailOutput, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem publicar versões", errs.ErrForbidden)
	}

	var created *domain.WorkoutVersion
	var orgID uuid.UUID

	err := u.runInTx(ctx, func(txCtx context.Context) error {
		tpl, err := u.loadAuthorizedTemplate(txCtx, requesterID, templateID)
		if err != nil {
			return err
		}
		if tpl.IsArchived {
			return fmt.Errorf("%w: não é possível publicar um template arquivado", errs.ErrInvalidInput)
		}
		if len(tpl.DraftStructure.Sessions) == 0 {
			return fmt.Errorf("%w: adicione ao menos uma sessão antes de publicar", errs.ErrInvalidInput)
		}

		nextNum, err := u.repo.NextVersionNumber(txCtx, tpl.ID)
		if err != nil {
			return errors.New("falha ao gerar número da versão")
		}

		snapshotCopy := domain.CloneStructure(tpl.DraftStructure)
		version := &domain.WorkoutVersion{
			ID:                uuid.New(),
			TemplateID:        tpl.ID,
			VersionNumber:     nextNum,
			Title:             tpl.Title,
			Description:       tpl.Description,
			Goal:              tpl.Goal,
			SnapshotStructure: snapshotCopy,
			PublishedByUserID: &requesterID,
			PublishedAt:       time.Now().UTC(),
		}

		if err := u.repo.CreateVersion(txCtx, version); err != nil {
			return errors.New("falha ao salvar snapshot imutável do treino")
		}

		created = version
		orgID = tpl.OrganizationID
		return nil
	})
	if err != nil {
		return nil, err
	}

	u.recordAudit(ctx, &requesterID, &orgID, "WORKOUT_VERSION_PUBLISHED", "workout_version", &created.ID, map[string]any{
		"template_id":    created.TemplateID.String(),
		"version_number": created.VersionNumber,
	}, clientIP)

	return toVersionDetailOutput(created), nil
}

func (u *BuilderUseCase) ListVersions(ctx context.Context, requesterID uuid.UUID, role string, templateID uuid.UUID) ([]VersionSummaryOutput, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem listar versões", errs.ErrForbidden)
	}

	tpl, err := u.loadAuthorizedTemplate(ctx, requesterID, templateID)
	if err != nil {
		return nil, err
	}

	versions, err := u.repo.ListVersionsByTemplate(ctx, tpl.ID)
	if err != nil {
		return nil, errors.New("falha ao listar versões do template")
	}

	out := make([]VersionSummaryOutput, 0, len(versions))
	for _, v := range versions {
		out = append(out, VersionSummaryOutput{
			ID:            v.ID.String(),
			TemplateID:    v.TemplateID.String(),
			VersionNumber: v.VersionNumber,
			Title:         v.Title,
			Goal:          v.Goal,
			SessionCount:  len(v.SnapshotStructure.Sessions),
			PublishedAt:   v.PublishedAt,
		})
	}
	return out, nil
}

func (u *BuilderUseCase) GetVersion(ctx context.Context, requesterID uuid.UUID, role string, templateID, versionID uuid.UUID) (*VersionDetailOutput, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem consultar versões", errs.ErrForbidden)
	}

	tpl, err := u.loadAuthorizedTemplate(ctx, requesterID, templateID)
	if err != nil {
		return nil, err
	}

	version, err := u.repo.GetVersionByID(ctx, versionID)
	if err != nil {
		return nil, errors.New("falha ao consultar versão do template")
	}
	if version == nil || version.TemplateID != tpl.ID {
		return nil, fmt.Errorf("%w: versão não encontrada para este template", errs.ErrNotFound)
	}
	return toVersionDetailOutput(version), nil
}

func (u *BuilderUseCase) AssignTemplate(ctx context.Context, requesterID uuid.UUID, role string, templateID uuid.UUID, input AssignTemplateInput) (*AssignmentDetailOutput, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem atribuir treinos", errs.ErrForbidden)
	}

	studentID, err := uuid.Parse(strings.TrimSpace(input.StudentID))
	if err != nil {
		return nil, fmt.Errorf("%w: student_id inválido", errs.ErrInvalidInput)
	}

	validFrom := time.Now().UTC()
	if strings.TrimSpace(input.ValidFrom) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(input.ValidFrom))
		if err != nil {
			return nil, fmt.Errorf("%w: valid_from deve usar formato YYYY-MM-DD", errs.ErrInvalidInput)
		}
		validFrom = parsed
	}

	var validUntil *time.Time
	if strings.TrimSpace(input.ValidUntil) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(input.ValidUntil))
		if err != nil {
			return nil, fmt.Errorf("%w: valid_until deve usar formato YYYY-MM-DD", errs.ErrInvalidInput)
		}
		if parsed.Before(validFrom) {
			return nil, fmt.Errorf("%w: valid_until não pode ser anterior a valid_from", errs.ErrInvalidInput)
		}
		validUntil = &parsed
	}

	stud, err := u.studRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("falha ao consultar aluno")
	}
	if stud == nil {
		return nil, fmt.Errorf("%w: aluno não encontrado", errs.ErrNotFound)
	}

	activeRel, err := u.orgPort.HasActiveRelationship(ctx, requesterID, studentID)
	if err != nil {
		return nil, errors.New("falha ao verificar vínculo com o aluno")
	}
	if !activeRel {
		return nil, fmt.Errorf("%w: este aluno não possui vínculo ativo com o treinador", errs.ErrForbidden)
	}

	var assignedID uuid.UUID
	var orgID uuid.UUID

	err = u.runInTx(ctx, func(txCtx context.Context) error {
		tpl, err := u.loadAuthorizedTemplate(txCtx, requesterID, templateID)
		if err != nil {
			return err
		}
		if tpl.IsArchived {
			return fmt.Errorf("%w: não é possível atribuir um template arquivado", errs.ErrInvalidInput)
		}
		orgID = tpl.OrganizationID

		var version *domain.WorkoutVersion
		if trimmedVID := strings.TrimSpace(input.WorkoutVersionID); trimmedVID != "" {
			parsedVID, err := uuid.Parse(trimmedVID)
			if err != nil {
				return fmt.Errorf("%w: workout_version_id inválido", errs.ErrInvalidInput)
			}
			v, err := u.repo.GetVersionByID(txCtx, parsedVID)
			if err != nil {
				return errors.New("falha ao consultar versão")
			}
			if v == nil || v.TemplateID != tpl.ID {
				return fmt.Errorf("%w: versão informada não pertence a este template", errs.ErrNotFound)
			}
			version = v
		} else {
			shouldPublish := tpl.LatestVersionID == nil || (input.PublishIfNeeded != nil && *input.PublishIfNeeded)
			if shouldPublish {
				if len(tpl.DraftStructure.Sessions) == 0 {
					return fmt.Errorf("%w: adicione ao menos uma sessão antes de publicar/atribuir", errs.ErrInvalidInput)
				}
				nextNum, err := u.repo.NextVersionNumber(txCtx, tpl.ID)
				if err != nil {
					return errors.New("falha ao gerar número da versão")
				}
				v := &domain.WorkoutVersion{
					ID:                uuid.New(),
					TemplateID:        tpl.ID,
					VersionNumber:     nextNum,
					Title:             tpl.Title,
					Description:       tpl.Description,
					Goal:              tpl.Goal,
					SnapshotStructure: domain.CloneStructure(tpl.DraftStructure),
					PublishedByUserID: &requesterID,
					PublishedAt:       time.Now().UTC(),
				}
				if err := u.repo.CreateVersion(txCtx, v); err != nil {
					return errors.New("falha ao publicar snapshot do treino")
				}
				version = v
			} else {
				v, err := u.repo.GetVersionByID(txCtx, *tpl.LatestVersionID)
				if err != nil || v == nil {
					return errors.New("falha ao recuperar versão mais recente do template")
				}
				version = v
			}
		}

		now := time.Now().UTC()
		assignment := &domain.WorkoutAssignment{
			ID:               uuid.New(),
			OrganizationID:   tpl.OrganizationID,
			TemplateID:       tpl.ID,
			WorkoutVersionID: version.ID,
			StudentID:        studentID,
			TrainerUserID:    requesterID,
			ValidFrom:        validFrom,
			ValidUntil:       validUntil,
			Status:           domain.AssignmentStatusActive,
			Notes:            strings.TrimSpace(input.Notes),
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := u.repo.CreateAssignment(txCtx, assignment); err != nil {
			return errors.New("falha ao atribuir treino ao aluno")
		}
		assignedID = assignment.ID
		return nil
	})
	if err != nil {
		return nil, err
	}

	full, err := u.repo.GetAssignmentByID(ctx, assignedID)
	if err != nil || full == nil {
		return nil, errors.New("falha ao carregar atribuição criada")
	}

	u.recordAudit(ctx, &requesterID, &orgID, "WORKOUT_ASSIGNMENT_CREATED", "workout_assignment", &full.ID, map[string]any{
		"template_id":        full.TemplateID.String(),
		"workout_version_id": full.WorkoutVersionID.String(),
		"student_id":         full.StudentID.String(),
	}, input.ClientIP)

	return toAssignmentDetailOutput(full), nil
}

func (u *BuilderUseCase) ListAssignments(ctx context.Context, requesterID uuid.UUID, role string, studentIDRaw string) ([]AssignmentSummaryOutput, error) {
	activeOrgIDs, err := u.orgPort.ListActiveOrgIDsForUser(ctx, requesterID, role)
	if err != nil {
		return nil, errors.New("falha ao verificar organizações autorizadas")
	}
	allowedOrgs := make(map[uuid.UUID]bool, len(activeOrgIDs))
	for _, id := range activeOrgIDs {
		allowedOrgs[id] = true
	}

	out := make([]AssignmentSummaryOutput, 0)
	switch role {
	case domainU.RoleStudent:
		stud, err := u.studRepo.GetByUserID(ctx, requesterID)
		if err != nil {
			return nil, errors.New("falha ao resolver perfil do aluno")
		}
		if stud == nil || len(allowedOrgs) == 0 {
			return out, nil
		}
		items, err := u.repo.ListAssignmentsByStudent(ctx, stud.ID)
		if err != nil {
			return nil, errors.New("falha ao listar treinos atribuídos")
		}
		for _, a := range items {
			if !allowedOrgs[a.OrganizationID] {
				continue
			}
			out = append(out, toAssignmentSummaryOutput(a))
		}
		return out, nil
	case domainU.RoleTrainer:
		var studentFilter *uuid.UUID
		if trimmed := strings.TrimSpace(studentIDRaw); trimmed != "" {
			parsed, err := uuid.Parse(trimmed)
			if err != nil {
				return nil, fmt.Errorf("%w: student_id inválido", errs.ErrInvalidInput)
			}
			studentFilter = &parsed
		}
		for _, orgID := range activeOrgIDs {
			items, err := u.repo.ListAssignmentsByOrganization(ctx, orgID, studentFilter)
			if err != nil {
				return nil, errors.New("falha ao listar atribuições de treino")
			}
			for _, a := range items {
				out = append(out, toAssignmentSummaryOutput(a))
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%w: acesso negado", errs.ErrForbidden)
	}
}

func (u *BuilderUseCase) GetAssignment(ctx context.Context, requesterID uuid.UUID, role string, assignmentID uuid.UUID) (*AssignmentDetailOutput, error) {
	a, err := u.repo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return nil, errors.New("falha ao consultar atribuição de treino")
	}
	if a == nil {
		return nil, fmt.Errorf("%w: atribuição de treino não encontrada", errs.ErrNotFound)
	}

	switch role {
	case domainU.RoleStudent:
		stud, err := u.studRepo.GetByUserID(ctx, requesterID)
		if err != nil {
			return nil, errors.New("falha ao verificar perfil do aluno")
		}
		if stud == nil || a.StudentID != stud.ID {
			return nil, fmt.Errorf("%w: acesso negado", errs.ErrForbidden)
		}
		activeOrgIDs, err := u.orgPort.ListActiveOrgIDsForUser(ctx, requesterID, role)
		if err != nil {
			return nil, errors.New("falha ao verificar vínculo ativo do aluno")
		}
		hasOrg := false
		for _, oid := range activeOrgIDs {
			if oid == a.OrganizationID {
				hasOrg = true
				break
			}
		}
		if !hasOrg {
			return nil, fmt.Errorf("%w: vínculo com o treinador encerrado", errs.ErrForbidden)
		}
	case domainU.RoleTrainer:
		mem, err := u.orgPort.GetMembership(ctx, a.OrganizationID, requesterID)
		if err != nil {
			return nil, errors.New("falha ao validar organização")
		}
		if mem == nil || mem.Status != domainOrg.MemberStatusActive {
			return nil, fmt.Errorf("%w: acesso negado", errs.ErrForbidden)
		}
	default:
		return nil, fmt.Errorf("%w: acesso negado", errs.ErrForbidden)
	}

	return toAssignmentDetailOutput(a), nil
}

func (u *BuilderUseCase) UpdateAssignmentStatus(ctx context.Context, requesterID uuid.UUID, role string, assignmentID uuid.UUID, input UpdateAssignmentStatusInput) (*AssignmentSummaryOutput, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem alterar status de atribuições", errs.ErrForbidden)
	}

	status := strings.ToUpper(strings.TrimSpace(input.Status))
	if status != domain.AssignmentStatusActive && status != domain.AssignmentStatusArchived {
		return nil, fmt.Errorf("%w: status inválido; use ACTIVE ou ARCHIVED", errs.ErrInvalidInput)
	}

	a, err := u.repo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return nil, errors.New("falha ao consultar atribuição")
	}
	if a == nil {
		return nil, fmt.Errorf("%w: atribuição não encontrada", errs.ErrNotFound)
	}

	mem, err := u.orgPort.GetMembership(ctx, a.OrganizationID, requesterID)
	if err != nil {
		return nil, errors.New("falha ao verificar vínculo organizacional")
	}
	if mem == nil || mem.Status != domainOrg.MemberStatusActive {
		return nil, fmt.Errorf("%w: você não possui permissão sobre esta atribuição", errs.ErrForbidden)
	}

	if err := u.repo.UpdateAssignmentStatus(ctx, a.ID, status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: atribuição não encontrada", errs.ErrNotFound)
		}
		return nil, errors.New("falha ao atualizar status da atribuição")
	}

	a.Status = status
	a.UpdatedAt = time.Now().UTC()

	u.recordAudit(ctx, &requesterID, &a.OrganizationID, "WORKOUT_ASSIGNMENT_STATUS_UPDATED", "workout_assignment", &a.ID, map[string]any{
		"status": status,
	}, input.ClientIP)

	out := toAssignmentSummaryOutput(a)
	return &out, nil
}

func (u *BuilderUseCase) buildAndValidateStructure(ctx context.Context, orgID uuid.UUID, inputs []TemplateSessionInput) (domain.WorkoutStructure, error) {
	sessions := make([]domain.TemplateSession, 0, len(inputs))
	for sIdx, rawSess := range inputs {
		sessPos := rawSess.Position
		if sessPos <= 0 {
			sessPos = sIdx + 1
		}

		exercises := make([]domain.TemplateExercise, 0, len(rawSess.Exercises))
		for eIdx, rawEx := range rawSess.Exercises {
			exPos := rawEx.Position
			if exPos <= 0 {
				exPos = eIdx + 1
			}

			var catExID *uuid.UUID
			var mediaID *uuid.UUID
			name := strings.TrimSpace(rawEx.Name)
			primaryMuscle := strings.TrimSpace(rawEx.PrimaryMuscle)
			equipment := strings.TrimSpace(rawEx.Equipment)

			if trimmedCat := strings.TrimSpace(rawEx.CatalogExerciseID); trimmedCat != "" {
				parsedCatID, err := uuid.Parse(trimmedCat)
				if err != nil {
					return domain.WorkoutStructure{}, fmt.Errorf("%w: catalog_exercise_id inválido", errs.ErrInvalidInput)
				}
				catExID = &parsedCatID

				if u.catalogRepo != nil {
					catEx, err := u.catalogRepo.GetExerciseByID(ctx, parsedCatID)
					if err != nil {
						return domain.WorkoutStructure{}, errors.New("falha ao consultar exercício no catálogo")
					}
					if catEx == nil {
						return domain.WorkoutStructure{}, fmt.Errorf("%w: exercício do catálogo não encontrado", errs.ErrNotFound)
					}
					if catEx.OrganizationID != nil && *catEx.OrganizationID != orgID {
						return domain.WorkoutStructure{}, fmt.Errorf("%w: exercício customizado pertence a outra organização", errs.ErrForbidden)
					}
					if name == "" {
						name = catEx.NamePT
					}
					if primaryMuscle == "" {
						primaryMuscle = catEx.PrimaryMuscle
					}
					if equipment == "" {
						equipment = catEx.Equipment
					}

					if trimmedMedia := strings.TrimSpace(rawEx.MediaID); trimmedMedia != "" {
						parsedMediaID, err := uuid.Parse(trimmedMedia)
						if err != nil {
							return domain.WorkoutStructure{}, fmt.Errorf("%w: media_id inválido", errs.ErrInvalidInput)
						}
						media, err := u.catalogRepo.GetMediaByID(ctx, parsedMediaID)
						if err != nil {
							return domain.WorkoutStructure{}, errors.New("falha ao validar mídia do exercício")
						}
						if media == nil || media.CatalogExerciseID != catEx.ID || media.Status != domainCatalog.MediaStatusConfirmed {
							return domain.WorkoutStructure{}, fmt.Errorf("%w: mídia não encontrada ou não confirmada para este exercício", errs.ErrInvalidInput)
						}
						mediaID = &parsedMediaID
					} else {
						confirmedMedia, err := u.catalogRepo.ListMediaByExercise(ctx, catEx.ID, true)
						if err == nil && len(confirmedMedia) > 0 {
							id := confirmedMedia[0].ID
							mediaID = &id
						}
					}
				}
			} else if trimmedMedia := strings.TrimSpace(rawEx.MediaID); trimmedMedia != "" {
				parsedMediaID, err := uuid.Parse(trimmedMedia)
				if err != nil {
					return domain.WorkoutStructure{}, fmt.Errorf("%w: media_id inválido", errs.ErrInvalidInput)
				}
				mediaID = &parsedMediaID
			}

			exercises = append(exercises, domain.TemplateExercise{
				ID:                uuid.New(),
				CatalogExerciseID: catExID,
				MediaID:           mediaID,
				Name:              name,
				PrimaryMuscle:     primaryMuscle,
				Equipment:         equipment,
				Sets:              rawEx.Sets,
				Reps:              rawEx.Reps,
				RestSeconds:       rawEx.RestSeconds,
				Notes:             rawEx.Notes,
				Position:          exPos,
			})
		}

		sessions = append(sessions, domain.TemplateSession{
			ID:        uuid.New(),
			Title:     rawSess.Title,
			DayOfWeek: rawSess.DayOfWeek,
			Notes:     rawSess.Notes,
			Position:  sessPos,
			Exercises: exercises,
		})
	}

	return domain.NormalizeAndValidateStructure(domain.WorkoutStructure{Sessions: sessions})
}

func (u *BuilderUseCase) loadAuthorizedTemplate(ctx context.Context, trainerUserID, templateID uuid.UUID) (*domain.WorkoutTemplate, error) {
	tpl, err := u.repo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return nil, errors.New("falha ao consultar template de treino")
	}
	if tpl == nil {
		return nil, fmt.Errorf("%w: template de treino não encontrado", errs.ErrNotFound)
	}

	mem, err := u.orgPort.GetMembership(ctx, tpl.OrganizationID, trainerUserID)
	if err != nil {
		return nil, errors.New("falha ao validar organização do template")
	}
	if mem == nil || mem.Status != domainOrg.MemberStatusActive {
		return nil, fmt.Errorf("%w: você não possui permissão sobre este template", errs.ErrForbidden)
	}
	return tpl, nil
}

func (u *BuilderUseCase) resolveTrainerOrg(ctx context.Context, trainerUserID uuid.UUID, rawOrgID string) (uuid.UUID, error) {
	if trimmed := strings.TrimSpace(rawOrgID); trimmed != "" {
		orgID, err := uuid.Parse(trimmed)
		if err != nil {
			return uuid.Nil, fmt.Errorf("%w: organization_id inválido", errs.ErrInvalidInput)
		}
		mem, err := u.orgPort.GetMembership(ctx, orgID, trainerUserID)
		if err != nil {
			return uuid.Nil, errors.New("falha ao validar organização")
		}
		if mem == nil || mem.Status != domainOrg.MemberStatusActive {
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
	org, err := u.orgPort.EnsureTrainerDefaultOrg(ctx, trainerUserID, trainerName)
	if err != nil {
		return uuid.Nil, errors.New("falha ao resolver organização padrão do treinador")
	}
	return org.ID, nil
}

func (u *BuilderUseCase) runInTx(ctx context.Context, fn func(context.Context) error) error {
	if u.txManager != nil {
		return u.txManager.RunInTx(ctx, fn)
	}
	return fn(ctx)
}

func (u *BuilderUseCase) recordAudit(
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

func toStructureOutput(st domain.WorkoutStructure) WorkoutStructureOutput {
	sessions := make([]TemplateSessionOutput, 0, len(st.Sessions))
	for _, s := range st.Sessions {
		exercises := make([]TemplateExerciseOutput, 0, len(s.Exercises))
		for _, e := range s.Exercises {
			exercises = append(exercises, TemplateExerciseOutput{
				ID:                e.ID.String(),
				CatalogExerciseID: uuidPtrToStr(e.CatalogExerciseID),
				LegacyExerciseID:  uuidPtrToStr(e.LegacyExerciseID),
				MediaID:           uuidPtrToStr(e.MediaID),
				Name:              e.Name,
				PrimaryMuscle:     e.PrimaryMuscle,
				Equipment:         e.Equipment,
				Sets:              e.Sets,
				Reps:              e.Reps,
				RestSeconds:       e.RestSeconds,
				Notes:             e.Notes,
				Position:          e.Position,
				HasMedia:          e.MediaID != nil || e.LegacyExerciseID != nil,
			})
		}
		sessions = append(sessions, TemplateSessionOutput{
			ID:        s.ID.String(),
			Title:     s.Title,
			DayOfWeek: s.DayOfWeek,
			Notes:     s.Notes,
			Position:  s.Position,
			Exercises: exercises,
		})
	}
	return WorkoutStructureOutput{Sessions: sessions}
}

func toTemplateSummaryOutput(tpl *domain.WorkoutTemplate) TemplateSummaryOutput {
	exCount := 0
	for _, s := range tpl.DraftStructure.Sessions {
		exCount += len(s.Exercises)
	}
	return TemplateSummaryOutput{
		ID:                   tpl.ID.String(),
		OrganizationID:       tpl.OrganizationID.String(),
		TrainerUserID:        tpl.TrainerUserID.String(),
		Title:                tpl.Title,
		Description:          tpl.Description,
		Goal:                 tpl.Goal,
		ClonedFromTemplateID: uuidPtrToStr(tpl.ClonedFromTemplateID),
		IsArchived:           tpl.IsArchived,
		LatestVersionNumber:  tpl.LatestVersionNumber,
		LatestVersionID:      uuidPtrToStr(tpl.LatestVersionID),
		SessionCount:         len(tpl.DraftStructure.Sessions),
		ExerciseCount:        exCount,
		CreatedAt:            tpl.CreatedAt,
		UpdatedAt:            tpl.UpdatedAt,
	}
}

func toTemplateDetailOutput(tpl *domain.WorkoutTemplate, versions []*domain.WorkoutVersion) *TemplateDetailOutput {
	verSummaries := make([]VersionSummaryOutput, 0, len(versions))
	for _, v := range versions {
		verSummaries = append(verSummaries, VersionSummaryOutput{
			ID:            v.ID.String(),
			TemplateID:    v.TemplateID.String(),
			VersionNumber: v.VersionNumber,
			Title:         v.Title,
			Goal:          v.Goal,
			SessionCount:  len(v.SnapshotStructure.Sessions),
			PublishedAt:   v.PublishedAt,
		})
	}
	return &TemplateDetailOutput{
		TemplateSummaryOutput: toTemplateSummaryOutput(tpl),
		DraftStructure:        toStructureOutput(tpl.DraftStructure),
		Versions:              verSummaries,
	}
}

func toVersionDetailOutput(v *domain.WorkoutVersion) *VersionDetailOutput {
	return &VersionDetailOutput{
		ID:                v.ID.String(),
		TemplateID:        v.TemplateID.String(),
		VersionNumber:     v.VersionNumber,
		Title:             v.Title,
		Description:       v.Description,
		Goal:              v.Goal,
		SnapshotStructure: toStructureOutput(v.SnapshotStructure),
		PublishedByUserID: uuidPtrToStr(v.PublishedByUserID),
		PublishedAt:       v.PublishedAt,
	}
}

func toAssignmentSummaryOutput(a *domain.WorkoutAssignment) AssignmentSummaryOutput {
	var validUntil *string
	if a.ValidUntil != nil {
		s := a.ValidUntil.Format("2006-01-02")
		validUntil = &s
	}
	return AssignmentSummaryOutput{
		ID:                  a.ID.String(),
		OrganizationID:      a.OrganizationID.String(),
		TemplateID:          a.TemplateID.String(),
		WorkoutVersionID:    a.WorkoutVersionID.String(),
		StudentID:           a.StudentID.String(),
		StudentName:         a.StudentName,
		TrainerUserID:       a.TrainerUserID.String(),
		LegacyWorkoutPlanID: uuidPtrToStr(a.LegacyWorkoutPlanID),
		TemplateTitle:       a.TemplateTitle,
		Goal:                a.Goal,
		VersionNumber:       a.VersionNumber,
		ValidFrom:           a.ValidFrom.Format("2006-01-02"),
		ValidUntil:          validUntil,
		Status:              a.Status,
		Notes:               a.Notes,
		SessionCount:        len(a.SnapshotStructure.Sessions),
		CreatedAt:           a.CreatedAt,
		UpdatedAt:           a.UpdatedAt,
	}
}

func toAssignmentDetailOutput(a *domain.WorkoutAssignment) *AssignmentDetailOutput {
	return &AssignmentDetailOutput{
		AssignmentSummaryOutput: toAssignmentSummaryOutput(a),
		SnapshotStructure:       toStructureOutput(a.SnapshotStructure),
	}
}

func uuidPtrToStr(in *uuid.UUID) *string {
	if in == nil || *in == uuid.Nil {
		return nil
	}
	s := in.String()
	return &s
}
