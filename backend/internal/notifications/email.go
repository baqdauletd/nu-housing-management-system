package notifications

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"

	"nu-housing-management-system/backend/internal/config"
)

type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

type EmailSender struct {
	cfg *config.Config
}

func NewEmailSender(cfg *config.Config) EmailSender {
	return EmailSender{cfg: cfg}
}

func (s EmailSender) Configured() bool {
	if s.cfg == nil {
		return false
	}
	return strings.TrimSpace(s.cfg.SMTPHost) != "" &&
		s.cfg.SMTPPort > 0 &&
		strings.TrimSpace(s.cfg.SMTPFrom) != "" &&
		len(s.cfg.SMTPAllowedRecipients) > 0
}

func (s EmailSender) Send(message EmailMessage) error {
	if !s.Configured() {
		return nil
	}

	from := strings.TrimSpace(s.cfg.SMTPFrom)
	to := strings.TrimSpace(message.To)
	if to == "" {
		return nil
	}
	if !s.cfg.IsEmailRecipientAllowed(to) {
		return nil
	}

	var auth smtp.Auth
	username := strings.TrimSpace(s.cfg.SMTPUsername)
	password := strings.TrimSpace(s.cfg.SMTPPassword)
	if username != "" || password != "" {
		auth = smtp.PlainAuth("", username, password, strings.TrimSpace(s.cfg.SMTPHost))
	}

	var payload bytes.Buffer
	payload.WriteString("From: " + from + "\r\n")
	payload.WriteString("To: " + to + "\r\n")
	payload.WriteString("Subject: " + sanitizeHeader(message.Subject) + "\r\n")
	payload.WriteString("MIME-Version: 1.0\r\n")
	payload.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	payload.WriteString("\r\n")
	payload.WriteString(message.Body)
	payload.WriteString("\r\n")

	address := fmt.Sprintf("%s:%d", strings.TrimSpace(s.cfg.SMTPHost), s.cfg.SMTPPort)
	return smtp.SendMail(address, auth, from, []string{to}, payload.Bytes())
}

func sanitizeHeader(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}
