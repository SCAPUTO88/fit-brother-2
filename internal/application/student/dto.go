package student

import "github.com/google/uuid"

// CreateStudentInput representa os dados que chegam do HTTP
type CreateStudentInput struct {
	TrainerID uuid.UUID `json:"-"`
	Email     string    `json:"email" binding:"required,email"`
	FullName  string    `json:"full_name" binding:"required,min=3"`
	Password  string    `json:"password" binding:"required,min=6"`
	CPF       string    `json:"cpf" binding:"omitempty"`
}

// StudentOutput representa a saida padrao segura (sem dados sensíveis)
type StudentOutput struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	CreatedAt string `json:"created_at"`
}

// StudentDetailOutput representa uma leitura profunda que descriptografa informações confidenciais
type StudentDetailOutput struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	CPF       string `json:"cpf,omitempty"`
	CreatedAt string `json:"created_at"`
}
