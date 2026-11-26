package invitation

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/application/auth"
	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainInv "github.com/SCAPUTO88/fit-brother/internal/domain/invitation"
	domainMailer "github.com/SCAPUTO88/fit-brother/internal/domain/mailer"
	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	"github.com/SCAPUTO88/fit-brother/internal/domain/tx"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/jwt"
)

const defaultInvitationTTL = 72 * time.Hour

type CreateInvitationInput struct {
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email" binding:"required,email"`
	FullName       string `json:"full_name" binding:"required,min=3"`
	TrainerUserID  uuid.UUID
	ClientIP       string `json:"-"`
}

type AcceptInvitationInput struct {
	Token             string `json:"token" binding:"required"`
	Password          string `json:"password" binding:"required,min=8"`
	ConsentTerms      bool   `json:"consent_terms"`
	ConsentHealthData bool   `json:"consent_health_data"`
	PolicyVersion     string `json:"policy_version"`
	ClientIP          string `json:"-"`
	UserAgent         string `json:"-"`
}

type InvitationOutput struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	TrainerUserID  string  `json:"trainer_user_id"`
	Email          string  `json:"email"`
	FullName       string  `json:"full_name"`
	Status         string  `json:"status"`
	ExpiresAt      string  `json:"expires_at"`
	AcceptedAt     *string `json:"accepted_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

type AcceptInvitationOutput struct {
	StudentID      string           `json:"student_id"`
	OrganizationID string           `json:"organization_id"`
	Session        auth.TokenOutput `json:"session"`
}

type UseCase struct {
	invRepo     domainInv.Repository
	orgRepo     domainOrg.Repository
	userRepo    domainU.Repository
	studRepo    domainS.Repository
	privacyRepo domainPrivacy.Repository
	txManager   tx.Manager
	passSvc     *crypto.PasswordService
	jwtSvc      *jwt.Service
	mailer      domainMailer.Port
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

func NewUseCase(
	invRepo domainInv.Repository,
	orgRepo domainOrg.Repository,
	userRepo domainU.Repository,
	studRepo domainS.Repository,
	privacyRepo domainPrivacy.Repository,
	txManager tx.Manager,
	passSvc *crypto.PasswordService,
	jwtSvc *jwt.Service,
	mailer domainMailer.Port,
	accessTTL, refreshTTL time.Duration,
) *UseCase {
	return &UseCase{
		invRepo:     invRepo,
		orgRepo:     orgRepo,
		userRepo:    userRepo,
		studRepo:    studRepo,
		privacyRepo: privacyRepo,
		txManager:   txManager,
		passSvc:     passSvc,
		jwtSvc:      jwtSvc,
		mailer:      mailer,
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
	}
}

func (u *UseCase) Create(ctx context.Context, role string, input CreateInvitationInput) (*InvitationOutput, error) {
	if role != domainU.RoleTrainer || input.TrainerUserID == uuid.Nil {
		return nil, fmt.Errorf("%w: apenas treinadores podem enviar convites", errs.ErrForbidden)
	}

	trainer, err := u.userRepo.GetByID(ctx, input.TrainerUserID)
	if err != nil || trainer == nil || !trainer.IsActive {
		return nil, fmt.Errorf("%w: treinador inválido", errs.ErrUnauthorized)
	}

	var org *domainOrg.Organization
	if strings.TrimSpace(input.OrganizationID) != "" {
		orgID, err := uuid.Parse(strings.TrimSpace(input.OrganizationID))
		if err != nil {
			return nil, fmt.Errorf("%w: organization_id inválido", errs.ErrInvalidInput)
		}
		mem, err := u.orgRepo.GetMembership(ctx, orgID, input.TrainerUserID)
		if err != nil {
			return nil, errors.New("falha ao verificar acesso à organização")
		}
		if mem == nil || mem.Status != domainOrg.MembershipActive {
			return nil, fmt.Errorf("%w: você não pertence a esta organização", errs.ErrForbidden)
		}
		org, err = u.orgRepo.GetOrganizationByID(ctx, orgID)
		if err != nil || org == nil {
			return nil, fmt.Errorf("%w: organização não encontrada", errs.ErrNotFound)
		}
	} else {
		org, err = u.orgRepo.EnsureTrainerDefaultOrg(ctx, input.TrainerUserID, trainer.FullName)
		if err != nil {
			return nil, errors.New("falha ao resolver organização padrão do treinador")
		}
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	existingUser, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("falha ao verificar usuário cadastrado")
	}
	if existingUser != nil {
		return nil, fmt.Errorf("%w: já existe uma conta ativa com este e-mail", errs.ErrConflict)
	}

	rawToken, err := generateOpaqueToken()
	if err != nil {
		return nil, errors.New("falha ao gerar token de convite")
	}

	now := time.Now().UTC()
	inv := &domainInv.Invitation{
		ID:             uuid.New(),
		OrganizationID: org.ID,
		TrainerUserID:  input.TrainerUserID,
		Email:          email,
		FullName:       strings.TrimSpace(input.FullName),
		TokenHash:      crypto.SHA256Raw(rawToken),
		Status:         domainInv.StatusPending,
		ExpiresAt:      now.Add(defaultInvitationTTL),
		CreatedAt:      now,
	}
	if err := inv.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrInvalidInput, err)
	}

	run := func(execCtx context.Context) error {
		if err := u.invRepo.RevokePendingByEmailAndOrg(execCtx, org.ID, inv.Email); err != nil {
			return errors.New("falha ao substituir convites anteriores")
		}
		if err := u.invRepo.Create(execCtx, inv); err != nil {
			return errors.New("falha ao persistir convite")
		}
		u.audit(execCtx, &input.TrainerUserID, &org.ID, "INVITATION_CREATED", "invitation", &inv.ID, input.ClientIP)
		return nil
	}

	if u.txManager != nil {
		if err := u.txManager.RunInTx(ctx, run); err != nil {
			return nil, err
		}
	} else if err := run(ctx); err != nil {
		return nil, err
	}

	if u.mailer != nil {
		if err := u.mailer.SendInvitation(ctx, domainMailer.InvitationEmail{
			ToEmail:     inv.Email,
			StudentName: inv.FullName,
			TrainerName: trainer.FullName,
			OrgName:     org.Name,
			RawToken:    rawToken,
			ExpiresAt:   inv.ExpiresAt,
		}); err != nil {
			return nil, errors.New("convite criado, mas houve falha ao enviar e-mail")
		}
	}

	return toOutput(inv), nil
}

func (u *UseCase) List(ctx context.Context, trainerID uuid.UUID, role string) ([]*InvitationOutput, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem listar convites", errs.ErrForbidden)
	}
	items, err := u.invRepo.ListByTrainer(ctx, trainerID)
	if err != nil {
		return nil, errors.New("falha ao listar convites")
	}
	out := make([]*InvitationOutput, 0, len(items))
	for _, item := range items {
		out = append(out, toOutput(item))
	}
	return out, nil
}

func (u *UseCase) Revoke(ctx context.Context, invitationID, trainerID uuid.UUID, role, clientIP string) error {
	if role != domainU.RoleTrainer {
		return fmt.Errorf("%w: apenas treinadores podem revogar convites", errs.ErrForbidden)
	}
	inv, err := u.invRepo.GetByID(ctx, invitationID)
	if err != nil {
		return errors.New("falha ao buscar convite")
	}
	if inv == nil {
		return fmt.Errorf("%w: convite não encontrado", errs.ErrNotFound)
	}
	if inv.TrainerUserID != trainerID {
		return fmt.Errorf("%w: acesso negado a este convite", errs.ErrForbidden)
	}
	if inv.Status != domainInv.StatusPending {
		return fmt.Errorf("%w: apenas convites pendentes podem ser revogados", errs.ErrConflict)
	}
	if err := u.invRepo.UpdateStatus(ctx, inv.ID, domainInv.StatusRevoked); err != nil {
		return errors.New("falha ao revogar convite")
	}
	u.audit(ctx, &trainerID, &inv.OrganizationID, "INVITATION_REVOKED", "invitation", &inv.ID, clientIP)
	return nil
}

func (u *UseCase) Accept(ctx context.Context, input AcceptInvitationInput) (*AcceptInvitationOutput, error) {
	rawToken := strings.TrimSpace(input.Token)
	if rawToken == "" || len(input.Password) < 8 {
		return nil, fmt.Errorf("%w: token e senha com no mínimo 8 caracteres são obrigatórios", errs.ErrInvalidInput)
	}
	if !input.ConsentTerms || !input.ConsentHealthData {
		return nil, fmt.Errorf("%w: aceite dos Termos de Uso e do tratamento de dados de saúde (LGPD) é obrigatório", errs.ErrInvalidInput)
	}

	policyVersion := strings.TrimSpace(input.PolicyVersion)
	if policyVersion == "" {
		policyVersion = "2026-11"
	}

	hashedPass, err := u.passSvc.Hash(input.Password)
	if err != nil {
		return nil, errors.New("falha interna ao processar senha")
	}
	rawRefreshToken, err := generateOpaqueToken()
	if err != nil {
		return nil, errors.New("falha interna ao gerar sessão")
	}
	tokenHash := crypto.SHA256Raw(rawToken)
	now := time.Now().UTC()

	var result *AcceptInvitationOutput
	run := func(execCtx context.Context) error {
		inv, err := u.invRepo.GetByTokenHash(execCtx, tokenHash)
		if err != nil {
			return errors.New("falha ao validar convite")
		}
		if inv == nil {
			return fmt.Errorf("%w: convite inválido ou expirado", errs.ErrUnauthorized)
		}
		if !inv.CanAccept(now) {
			if inv.Status == domainInv.StatusPending && inv.IsExpired(now) {
				_ = u.invRepo.UpdateStatus(execCtx, inv.ID, domainInv.StatusExpired)
			}
			return fmt.Errorf("%w: convite inválido, expirado ou já utilizado", errs.ErrUnauthorized)
		}

		existingUser, err := u.userRepo.GetByEmail(execCtx, inv.Email)
		if err != nil {
			return errors.New("falha ao verificar e-mail do convite")
		}
		if existingUser != nil {
			return fmt.Errorf("%w: e-mail já associado a uma conta", errs.ErrConflict)
		}

		userDomain := &domainU.User{
			ID:           uuid.New(),
			Email:        inv.Email,
			FullName:     inv.FullName,
			PasswordHash: hashedPass,
			Role:         domainU.RoleStudent,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := userDomain.Validate(); err != nil {
			return fmt.Errorf("%w: %v", errs.ErrInvalidInput, err)
		}
		if err := u.userRepo.Create(execCtx, userDomain); err != nil {
			return errors.New("falha ao criar usuário do aluno")
		}

		studDomain, err := domainS.NewStudent(userDomain.ID, inv.TrainerUserID, "")
		if err != nil {
			return fmt.Errorf("%w: %v", errs.ErrInvalidInput, err)
		}
		studDomain.ClearSensitiveData()
		if err := u.studRepo.Create(execCtx, studDomain); err != nil {
			return errors.New("falha ao criar perfil do aluno")
		}

		rel := &domainOrg.TrainerStudentRelationship{
			ID:             uuid.New(),
			OrganizationID: inv.OrganizationID,
			TrainerUserID:  inv.TrainerUserID,
			StudentID:      studDomain.ID,
			Status:         domainOrg.RelActive,
			StartedAt:      now,
		}
		if err := u.orgRepo.CreateRelationship(execCtx, rel); err != nil {
			return errors.New("falha ao vincular aluno à organização")
		}

		if err := u.invRepo.MarkAccepted(execCtx, inv.ID); err != nil {
			return errors.New("falha ao finalizar convite")
		}

		if u.privacyRepo != nil {
			ipPtr := optionalNonEmpty(input.ClientIP)
			uaPtr := optionalNonEmpty(input.UserAgent)
			for _, purpose := range []string{domainPrivacy.PurposeTerms, domainPrivacy.PurposeHealthData} {
				if err := u.privacyRepo.CreateConsent(execCtx, &domainPrivacy.ConsentRecord{
					ID:            uuid.New(),
					UserID:        userDomain.ID,
					Purpose:       purpose,
					PolicyVersion: policyVersion,
					Granted:       true,
					IPAddress:     ipPtr,
					UserAgent:     uaPtr,
					RecordedAt:    now,
				}); err != nil {
					return errors.New("falha ao registrar consentimento LGPD")
				}
			}
			u.audit(execCtx, &userDomain.ID, &inv.OrganizationID, "INVITATION_ACCEPTED", "student", &studDomain.ID, input.ClientIP)
		}

		rtRecord := &domainU.RefreshToken{
			UserID:    userDomain.ID,
			TokenHash: crypto.SHA256Raw(rawRefreshToken),
			FamilyID:  uuid.New(),
			ExpiresAt: now.Add(u.refreshTTL),
			CreatedIP: input.ClientIP,
		}
		if err := u.userRepo.CreateRefreshToken(execCtx, rtRecord); err != nil {
			return errors.New("falha ao registrar sessão do aluno")
		}

		accessToken, err := u.jwtSvc.GenerateAccessToken(userDomain.ID, userDomain.Role, uuid.NewString(), u.accessTTL)
		if err != nil {
			return errors.New("falha ao emitir token de acesso")
		}

		result = &AcceptInvitationOutput{
			StudentID:      studDomain.ID.String(),
			OrganizationID: inv.OrganizationID.String(),
			Session: auth.TokenOutput{
				AccessToken:  accessToken,
				RefreshToken: rawRefreshToken,
				ExpiresIn:    int(u.accessTTL.Seconds()),
				Role:         userDomain.Role,
			},
		}
		return nil
	}

	if u.txManager != nil {
		if err := u.txManager.RunInTx(ctx, run); err != nil {
			return nil, err
		}
	} else if err := run(ctx); err != nil {
		return nil, err
	}

	return result, nil
}

func (u *UseCase) audit(ctx context.Context, actorID, orgID *uuid.UUID, action, targetType string, targetID *uuid.UUID, clientIP string) {
	if u.privacyRepo == nil {
		return
	}
	_ = u.privacyRepo.RecordAuditEvent(ctx, &domainPrivacy.AuditEvent{
		ID:             uuid.New(),
		ActorUserID:    actorID,
		OrganizationID: orgID,
		Action:         action,
		TargetType:     targetType,
		TargetID:       targetID,
		MetadataJSON:   "{}",
		IPAddress:      optionalNonEmpty(clientIP),
		CreatedAt:      time.Now().UTC(),
	})
}

func toOutput(inv *domainInv.Invitation) *InvitationOutput {
	var acceptedAt *string
	if inv.AcceptedAt != nil {
		s := inv.AcceptedAt.UTC().Format(time.RFC3339)
		acceptedAt = &s
	}
	return &InvitationOutput{
		ID:             inv.ID.String(),
		OrganizationID: inv.OrganizationID.String(),
		TrainerUserID:  inv.TrainerUserID.String(),
		Email:          inv.Email,
		FullName:       inv.FullName,
		Status:         inv.Status,
		ExpiresAt:      inv.ExpiresAt.UTC().Format(time.RFC3339),
		AcceptedAt:     acceptedAt,
		CreatedAt:      inv.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func optionalNonEmpty(v string) *string {
	t := strings.TrimSpace(v)
	if t == "" {
		return nil
	}
	return &t
}

func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}
