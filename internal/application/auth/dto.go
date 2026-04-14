package auth

// Entradas de Login
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	ClientIP string `json:"-"`
}

// Saídas padrão de Token
type TokenOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // em segundos
	Role         string `json:"role"`
}

// Entrada para o Refresh Token
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	ClientIP     string `json:"-"`
}

// Entrada para Logout
type LogoutInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
