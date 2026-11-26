package privacy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/SCAPUTO88/fit-brother/internal/domain/errs"
	domainOrg "github.com/SCAPUTO88/fit-brother/internal/domain/organization"
	domainPrivacy "github.com/SCAPUTO88/fit-brother/internal/domain/privacy"
	domainS "github.com/SCAPUTO88/fit-brother/internal/domain/student"
	"github.com/SCAPUTO88/fit-brother/internal/domain/tx"
	domainU "github.com/SCAPUTO88/fit-brother/internal/domain/user"
	domainW "github.com/SCAPUTO88/fit-brother/internal/domain/workout"
)

type RecordConsentInput struct {
	Purpose       string `json:"purpose" binding:"required"`
	PolicyVersion string `json:"policy_version" binding:"required"`
	Granted       bool   `json:"granted"`
	ClientIP      string `json:"-"`
	UserAgent     string `json:"-"`
}

type CreatePrivacyRequestInput struct {
	RequestType string `json:"request_type" binding:"required"`
	Details     string `json:"details"`
	ClientIP    string `json:"-"`
}

type ExportUserProfile struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type ExportStudentProfile struct {
	ID        string  `json:"id"`
	TrainerID *string `json:"trainer_id,omitempty"`
	HasCPF    bool    `json:"has_cpf"`
	CreatedAt string  `json:"created_at"`
}

type ExportWorkoutSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
}

type ExportDataOutput struct {
	ExportedAt      string                           `json:"exported_at"`
	User            ExportUserProfile                `json:"user"`
	Student         *ExportStudentProfile            `json:"student,omitempty"`
	Organizations   []*domainOrg.OrganizationSummary `json:"organizations"`
	Workouts        []ExportWorkoutSummary           `json:"workouts"`
	Consents        []*domainPrivacy.ConsentRecord   `json:"consents"`
	PrivacyRequests []*domainPrivacy.PrivacyRequest  `json:"privacy_requests"`
	AuditEvents     []*domainPrivacy.AuditEvent      `json:"audit_events"`
}

type UseCase struct {
	privacyRepo domainPrivacy.Repository
	userRepo    domainU.Repository
	mfaRepo     domainU.MFARepository
	studRepo    domainS.Repository
	orgRepo     domainOrg.Repository
	workoutRepo domainW.Repository
	txManager   tx.Manager
}

func NewUseCase(
	privacyRepo domainPrivacy.Repository,
	userRepo domainU.Repository,
	mfaRepo domainU.MFARepository,
	studRepo domainS.Repository,
	orgRepo domainOrg.Repository,
	workoutRepo domainW.Repository,
	txManager tx.Manager,
) *UseCase {
	return &UseCase{
		privacyRepo: privacyRepo,
		userRepo:    userRepo,
		mfaRepo:     mfaRepo,
		studRepo:    studRepo,
		orgRepo:     orgRepo,
		workoutRepo: workoutRepo,
		txManager:   txManager,
	}
}

func (u *UseCase) RecordConsent(ctx context.Context, userID uuid.UUID, input RecordConsentInput) (*domainPrivacy.ConsentRecord, error) {
	rec := &domainPrivacy.ConsentRecord{
		ID:            uuid.New(),
		UserID:        userID,
		Purpose:       strings.ToUpper(strings.TrimSpace(input.Purpose)),
		PolicyVersion: strings.TrimSpace(input.PolicyVersion),
		Granted:       input.Granted,
		IPAddress:     optionalNonEmpty(input.ClientIP),
		UserAgent:     optionalNonEmpty(input.UserAgent),
		RecordedAt:    time.Now().UTC(),
	}
	if err := rec.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrInvalidInput, err)
	}
	if err := u.privacyRepo.CreateConsent(ctx, rec); err != nil {
		return nil, errors.New("falha ao registrar consentimento LGPD")
	}
	u.audit(ctx, userID, "CONSENT_RECORDED", "consent_record", &rec.ID, input.ClientIP)
	return rec, nil
}

func (u *UseCase) ListConsents(ctx context.Context, userID uuid.UUID) ([]*domainPrivacy.ConsentRecord, error) {
	items, err := u.privacyRepo.ListConsentsByUser(ctx, userID)
	if err != nil {
		return nil, errors.New("falha ao listar consentimentos")
	}
	return items, nil
}

func (u *UseCase) CreateRequest(ctx context.Context, userID uuid.UUID, input CreatePrivacyRequestInput) (*domainPrivacy.PrivacyRequest, error) {
	req := &domainPrivacy.PrivacyRequest{
		ID:          uuid.New(),
		UserID:      userID,
		RequestType: strings.ToUpper(strings.TrimSpace(input.RequestType)),
		Status:      domainPrivacy.StatusOpen,
		Details:     optionalNonEmpty(input.Details),
		RequestedAt: time.Now().UTC(),
	}
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrInvalidInput, err)
	}
	if err := u.privacyRepo.CreatePrivacyRequest(ctx, req); err != nil {
		return nil, errors.New("falha ao registrar solicitação LGPD")
	}
	u.audit(ctx, userID, "PRIVACY_REQUEST_CREATED", "privacy_request", &req.ID, input.ClientIP)
	return req, nil
}

func (u *UseCase) ListRequests(ctx context.Context, userID uuid.UUID) ([]*domainPrivacy.PrivacyRequest, error) {
	items, err := u.privacyRepo.ListPrivacyRequestsByUser(ctx, userID)
	if err != nil {
		return nil, errors.New("falha ao listar solicitações LGPD")
	}
	return items, nil
}

func (u *UseCase) ExportMyData(ctx context.Context, userID uuid.UUID, clientIP string) (*ExportDataOutput, error) {
	usr, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("falha ao consultar usuário")
	}
	if usr == nil || !usr.IsActive {
		return nil, fmt.Errorf("%w: usuário não encontrado ou inativo", errs.ErrNotFound)
	}

	var studOut *ExportStudentProfile
	workoutsOut := make([]ExportWorkoutSummary, 0)
	if u.studRepo != nil {
		stud, err := u.studRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, errors.New("falha ao consultar perfil de aluno")
		}
		if stud != nil {
			var tidStr *string
			if stud.TrainerID != nil {
				v := stud.TrainerID.String()
				tidStr = &v
			}
			studOut = &ExportStudentProfile{
				ID:        stud.ID.String(),
				TrainerID: tidStr,
				HasCPF:    stud.CPFEncrypted != nil && *stud.CPFEncrypted != "",
				CreatedAt: stud.CreatedAt.UTC().Format(time.RFC3339),
			}
			if u.workoutRepo != nil {
				plans, err := u.workoutRepo.ListPlansByStudent(ctx, stud.ID)
				if err != nil {
					return nil, errors.New("falha ao consultar planos de treino")
				}
				for _, p := range plans {
					workoutsOut = append(workoutsOut, ExportWorkoutSummary{
						ID:          p.ID.String(),
						Title:       p.Title,
						Description: p.Description,
						IsActive:    p.IsActive,
						CreatedAt:   p.CreatedAt.UTC().Format(time.RFC3339),
					})
				}
			}
		}
	}

	if usr.Role == domainU.RoleTrainer && u.workoutRepo != nil {
		plans, err := u.workoutRepo.ListPlansByTrainer(ctx, userID)
		if err != nil {
			return nil, errors.New("falha ao consultar planos criados")
		}
		for _, p := range plans {
			workoutsOut = append(workoutsOut, ExportWorkoutSummary{
				ID:          p.ID.String(),
				Title:       p.Title,
				Description: p.Description,
				IsActive:    p.IsActive,
				CreatedAt:   p.CreatedAt.UTC().Format(time.RFC3339),
			})
		}
	}

	orgs := make([]*domainOrg.OrganizationSummary, 0)
	if u.orgRepo != nil {
		list, err := u.orgRepo.ListOrganizationsByUser(ctx, userID)
		if err != nil {
			return nil, errors.New("falha ao consultar organizações")
		}
		orgs = list
	}

	consents, err := u.privacyRepo.ListConsentsByUser(ctx, userID)
	if err != nil {
		return nil, errors.New("falha ao consultar consentimentos")
	}

	now := time.Now().UTC()
	exportReq := &domainPrivacy.PrivacyRequest{
		ID:          uuid.New(),
		UserID:      userID,
		RequestType: domainPrivacy.RequestExport,
		Status:      domainPrivacy.StatusCompleted,
		RequestedAt: now,
		CompletedAt: &now,
	}
	_ = u.privacyRepo.CreatePrivacyRequest(ctx, exportReq)
	u.audit(ctx, userID, "PRIVACY_DATA_EXPORTED", "user", &userID, clientIP)

	reqs, err := u.privacyRepo.ListPrivacyRequestsByUser(ctx, userID)
	if err != nil {
		return nil, errors.New("falha ao consultar histórico de solicitações")
	}
	events, err := u.privacyRepo.ListAuditEventsForUser(ctx, userID, 50)
	if err != nil {
		return nil, errors.New("falha ao consultar trilha de auditoria")
	}

	return &ExportDataOutput{
		ExportedAt: now.Format(time.RFC3339),
		User: ExportUserProfile{
			ID:        usr.ID.String(),
			Email:     usr.Email,
			FullName:  usr.FullName,
			Role:      usr.Role,
			IsActive:  usr.IsActive,
			CreatedAt: usr.CreatedAt.UTC().Format(time.RFC3339),
		},
		Student:         studOut,
		Organizations:   orgs,
		Workouts:        workoutsOut,
		Consents:        consents,
		PrivacyRequests: reqs,
		AuditEvents:     events,
	}, nil
}

func (u *UseCase) AnonymizeSelf(ctx context.Context, userID uuid.UUID, clientIP string) error {
	now := time.Now().UTC()
	run := func(execCtx context.Context) error {
		usr, err := u.userRepo.GetByID(execCtx, userID)
		if err != nil {
			return errors.New("falha ao consultar usuário")
		}
		if usr == nil || !usr.IsActive {
			return fmt.Errorf("%w: usuário não encontrado ou já anonimizado", errs.ErrNotFound)
		}

		if u.studRepo != nil {
			stud, err := u.studRepo.GetByUserID(execCtx, usr.ID)
			if err != nil {
				return errors.New("falha ao consultar perfil de aluno")
			}
			if stud != nil {
				if u.orgRepo != nil {
					if err := u.orgRepo.EndAllByStudent(execCtx, stud.ID); err != nil {
						return errors.New("falha ao encerrar vínculos do aluno")
					}
				}
				if err := u.studRepo.Delete(execCtx, stud.ID); err != nil {
					return errors.New("falha ao remover dados do aluno")
				}
			}
		}

		if u.mfaRepo != nil {
			_ = u.mfaRepo.DisableMFA(execCtx, usr.ID)
		}

		usr.Email = "deleted-" + usr.ID.String() + "@anon.local"
		usr.FullName = "Usuário Anonimizado"
		usr.PasswordHash = "ANONIMIZADO"
		usr.IsActive = false
		usr.DeletedAt = &now

		if err := u.userRepo.Update(execCtx, usr); err != nil {
			return errors.New("falha ao anonimizar usuário")
		}
		if err := u.userRepo.RevokeAllByUserID(execCtx, usr.ID); err != nil {
			return errors.New("falha ao revogar sessões")
		}

		_ = u.privacyRepo.CompletePrivacyRequestsByUserAndType(execCtx, usr.ID, domainPrivacy.RequestAnonymize)
		u.audit(execCtx, usr.ID, "SELF_ANONYMIZED", "user", &usr.ID, clientIP)
		return nil
	}

	if u.txManager != nil {
		return u.txManager.RunInTx(ctx, run)
	}
	return run(ctx)
}

func (u *UseCase) ListAuditEvents(ctx context.Context, userID uuid.UUID, role string) ([]*domainPrivacy.AuditEvent, error) {
	if role != domainU.RoleTrainer {
		return nil, fmt.Errorf("%w: apenas treinadores podem consultar eventos de auditoria organizacional", errs.ErrForbidden)
	}
	items, err := u.privacyRepo.ListAuditEventsForUser(ctx, userID, 100)
	if err != nil {
		return nil, errors.New("falha ao listar eventos de auditoria")
	}
	return items, nil
}

func (u *UseCase) audit(ctx context.Context, actorID uuid.UUID, action, targetType string, targetID *uuid.UUID, clientIP string) {
	_ = u.privacyRepo.RecordAuditEvent(ctx, &domainPrivacy.AuditEvent{
		ID:           uuid.New(),
		ActorUserID:  &actorID,
		Action:       action,
		TargetType:   targetType,
		TargetID:     targetID,
		MetadataJSON: "{}",
		IPAddress:    optionalNonEmpty(clientIP),
		CreatedAt:    time.Now().UTC(),
	})
}

func optionalNonEmpty(v string) *string {
	t := strings.TrimSpace(v)
	if t == "" {
		return nil
	}
	return &t
}
