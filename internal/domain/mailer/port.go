package mailer

import (
	"context"
	"time"
)

// InvitationEmail contém os dados necessários para disparo do convite ao aluno.
type InvitationEmail struct {
	ToEmail     string
	StudentName string
	TrainerName string
	OrgName     string
	RawToken    string
	ExpiresAt   time.Time
}

// PasswordResetEmail contém os dados necessários para redefinição de senha.
type PasswordResetEmail struct {
	ToEmail   string
	FullName  string
	RawToken  string
	ExpiresAt time.Time
}

// Port define o contrato de envio de e-mails transacionais (Arquitetura Hexagonal).
type Port interface {
	SendInvitation(ctx context.Context, msg InvitationEmail) error
	SendPasswordReset(ctx context.Context, msg PasswordResetEmail) error
}
