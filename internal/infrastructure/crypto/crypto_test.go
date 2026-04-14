package crypto_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/crypto"
)

func TestAESEncryptor(t *testing.T) {
	// Chave de 32 bytes (256 bits) codificada em base64
	validKeyBase64 := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))

	encryptor, err := crypto.NewAESEncryptor(validKeyBase64)
	require.NoError(t, err)

	t.Run("Deve criptografar e decriptografar com sucesso (GCM Mode)", func(t *testing.T) {
		plaintext := "dado-secreto-123"
		aad := []byte("student-id-999") // AAD impede que esse campo seja copiado pra outro ID no banco

		// 1. Cifrar
		cipherString, err := encryptor.Encrypt(plaintext, aad)
		assert.NoError(t, err)
		assert.NotEmpty(t, cipherString)
		assert.NotEqual(t, plaintext, cipherString)

		// 2. Decifrar
		decryptedText, err := encryptor.Decrypt(cipherString, aad)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, decryptedText)
	})

	t.Run("Deve falhar ao decriptografar se o AAD estiver incorreto", func(t *testing.T) {
		plaintext := "dado-secreto-123"
		originalAAD := []byte("student-id-999")
		wrongAAD := []byte("student-id-000") // Hacker copiou o dado pro perfil 000!

		cipherString, err := encryptor.Encrypt(plaintext, originalAAD)
		require.NoError(t, err)

		// Deve falhar e não revelar o texto!
		decryptedText, err := encryptor.Decrypt(cipherString, wrongAAD)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "falha na descriptografia")
		assert.Empty(t, decryptedText)
	})

	t.Run("Deve falhar se a chave inicial nao tiver 32 bytes", func(t *testing.T) {
		invalidKey := base64.StdEncoding.EncodeToString([]byte("chave-curta"))
		badEncryptor, err := crypto.NewAESEncryptor(invalidKey)
		assert.Error(t, err)
		assert.Nil(t, badEncryptor)
	})

	t.Run("Deve criar hashes unicos e aleatorios (Nonce Verification)", func(t *testing.T) {
		plaintext := "cpf-mesmo-valor"
		aad := []byte("student-123")

		cipher1, _ := encryptor.Encrypt(plaintext, aad)
		cipher2, _ := encryptor.Encrypt(plaintext, aad)

		// Mesmo que o AAD e o dado sejam idênticos, o Nonce sendo aleatório gera um ciphertext totalmente diferente.
		// Isso previne ataques de dicionário!
		assert.NotEqual(t, cipher1, cipher2)
	})
}

func TestHash(t *testing.T) {
	t.Run("Deve gerar um SHA256 puro consistentemente", func(t *testing.T) {
		hash1 := crypto.SHA256Raw("meu-token-123")
		hash2 := crypto.SHA256Raw("meu-token-123")
		assert.Equal(t, hash1, hash2)
		assert.Equal(t, 64, len(hash1)) // SHA256 em HEX tem 64 chars
	})

	t.Run("Deve gerar um HMAC com pepper diferindo do puro", func(t *testing.T) {
		secretBase64 := base64.StdEncoding.EncodeToString([]byte("super-secret-pepper-for-hmac----"))
		hmacGen, err := crypto.NewHash(secretBase64)
		require.NoError(t, err)

		rawHash := crypto.SHA256Raw("valor")
		macHash := hmacGen.Generate("valor")

		assert.NotEqual(t, rawHash, macHash)
		// E ainda ser consistente
		assert.Equal(t, macHash, hmacGen.Generate("valor"))
	})
}

func TestPasswordService(t *testing.T) {
	pepper := "my-secret-global-pepper"
	svc, err := crypto.NewPasswordService(pepper)
	require.NoError(t, err)

	t.Run("Fazer Hash e Verificar a senha corretamente", func(t *testing.T) {
		pass := "senhaSegura!@123"

		hashResult, err := svc.Hash(pass)
		require.NoError(t, err)
		assert.NotEmpty(t, hashResult)
		assert.True(t, strings.HasPrefix(hashResult, "$2a$")) // Prefixo do bcrypt

		isValid := svc.Check(pass, hashResult)
		assert.True(t, isValid)
	})

	t.Run("Falhar ao fornecer senha errada", func(t *testing.T) {
		hashResult, _ := svc.Hash("senhaCorreta")
		isValid := svc.Check("senhaErrada", hashResult)
		assert.False(t, isValid)
	})
}
