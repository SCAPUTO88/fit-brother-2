package crypto

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type PasswordService struct {
	pepper []byte
}

const bcryptCost = 12

func NewPasswordService(pepper string) (*PasswordService, error) {
	if pepper == "" {
		return nil, errors.New("o password pepper não pode ser vazio")
	}
	return &PasswordService{pepper: []byte(pepper)}, nil
}

func (s *PasswordService) Hash(password string) (string, error) {
	pepperedPassword := append([]byte(password), s.pepper...)
	hash, err := bcrypt.GenerateFromPassword(pepperedPassword, bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *PasswordService) Check(password, hash string) bool {
	pepperedPassword := append([]byte(password), s.pepper...)
	err := bcrypt.CompareHashAndPassword([]byte(hash), pepperedPassword)
	return err == nil
}
