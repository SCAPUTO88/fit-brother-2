package user

import (
	"errors"
	"net/mail"
	"time"

	"github.com/google/uuid"
)

// Constantes de regra de negócio (Roles)
const (
	RoleTrainer = "TRAINER"
	RoleStudent = "STUDENT"
	// RoleAdmin = "ADMIN" (A definir no futuro se necessário)
)

// User representa a entidade purista de Domínio para Usuários
// O Domínio não conhece banco de dados, nem JSON.
type User struct {
	ID           uuid.UUID
	Email        string
	FullName     string
	PasswordHash string // Guardamos apenas o Hash c/ Pepper no contexto do domínio
	Role         string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// Validate executará a validação central de negócio para a criação/atualização
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

// Activate / Deactivate controlam o acesso sem deletar registros
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

// RecordTokens encapsula a representação de um Refresh Token no domínio
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string // Hash bruto (ex. SHA256) do token opaco
	FamilyID  uuid.UUID
	IsRevoked bool
	ExpiresAt time.Time
	CreatedAt time.Time
	CreatedIP string
}

// IsValid checa se um refresh token pode ser rotacionado
func (r *RefreshToken) IsValid() bool {
	return !r.IsRevoked && time.Now().Before(r.ExpiresAt)
}
