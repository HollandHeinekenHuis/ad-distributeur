package main

import (
	"fmt"
	"log"
	"mime"
	"net/smtp"
	"strings"
)

// Mailer sends plain-text emails.
type Mailer interface {
	Send(to, subject, body string) error
}

// newMailer returns an SMTP mailer when SMTP_HOST is set, otherwise a mailer
// that only logs, so bookings work before email is configured.
func newMailer(cfg Config) Mailer {
	if cfg.SMTPHost == "" {
		log.Println("SMTP_HOST not set: booking emails will be logged, not sent")
		return logMailer{}
	}
	return smtpMailer{cfg: cfg}
}

type logMailer struct{}

func (logMailer) Send(to, subject, body string) error {
	log.Printf("[email not sent: SMTP not configured] to=%s subject=%q\n%s", to, subject, body)
	return nil
}

type smtpMailer struct{ cfg Config }

func (m smtpMailer) Send(to, subject, body string) error {
	from := sanitizeHeader(m.cfg.MailFrom)
	to = sanitizeHeader(to)
	var sb strings.Builder
	fmt.Fprintf(&sb, "From: %s\r\n", from)
	fmt.Fprintf(&sb, "To: %s\r\n", to)
	fmt.Fprintf(&sb, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", sanitizeHeader(subject)))
	sb.WriteString("MIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n")
	sb.WriteString(strings.ReplaceAll(body, "\n", "\r\n"))

	var auth smtp.Auth
	if m.cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPassword, m.cfg.SMTPHost)
	}
	addr := m.cfg.SMTPHost + ":" + m.cfg.SMTPPort
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(sb.String()))
}

// sanitizeHeader strips CR/LF to prevent header injection.
func sanitizeHeader(s string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}
