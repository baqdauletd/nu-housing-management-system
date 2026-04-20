package notifications

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"nu-housing-management-system/backend/internal/config"
)

var ErrEmailRecipientNotAllowed = errors.New("email recipient is not allowed")

const sendTimeout = 10 * time.Second

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
		strings.TrimSpace(s.cfg.SMTPFrom) != ""
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
		return fmt.Errorf("%w: %s", ErrEmailRecipientNotAllowed, to)
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

	host := strings.TrimSpace(s.cfg.SMTPHost)
	address := fmt.Sprintf("%s:%d", host, s.cfg.SMTPPort)
	dialer := net.Dialer{Timeout: sendTimeout}
	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(sendTimeout)); err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(payload.Bytes()); err != nil {
		writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func sanitizeHeader(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}
