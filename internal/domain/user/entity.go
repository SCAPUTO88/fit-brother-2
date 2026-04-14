package user

import (
	"errors"
	"net/mail"
	"time"

	"github.com/google/uuid"
)

const (
	RoleTrainer = "TRAINER"
	RoleStudent = "STUDENT"
)

type User struct {
	ID           uuid.UUID
	Email        string
	FullName     string
	PasswordHash string
	Role         string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

func (u *User) Validate() error {
	if u.FullName == "" {
		return errors.New("o nome completo é obrigatório")
	}
	if len(u.FullName) < 3 {
		return errors.New("o nome completo deve ter ao menos 3 caracteres")
	}
	_, err := mail.ParseAddress(u.Email)
	if err != nil {
		return errors.New("email com formato inválido")
	}
	if u.Role != RoleTrainer && u.Role != RoleStudent {
		return errors.New("papel/role inválido para o usuário")
	}
	return nil
}

func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	FamilyID  uuid.UUID
	IsRevoked bool
	ExpiresAt time.Time
	CreatedAt time.Time
	CreatedIP string
}

func (r *RefreshToken) IsValid() bool {
	return !r.IsRevoked && time.Now().Before(r.ExpiresAt)
}
