package totp

import (
	"crypto/hmac"
	"crypto/rand"
	// RFC 6238 TOTP exige HMAC-SHA1 para interoperabilidade com apps autenticadores.
	// nosemgrep: go.lang.security.audit.crypto.sha1.use-of-sha1
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	PeriodSeconds = 30
	Digits        = 6
)

var base32NoPad = base32.StdEncoding.WithPadding(base32.NoPadding)

// GenerateSecret cria um segredo Base32 aleatório de 160 bits (20 bytes) conforme RFC 4226/6238.
func GenerateSecret() (string, error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base32NoPad.EncodeToString(raw), nil
}

// BuildOTPAuthURI gera a URI padrão otpauth:// para leitura por aplicativos autenticadores.
func BuildOTPAuthURI(issuer, accountEmail, secret string) string {
	label := url.PathEscape(fmt.Sprintf("%s:%s", issuer, accountEmail))
	q := url.Values{}
	q.Set("secret", strings.ToUpper(strings.TrimSpace(secret)))
	q.Set("issuer", issuer)
	q.Set("algorithm", "SHA1")
	q.Set("digits", fmt.Sprintf("%d", Digits))
	q.Set("period", fmt.Sprintf("%d", PeriodSeconds))
	return fmt.Sprintf("otpauth://totp/%s?%s", label, q.Encode())
}

// VerifyCode valida um código TOTP de 6 dígitos permitindo janela de ±1 passo (30s)
// e retorna o contador de passo correspondente para prevenir reuso no mesmo intervalo.
func VerifyCode(secret, code string, now time.Time) (int64, bool) {
	trimmedCode := strings.TrimSpace(code)
	if len(trimmedCode) != Digits {
		return 0, false
	}

	normalizedSecret := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(secret), " ", ""))
	key, err := base32NoPad.DecodeString(normalizedSecret)
	if err != nil {
		return 0, false
	}

	currentStep := now.UTC().Unix() / PeriodSeconds
	for offset := int64(-1); offset <= 1; offset++ {
		step := currentStep + offset
		if step < 0 {
			continue
		}
		expected := GenerateCodeForStep(key, step)
		if subtle.ConstantTimeCompare([]byte(trimmedCode), []byte(expected)) == 1 {
			return step, true
		}
	}
	return 0, false
}

// GenerateCodeForStep calcula o código TOTP de 6 dígitos para um passo específico.
func GenerateCodeForStep(key []byte, step int64) string {
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], uint64(step))

	// nosemgrep: go.lang.security.audit.crypto.sha1.use-of-sha1
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(msg[:])
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	binCode := (uint32(sum[offset]&0x7f) << 24) |
		(uint32(sum[offset+1]) << 16) |
		(uint32(sum[offset+2]) << 8) |
		uint32(sum[offset+3])

	otp := binCode % 1000000
	return fmt.Sprintf("%06d", otp)
}
