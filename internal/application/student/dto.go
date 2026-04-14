package student

// CreateStudentInput representa os dados que chegam do HTTP
type CreateStudentInput struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
	CPF      string `json:"cpf" binding:"required"`
}

// StudentOutput representa a saida segura (sem dados sensíveis)
type StudentOutput struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	CreatedAt string `json:"created_at"`
}
