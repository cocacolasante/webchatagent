package notification

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"

	"go.uber.org/zap"
)

// EmailSender sends emails via SMTP.
type EmailSender struct {
	host      string
	port      int
	user      string
	password  string
	fromEmail string
	fromName  string
	log       *zap.Logger
}

// NewEmailSender creates a new EmailSender.
func NewEmailSender(host string, port int, user, password, fromEmail, fromName string, log *zap.Logger) *EmailSender {
	return &EmailSender{
		host:      host,
		port:      port,
		user:      user,
		password:  password,
		fromEmail: fromEmail,
		fromName:  fromName,
		log:       log,
	}
}

// Send sends a plain-text email.
func (s *EmailSender) Send(ctx context.Context, to, subject, body string) error {
	if s.host == "" {
		s.log.Debug("email not configured, skipping", zap.String("to", to))
		return nil
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.password, s.host)

	msg := fmt.Sprintf("From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.fromName, s.fromEmail, to, subject, body)

	// TLS config for secure connection
	tlsConfig := &tls.Config{
		ServerName: s.host,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		// Fall back to non-TLS for development
		return smtp.SendMail(addr, auth, s.fromEmail, []string{to}, []byte(msg))
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err = client.Mail(s.fromEmail); err != nil {
		return fmt.Errorf("smtp from: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	defer wc.Close()

	if _, err = fmt.Fprint(wc, msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	return nil
}
