package auth

// Entradas de Login
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	TOTPCode string `json:"totp_code"`
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
	RefreshToken string `json:"refresh_token"`
	ClientIP     string `json:"-"`
}

// Entrada para Logout
type LogoutInput struct {
	RefreshToken string `json:"refresh_token"`
}

type PasswordResetRequestInput struct {
	Email    string `json:"email" binding:"required,email"`
	ClientIP string `json:"-"`
}

type PasswordResetConfirmInput struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
	ClientIP    string `json:"-"`
}

type MFAStatusOutput struct {
	Enabled    bool    `json:"enabled"`
	VerifiedAt *string `json:"verified_at,omitempty"`
}

type MFASetupOutput struct {
	Secret     string `json:"secret"`
	OTPAuthURI string `json:"otpauth_uri"`
}

type MFAVerifyInput struct {
	Code string `json:"code" binding:"required,len=6"`
}

type MFADisableInput struct {
	Password string `json:"password" binding:"required"`
	Code     string `json:"code" binding:"required,len=6"`
}
