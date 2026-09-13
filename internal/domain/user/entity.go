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
	ID           uuid.UUID  `db:"id"`
	Email        string     `db:"email"`
	FullName     string     `db:"full_name"`
	PasswordHash string     `db:"password_hash"`
	Role         string     `db:"role"`
	IsActive     bool       `db:"is_active"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
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
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	TokenHash string    `db:"token_hash"`
	FamilyID  uuid.UUID `db:"family_id"`
	IsRevoked bool      `db:"is_revoked"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
	CreatedIP string    `db:"created_ip"`
}

func (r *RefreshToken) IsValid() bool {
	return !r.IsRevoked && time.Now().Before(r.ExpiresAt)
}
