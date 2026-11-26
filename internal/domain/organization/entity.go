package organization

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	TypeSolo   = "SOLO"
	TypeStudio = "STUDIO"

	RoleOwner   = "OWNER"
	RoleTrainer = "TRAINER"
	RoleAdmin   = "ADMIN"

	MembershipActive    = "ACTIVE"
	MembershipSuspended = "SUSPENDED"

	RelActive = "ACTIVE"
	RelPaused = "PAUSED"
	RelEnded  = "ENDED"
)

var slugCleanRegex = regexp.MustCompile(`[^a-z0-9-]+`)

type Organization struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Slug      string    `db:"slug" json:"slug"`
	Type      string    `db:"type" json:"type"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (o *Organization) Validate() error {
	o.Name = strings.TrimSpace(o.Name)
	if len(o.Name) < 2 {
		return errors.New("o nome da organização deve ter pelo menos 2 caracteres")
	}
	if o.Type != TypeSolo && o.Type != TypeStudio {
		return errors.New("tipo de organização inválido")
	}
	if strings.TrimSpace(o.Slug) == "" {
		o.Slug = NormalizeSlug(o.Name)
	}
	if len(o.Slug) < 2 {
		return errors.New("slug da organização inválido")
	}
	return nil
}

func NormalizeSlug(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, " ", "-")
	s = slugCleanRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

type Membership struct {
	ID             uuid.UUID `db:"id" json:"id"`
	OrganizationID uuid.UUID `db:"organization_id" json:"organization_id"`
	UserID         uuid.UUID `db:"user_id" json:"user_id"`
	Role           string    `db:"role" json:"role"`
	Status         string    `db:"status" json:"status"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type OrganizationSummary struct {
	OrganizationID   uuid.UUID `db:"organization_id" json:"organization_id"`
	OrganizationName string    `db:"organization_name" json:"organization_name"`
	OrganizationSlug string    `db:"organization_slug" json:"organization_slug"`
	OrganizationType string    `db:"organization_type" json:"organization_type"`
	MemberRole       string    `db:"member_role" json:"member_role"`
	MemberStatus     string    `db:"member_status" json:"member_status"`
}

type MemberSummary struct {
	ID       uuid.UUID `db:"id" json:"id"`
	UserID   uuid.UUID `db:"user_id" json:"user_id"`
	FullName string    `db:"full_name" json:"full_name"`
	Email    string    `db:"email" json:"email"`
	Role     string    `db:"role" json:"role"`
	Status   string    `db:"status" json:"status"`
}

type TrainerStudentRelationship struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	OrganizationID uuid.UUID  `db:"organization_id" json:"organization_id"`
	TrainerUserID  uuid.UUID  `db:"trainer_user_id" json:"trainer_user_id"`
	StudentID      uuid.UUID  `db:"student_id" json:"student_id"`
	Status         string     `db:"status" json:"status"`
	StartedAt      time.Time  `db:"started_at" json:"started_at"`
	EndedAt        *time.Time `db:"ended_at" json:"ended_at,omitempty"`
}

func IsValidRelationshipStatus(status string) bool {
	return status == RelActive || status == RelPaused || status == RelEnded
}

func IsValidMembershipRole(role string) bool {
	return role == RoleOwner || role == RoleTrainer || role == RoleAdmin
}
