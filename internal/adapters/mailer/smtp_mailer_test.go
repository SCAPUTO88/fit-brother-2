package mailer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
)

func TestBuildMessageSanitizesCRLFHeaderInjection(t *testing.T) {
	msg := BuildMessage(
		"no-reply@fitbrother.test\r\nBcc: evil@attacker.test",
		"student@fitbrother.test\nCc: leak@attacker.test",
		"Convite\r\nX-Injected: true",
		"Corpo seguro",
	)

	headerPart := strings.SplitN(msg, "\r\n\r\n", 2)[0]
	assert.NotContains(t, headerPart, "\nBcc:")
	assert.NotContains(t, headerPart, "\nCc:")
	assert.NotContains(t, headerPart, "\nX-Injected:")
	assert.Len(t, strings.Split(headerPart, "\r\n"), 5)
	assert.Contains(t, headerPart, "From: no-reply@fitbrother.testBcc: evil@attacker.test")
}

func TestNewSMTPMailerValidation(t *testing.T) {
	_, err := NewSMTPMailer(config.SMTPConfig{}, "http://localhost:5173")
	require.Error(t, err)

	m, err := NewSMTPMailer(config.SMTPConfig{
		Host:     "smtp.resend.com",
		Port:     587,
		From:     "no-reply@fitbrother.app",
		StartTLS: true,
	}, "https://app.fitbrother.app/")
	require.NoError(t, err)
	assert.Equal(t, "https://app.fitbrother.app", m.frontendURL)
}
