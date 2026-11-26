package totp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAndVerifyTOTP(t *testing.T) {
	secret, err := GenerateSecret()
	require.NoError(t, err)
	assert.Len(t, secret, 32)

	uri := BuildOTPAuthURI("FitBrother", "trainer@fit.test", secret)
	assert.Contains(t, uri, "otpauth://totp/FitBrother:trainer@fit.test?")
	assert.Contains(t, uri, "secret="+secret)

	key, err := base32NoPad.DecodeString(secret)
	require.NoError(t, err)

	now := time.Unix(1732570000, 0).UTC()
	step := now.Unix() / PeriodSeconds
	code := GenerateCodeForStep(key, step)

	matchedStep, ok := VerifyCode(secret, code, now)
	assert.True(t, ok)
	assert.Equal(t, step, matchedStep)

	// Aceita passo adjacente (±30s)
	prevCode := GenerateCodeForStep(key, step-1)
	prevStep, ok := VerifyCode(secret, prevCode, now)
	assert.True(t, ok)
	assert.Equal(t, step-1, prevStep)

	// Rejeita código inválido ou fora da janela
	oldCode := GenerateCodeForStep(key, step-5)
	_, ok = VerifyCode(secret, oldCode, now)
	assert.False(t, ok)

	_, ok = VerifyCode(secret, "00000", now)
	assert.False(t, ok)
}
