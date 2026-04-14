package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type Hash struct {
	secret []byte
}

func NewHash(secretBase64 string) (*Hash, error) {
	if secretBase64 == "" {
		return nil, errors.New("secret do hash não pode ser vazio")
	}
	return &Hash{secret: []byte(secretBase64)}, nil
}

func (h *Hash) Generate(plaintext string) string {
	mac := hmac.New(sha256.New, h.secret)
	mac.Write([]byte(plaintext))
	return hex.EncodeToString(mac.Sum(nil))
}

func SHA256Raw(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}
