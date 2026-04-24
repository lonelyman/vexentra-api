package mailer

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"vexentra-api/internal/config"
	"vexentra-api/pkg/logger"
)

type Mailer interface {
	IsEnabled() bool
	Send(to, subject, htmlBody string) error
}

type smtpMailer struct {
	cfg config.MailerConfig
	log logger.Logger
}

func New(cfg config.MailerConfig, l logger.Logger) Mailer {
	if l == nil {
		l = logger.Get()
	}
	return &smtpMailer{cfg: cfg, log: l}
}

func (m *smtpMailer) IsEnabled() bool {
	return m.cfg.Host != "" && m.cfg.Port > 0 && m.cfg.Username != "" && m.cfg.Password != ""
}

func (m *smtpMailer) Send(to, subject, htmlBody string) error {
	if !m.IsEnabled() {
		return fmt.Errorf("mailer is not configured")
	}

	fromAddress := m.cfg.Username
	displayName := strings.TrimSpace(m.cfg.Name)
	fromHeader := fromAddress
	if displayName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", displayName, fromAddress)
	}

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	headers := []string{
		fmt.Sprintf("From: %s", fromHeader),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}
	msg := strings.Join(headers, "\r\n") + "\r\n\r\n" + htmlBody

	// Common TLS SMTP port.
	if m.cfg.Port == 465 {
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: m.cfg.Host})
		if err != nil {
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, m.cfg.Host)
		if err != nil {
			return err
		}
		defer client.Quit()

		if err = client.Auth(auth); err != nil {
			return err
		}
		if err = client.Mail(fromAddress); err != nil {
			return err
		}
		if err = client.Rcpt(to); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		if _, err = w.Write([]byte(msg)); err != nil {
			_ = w.Close()
			return err
		}
		return w.Close()
	}

	// STARTTLS / plain SMTP (typically 587 / local test).
	return smtp.SendMail(addr, auth, fromAddress, []string{to}, []byte(msg))
}
