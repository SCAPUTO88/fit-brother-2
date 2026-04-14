package crypto

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// PasswordService lida com hashing e verificação de senhas usando bcrypt e pepper.
type PasswordService struct {
	pepper []byte
}

// O custo do bcrypt determina a lentidão. Custo 12 leva ~250ms,
// equilibrando experiência de usuário com proteção contra brute-force.
const bcryptCost = 12

// NewPasswordService inicializa o serviço com um pepper que ficará oculto
// do banco de dados (ex: via variáveis de ambiente).
func NewPasswordService(pepper string) (*PasswordService, error) {
	if pepper == "" {
		return nil, errors.New("o password pepper não pode ser vazio")
	}
	return &PasswordService{pepper: []byte(pepper)}, nil
}

// Hash gera o hash bcrypt da senha combinada com o pepper.
func (s *PasswordService) Hash(password string) (string, error) {
	pepperedPassword := append([]byte(password), s.pepper...)
	hash, err := bcrypt.GenerateFromPassword(pepperedPassword, bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Check verifica se a senha fornecida bate com o hash armazenado.
func (s *PasswordService) Check(password, hash string) bool {
	pepperedPassword := append([]byte(password), s.pepper...)
	err := bcrypt.CompareHashAndPassword([]byte(hash), pepperedPassword)
	return err == nil
}
