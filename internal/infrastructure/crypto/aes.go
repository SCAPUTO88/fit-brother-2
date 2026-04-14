package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// AESEncryptor implementa criptografia autenticada AES-256-GCM.
// O modo GCM garante confidencialidade E integridade — qualquer adulteração
// no dado cifrado resulta em falha na descriptografia.
type AESEncryptor struct {
	key []byte // exatamente 32 bytes para AES-256
}

// NewAESEncryptor cria um encryptor a partir de uma chave base64.
// A chave deve ser mantida em variável de ambiente / KMS, nunca hardcoded.
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

// Encrypt cifra o plaintext com AES-256-GCM.
//
// O parâmetro aad (Additional Authenticated Data) vincula o dado cifrado
// ao contexto — por exemplo, ao ID do aluno. Isso impede que um CPF cifrado
// seja copiado para o registro de outro aluno sem ser detectado.
//
// Formato do output: base64(nonce || ciphertext || tag_GCM)
func (e *AESEncryptor) Encrypt(plaintext string, aad []byte) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Nonce aleatório de 12 bytes (96 bits) — nunca reutilizar com a mesma chave
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// gcm.Seal: nonce como prefixo, ciphertext+tag como resto
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), aad)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt descriptografa e verifica a autenticidade do dado cifrado.
// Retorna erro se o dado estiver corrompido ou se o AAD não corresponder.
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
		// Intencionalmente genérico — não vazar informação sobre o motivo da falha
		return "", errors.New("falha na descriptografia: dado inválido ou adulterado")
	}

	return string(plaintext), nil
}
