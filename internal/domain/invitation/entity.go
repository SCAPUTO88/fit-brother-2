package invitation

import (
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending  = "PENDING"
	StatusAccepted = "ACCEPTED"
	StatusExpired  = "EXPIRED"
	StatusRevoked  = "REVOKED"
)

type Invitation struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	OrganizationID uuid.UUID  `db:"organization_id" json:"organization_id"`
	TrainerUserID  uuid.UUID  `db:"trainer_user_id" json:"trainer_user_id"`
	Email          string     `db:"email" json:"email"`
	FullName       string     `db:"full_name" json:"full_name"`
	TokenHash      string     `db:"token_hash" json:"-"`
	Status         string     `db:"status" json:"status"`
	ExpiresAt      time.Time  `db:"expires_at" json:"expires_at"`
	AcceptedAt     *time.Time `db:"accepted_at" json:"accepted_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}

func (i *Invitation) Validate() error {
	i.Email = strings.ToLower(strings.TrimSpace(i.Email))
	i.FullName = strings.TrimSpace(i.FullName)
	if len(i.FullName) < 3 {
		return errors.New("o nome completo deve ter ao menos 3 caracteres")
	}
	if _, err := mail.ParseAddress(i.Email); err != nil {
		return errors.New("email com formato inválido")
	}
	if i.OrganizationID == uuid.Nil || i.TrainerUserID == uuid.Nil {
		return errors.New("organização e treinador são obrigatórios")
	}
	return nil
}

func (i *Invitation) IsExpired(now time.Time) bool {
	return !now.Before(i.ExpiresAt)
}

func (i *Invitation) CanAccept(now time.Time) bool {
	return i.Status == StatusPending && !i.IsExpired(now)
}
