package organization

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	"github.com/SCAPUTO88/fit-brother/internal/domain/tx"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
)

type CreateOrganizationInput struct {
	Name     string `json:"name" binding:"required,min=2"`
	Slug     string `json:"slug"`
	Type     string `json:"type"`
	ClientIP string `json:"-"`
}

type AddMemberInput struct {
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required"`
	ClientIP string `json:"-"`
}

type UpdateRelationshipStatusInput struct {
	Status   string `json:"status" binding:"required"`
	ClientIP string `json:"-"`
}

type UseCase struct {
	orgRepo     domainOrg.Repository
	userRepo    domainU.Repository
	privacyRepo domainPrivacy.Repository
	txManager   tx.Manager
}

func NewUseCase(
	orgRepo domainOrg.Repository,
	userRepo domainU.Repository,
	privacyRepo domainPrivacy.Repository,
	txManager tx.Manager,
) *UseCase {
	return &UseCase{
		orgRepo:     orgRepo,
		userRepo:    userRepo,
		privacyRepo: privacyRepo,
		txManager:   txManager,
	}
}

func (u *UseCase) CreateOrganization(ctx context.Context, requesterID uuid.UUID, role string, input CreateOrganizationInput) (*domainOrg.Organization, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem criar organizações", errs.ErrForbidden)
	}

	orgType := strings.ToUpper(strings.TrimSpace(input.Type))
	if orgType == "" {
		orgType = domainOrg.TypeStudio
	}

	slug := domainOrg.NormalizeSlug(input.Slug)
	if strings.TrimSpace(input.Slug) == "" {
		base := domainOrg.NormalizeSlug(input.Name)
		if len(base) < 2 {
			base = "studio"
		}
		slug = fmt.Sprintf("%s-%s", base, uuid.NewString()[:6])
	}

	now := time.Now().UTC()
	org := &domainOrg.Organization{
		ID:        uuid.New(),
		Name:      strings.TrimSpace(input.Name),
		Slug:      slug,
		Type:      orgType,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := org.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrInvalidInput, err)
	}

	membership := &domainOrg.Membership{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		UserID:         requesterID,
		Role:           domainOrg.RoleOwner,
		Status:         domainOrg.MembershipActive,
		CreatedAt:      now,
	}

	run := func(execCtx context.Context) error {
		if err := u.orgRepo.CreateOrganizationWithOwner(execCtx, org, membership); err != nil {
			if isUniqueViolation(err) {
				return fmt.Errorf("%w: slug de organização já está em uso", errs.ErrConflict)
			}
			return errors.New("falha ao criar organização")
		}
		u.audit(execCtx, requesterID, &org.ID, "ORGANIZATION_CREATED", "organization", &org.ID, input.ClientIP)
		return nil
	}

	if u.txManager != nil {
		if err := u.txManager.RunInTx(ctx, run); err != nil {
			return nil, err
		}
	} else if err := run(ctx); err != nil {
		return nil, err
	}

	return org, nil
}

func (u *UseCase) ListMyOrganizations(ctx context.Context, userID uuid.UUID) ([]*domainOrg.OrganizationSummary, error) {
	items, err := u.orgRepo.ListOrganizationsByUser(ctx, userID)
	if err != nil {
		return nil, errors.New("falha ao listar organizações")
	}
	return items, nil
}

func (u *UseCase) ListMembers(ctx context.Context, orgID, requesterID uuid.UUID) ([]*domainOrg.MemberSummary, error) {
	mem, err := u.orgRepo.GetMembership(ctx, orgID, requesterID)
	if err != nil {
		return nil, errors.New("falha ao verificar vínculo na organização")
	}
	if mem == nil || mem.Status != domainOrg.MembershipActive {
		return nil, fmt.Errorf("%w: acesso negado à organização", errs.ErrForbidden)
	}

	members, err := u.orgRepo.ListMembers(ctx, orgID)
	if err != nil {
		return nil, errors.New("falha ao listar membros da organização")
	}
	return members, nil
}

func (u *UseCase) AddMember(ctx context.Context, orgID, requesterID uuid.UUID, input AddMemberInput) (*domainOrg.MemberSummary, error) {
	reqMem, err := u.orgRepo.GetMembership(ctx, orgID, requesterID)
	if err != nil {
		return nil, errors.New("falha ao verificar permissões na organização")
	}
	if reqMem == nil || reqMem.Status != domainOrg.MembershipActive {
		return nil, fmt.Errorf("%w: acesso negado à organização", errs.ErrForbidden)
	}
	if reqMem.Role != domainOrg.RoleOwner && reqMem.Role != domainOrg.RoleAdmin {
		return nil, fmt.Errorf("%w: apenas proprietários ou administradores podem adicionar membros", errs.ErrForbidden)
	}

	targetRole := strings.ToUpper(strings.TrimSpace(input.Role))
	if targetRole != domainOrg.RoleTrainer && targetRole != domainOrg.RoleAdmin {
		return nil, fmt.Errorf("%w: papel permitido para novos membros é TRAINER ou ADMIN", errs.ErrInvalidInput)
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	targetUser, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("falha ao buscar usuário convidado")
	}
	if targetUser == nil || !targetUser.IsActive {
		return nil, fmt.Errorf("%w: usuário não encontrado ou inativo", errs.ErrNotFound)
	}
	if targetUser.Role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem integrar membros de uma organização", errs.ErrInvalidInput)
	}

	existing, err := u.orgRepo.GetMembership(ctx, orgID, targetUser.ID)
	if err != nil {
		return nil, errors.New("falha ao verificar vínculo existente")
	}
	if existing != nil {
		return nil, fmt.Errorf("%w: usuário já é membro desta organização", errs.ErrConflict)
	}

	now := time.Now().UTC()
	mem := &domainOrg.Membership{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         targetUser.ID,
		Role:           targetRole,
		Status:         domainOrg.MembershipActive,
		CreatedAt:      now,
	}

	if err := u.orgRepo.AddMember(ctx, mem); err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("%w: usuário já é membro desta organização", errs.ErrConflict)
		}
		return nil, errors.New("falha ao adicionar membro na organização")
	}

	u.audit(ctx, requesterID, &orgID, "ORGANIZATION_MEMBER_ADDED", "membership", &mem.ID, input.ClientIP)

	return &domainOrg.MemberSummary{
		ID:       mem.ID,
		UserID:   targetUser.ID,
		FullName: targetUser.FullName,
		Email:    targetUser.Email,
		Role:     mem.Role,
		Status:   mem.Status,
	}, nil
}

func (u *UseCase) UpdateRelationshipStatus(
	ctx context.Context,
	orgID, requesterID uuid.UUID,
	role string,
	studentID uuid.UUID,
	input UpdateRelationshipStatusInput,
) error {
	if role != domainU.RoleTrainer {
		return fmt.Errorf("%w: apenas treinadores podem gerenciar vínculos com alunos", errs.ErrForbidden)
	}

	status := strings.ToUpper(strings.TrimSpace(input.Status))
	if !domainOrg.IsValidRelationshipStatus(status) {
		return fmt.Errorf("%w: status de vínculo inválido", errs.ErrInvalidInput)
	}

	mem, err := u.orgRepo.GetMembership(ctx, orgID, requesterID)
	if err != nil {
		return errors.New("falha ao verificar vínculo na organização")
	}
	if mem == nil || mem.Status != domainOrg.MembershipActive {
		return fmt.Errorf("%w: acesso negado à organização", errs.ErrForbidden)
	}

	if err := u.orgRepo.UpdateRelationshipStatus(ctx, orgID, requesterID, studentID, status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: vínculo treinador-aluno não encontrado nesta organização", errs.ErrNotFound)
		}
		return errors.New("falha ao atualizar status do vínculo")
	}

	u.audit(ctx, requesterID, &orgID, "RELATIONSHIP_STATUS_UPDATED", "student", &studentID, input.ClientIP)
	return nil
}

func (u *UseCase) audit(ctx context.Context, actorID uuid.UUID, orgID *uuid.UUID, action, targetType string, targetID *uuid.UUID, clientIP string) {
	if u.privacyRepo == nil {
		return
	}
	var ipPtr *string
	if strings.TrimSpace(clientIP) != "" {
		ip := strings.TrimSpace(clientIP)
		ipPtr = &ip
	}
	_ = u.privacyRepo.RecordAuditEvent(ctx, &domainPrivacy.AuditEvent{
		ID:             uuid.New(),
		ActorUserID:    &actorID,
		OrganizationID: orgID,
		Action:         action,
		TargetType:     targetType,
		TargetID:       targetID,
		MetadataJSON:   "{}",
		IPAddress:      ipPtr,
		CreatedAt:      time.Now().UTC(),
	})
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "23505") || strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique")
}
