package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

type AESEncryptor struct {
	key []byte
}

func NewAESEncryptor(keyBase64 string) (*AESEncryptor, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, errors.New("AES_KEY inválida: deve ser base64 de 32 bytes")
	}
	if len(key) != 32 {
		return nil, errors.New("AES_KEY deve ter exatamente 32 bytes (AES-256)")
	}
	return &AESEncryptor{key: key}, nil
}

func (e *AESEncryptor) Encrypt(plaintext string, aad []byte) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), aad)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *AESEncryptor) Decrypt(encoded string, aad []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.New("dado cifrado inválido: erro de base64")
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("dado cifrado corrompido: tamanho inválido")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return "", errors.New("falha na descriptografia: dado inválido ou adulterado")
	}

	return string(plaintext), nil
}
