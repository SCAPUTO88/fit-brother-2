package privacy

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateConsent(ctx context.Context, record *ConsentRecord) error
	ListConsentsByUser(ctx context.Context, userID uuid.UUID) ([]*ConsentRecord, error)

	CreatePrivacyRequest(ctx context.Context, req *PrivacyRequest) error
	ListPrivacyRequestsByUser(ctx context.Context, userID uuid.UUID) ([]*PrivacyRequest, error)
	CompletePrivacyRequestsByUserAndType(ctx context.Context, userID uuid.UUID, reqType string) error

	RecordAuditEvent(ctx context.Context, evt *AuditEvent) error
	ListAuditEventsForUser(ctx context.Context, userID uuid.UUID, limit int) ([]*AuditEvent, error)
}
