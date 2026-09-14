package notification

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// SMTPConfig holds the SMTP connection parameters.
type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

// SMTPProvider sends email via SMTP with STARTTLS.
type SMTPProvider struct {
	cfg SMTPConfig
}

// NewSMTPProvider creates a new SMTP-backed Provider.
// Returns a NoopProvider if Host is empty.
func NewSMTPProvider(cfg SMTPConfig) Provider {
	if cfg.Host == "" {
		return &NoopProvider{}
	}
	return &SMTPProvider{cfg: cfg}
}

func (s *SMTPProvider) Send(_ context.Context, msg EmailMessage) error {
	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)

	auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)

	body := buildMIME(s.cfg.From, msg)

	// Connect and upgrade to TLS via STARTTLS
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp: dial %s: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp: new client: %w", err)
	}
	defer client.Close()

	tlsCfg := &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}
	if err := client.StartTLS(tlsCfg); err != nil {
		return fmt.Errorf("smtp: starttls: %w", err)
	}

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp: auth: %w", err)
	}

	if err := client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("smtp: mail from: %w", err)
	}

	recipients := append(msg.To, msg.CC...)
	for _, r := range recipients {
		if err := client.Rcpt(r); err != nil {
			return fmt.Errorf("smtp: rcpt %s: %w", r, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp: data: %w", err)
	}

	if _, err := fmt.Fprint(w, body); err != nil {
		return fmt.Errorf("smtp: write body: %w", err)
	}

	return w.Close()
}

// buildMIME constructs a minimal multipart/alternative MIME message.
func buildMIME(from string, msg EmailMessage) string {
	boundary := "=_cankora_boundary_"
	var sb strings.Builder

	sb.WriteString("From: " + from + "\r\n")
	sb.WriteString("To: " + strings.Join(msg.To, ", ") + "\r\n")
	if len(msg.CC) > 0 {
		sb.WriteString("Cc: " + strings.Join(msg.CC, ", ") + "\r\n")
	}
	sb.WriteString("Subject: " + msg.Subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString(`Content-Type: multipart/alternative; boundary="` + boundary + `"` + "\r\n\r\n")

	// Plain text part
	sb.WriteString("--" + boundary + "\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	sb.WriteString(msg.TextBody + "\r\n\r\n")

	// HTML part
	sb.WriteString("--" + boundary + "\r\n")
	sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	sb.WriteString(msg.HTMLBody + "\r\n\r\n")

	sb.WriteString("--" + boundary + "--\r\n")
	return sb.String()
}
