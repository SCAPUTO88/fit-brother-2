package privacy

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	PurposeHealthData  = "HEALTH_DATA"
	PurposeMediaUpload = "MEDIA_UPLOAD"
	PurposeTerms       = "TERMS"

	RequestExport    = "EXPORT"
	RequestAnonymize = "ANONYMIZE"
	RequestRectify   = "RECTIFY"

	StatusOpen      = "OPEN"
	StatusCompleted = "COMPLETED"
	StatusRejected  = "REJECTED"
)

type ConsentRecord struct {
	ID            uuid.UUID `db:"id" json:"id"`
	UserID        uuid.UUID `db:"user_id" json:"user_id"`
	Purpose       string    `db:"purpose" json:"purpose"`
	PolicyVersion string    `db:"policy_version" json:"policy_version"`
	Granted       bool      `db:"granted" json:"granted"`
	IPAddress     *string   `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent     *string   `db:"user_agent" json:"user_agent,omitempty"`
	RecordedAt    time.Time `db:"recorded_at" json:"recorded_at"`
}

func (c *ConsentRecord) Validate() error {
	if c.UserID == uuid.Nil {
		return errors.New("usuário obrigatório")
	}
	if !IsValidPurpose(c.Purpose) {
		return errors.New("finalidade de consentimento inválida")
	}
	c.PolicyVersion = strings.TrimSpace(c.PolicyVersion)
	if c.PolicyVersion == "" {
		return errors.New("versão da política obrigatória")
	}
	return nil
}

func IsValidPurpose(p string) bool {
	return p == PurposeHealthData || p == PurposeMediaUpload || p == PurposeTerms
}

type PrivacyRequest struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	UserID      uuid.UUID  `db:"user_id" json:"user_id"`
	RequestType string     `db:"request_type" json:"request_type"`
	Status      string     `db:"status" json:"status"`
	Details     *string    `db:"details" json:"details,omitempty"`
	RequestedAt time.Time  `db:"requested_at" json:"requested_at"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
}

func (r *PrivacyRequest) Validate() error {
	if r.UserID == uuid.Nil {
		return errors.New("usuário obrigatório")
	}
	if !IsValidRequestType(r.RequestType) {
		return errors.New("tipo de solicitação LGPD inválido")
	}
	return nil
}

func IsValidRequestType(t string) bool {
	return t == RequestExport || t == RequestAnonymize || t == RequestRectify
}

type AuditEvent struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	ActorUserID    *uuid.UUID `db:"actor_user_id" json:"actor_user_id,omitempty"`
	OrganizationID *uuid.UUID `db:"organization_id" json:"organization_id,omitempty"`
	Action         string     `db:"action" json:"action"`
	TargetType     string     `db:"target_type" json:"target_type"`
	TargetID       *uuid.UUID `db:"target_id" json:"target_id,omitempty"`
	MetadataJSON   string     `db:"metadata" json:"metadata"`
	IPAddress      *string    `db:"ip_address" json:"ip_address,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}
