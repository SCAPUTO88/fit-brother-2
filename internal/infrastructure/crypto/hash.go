package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// Hash implementa funções de hash HMAC para dados que precisam
// ser pesquisáveis no banco (como CPF), mas não podem ser armazenados em texto plano.
type Hash struct {
	secret []byte
}

// NewHash cria um novo gerador de HMAC usando um secret (pepper).
func NewHash(secretBase64 string) (*Hash, error) {
	if secretBase64 == "" {
		return nil, errors.New("secret do hash não pode ser vazio")
	}
	return &Hash{secret: []byte(secretBase64)}, nil
}

// Generate cria um HMAC-SHA256 do texto de entrada.
// Retorna a string em hexadecimal para armazenamento.
func (h *Hash) Generate(plaintext string) string {
	mac := hmac.New(sha256.New, h.secret)
	mac.Write([]byte(plaintext))
	return hex.EncodeToString(mac.Sum(nil))
}

// SHA256Raw gera um hash criptográfico sem secret.
// Útil para tokens de refresh, que já são 100% aleatórios.
func SHA256Raw(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}
