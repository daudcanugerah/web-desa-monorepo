package email

import (
	"fmt"
	"net/smtp"

	"webdesa/api/config"
)

// SMTPSender is a low-level SMTP client wrapper used by all email services
// (PPID, Auth, etc.) so that SMTP configuration and connection handling
// live in one place.
type SMTPSender struct {
	cfg *config.SMTPConfig
}

// NewSMTPSender constructs an SMTPSender. The cfg pointer may be nil or
// unconfigured; IsConfigured() reports back whether the caller should
// actually try to deliver.
func NewSMTPSender(cfg *config.SMTPConfig) *SMTPSender {
	return &SMTPSender{cfg: cfg}
}

// IsConfigured reports whether SMTP credentials are available.
func (s *SMTPSender) IsConfigured() bool {
	return s.cfg != nil && s.cfg.IsConfigured()
}

// Send delivers a single plain-HTML email. It returns a descriptive error
// if SMTP is not configured or the underlying SendMail call fails.
func (s *SMTPSender) Send(to, subject, htmlBody, fromName string) error {
	if !s.IsConfigured() {
		return fmt.Errorf("SMTP not configured")
	}

	from := s.cfg.FromEmail
	if from == "" {
		from = s.cfg.Username
	}
	if fromName == "" {
		fromName = "Village Administration"
	}

	headers := fmt.Sprintf(
		"From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n",
		fromName, from, to, subject,
	)

	msg := []byte(headers + "\r\n" + htmlBody)
	addr := s.cfg.GetSMTPAddress()

	if err := smtp.SendMail(
		addr,
		smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host),
		from,
		[]string{to},
		msg,
	); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}