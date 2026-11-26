package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"net/url"
	"strings"
	"time"

	domainMailer "github.com/SCAPUTO88/fit-brother/internal/domain/mailer"
	"github.com/SCAPUTO88/fit-brother/internal/infrastructure/config"
)

// NoopMailer evita bloqueio de startup quando SMTP não está configurado em dev/testes.
type NoopMailer struct{}

func NewNoopMailer() *NoopMailer {
	return &NoopMailer{}
}

func (m *NoopMailer) SendInvitation(_ context.Context, _ domainMailer.InvitationEmail) error {
	return nil
}

func (m *NoopMailer) SendPasswordReset(_ context.Context, _ domainMailer.PasswordResetEmail) error {
	return nil
}

// SMTPMailer envia e-mails transacionais via SMTP com TLS/STARTTLS e proteção contra header injection.
type SMTPMailer struct {
	host        string
	port        int
	username    string
	password    string
	from        string
	useStartTLS bool
	frontendURL string
}

func NewSMTPMailer(cfg config.SMTPConfig, frontendURL string) (*SMTPMailer, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, fmt.Errorf("SMTP_HOST obrigatório")
	}
	if strings.TrimSpace(cfg.From) == "" {
		return nil, fmt.Errorf("SMTP_FROM obrigatório")
	}
	port := cfg.Port
	if port <= 0 {
		port = 587
	}
	return &SMTPMailer{
		host:        strings.TrimSpace(cfg.Host),
		port:        port,
		username:    strings.TrimSpace(cfg.Username),
		password:    cfg.Password,
		from:        sanitizeHeaderValue(cfg.From),
		useStartTLS: cfg.StartTLS,
		frontendURL: strings.TrimRight(strings.TrimSpace(frontendURL), "/"),
	}, nil
}

func (m *SMTPMailer) SendInvitation(ctx context.Context, msg domainMailer.InvitationEmail) error {
	link := fmt.Sprintf("%s/#/activate?token=%s", m.frontendURL, url.QueryEscape(msg.RawToken))
	subject := "Convite para acessar o Fit Brother"
	body := fmt.Sprintf(
		"Olá %s,\r\n\r\n%s (%s) convidou você para acessar o Fit Brother.\r\n"+
			"Ative sua conta em: %s\r\n"+
			"Este convite expira em %s.\r\n",
		sanitizeHeaderValue(msg.StudentName),
		sanitizeHeaderValue(msg.TrainerName),
		sanitizeHeaderValue(msg.OrgName),
		link,
		msg.ExpiresAt.UTC().Format(time.RFC3339),
	)
	return m.send(ctx, msg.ToEmail, subject, body)
}

func (m *SMTPMailer) SendPasswordReset(ctx context.Context, msg domainMailer.PasswordResetEmail) error {
	link := fmt.Sprintf("%s/#/forgot-password?token=%s", m.frontendURL, url.QueryEscape(msg.RawToken))
	subject := "Redefinição de senha — Fit Brother"
	body := fmt.Sprintf(
		"Olá %s,\r\n\r\nRecebemos uma solicitação para redefinir sua senha.\r\n"+
			"Use o link abaixo para concluir a redefinição (válido até %s):\r\n%s\r\n\r\n"+
			"Se você não solicitou esta alteração, ignore este e-mail.\r\n",
		sanitizeHeaderValue(msg.FullName),
		msg.ExpiresAt.UTC().Format(time.RFC3339),
		link,
	)
	return m.send(ctx, msg.ToEmail, subject, body)
}

func (m *SMTPMailer) send(ctx context.Context, toEmail, subject, body string) error {
	to := sanitizeHeaderValue(toEmail)
	if to == "" || strings.ContainsAny(to, "\r\n") {
		return fmt.Errorf("destinatário inválido")
	}

	payload := BuildMessage(m.from, to, subject, body)
	addr := net.JoinHostPort(m.host, fmt.Sprintf("%d", m.port))

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	var conn net.Conn
	var err error
	if m.port == 465 {
		tlsDialer := &tls.Dialer{
			NetDialer: dialer,
			Config:    &tls.Config{ServerName: m.host, MinVersion: tls.VersionTLS12},
		}
		conn, err = tlsDialer.DialContext(ctx, "tcp", addr)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer func() { _ = client.Close() }()

	if m.port != 465 && m.useStartTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: m.host, MinVersion: tls.VersionTLS12}); err != nil {
				return err
			}
		}
	}

	if m.username != "" {
		auth := smtp.PlainAuth("", m.username, m.password, m.host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(m.from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	wc, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := wc.Write([]byte(payload)); err != nil {
		_ = wc.Close()
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return client.Quit()
}

// BuildMessage formata cabeçalhos MIME sanitizados contra injeção CRLF.
func BuildMessage(from, to, subject, body string) string {
	cleanFrom := sanitizeHeaderValue(from)
	cleanTo := sanitizeHeaderValue(to)
	cleanSubject := sanitizeHeaderValue(subject)
	return fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		cleanFrom, cleanTo, cleanSubject, body,
	)
}

func sanitizeHeaderValue(v string) string {
	r := strings.NewReplacer("\r", "", "\n", "")
	return strings.TrimSpace(r.Replace(v))
}
